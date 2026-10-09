package communication

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/websocket"
)

// PollingEngine coordinates concurrent, isolated polling loops for all active edge devices
type PollingEngine struct {
	connManager      *ConnectionManager
	deviceService    *service.DeviceService
	telemetryService *service.TelemetryService
	hub              *websocket.Hub
	holdTracker      *AnomalyHoldTracker
	workersMu        sync.Mutex
	workers          map[uint]context.CancelFunc
	ctx              context.Context
	cancel           context.CancelFunc
	stopOnce         sync.Once
}

// NewPollingEngine constructs the polling coordinator
func NewPollingEngine(connManager *ConnectionManager, deviceService *service.DeviceService, hub *websocket.Hub) *PollingEngine {
	ctx, cancel := context.WithCancel(context.Background())
	return &PollingEngine{
		connManager:   connManager,
		deviceService: deviceService,
		hub:           hub,
		holdTracker:   NewAnomalyHoldTracker(),
		workers:       make(map[uint]context.CancelFunc),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// GetHoldTracker returns the AnomalyHoldTracker instance
func (p *PollingEngine) GetHoldTracker() *AnomalyHoldTracker {
	return p.holdTracker
}

// SetTelemetryService injects the Phase 3.1 Telemetry Ingestion Pipeline
func (p *PollingEngine) SetTelemetryService(ts *service.TelemetryService) {
	p.telemetryService = ts
}

// Start spawns polling workers for all enabled devices
func (p *PollingEngine) Start() {
	if p.deviceService == nil {
		logger.Warn("PollingEngine started without deviceService, skipping device discovery")
		return
	}

	devices, _, err := p.deviceService.ListDevices(repository.DeviceFilterParams{
		Page:     1,
		PageSize: 500,
	})
	if err != nil {
		logger.Error("PollingEngine failed to query active devices: %v", err)
		return
	}

	started := 0
	for i := range devices {
		dev := &devices[i]
		if dev.Enabled && dev.Status == model.DeviceStatusActive && dev.Connection != nil && dev.Connection.Enabled {
			p.StartDeviceWorker(dev)
			started++
		}
	}

	logger.Info("PollingEngine started with %d isolated device workers", started)
}

// Stop terminates all polling goroutines and releases transports
func (p *PollingEngine) Stop() {
	p.stopOnce.Do(func() {
		p.cancel()

		p.workersMu.Lock()
		for id, cancelFunc := range p.workers {
			cancelFunc()
			delete(p.workers, id)
		}
		p.workersMu.Unlock()

		p.connManager.CloseAll()
		logger.Info("PollingEngine stopped gracefully")
	})
}

// StartDeviceWorker registers an isolated worker goroutine for a specific device
func (p *PollingEngine) StartDeviceWorker(device *model.Device) {
	if device == nil || device.Connection == nil {
		return
	}

	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	// Cancel existing worker for this device if already running
	if existingCancel, exists := p.workers[device.ID]; exists {
		existingCancel()
		delete(p.workers, device.ID)
	}

	workerCtx, workerCancel := context.WithCancel(p.ctx)
	p.workers[device.ID] = workerCancel

	go p.runWorker(workerCtx, device.ID)
}

// StopDeviceWorker stops polling for a specific device
func (p *PollingEngine) StopDeviceWorker(deviceID uint) {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	if cancel, exists := p.workers[deviceID]; exists {
		cancel()
		delete(p.workers, deviceID)
	}
	_ = p.connManager.DisconnectDevice(deviceID)
}

// ActiveWorkerCount returns the number of active polling workers currently running
func (p *PollingEngine) ActiveWorkerCount() int {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()
	return len(p.workers)
}

// runWorker is the isolated loop executing on a per-device schedule with in-memory caching & adaptive backoff
func (p *PollingEngine) runWorker(ctx context.Context, deviceID uint) {
	if p.deviceService == nil {
		return
	}

	// Query device to get interval
	dev, err := p.deviceService.GetDeviceByID(deviceID)
	if err != nil || dev == nil || dev.Connection == nil {
		return
	}

	baseIntervalMs := dev.Connection.PollingInterval
	if baseIntervalMs < 200 {
		baseIntervalMs = 1000
	}
	currentIntervalMs := baseIntervalMs
	ticker := time.NewTicker(time.Duration(currentIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	logger.Debug("Worker for device %s (%d) started (interval: %dms)", dev.DeviceCode, dev.ID, currentIntervalMs)

	// In-memory cached device & parameters to avoid hitting MariaDB on every poll cycle
	cachedDev := dev
	var cachedParams []model.Parameter
	if params, pErr := p.deviceService.ListParameters(deviceID); pErr == nil {
		for _, param := range params {
			if param.Enabled {
				cachedParams = append(cachedParams, param)
			}
		}
	}

	cyclesSinceRefresh := 0
	consecutiveFailures := 0

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Worker for device %d stopped", deviceID)
			return

		case <-ticker.C:
			cyclesSinceRefresh++

			// Periodically (every 10 cycles, ~10-20s) re-sync device config & parameters from DB
			if cyclesSinceRefresh >= 10 {
				cyclesSinceRefresh = 0
				freshDev, err := p.deviceService.GetDeviceByID(deviceID)
				if err != nil {
					if strings.Contains(strings.ToLower(err.Error()), "not found") {
						logger.Info("Device %d deleted or removed, terminating polling worker", deviceID)
						return
					}
				} else if freshDev != nil {
					cachedDev = freshDev
					if freshDev.Connection != nil && freshDev.Connection.PollingInterval >= 200 {
						baseIntervalMs = freshDev.Connection.PollingInterval
					}
					if params, pErr := p.deviceService.ListParameters(deviceID); pErr == nil {
						cachedParams = cachedParams[:0]
						for _, param := range params {
							if param.Enabled {
								cachedParams = append(cachedParams, param)
							}
						}
					}
				}
			}

			// Device Isolation & Admin Status Check
			if cachedDev == nil || !cachedDev.Enabled || cachedDev.Status != model.DeviceStatusActive || cachedDev.Connection == nil || !cachedDev.Connection.Enabled {
				// Standby mode without heavy polling: check every 5s
				if currentIntervalMs != 5000 {
					currentIntervalMs = 5000
					ticker.Reset(5 * time.Second)
				}
				continue
			}

			// Execute cached parameter poll cycle
			cycleSuccess := p.pollDevice(ctx, cachedDev, cachedParams)
			if cycleSuccess {
				if consecutiveFailures > 0 {
					consecutiveFailures = 0
					currentIntervalMs = baseIntervalMs
					ticker.Reset(time.Duration(currentIntervalMs) * time.Millisecond)
					logger.Info("Device %s (%d) communication active, restored normal polling interval (%dms)",
						cachedDev.DeviceCode, deviceID, baseIntervalMs)
				}
			} else {
				consecutiveFailures++
				// Auto-recovery (Phase 4.1): Trigger asynchronous deduplicated reconnection attempt
				go func(d *model.Device) {
					reconnCtx, reconnCancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer reconnCancel()
					_ = p.connManager.ReconnectDevice(reconnCtx, d)
				}(cachedDev)

				// Adaptive backoff on consecutive connection failures to preserve edge CPU and reduce log spam
				var backoffMs int
				if consecutiveFailures >= 5 {
					backoffMs = 15000 // 15s backoff
				} else if consecutiveFailures >= 2 {
					backoffMs = 5000 // 5s backoff
				}
				if backoffMs > 0 && currentIntervalMs != backoffMs {
					currentIntervalMs = backoffMs
					ticker.Reset(time.Duration(currentIntervalMs) * time.Millisecond)
					logger.Debug("Device %s (%d) offline/unreachable (%d failures), backing off polling to %dms",
						cachedDev.DeviceCode, deviceID, consecutiveFailures, backoffMs)
				}
			}
		}
	}
}

// pollDevice executes the parameter read loop using pre-cached parameters
func (p *PollingEngine) pollDevice(ctx context.Context, device *model.Device, enabledParams []model.Parameter) bool {
	if len(enabledParams) == 0 {
		// No parameters configured, perform health check
		_ = p.connManager.ConnectDevice(ctx, device)
		return true
	}

	anySuccess := false
	for _, param := range enabledParams {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		res, err := p.ReadParameter(ctx, device, &param)
		if err != nil {
			// Broadcast error event
			if p.hub != nil {
				p.hub.Broadcast(websocket.WSMessage{
					Type:      "device.communication.error",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"device_id":      device.ID,
						"device_code":    device.DeviceCode,
						"parameter_id":   param.ID,
						"parameter_code": param.ParameterCode,
						"error":          err.Error(),
					},
				})
			}

			// If device transport itself is dead, skip subsequent parameters in this cycle to avoid blocking
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "offline") ||
				strings.Contains(errStr, "connection refused") ||
				strings.Contains(errStr, "dial failed") ||
				strings.Contains(errStr, "cannot find") ||
				strings.Contains(errStr, "timeout") {
				return false
			}
			continue
		}

		anySuccess = true
		// Broadcast success event
		if p.hub != nil {
			p.hub.Broadcast(websocket.WSMessage{
				Type:      "device.communication.success",
				Timestamp: res.Timestamp,
				Data:      res,
			})
		}
	}

	return anySuccess
}

// ReadParameter reads and decodes a single parameter from the device
func (p *PollingEngine) ReadParameter(
	ctx context.Context,
	device *model.Device,
	param *model.Parameter,
) (*CommunicationResult, error) {
	if device == nil || device.Connection == nil || param == nil {
		return nil, fmt.Errorf("invalid device or parameter")
	}

	// 1. Resolve Register Address and Function Code
	res, err := modbus.ResolveRegisterAddress(param.RegisterType, param.RegisterAddress)
	if err != nil {
		return nil, fmt.Errorf("failed resolving address: %w", err)
	}

	// 2. Determine Register Count
	qty := modbus.RequiredRegisterCount(param.DataType)
	if res.FunctionCode == FunctionReadCoils || res.FunctionCode == FunctionReadDiscreteInputs {
		qty = 1
	}

	slaveID := byte(device.Connection.SlaveID)
	if slaveID == 0 {
		slaveID = 1
	}

	req := ModbusReadRequest{
		SlaveID:         slaveID,
		FunctionCode:    res.FunctionCode,
		StartingAddress: res.PDUAddress,
		Quantity:        qty,
	}

	// 3. Execute with Retry
	start := time.Now()
	resp, err := p.connManager.ExecuteWithRetry(ctx, device, req)
	elapsed := time.Since(start)

	if err != nil {
		source := ""
		if device.Connection != nil {
			source = string(device.Connection.Protocol)
		}

		// Evaluate anomaly hold tracker
		var effVal float64
		var formulaVal *float64
		var isHeld bool
		var quality model.TelemetryQuality = model.QualityBad
		var pErr error = err

		if p.holdTracker != nil {
			effVal, formulaVal, isHeld, quality, pErr = p.holdTracker.ProcessReading(param, 0, 0, err)
		}

		if isHeld {
			// Sensor anomaly/timeout suppressed by hold-last-good-value grace period!
			finalHeldVal := effVal
			if formulaVal != nil {
				finalHeldVal = *formulaVal
			}

			if p.deviceService != nil {
				_ = p.deviceService.UpdateParameterCurrentValue(param.ID, finalHeldVal, formulaVal, isHeld)
			}

			if p.telemetryService != nil {
				_ = p.telemetryService.Ingest(&service.TelemetryIngestPayload{
					DeviceID:      device.ID,
					DeviceCode:    device.DeviceCode,
					DeviceName:    device.DeviceName,
					ParameterID:   param.ID,
					ParameterCode: param.ParameterCode,
					ParameterName: param.ParameterName,
					Unit:          param.Unit,
					DataType:      param.DataType,
					RawValue:      effVal,
					Value:         finalHeldVal,
					FormulaValue:  formulaVal,
					Formula:       param.Formula,
					IsHeldValue:   true,
					Quality:       quality,
					QualityReason: model.ReasonHoldAnomalyGrace,
					QualityFlags:  model.FlagHeldAnomaly,
					Source:        source,
					ReceivedAt:    time.Now().UTC(),
					Parameter:     param,
					ReadError:     pErr,
				})
			}

			heldRes := NewSuccessResult(
				device.ID,
				device.Connection.ID,
				param.ID,
				param.ParameterCode,
				res.FunctionCode,
				res.PDUAddress,
				qty,
				nil,
				effVal,
				finalHeldVal,
				elapsed,
			)
			heldRes.ScaledValue = effVal
			heldRes.Formula = param.Formula
			heldRes.FormulaValue = formulaVal
			heldRes.IsHeld = true
			heldRes.Quality = quality
			heldRes.QualityReason = model.ReasonHoldAnomalyGrace
			heldRes.QualityFlags = model.FlagHeldAnomaly
			heldRes.ProcessedValue = finalHeldVal
			return heldRes, nil
		}

		lastKnownVal := 0.0
		if param.CurrentProcessedValue != nil {
			lastKnownVal = *param.CurrentProcessedValue
		} else if param.CurrentValue != nil {
			lastKnownVal = *param.CurrentValue
		}

		errRes := NewErrorResult(
			device.ID,
			device.Connection.ID,
			param.ID,
			param.ParameterCode,
			res.FunctionCode,
			res.PDUAddress,
			qty,
			"READ_ERROR",
			err.Error(),
			elapsed,
		)
		errRes.Quality = quality
		errRes.QualityReason = model.ReasonCommunicationError
		errRes.QualityFlags = model.FlagCommError
		errRes.ProcessedValue = lastKnownVal

		if p.telemetryService != nil {
			_ = p.telemetryService.Ingest(&service.TelemetryIngestPayload{
				DeviceID:      device.ID,
				DeviceCode:    device.DeviceCode,
				DeviceName:    device.DeviceName,
				ParameterID:   param.ID,
				ParameterCode: param.ParameterCode,
				ParameterName: param.ParameterName,
				Unit:          param.Unit,
				DataType:      param.DataType,
				RawValue:      lastKnownVal,
				Value:         lastKnownVal,
				Quality:       quality,
				QualityReason: model.ReasonCommunicationError,
				QualityFlags:  model.FlagCommError,
				IsHeldValue:   false,
				Source:        source,
				ErrorMessage:  err.Error(),
				ReceivedAt:    errRes.Timestamp,
				Parameter:     param,
				ReadError:     err,
			})
		}

		return errRes, pErr
	}

	// 4. Decode Register Values
	byteOrder := param.ByteOrder
	if byteOrder == "" && device.Connection != nil {
		byteOrder = device.Connection.ByteOrder
	}
	if byteOrder == "" {
		byteOrder = "ABCD"
	}

	decoded, err := modbus.DecodeRegisters(resp.Data, param.DataType, byteOrder)
	if err != nil {
		return nil, fmt.Errorf("register decode failed: %w", err)
	}

	// 5. Apply Scale & Offset
	engVal := modbus.ApplyScaleAndOffset(decoded.RawValue, param.Scale, param.Offset, param.Precision)

	// 5b. Evaluate Anomaly Hold Tracker & Custom Formula
	effVal := engVal
	var formulaVal *float64
	var isHeld bool
	quality := model.QualityGood

	if p.holdTracker != nil {
		effVal, formulaVal, isHeld, quality, _ = p.holdTracker.ProcessReading(param, decoded.RawValue, engVal, nil)
	}

	// 6. Update Device and Parameter State
	finalVal := effVal
	if formulaVal != nil {
		finalVal = *formulaVal
	}

	if p.deviceService != nil {
		_ = p.deviceService.UpdateParameterCurrentValue(param.ID, finalVal, formulaVal, isHeld)
		_ = p.deviceService.UpdateLastData(device.ID)
	}

	result := NewSuccessResult(
		device.ID,
		device.Connection.ID,
		param.ID,
		param.ParameterCode,
		res.FunctionCode,
		res.PDUAddress,
		qty,
		resp.Data,
		decoded.RawValue,
		finalVal,
		resp.ResponseTime,
	)
	result.ScaledValue = engVal
	result.Formula = param.Formula
	result.FormulaValue = formulaVal
	result.IsHeld = isHeld
	result.Quality = quality
	result.ProcessedValue = finalVal

	// 7. Pipeline Telemetry Ingestion (Phase 3.1 & 3.2)
	if p.telemetryService != nil {
		source := ""
		if device.Connection != nil {
			source = string(device.Connection.Protocol)
		}
		_ = p.telemetryService.Ingest(&service.TelemetryIngestPayload{
			DeviceID:      device.ID,
			DeviceCode:    device.DeviceCode,
			DeviceName:    device.DeviceName,
			ParameterID:   param.ID,
			ParameterCode: param.ParameterCode,
			ParameterName: param.ParameterName,
			Unit:          param.Unit,
			DataType:      param.DataType,
			RawBytes:      resp.Data,
			RawHex:        result.RawHex,
			RawValue:      decoded.RawValue,
			Value:         finalVal,
			FormulaValue:  formulaVal,
			Formula:       param.Formula,
			IsHeldValue:   isHeld,
			Quality:       quality,
			Source:        source,
			ReceivedAt:    result.Timestamp,
			Parameter:     param,
			ReadError:     nil,
		})
	}

	return result, nil
}

// ReadParameterOnDemand performs a manual diagnostic read for a specific parameter
func (p *PollingEngine) ReadParameterOnDemand(
	ctx context.Context,
	deviceID, paramID uint,
) (*CommunicationResult, error) {
	dev, err := p.deviceService.GetDeviceByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	param, err := p.deviceService.GetParameter(deviceID, paramID)
	if err != nil {
		return nil, fmt.Errorf("parameter not found: %w", err)
	}

	return p.ReadParameter(ctx, dev, param)
}

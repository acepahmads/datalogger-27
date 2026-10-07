package communication

import (
	"context"
	"fmt"
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
	connManager   *ConnectionManager
	deviceService *service.DeviceService
	hub           *websocket.Hub
	workersMu     sync.Mutex
	workers       map[uint]context.CancelFunc
	ctx           context.Context
	cancel        context.CancelFunc
	stopOnce      sync.Once
}

// NewPollingEngine constructs the polling coordinator
func NewPollingEngine(connManager *ConnectionManager, deviceService *service.DeviceService, hub *websocket.Hub) *PollingEngine {
	ctx, cancel := context.WithCancel(context.Background())
	return &PollingEngine{
		connManager:   connManager,
		deviceService: deviceService,
		hub:           hub,
		workers:       make(map[uint]context.CancelFunc),
		ctx:           ctx,
		cancel:        cancel,
	}
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

// runWorker is the isolated loop executing on a per-device schedule
func (p *PollingEngine) runWorker(ctx context.Context, deviceID uint) {
	// Query device to get interval
	dev, err := p.deviceService.GetDeviceByID(deviceID)
	if err != nil || dev == nil || dev.Connection == nil {
		return
	}

	intervalMs := dev.Connection.PollingInterval
	if intervalMs < 200 {
		intervalMs = 1000
	}
	ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	defer ticker.Stop()

	logger.Debug("Worker for device %s (%d) started (interval: %dms)", dev.DeviceCode, dev.ID, intervalMs)

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Worker for device %d stopped", deviceID)
			return

		case <-ticker.C:
			// Re-fetch device in case status changed
			freshDev, err := p.deviceService.GetDeviceByID(deviceID)
			if err != nil || freshDev == nil {
				return
			}

			// Device Isolation & Admin Status Check
			if !freshDev.Enabled || freshDev.Status != model.DeviceStatusActive || freshDev.Connection == nil || !freshDev.Connection.Enabled {
				// Device disabled or in maintenance, standby without heavy polling
				continue
			}

			p.pollDevice(ctx, freshDev)
		}
	}
}

// pollDevice executes the parameter read loop for one device cycle
func (p *PollingEngine) pollDevice(ctx context.Context, device *model.Device) {
	params, err := p.deviceService.ListParameters(device.ID)
	if err != nil {
		return
	}

	var enabledParams []model.Parameter
	for _, param := range params {
		if param.Enabled {
			enabledParams = append(enabledParams, param)
		}
	}

	if len(enabledParams) == 0 {
		// No parameters configured, perform health check
		_ = p.connManager.ConnectDevice(ctx, device)
		return
	}

	for _, param := range enabledParams {
		select {
		case <-ctx.Done():
			return
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
			continue
		}

		// Broadcast success event
		if p.hub != nil {
			p.hub.Broadcast(websocket.WSMessage{
				Type:      "device.communication.success",
				Timestamp: res.Timestamp,
				Data:      res,
			})
		}
	}
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
		return errRes, err
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

	// 6. Update Device and Parameter State
	if p.deviceService != nil {
		_ = p.deviceService.UpdateParameterCurrentValue(param.ID, engVal)
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
		engVal,
		resp.ResponseTime,
	)

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

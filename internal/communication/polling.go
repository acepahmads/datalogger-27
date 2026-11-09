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
	alarmEvaluator   *service.AlarmEvaluator
	hub              *websocket.Hub
	holdTracker      *AnomalyHoldTracker
	workersMu        sync.Mutex
	workers          map[uint]context.CancelFunc
	ctx              context.Context
	cancel           context.CancelFunc
	stopOnce         sync.Once
	batchFailuresMu      sync.Mutex
	batchFailures        map[string]int
	activityMonitor      *ActivityMonitor
	sensorPollingEnabled bool
}

// NewPollingEngine constructs the polling coordinator
func NewPollingEngine(connManager *ConnectionManager, deviceService *service.DeviceService, hub *websocket.Hub) *PollingEngine {
	ctx, cancel := context.WithCancel(context.Background())
	pe := &PollingEngine{
		connManager:          connManager,
		deviceService:        deviceService,
		hub:                   hub,
		holdTracker:          NewAnomalyHoldTracker(),
		workers:              make(map[uint]context.CancelFunc),
		batchFailures:        make(map[string]int),
		ctx:                  ctx,
		cancel:               cancel,
		sensorPollingEnabled: true,
	}

	tracker := GetSharedAdapterTracker()
	tracker.SetDeviceService(deviceService)
	tracker.SetHub(hub)
	tracker.SetConnectionManager(connManager)

	return pe
}

// GetHoldTracker returns the AnomalyHoldTracker instance
func (p *PollingEngine) GetHoldTracker() *AnomalyHoldTracker {
	return p.holdTracker
}

// SetTelemetryService injects the Phase 3.1 Telemetry Ingestion Pipeline
func (p *PollingEngine) SetTelemetryService(ts *service.TelemetryService) {
	p.telemetryService = ts
}

// GetTelemetryService retrieves the Phase 3.1 Telemetry Ingestion Pipeline
func (p *PollingEngine) GetTelemetryService() *service.TelemetryService {
	return p.telemetryService
}

// SetAlarmEvaluator injects the Phase 6 Alarm Evaluator
func (p *PollingEngine) SetAlarmEvaluator(ae *service.AlarmEvaluator) {
	p.alarmEvaluator = ae
}

// GetAlarmEvaluator retrieves the Phase 6 Alarm Evaluator
func (p *PollingEngine) GetAlarmEvaluator() *service.AlarmEvaluator {
	return p.alarmEvaluator
}

// SetActivityMonitor attaches the system activity monitor
func (p *PollingEngine) SetActivityMonitor(mon *ActivityMonitor) {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()
	p.activityMonitor = mon
	GetSharedAdapterTracker().SetActivityMonitor(mon)
}

// getActivityMonitor retrieves the attached or global activity monitor
func (p *PollingEngine) getActivityMonitor() *ActivityMonitor {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()
	if p.activityMonitor != nil {
		return p.activityMonitor
	}
	return GetGlobalActivityMonitor()
}

func (p *PollingEngine) broadcastForDevice(device *model.Device, msg websocket.WSMessage) {
	if p.hub == nil {
		return
	}
	if device != nil {
		msg.TenantID = device.TenantID
		msg.SiteID = device.SiteID
	}
	p.hub.Broadcast(msg)
}

// toHealthInput converts a ConnectionHealth struct to service.DeviceHealthInput
func toHealthInput(h ConnectionHealth) *service.DeviceHealthInput {
	return &service.DeviceHealthInput{
		DeviceID:             h.DeviceID,
		State:                string(h.State),
		ConsecutiveFailures:  h.ConsecutiveFailures,
		ConsecutiveSuccesses: h.ConsecutiveSuccesses,
		LastCommSuccess:      h.LastCommSuccess,
		LastCommFailure:      h.LastCommFailure,
		LastErrorCategory:    h.LastErrorCategory,
		LastError:            h.LastError,
	}
}

// IsSensorPollingEnabled reports whether physical sensor polling (Modbus/Serial/TCP) is active.
func (p *PollingEngine) IsSensorPollingEnabled() bool {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()
	return p.sensorPollingEnabled
}

// SetSensorPollingEnabled dynamically enables or pauses physical sensor polling without stopping HTTP/MQTT receivers.
func (p *PollingEngine) SetSensorPollingEnabled(enabled bool) {
	p.workersMu.Lock()
	changed := p.sensorPollingEnabled != enabled
	p.sensorPollingEnabled = enabled
	p.workersMu.Unlock()

	if !enabled {
		p.PauseAllWorkers()
		logger.Info("Physical sensor polling disabled by configuration; HTTP API & MQTT telemetry receivers remain active")
		return
	}
	if changed && p.ctx.Err() == nil {
		p.Start()
	}
}

// Start spawns polling workers for all enabled devices
func (p *PollingEngine) Start() {
	if !p.IsSensorPollingEnabled() {
		logger.Info("PollingEngine startup skipped physical sensor workers (sensor_polling_enabled=false)")
		return
	}
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

// PauseAllWorkers temporarily suspends all active device polling goroutines and closes transports
// (e.g. during full database restore so background telemetry/device writes do not contend with DDL/LOCK TABLES)
// without permanently cancelling the engine context.
func (p *PollingEngine) PauseAllWorkers() {
	p.workersMu.Lock()
	paused := len(p.workers)
	for id, cancelFunc := range p.workers {
		cancelFunc()
		delete(p.workers, id)
	}
	p.workersMu.Unlock()

	if p.connManager != nil {
		p.connManager.CloseAll()
	}
	logger.Info("PollingEngine paused %d device workers for maintenance/restore", paused)
}

// ResumeAllWorkers reloads active devices from the database and restarts device polling workers.
func (p *PollingEngine) ResumeAllWorkers() {
	if p.ctx.Err() != nil {
		return
	}
	p.Start()
}

// StartDeviceWorker registers an isolated worker goroutine for a specific device
func (p *PollingEngine) StartDeviceWorker(device *model.Device) {
	if device == nil || device.Connection == nil {
		return
	}
	if strings.EqualFold(device.DeviceType, string(model.ProtocolEdgePush)) ||
		device.Connection.Protocol == model.ProtocolEdgePush ||
		strings.EqualFold(device.DeviceType, string(model.ProtocolMQTT)) ||
		device.Connection.Protocol == model.ProtocolMQTT {
		return
	}
	if !p.IsSensorPollingEnabled() {
		return
	}

	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	// Cancel existing worker for this device if already running
	if existingCancel, exists := p.workers[device.ID]; exists {
		existingCancel()
		delete(p.workers, device.ID)
	}

	if device.Connection != nil && device.Connection.SerialPort != "" {
		GetSharedAdapterTracker().RegisterDevice(device.ID, device.DeviceCode, device.DeviceName, device.Connection.SerialPort)
	}
	hasFormula := deviceHasActiveFormula(device)
	GetSharedAdapterTracker().SetDeviceFormulaActive(device.ID, hasFormula)
	if hasFormula && p.deviceService != nil {
		_ = p.deviceService.SetOnline(device.ID)
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
	GetSharedAdapterTracker().UnregisterDevice(deviceID)
	_ = p.connManager.DisconnectDevice(deviceID)
	if p.alarmEvaluator != nil {
		p.alarmEvaluator.ClearDeviceAlarms(p.ctx, deviceID)
	}
	if p.telemetryService != nil {
		p.telemetryService.ClearDeviceCache(deviceID)
	}
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
	if baseIntervalMs > 0 && baseIntervalMs <= 60 {
		baseIntervalMs = baseIntervalMs * 1000
	}
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
		cachedDev.Parameters = cachedParams
	}
	initFormula := deviceHasActiveFormula(cachedDev)
	GetSharedAdapterTracker().SetDeviceFormulaActive(deviceID, initFormula)
	if initFormula && p.deviceService != nil {
		_ = p.deviceService.SetOnline(deviceID)
	}

	currentIntervalMs = calculateDeviceTickIntervalMs(baseIntervalMs, cachedParams)
	ticker.Reset(time.Duration(currentIntervalMs) * time.Millisecond)
	paramLastPolled := make(map[uint]time.Time)

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
					if freshDev.Connection != nil {
						newInt := freshDev.Connection.PollingInterval
						if newInt > 0 && newInt <= 60 {
							newInt = newInt * 1000
						}
						if newInt >= 200 && newInt != baseIntervalMs {
							baseIntervalMs = newInt
							currentIntervalMs = baseIntervalMs
							ticker.Reset(time.Duration(currentIntervalMs) * time.Millisecond)
							logger.Info("Device %s (%d) dynamic polling interval updated to %dms", freshDev.DeviceCode, deviceID, baseIntervalMs)
						}
					}
					if params, pErr := p.deviceService.ListParameters(deviceID); pErr == nil {
						cachedParams = cachedParams[:0]
						for _, param := range params {
							if param.Enabled {
								cachedParams = append(cachedParams, param)
							}
						}
						cachedDev.Parameters = cachedParams
					}
					GetSharedAdapterTracker().SetDeviceFormulaActive(deviceID, deviceHasActiveFormula(cachedDev))
					targetTick := calculateDeviceTickIntervalMs(baseIntervalMs, cachedParams)
					if targetTick != currentIntervalMs {
						currentIntervalMs = targetTick
						ticker.Reset(time.Duration(currentIntervalMs) * time.Millisecond)
					}
				}
			}

			// Device Isolation & Admin Status Check
			if cachedDev == nil || !cachedDev.Enabled || cachedDev.Status != model.DeviceStatusActive || cachedDev.Connection == nil || !cachedDev.Connection.Enabled {
				// Standby mode without heavy polling: check every 5s
				if p.alarmEvaluator != nil && cachedDev != nil {
					var hIn *service.DeviceHealthInput
					if p.connManager != nil {
						h := p.connManager.GetDeviceHealth(cachedDev.ID)
						hIn = toHealthInput(h)
					}
					p.alarmEvaluator.EvaluateDeviceOffline(ctx, cachedDev, hIn)
				}
				if currentIntervalMs != 5000 {
					currentIntervalMs = 5000
					ticker.Reset(5 * time.Second)
				}
				continue
			}

			// Execute cached parameter poll cycle with active observability
			pollStart := time.Now()
			if mon := p.getActivityMonitor(); mon != nil {
				mon.MarkPollStarted(deviceID)
				mon.UpdateDevicePollingSnapshot(deviceID, func(s *DevicePollingSnapshot) {
					s.PollingState = PollingStatePolling
					s.LastPollingStarted = &pollStart
				})
			}

			var dueParams []model.Parameter
			for _, param := range cachedParams {
				pIntMs := ResolveParamPollingIntervalMs(&param, baseIntervalMs)
				lastPoll, hasPolled := paramLastPolled[param.ID]
				if !hasPolled || pollStart.Sub(lastPoll) >= time.Duration(pIntMs-currentIntervalMs/2)*time.Millisecond {
					dueParams = append(dueParams, param)
				}
			}

			if len(dueParams) == 0 && len(cachedParams) > 0 {
				continue
			}

			cycleSuccess := p.pollDevice(ctx, cachedDev, dueParams)
			if cycleSuccess {
				for _, param := range dueParams {
					paramLastPolled[param.ID] = pollStart
				}
			}
			pollDurationMs := int(time.Since(pollStart).Milliseconds())

			if mon := p.getActivityMonitor(); mon != nil {
				mon.MarkPollEnded(deviceID)
				pollEnd := time.Now()
				mon.UpdateDevicePollingSnapshot(deviceID, func(s *DevicePollingSnapshot) {
					s.LastPollingCompleted = &pollEnd
					s.PollingDurationMs = pollDurationMs
				})
			}

			hasFormula := deviceHasActiveFormula(cachedDev)
			GetSharedAdapterTracker().SetDeviceFormulaActive(cachedDev.ID, hasFormula)
			if cycleSuccess || hasFormula {
				now := time.Now().UTC()
				cachedDev.LastDataAt = &now
			}
			if p.alarmEvaluator != nil {
				if cycleSuccess || hasFormula {
					p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, cachedDev, nil)
				} else {
					var hIn *service.DeviceHealthInput
					if p.connManager != nil {
						h := p.connManager.GetDeviceHealth(cachedDev.ID)
						hIn = toHealthInput(h)
					}
					p.alarmEvaluator.EvaluateDeviceOffline(ctx, cachedDev, hIn)
				}
			}
			if cycleSuccess || hasFormula {
				if !cycleSuccess && hasFormula && p.deviceService != nil {
					_ = p.deviceService.SetOnline(cachedDev.ID)
				}
				if p.connManager != nil && p.connManager.GetDeviceHealth(cachedDev.ID).ConsecutiveFailures > 0 {
					go func(d *model.Device) {
						reconnCtx, reconnCancel := context.WithTimeout(context.Background(), 45*time.Second)
						defer reconnCancel()
						_ = p.connManager.ReconnectDevice(reconnCtx, d)
					}(cachedDev)
				}
				wasFailing := consecutiveFailures > 0
				consecutiveFailures = 0
				if wasFailing || currentIntervalMs != baseIntervalMs {
					currentIntervalMs = baseIntervalMs
					ticker.Reset(time.Duration(currentIntervalMs) * time.Millisecond)
					logger.Info("Device %s (%d) communication active, restored normal polling interval (%dms)",
						cachedDev.DeviceCode, deviceID, baseIntervalMs)
				}

				GetSharedAdapterTracker().OnCommunicationSuccess(cachedDev.ID, pollDurationMs)
			} else {
				// Check physical serial port presence for Modbus RTU devices
				isSerial := cachedDev.Connection != nil && cachedDev.Connection.Protocol == model.ProtocolModbusRTU
				portMissing := false
				targetPort := ""
				if isSerial && cachedDev.Connection.SerialPort != "" {
					targetPort = modbus.ResolvePortAddress(cachedDev.Connection.SerialPort)
					if targetPort == "" {
						targetPort = cachedDev.Connection.SerialPort
					}
					portMissing = !modbus.IsPortPhysicallyPresent(targetPort)
				}

				if portMissing {
					consecutiveFailures++
					GetSharedAdapterTracker().OnPortMissing(cachedDev.ID, targetPort, fmt.Sprintf("Serial port %s disconnected/missing", targetPort))

					// Keep checking at 1500ms intervals so that hot-plug reinsertion is detected immediately!
					if currentIntervalMs != 1500 {
						currentIntervalMs = 1500
						ticker.Reset(1500 * time.Millisecond)
					}
				} else {
					consecutiveFailures++
					// If the port is present (e.g. re-plugged), notify recovery of physical link
					if isSerial {
						GetSharedAdapterTracker().OnPortRecovered(cachedDev.ID, targetPort)
						if p.connManager != nil {
							p.connManager.ResetDeviceBackoff(cachedDev.ID)
						}
					}
					if consecutiveFailures >= 2 {
						if p.deviceService != nil {
							_ = p.deviceService.SetConnectionError(cachedDev.ID, "Device communication failure")
						}
						if p.hub != nil {
							p.broadcastForDevice(cachedDev, websocket.WSMessage{
								Type:      "device.connection.changed",
								Timestamp: time.Now(),
								Data: map[string]interface{}{
									"device_id":         cachedDev.ID,
									"id":                cachedDev.ID,
									"device_code":       cachedDev.DeviceCode,
									"connection_status": CommOffline,
									"transport_status":  TransportAvailable,
									"status":            CommOffline,
									"reason":            "COMMUNICATION_FAILURE",
								},
							})
						}
						if mon := p.getActivityMonitor(); mon != nil {
							mon.RecordEvent(ActivityEvent{
								Severity:    SeverityError,
								Category:    CategoryCommunication,
								Component:   "PollingEngine",
								DeviceID:    cachedDev.ID,
								DeviceCode:  cachedDev.DeviceCode,
								DeviceName:  cachedDev.DeviceName,
								EventName:   "COMMUNICATION_TIMEOUT",
								Description: fmt.Sprintf("Device %s communication timeout (%d consecutive failures)", cachedDev.DeviceCode, consecutiveFailures),
								ErrorDetail: "Awaiting automatic retry/reconnect recovery",
							})
							mon.RecordEvent(ActivityEvent{
								Severity:    SeverityError,
								Category:    CategoryCommunication,
								Component:   "PollingEngine",
								DeviceID:    cachedDev.ID,
								DeviceCode:  cachedDev.DeviceCode,
								DeviceName:  cachedDev.DeviceName,
								EventName:   "DEVICE_OFFLINE",
								Description: fmt.Sprintf("Device %s marked OFFLINE: consecutive communication failures threshold reached", cachedDev.DeviceCode),
								ErrorDetail: "Awaiting automatic retry/reconnect recovery",
							})
							mon.UpdateDevicePollingSnapshot(cachedDev.ID, func(s *DevicePollingSnapshot) {
								s.CommunicationStatus = CommOffline
								s.TransportStatus = TransportAvailable
								s.PollingState = PollingStateRetrying
								s.ConsecutiveFailures = consecutiveFailures
								s.RecoveryState = "RECONNECTING"
								s.LastError = "Communication failure (threshold reached)"
								s.IsFresh = false
							})
						}
					} else {
						if mon := p.getActivityMonitor(); mon != nil {
							mon.UpdateDevicePollingSnapshot(cachedDev.ID, func(s *DevicePollingSnapshot) {
								s.CommunicationStatus = CommDegraded
								s.PollingState = PollingStateRetrying
								s.ConsecutiveFailures = consecutiveFailures
								s.RecoveryState = "RETRYING"
							})
						}
					}

					// Auto-recovery (Phase 4.1): Trigger asynchronous deduplicated reconnection attempt
					go func(d *model.Device) {
						reconnCtx, reconnCancel := context.WithTimeout(context.Background(), 45*time.Second)
						defer reconnCancel()
						_ = p.connManager.ReconnectDevice(reconnCtx, d)
					}(cachedDev)

					// Bounded adaptive backoff: capped at 5s max (instead of 15s) so sensor replug is detected promptly
					var backoffMs int
					if consecutiveFailures >= 5 {
						backoffMs = 5000 // 5s backoff
					} else if consecutiveFailures >= 2 {
						backoffMs = 3000 // 3s backoff
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
}

// pollDevice dispatches polling execution to batch mode or single read mode
func (p *PollingEngine) pollDevice(ctx context.Context, device *model.Device, enabledParams []model.Parameter) bool {
	if len(enabledParams) == 0 {
		err := p.connManager.ConnectDevice(ctx, device)
		return err == nil
	}

	if device.Connection != nil && device.Connection.BatchReadEnabled {
		return p.pollDeviceBatch(ctx, device, enabledParams)
	}

	return p.pollDeviceSingle(ctx, device, enabledParams)
}

// pollDeviceSingle executes the parameter read loop sequentially (Single Read mode)
func (p *PollingEngine) pollDeviceSingle(ctx context.Context, device *model.Device, enabledParams []model.Parameter) bool {
	if device != nil {
		cp := *device
		if len(cp.Parameters) == 0 {
			cp.Parameters = enabledParams
		}
		GetSharedAdapterTracker().SetDeviceFormulaActive(device.ID, deviceHasActiveFormula(&cp))
	}
	anySuccess := false
	source := ""
	if device != nil && device.Connection != nil {
		source = string(device.Connection.Protocol)
	}

	for idx, param := range enabledParams {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		res, err := p.ReadParameter(ctx, device, &param)
		if err != nil {
			hasFormula := deviceHasActiveFormula(device)
			if !anySuccess && !hasFormula {
				if p.alarmEvaluator != nil && p.connManager != nil {
					h := p.connManager.GetDeviceHealth(device.ID)
					p.alarmEvaluator.EvaluateDeviceCommFailure(ctx, device, toHealthInput(h), err)
				}

				if p.hub != nil {
					p.broadcastForDevice(device, websocket.WSMessage{
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
			}

			if isFatalCommError(err) {
				if p.markRemainingParamsBad(device, enabledParams, idx+1, err, source, anySuccess || hasFormula) {
					anySuccess = true
				}
				break
			}
			continue
		}

		anySuccess = true
		if p.alarmEvaluator != nil && p.connManager != nil {
			h := p.connManager.GetDeviceHealth(device.ID)
			p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, device, toHealthInput(h))
		}

		if p.hub != nil {
			p.broadcastForDevice(device, websocket.WSMessage{
				Type:      "device.communication.success",
				Timestamp: res.Timestamp,
				Data:      res,
			})
		}

		// If physical transport suffered a fatal timeout/disconnect and only succeeded via
		// formula/hold fallback, evaluate remaining parameters in-memory immediately instead
		// of stalling on repeated physical serial timeouts for every channel in this cycle.
		if res != nil && isFatalCommError(res.TransportErr) {
			p.markRemainingParamsBad(device, enabledParams, idx+1, res.TransportErr, source, true)
			break
		}
	}

	if anySuccess {
		if p.deviceService != nil {
			_ = p.deviceService.SetOnline(device.ID)
		}
		if p.hub != nil {
			p.broadcastForDevice(device, websocket.WSMessage{
				Type:      "device.connection.changed",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"device_id":         device.ID,
					"id":                device.ID,
					"device_code":       device.DeviceCode,
					"connection_status": "ONLINE",
					"status":            "ONLINE",
				},
			})
		}
	}

	return anySuccess
}

// pollDeviceBatch executes consolidated Modbus batch read jobs with controlled single-read fallback
func (p *PollingEngine) pollDeviceBatch(ctx context.Context, device *model.Device, enabledParams []model.Parameter) bool {
	if device != nil {
		cp := *device
		if len(cp.Parameters) == 0 {
			cp.Parameters = enabledParams
		}
		GetSharedAdapterTracker().SetDeviceFormulaActive(device.ID, deviceHasActiveFormula(&cp))
	}
	plan := modbus.PlanDeviceBatches(device, enabledParams)
	if len(plan.Jobs) == 0 {
		return false
	}

	anySuccess := false
	polledIDs := make(map[uint]bool, len(enabledParams))
	source := ""
	if device != nil && device.Connection != nil {
		source = string(device.Connection.Protocol)
	}

jobLoop:
	for _, job := range plan.Jobs {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		// 1. Single-parameter job (either opt-out SINGLE mode or isolated register)
		if job.IsSingle && len(job.Parameters) == 1 {
			param := job.Parameters[0].Parameter
			polledIDs[param.ID] = true
			res, err := p.ReadParameter(ctx, device, param)
			if err != nil {
				hasFormula := deviceHasActiveFormula(device)
				if !anySuccess && !hasFormula {
					if p.alarmEvaluator != nil && p.connManager != nil {
						h := p.connManager.GetDeviceHealth(device.ID)
						p.alarmEvaluator.EvaluateDeviceCommFailure(ctx, device, toHealthInput(h), err)
					}
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
				}
				if isFatalCommError(err) {
					rem := filterUnpolledParams(enabledParams, polledIDs)
					if p.markRemainingParamsBad(device, rem, 0, err, source, anySuccess || hasFormula) {
						anySuccess = true
					}
					break jobLoop
				}
				continue
			}

			anySuccess = true
			if p.alarmEvaluator != nil && p.connManager != nil {
				h := p.connManager.GetDeviceHealth(device.ID)
				p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, device, toHealthInput(h))
			}
			if p.hub != nil {
				p.broadcastForDevice(device, websocket.WSMessage{
					Type:      "device.communication.success",
					Timestamp: res.Timestamp,
					Data:      res,
				})
			}
			if res != nil && isFatalCommError(res.TransportErr) {
				rem := filterUnpolledParams(enabledParams, polledIDs)
				p.markRemainingParamsBad(device, rem, 0, res.TransportErr, source, true)
				break jobLoop
			}
			continue
		}

		// 2. Multi-parameter Batch Job
		batchKey := fmt.Sprintf("%d_%s", device.ID, job.ID)
		p.batchFailuresMu.Lock()
		consecBatchFailures := p.batchFailures[batchKey]
		p.batchFailuresMu.Unlock()

		// Circuit Breaker: If batch failed 3 consecutive times, execute single-read fallback directly.
		// After 10 fallback cycles (total 13 failures), reset counter to 0 to trial batch read again (half-open).
		if consecBatchFailures >= 3 {
			if consecBatchFailures >= 13 {
				p.batchFailuresMu.Lock()
				p.batchFailures[batchKey] = 0
				p.batchFailuresMu.Unlock()
				logger.Info("Batch %s on device %s half-open trial: retrying batch read", job.ID, device.DeviceCode)
			} else {
				p.batchFailuresMu.Lock()
				p.batchFailures[batchKey]++
				p.batchFailuresMu.Unlock()
				logger.Warn("Batch %s on device %s circuit-broken (%d failures); executing single-read fallback",
					job.ID, device.DeviceCode, consecBatchFailures)
				for _, sp := range job.Parameters {
					_, ok, abort := p.executeSingleFallback(ctx, device, sp.Parameter, enabledParams, polledIDs, anySuccess, source)
					if ok {
						anySuccess = true
					}
					if abort {
						break jobLoop
					}
				}
				continue
			}
		}

		var bCodes []string
		var bNames []string
		for _, sp := range job.Parameters {
			c := sp.Parameter.ParameterCode
			if c == "" {
				c = sp.Parameter.Code
			}
			n := sp.Parameter.ParameterName
			if n == "" {
				n = sp.Parameter.Name
			}
			if c != "" {
				bCodes = append(bCodes, c)
				if n != "" && n != c {
					bNames = append(bNames, fmt.Sprintf("%s (%s)", c, n))
				} else {
					bNames = append(bNames, c)
				}
			}
		}

		req := modbus.ModbusReadRequest{
			SlaveID:         job.SlaveID,
			FunctionCode:    job.FunctionCode,
			StartingAddress: job.StartingAddress,
			Quantity:        job.Quantity,
			ParameterCode:   strings.Join(bCodes, ", "),
			ParameterName:   strings.Join(bNames, ", "),
		}

		resp, err := p.connManager.ExecuteWithRetry(ctx, device, req)
		if err != nil {
			// Batch Read Failed! Track failure and initiate controlled single-read fallback
			p.batchFailuresMu.Lock()
			p.batchFailures[batchKey]++
			p.batchFailuresMu.Unlock()

			logger.Warn("Batch read %s (dev %s, addr %d, qty %d) failed: %v. Initiating controlled single-read fallback...",
				job.ID, device.DeviceCode, job.StartingAddress, job.Quantity, err)

			if isFatalCommError(err) {
				rem := filterUnpolledParams(enabledParams, polledIDs)
				cp := *device
				if len(cp.Parameters) == 0 {
					cp.Parameters = enabledParams
				}
				if p.markRemainingParamsBad(device, rem, 0, err, source, anySuccess || deviceHasActiveFormula(&cp)) {
					anySuccess = true
				}
				break jobLoop
			}

			// Controlled fallback: read each parameter in this batch individually
			for _, sp := range job.Parameters {
				_, ok, abort := p.executeSingleFallback(ctx, device, sp.Parameter, enabledParams, polledIDs, anySuccess, source)
				if ok {
					anySuccess = true
				}
				if abort {
					break jobLoop
				}
			}
			continue
		}

		// Batch Read Succeeded! Reset failure counter
		p.batchFailuresMu.Lock()
		delete(p.batchFailures, batchKey)
		p.batchFailuresMu.Unlock()

		anySuccess = true
		for _, sp := range job.Parameters {
			polledIDs[sp.Parameter.ID] = true
		}
		if p.alarmEvaluator != nil && p.connManager != nil {
			h := p.connManager.GetDeviceHealth(device.ID)
			p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, device, toHealthInput(h))
		}

		// Slice response data and decode each parameter independently
		for _, sp := range job.Parameters {
			select {
			case <-ctx.Done():
				return false
			default:
			}

			var paramBytes []byte
			if job.FunctionCode == modbus.FunctionReadCoils || job.FunctionCode == modbus.FunctionReadDiscreteInputs {
				val, bitErr := modbus.DecodeCoilBit(resp.Data, sp.BitOffset)
				if bitErr != nil {
					logger.Error("Failed decoding coil bit for %s: %v", sp.Parameter.ParameterCode, bitErr)
					continue
				}
				if val {
					paramBytes = []byte{0x01}
				} else {
					paramBytes = []byte{0x00}
				}
			} else {
				if sp.ByteOffset+sp.ByteLength > len(resp.Data) {
					logger.Error("Slice offset [%d:%d] out of range for response length %d on param %s",
						sp.ByteOffset, sp.ByteOffset+sp.ByteLength, len(resp.Data), sp.Parameter.ParameterCode)
					continue
				}
				paramBytes = resp.Data[sp.ByteOffset : sp.ByteOffset+sp.ByteLength]
			}

			res, pErr := p.processDecodedParameter(
				ctx,
				device,
				sp.Parameter,
				paramBytes,
				job.FunctionCode,
				sp.PDUAddress,
				sp.RegisterCount,
				resp.ResponseTime,
			)
			if pErr != nil {
				logger.Error("Failed processing batch parameter %s: %v", sp.Parameter.ParameterCode, pErr)
				continue
			}

			if p.hub != nil {
				p.broadcastForDevice(device, websocket.WSMessage{
					Type:      "device.communication.success",
					Timestamp: res.Timestamp,
					Data:      res,
				})
			}
		}
	}

	if anySuccess {
		if p.deviceService != nil {
			_ = p.deviceService.SetOnline(device.ID)
		}
		if p.hub != nil {
			p.broadcastForDevice(device, websocket.WSMessage{
				Type:      "device.connection.changed",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"device_id":         device.ID,
					"id":                device.ID,
					"device_code":       device.DeviceCode,
					"connection_status": "ONLINE",
					"status":            "ONLINE",
				},
			})
		}
	}

	return anySuccess
}

func isFatalCommError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "offline") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "dial failed") ||
		strings.Contains(errStr, "cannot find") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "input/output error") ||
		strings.Contains(errStr, "bad file descriptor") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "pipe closed") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "eof") ||
		strings.Contains(errStr, "access is denied") ||
		strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "not functioning") ||
		strings.Contains(errStr, "not connected") ||
		strings.Contains(errStr, "serial port offline") ||
		strings.Contains(errStr, "failed after") ||
		strings.Contains(errStr, "no such file") ||
		strings.Contains(errStr, "busy")
}

func filterUnpolledParams(params []model.Parameter, polledIDs map[uint]bool) []model.Parameter {
	if len(polledIDs) == 0 {
		return params
	}
	rem := make([]model.Parameter, 0, len(params))
	for _, p := range params {
		if !polledIDs[p.ID] {
			rem = append(rem, p)
		}
	}
	return rem
}

func (p *PollingEngine) ingestFallbackReading(
	device *model.Device,
	param *model.Parameter,
	fc byte,
	pdu uint16,
	qty uint16,
	elapsed time.Duration,
	effVal float64,
	formulaVal *float64,
	isHeld bool,
	quality model.TelemetryQuality,
	pErr error,
	source string,
) *CommunicationResult {
	finalVal := effVal
	if formulaVal != nil {
		finalVal = *formulaVal
	}

	if p.deviceService != nil {
		_ = p.deviceService.UpdateParameterCurrentValue(param.ID, finalVal, formulaVal, isHeld)
		_ = p.deviceService.UpdateLastData(device.ID)
	}

	qReason := model.ReasonNone
	qFlags := ""
	if isHeld {
		qReason = model.ReasonHoldAnomalyGrace
		if formulaVal != nil {
			qFlags = model.FlagHeldAnomaly + "," + model.FlagFormulaFallback
		} else {
			qFlags = model.FlagHeldAnomaly
		}
	} else if formulaVal != nil {
		qFlags = model.FlagFormulaFallback
	}

	effectiveFormula := param.Formula
	if param.CalibrationEnabled || param.CalibrationSessionActive {
		effectiveFormula = ""
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
			Value:         finalVal,
			FormulaValue:  formulaVal,
			Formula:       effectiveFormula,
			IsHeldValue:   isHeld,
			Quality:       quality,
			QualityReason: qReason,
			QualityFlags:  qFlags,
			Source:        source,
			ReceivedAt:    time.Now().UTC(),
			Parameter:     param,
			ReadError:     pErr,
		})
	}

	connID := uint(0)
	if device.Connection != nil {
		connID = device.Connection.ID
	}

	res := NewSuccessResult(
		device.ID,
		connID,
		param.ID,
		param.ParameterCode,
		fc,
		pdu,
		qty,
		nil,
		effVal,
		finalVal,
		elapsed,
	)
	res.ScaledValue = effVal
	res.Formula = effectiveFormula
	res.FormulaValue = formulaVal
	res.IsHeld = isHeld
	res.Quality = quality
	res.QualityReason = qReason
	res.QualityFlags = qFlags
	res.ProcessedValue = finalVal
	return res
}

func (p *PollingEngine) markRemainingParamsBad(device *model.Device, params []model.Parameter, fromIdx int, err error, source string, priorSuccess bool) bool {
	nowErr := time.Now().UTC()
	anyRecovered := false
	for j := fromIdx; j < len(params); j++ {
		remParam := params[j]

		var effVal float64
		var formulaVal *float64
		var isHeld bool
		var quality model.TelemetryQuality = model.QualityBad
		var pErr error = err

		if p.holdTracker != nil {
			effVal, formulaVal, isHeld, quality, pErr = p.holdTracker.ProcessReading(&remParam, 0, 0, err)
		} else if paramHasActiveFormula(&remParam) {
			fallbackEng, fallbackRaw := resolveFormulaFallbackInputs(&remParam, nil)
			if fVal := evalParamFormula(remParam.Formula, fallbackEng, fallbackRaw); fVal != nil {
				effVal = fallbackEng
				formulaVal = fVal
				quality = model.QualityGood
				pErr = nil
			}
		}

		if isHeld || (formulaVal != nil && pErr == nil) {
			fc := byte(0x03)
			pdu := uint16(0)
			if r, rErr := modbus.ResolveRegisterAddress(remParam.RegisterType, remParam.RegisterAddress); rErr == nil {
				fc = r.FunctionCode
				pdu = r.PDUAddress
			}
			qty := modbus.RequiredRegisterCount(remParam.DataType)
			if fc == FunctionReadCoils || fc == FunctionReadDiscreteInputs {
				qty = 1
			}
			res := p.ingestFallbackReading(device, &remParam, fc, pdu, qty, 0, effVal, formulaVal, isHeld, quality, pErr, source)
			if p.hub != nil {
				p.broadcastForDevice(device, websocket.WSMessage{
					Type:      "device.communication.success",
					Timestamp: res.Timestamp,
					Data:      res,
				})
			}
			anyRecovered = true
			continue
		}

		lastKnownVal := 0.0
		if remParam.CurrentProcessedValue != nil {
			lastKnownVal = *remParam.CurrentProcessedValue
		} else if remParam.CurrentValue != nil {
			lastKnownVal = *remParam.CurrentValue
		}
		if p.telemetryService != nil {
			_ = p.telemetryService.Ingest(&service.TelemetryIngestPayload{
				DeviceID:      device.ID,
				DeviceCode:    device.DeviceCode,
				DeviceName:    device.DeviceName,
				ParameterID:   remParam.ID,
				ParameterCode: remParam.ParameterCode,
				ParameterName: remParam.ParameterName,
				Unit:          remParam.Unit,
				DataType:      remParam.DataType,
				RawValue:      lastKnownVal,
				Value:         lastKnownVal,
				Quality:       model.QualityBad,
				QualityReason: model.ReasonCommunicationError,
				QualityFlags:  model.FlagCommError,
				IsHeldValue:   false,
				Source:        source,
				ErrorMessage:  err.Error(),
				ReceivedAt:    nowErr,
				Parameter:     &remParam,
				ReadError:     err,
			})
		}
	}

	if !anyRecovered && !priorSuccess && !deviceHasActiveFormula(device) {
		if p.deviceService != nil {
			_ = p.deviceService.SetOffline(device.ID)
		}
		if p.hub != nil {
			p.broadcastForDevice(device, websocket.WSMessage{
				Type:      "device.connection.changed",
				Timestamp: nowErr,
				Data: map[string]interface{}{
					"device_id":         device.ID,
					"id":                device.ID,
					"device_code":       device.DeviceCode,
					"connection_status": "OFFLINE",
					"status":            "OFFLINE",
					"reason":            "PORT_DISCONNECTED",
					"error":             err.Error(),
				},
			})
		}
	}
	return anyRecovered
}

func (p *PollingEngine) executeSingleFallback(
	ctx context.Context,
	device *model.Device,
	param *model.Parameter,
	enabledParams []model.Parameter,
	polledIDs map[uint]bool,
	priorSuccess bool,
	source string,
) (*CommunicationResult, bool, bool) {
	select {
	case <-ctx.Done():
		return nil, false, true
	default:
	}

	if polledIDs != nil {
		polledIDs[param.ID] = true
	}

	res, err := p.ReadParameter(ctx, device, param)
	if err != nil {
		hasFormula := deviceHasActiveFormula(device)
		if !priorSuccess && !hasFormula {
			if p.alarmEvaluator != nil && p.connManager != nil {
				h := p.connManager.GetDeviceHealth(device.ID)
				p.alarmEvaluator.EvaluateDeviceCommFailure(ctx, device, toHealthInput(h), err)
			}
			if p.hub != nil {
				p.broadcastForDevice(device, websocket.WSMessage{
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
		}
		if isFatalCommError(err) {
			rem := filterUnpolledParams(enabledParams, polledIDs)
			recovered := p.markRemainingParamsBad(device, rem, 0, err, source, priorSuccess || hasFormula)
			return nil, recovered, true
		}
		return nil, false, false
	}

	if p.alarmEvaluator != nil && p.connManager != nil {
		h := p.connManager.GetDeviceHealth(device.ID)
		p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, device, toHealthInput(h))
	}
	if p.hub != nil {
		p.broadcastForDevice(device, websocket.WSMessage{
			Type:      "device.communication.success",
			Timestamp: res.Timestamp,
			Data:      res,
		})
	}
	if res != nil && isFatalCommError(res.TransportErr) {
		rem := filterUnpolledParams(enabledParams, polledIDs)
		p.markRemainingParamsBad(device, rem, 0, res.TransportErr, source, true)
		return res, true, true
	}
	return res, true, false
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
	if param.SlaveID > 0 {
		slaveID = byte(param.SlaveID)
	}
	if slaveID == 0 {
		slaveID = 1
	}

	pCode := param.ParameterCode
	if pCode == "" {
		pCode = param.Code
	}
	pName := param.ParameterName
	if pName == "" {
		pName = param.Name
	}
	if pName != "" && pName != pCode {
		pName = fmt.Sprintf("%s (%s)", pCode, pName)
	} else if pName == "" {
		pName = pCode
	}

	req := ModbusReadRequest{
		SlaveID:         slaveID,
		FunctionCode:    res.FunctionCode,
		StartingAddress: res.PDUAddress,
		Quantity:        qty,
		ParameterCode:   pCode,
		ParameterName:   pName,
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

		// Evaluate anomaly hold tracker & custom formula fallback
		var effVal float64
		var formulaVal *float64
		var isHeld bool
		var quality model.TelemetryQuality = model.QualityBad
		var pErr error = err

		if p.holdTracker != nil {
			effVal, formulaVal, isHeld, quality, pErr = p.holdTracker.ProcessReading(param, 0, 0, err)
		} else if paramHasActiveFormula(param) {
			fallbackEng, fallbackRaw := resolveFormulaFallbackInputs(param, nil)
			if fVal := evalParamFormula(param.Formula, fallbackEng, fallbackRaw); fVal != nil {
				effVal = fallbackEng
				formulaVal = fVal
				quality = model.QualityGood
				pErr = nil
			}
		}

		if isHeld || (formulaVal != nil && pErr == nil) {
			fallbackRes := p.ingestFallbackReading(
				device,
				param,
				res.FunctionCode,
				res.PDUAddress,
				qty,
				elapsed,
				effVal,
				formulaVal,
				isHeld,
				quality,
				pErr,
				source,
			)
			fallbackRes.TransportErr = err
			return fallbackRes, nil
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

	// 4. Delegate decoding and telemetry ingestion to shared processDecodedParameter
	return p.processDecodedParameter(
		ctx,
		device,
		param,
		resp.Data,
		res.FunctionCode,
		res.PDUAddress,
		qty,
		resp.ResponseTime,
	)
}

// processDecodedParameter decodes raw register bytes and executes the telemetry ingestion pipeline
func (p *PollingEngine) processDecodedParameter(
	ctx context.Context,
	device *model.Device,
	param *model.Parameter,
	rawBytes []byte,
	fc byte,
	pdu uint16,
	qty uint16,
	responseTime time.Duration,
) (*CommunicationResult, error) {
	byteOrder := param.ByteOrder
	if byteOrder == "" && device.Connection != nil {
		byteOrder = device.Connection.ByteOrder
	}
	if byteOrder == "" {
		byteOrder = "ABCD"
	}

	decoded, err := modbus.DecodeRegisters(rawBytes, param.DataType, byteOrder)
	if err != nil {
		return nil, fmt.Errorf("register decode failed: %w", err)
	}

	// Apply Base Register Scale & Offset (Pre-Calibration Sensor Value x)
	uncalVal := modbus.ApplyScaleAndOffset(decoded.RawValue, param.Scale, param.Offset, param.Precision)
	// Apply Multi-Point Sensor Calibration Curve (y = f(x)) when enabled
	engVal := param.ApplyCalibration(uncalVal)

	// Evaluate Anomaly Hold Tracker & Custom Formula
	effVal := engVal
	var formulaVal *float64
	var isHeld bool
	quality := model.QualityGood

	if p.holdTracker != nil {
		effVal, formulaVal, isHeld, quality, _ = p.holdTracker.ProcessReading(param, decoded.RawValue, engVal, nil)
	}

	// Update Device and Parameter State
	finalVal := effVal
	if formulaVal != nil {
		finalVal = *formulaVal
	}

	if p.deviceService != nil {
		_ = p.deviceService.UpdateParameterCurrentReadings(param.ID, &uncalVal, finalVal, formulaVal, isHeld)
		_ = p.deviceService.UpdateLastData(device.ID)
	}

	result := NewSuccessResult(
		device.ID,
		device.Connection.ID,
		param.ID,
		param.ParameterCode,
		fc,
		pdu,
		qty,
		rawBytes,
		decoded.RawValue,
		finalVal,
		responseTime,
	)
	result.ScaledValue = uncalVal
	result.Formula = param.Formula
	result.FormulaValue = formulaVal
	result.IsHeld = isHeld
	result.Quality = quality
	result.ProcessedValue = finalVal

	// Pipeline Telemetry Ingestion (Phase 3.1 & 3.2)
	if p.telemetryService != nil {
		source := ""
		if device.Connection != nil {
			source = string(device.Connection.Protocol)
		}
		_ = p.telemetryService.Ingest(&service.TelemetryIngestPayload{
			DeviceID:          device.ID,
			DeviceCode:        device.DeviceCode,
			DeviceName:        device.DeviceName,
			ParameterID:       param.ID,
			ParameterCode:     param.ParameterCode,
			ParameterName:     param.ParameterName,
			Unit:              param.Unit,
			DataType:          param.DataType,
			RawBytes:          rawBytes,
			RawHex:            result.RawHex,
			RawValue:          decoded.RawValue,
			UncalibratedValue: &uncalVal,
			Value:             finalVal,
			FormulaValue:      formulaVal,
			Formula:           param.Formula,
			IsHeldValue:       isHeld,
			Quality:           quality,
			Source:            source,
			ReceivedAt:        result.Timestamp,
			Parameter:         param,
			ReadError:         nil,
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

// ResolveParamPollingIntervalMs determines the effective polling interval for a parameter.
// If the parameter specifies a custom PollingIntervalMs > 0, that interval is used;
// otherwise, it inherits from the device connection's base polling interval.
func ResolveParamPollingIntervalMs(param *model.Parameter, devIntervalMs int) int {
	if param != nil && param.PollingIntervalMs != nil && *param.PollingIntervalMs > 0 {
		ms := *param.PollingIntervalMs
		if ms > 0 && ms <= 60 {
			ms = ms * 1000
		}
		if ms < 200 {
			ms = 200
		}
		return ms
	}
	if devIntervalMs < 200 {
		return 1000
	}
	return devIntervalMs
}

func calculateDeviceTickIntervalMs(devIntervalMs int, params []model.Parameter) int {
	minInterval := devIntervalMs
	for _, p := range params {
		if p.Enabled && p.PollingIntervalMs != nil && *p.PollingIntervalMs > 0 {
			pMs := *p.PollingIntervalMs
			if pMs > 0 && pMs <= 60 {
				pMs = pMs * 1000
			}
			if pMs >= 200 && pMs < minInterval {
				minInterval = pMs
			}
		}
	}
	if minInterval < 200 {
		minInterval = 200
	}
	return minInterval
}

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
	batchFailuresMu  sync.Mutex
	batchFailures    map[string]int
	activityMonitor  *ActivityMonitor
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
		batchFailures: make(map[string]int),
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

// SetAlarmEvaluator injects the Phase 6 Alarm Evaluator
func (p *PollingEngine) SetAlarmEvaluator(ae *service.AlarmEvaluator) {
	p.alarmEvaluator = ae
}

// SetActivityMonitor attaches the system activity monitor
func (p *PollingEngine) SetActivityMonitor(mon *ActivityMonitor) {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()
	p.activityMonitor = mon
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
	if p.alarmEvaluator != nil {
		p.alarmEvaluator.ClearDeviceAlarms(p.ctx, deviceID)
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
						cachedDev.Parameters = cachedParams
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

			cycleSuccess := p.pollDevice(ctx, cachedDev, cachedParams)
			pollDurationMs := int(time.Since(pollStart).Milliseconds())

			if mon := p.getActivityMonitor(); mon != nil {
				mon.MarkPollEnded(deviceID)
				pollEnd := time.Now()
				mon.UpdateDevicePollingSnapshot(deviceID, func(s *DevicePollingSnapshot) {
					s.LastPollingCompleted = &pollEnd
					s.PollingDurationMs = pollDurationMs
				})
			}

			if cycleSuccess {
				now := time.Now().UTC()
				cachedDev.LastDataAt = &now
			}
			if p.alarmEvaluator != nil {
				var hIn *service.DeviceHealthInput
				if p.connManager != nil {
					h := p.connManager.GetDeviceHealth(cachedDev.ID)
					hIn = toHealthInput(h)
				}
				p.alarmEvaluator.EvaluateDeviceOffline(ctx, cachedDev, hIn)
			}
			if cycleSuccess {
				wasFailing := consecutiveFailures > 0
				if wasFailing || currentIntervalMs != baseIntervalMs {
					consecutiveFailures = 0
					currentIntervalMs = baseIntervalMs
					ticker.Reset(time.Duration(currentIntervalMs) * time.Millisecond)
					logger.Info("Device %s (%d) communication active, restored normal polling interval (%dms)",
						cachedDev.DeviceCode, deviceID, baseIntervalMs)

					if mon := p.getActivityMonitor(); mon != nil {
						mon.RecordEvent(ActivityEvent{
							Severity:    SeverityInfo,
							Category:    CategoryCommunication,
							Component:   "PollingEngine",
							DeviceID:    cachedDev.ID,
							DeviceCode:  cachedDev.DeviceCode,
							DeviceName:  cachedDev.DeviceName,
							EventName:   "COMMUNICATION_RECOVERED",
							Description: fmt.Sprintf("Device %s communication fully restored: valid Modbus response received", cachedDev.DeviceCode),
							DurationMs:  pollDurationMs,
						})
					}

					if p.hub != nil {
						p.hub.Broadcast(websocket.WSMessage{
							Type:      "device.connection.changed",
							Timestamp: time.Now(),
							Data: map[string]interface{}{
								"device_id":         cachedDev.ID,
								"id":                cachedDev.ID,
								"device_code":       cachedDev.DeviceCode,
								"connection_status": "ONLINE",
								"status":            "ONLINE",
								"reason":            "COMMUNICATION_RESTORED",
							},
						})
					}
				}

				if mon := p.getActivityMonitor(); mon != nil {
					localNow := time.Now()
					mon.UpdateDevicePollingSnapshot(cachedDev.ID, func(s *DevicePollingSnapshot) {
						s.CommunicationStatus = "ONLINE"
						s.PollingState = PollingStateSuccess
						s.LastSuccessfulResponse = &localNow
						s.LastSuccessfulTelemetry = &localNow
						s.ConsecutiveFailures = 0
						s.RecoveryState = "NORMAL"
						s.LastError = ""
					})
				}
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
					if mon := p.getActivityMonitor(); mon != nil {
						mon.RecordEvent(ActivityEvent{
							Severity:    SeverityWarning,
							Category:    CategorySerial,
							Component:   "PollingEngine",
							DeviceID:    cachedDev.ID,
							DeviceCode:  cachedDev.DeviceCode,
							DeviceName:  cachedDev.DeviceName,
							EventName:   "SERIAL_PORT_DISCONNECTED",
							Description: fmt.Sprintf("Serial port %s disconnected/missing. Awaiting hardware reconnect.", targetPort),
						})
						mon.UpdateDevicePollingSnapshot(cachedDev.ID, func(s *DevicePollingSnapshot) {
							s.CommunicationStatus = "OFFLINE"
							s.PollingState = PollingStateStopped
							s.ConsecutiveFailures = consecutiveFailures
							s.RecoveryState = "RECONNECTING"
							s.LastError = fmt.Sprintf("Serial port %s missing", targetPort)
						})
					}

					if p.deviceService != nil {
						_ = p.deviceService.SetOffline(cachedDev.ID)
					}
					if p.hub != nil {
						p.hub.Broadcast(websocket.WSMessage{
							Type:      "device.connection.changed",
							Timestamp: time.Now(),
							Data: map[string]interface{}{
								"device_id":         cachedDev.ID,
								"id":                cachedDev.ID,
								"device_code":       cachedDev.DeviceCode,
								"connection_status": "OFFLINE",
								"status":            "OFFLINE",
								"reason":            "PORT_DISCONNECTED",
							},
						})
					}
					// Keep checking at 1500ms intervals so that hot-plug reinsertion is detected immediately!
					if currentIntervalMs != 1500 {
						currentIntervalMs = 1500
						ticker.Reset(1500 * time.Millisecond)
					}
				} else {
					consecutiveFailures++
					// If the port is present (e.g. re-plugged), reset backoff immediately for instant recovery
					if isSerial && p.connManager != nil {
						p.connManager.ResetDeviceBackoff(cachedDev.ID)
					}
					if consecutiveFailures >= 2 {
						if p.deviceService != nil {
							_ = p.deviceService.SetConnectionError(cachedDev.ID, "Device communication failure")
						}
						if p.hub != nil {
							p.hub.Broadcast(websocket.WSMessage{
								Type:      "device.connection.changed",
								Timestamp: time.Now(),
								Data: map[string]interface{}{
									"device_id":         cachedDev.ID,
									"id":                cachedDev.ID,
									"device_code":       cachedDev.DeviceCode,
									"connection_status": "ERROR",
									"status":            "ERROR",
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
								EventName:   "DEVICE_OFFLINE_DETECTED",
								Description: fmt.Sprintf("Device %s communication failed: threshold reached (%d consecutive failures)", cachedDev.DeviceCode, consecutiveFailures),
								ErrorDetail: "Awaiting automatic retry/reconnect recovery",
							})
							mon.UpdateDevicePollingSnapshot(cachedDev.ID, func(s *DevicePollingSnapshot) {
								s.CommunicationStatus = "ERROR"
								s.PollingState = PollingStateRetrying
								s.ConsecutiveFailures = consecutiveFailures
								s.RecoveryState = "RECONNECTING"
								s.LastError = "Communication failure (threshold reached)"
							})
						}
					} else {
						if mon := p.getActivityMonitor(); mon != nil {
							mon.UpdateDevicePollingSnapshot(cachedDev.ID, func(s *DevicePollingSnapshot) {
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
		_ = p.connManager.ConnectDevice(ctx, device)
		return true
	}

	if device.Connection != nil && device.Connection.BatchReadEnabled {
		return p.pollDeviceBatch(ctx, device, enabledParams)
	}

	return p.pollDeviceSingle(ctx, device, enabledParams)
}

// pollDeviceSingle executes the parameter read loop sequentially (Single Read mode)
func (p *PollingEngine) pollDeviceSingle(ctx context.Context, device *model.Device, enabledParams []model.Parameter) bool {
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

			if isFatalCommError(err) {
				p.markRemainingParamsBad(device, enabledParams, idx, err, source)
				return false
			}
			continue
		}

		anySuccess = true
		if p.alarmEvaluator != nil && p.connManager != nil {
			h := p.connManager.GetDeviceHealth(device.ID)
			p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, device, toHealthInput(h))
		}

		if p.hub != nil {
			p.hub.Broadcast(websocket.WSMessage{
				Type:      "device.communication.success",
				Timestamp: res.Timestamp,
				Data:      res,
			})
		}
	}

	if anySuccess {
		if p.deviceService != nil {
			_ = p.deviceService.SetOnline(device.ID)
		}
		if p.hub != nil {
			p.hub.Broadcast(websocket.WSMessage{
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
	plan := modbus.PlanDeviceBatches(device, enabledParams)
	if len(plan.Jobs) == 0 {
		return false
	}

	anySuccess := false
	source := ""
	if device != nil && device.Connection != nil {
		source = string(device.Connection.Protocol)
	}

	for _, job := range plan.Jobs {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		// 1. Single-parameter job (either opt-out SINGLE mode or isolated register)
		if job.IsSingle && len(job.Parameters) == 1 {
			param := job.Parameters[0].Parameter
			res, err := p.ReadParameter(ctx, device, param)
			if err != nil {
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
				if isFatalCommError(err) {
					p.markRemainingParamsBad(device, enabledParams, 0, err, source)
					return false
				}
				continue
			}

			anySuccess = true
			if p.alarmEvaluator != nil && p.connManager != nil {
				h := p.connManager.GetDeviceHealth(device.ID)
				p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, device, toHealthInput(h))
			}
			if p.hub != nil {
				p.hub.Broadcast(websocket.WSMessage{
					Type:      "device.communication.success",
					Timestamp: res.Timestamp,
					Data:      res,
				})
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
					_, ok, abort := p.executeSingleFallback(ctx, device, sp.Parameter, enabledParams, source)
					if abort {
						return false
					}
					if ok {
						anySuccess = true
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
				p.markRemainingParamsBad(device, enabledParams, 0, err, source)
				return false
			}

			// Controlled fallback: read each parameter in this batch individually
			for _, sp := range job.Parameters {
				_, ok, abort := p.executeSingleFallback(ctx, device, sp.Parameter, enabledParams, source)
				if abort {
					return false
				}
				if ok {
					anySuccess = true
				}
			}
			continue
		}

		// Batch Read Succeeded! Reset failure counter
		p.batchFailuresMu.Lock()
		delete(p.batchFailures, batchKey)
		p.batchFailuresMu.Unlock()

		anySuccess = true
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
				p.hub.Broadcast(websocket.WSMessage{
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
			p.hub.Broadcast(websocket.WSMessage{
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

func (p *PollingEngine) markRemainingParamsBad(device *model.Device, params []model.Parameter, fromIdx int, err error, source string) {
	nowErr := time.Now().UTC()
	for j := fromIdx; j < len(params); j++ {
		remParam := params[j]
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
	if p.deviceService != nil {
		_ = p.deviceService.SetOffline(device.ID)
	}
	if p.hub != nil {
		p.hub.Broadcast(websocket.WSMessage{
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

func (p *PollingEngine) executeSingleFallback(
	ctx context.Context,
	device *model.Device,
	param *model.Parameter,
	enabledParams []model.Parameter,
	source string,
) (*CommunicationResult, bool, bool) {
	select {
	case <-ctx.Done():
		return nil, false, true
	default:
	}

	res, err := p.ReadParameter(ctx, device, param)
	if err != nil {
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
		if isFatalCommError(err) {
			p.markRemainingParamsBad(device, enabledParams, 0, err, source)
			return nil, false, true
		}
		return nil, false, false
	}

	if p.alarmEvaluator != nil && p.connManager != nil {
		h := p.connManager.GetDeviceHealth(device.ID)
		p.alarmEvaluator.EvaluateDeviceCommSuccess(ctx, device, toHealthInput(h))
	}
	if p.hub != nil {
		p.hub.Broadcast(websocket.WSMessage{
			Type:      "device.communication.success",
			Timestamp: res.Timestamp,
			Data:      res,
		})
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

	// Apply Scale & Offset
	engVal := modbus.ApplyScaleAndOffset(decoded.RawValue, param.Scale, param.Offset, param.Precision)

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
		_ = p.deviceService.UpdateParameterCurrentValue(param.ID, finalVal, formulaVal, isHeld)
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
	result.ScaledValue = engVal
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
			DeviceID:      device.ID,
			DeviceCode:    device.DeviceCode,
			DeviceName:    device.DeviceName,
			ParameterID:   param.ID,
			ParameterCode: param.ParameterCode,
			ParameterName: param.ParameterName,
			Unit:          param.Unit,
			DataType:      param.DataType,
			RawBytes:      rawBytes,
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

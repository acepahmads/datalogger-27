package communication

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/service"
	"datalogger/internal/websocket"
)

// Shared USB serial adapter tracker coordinates disconnect detection,
// shared bus state updates, and status consistency across multiple logical devices.
type USBAdapterTracker struct {
	mu                  sync.RWMutex
	deviceAdapters      map[uint]string            // deviceID -> adapterKey
	deviceCodes         map[uint]string            // deviceID -> deviceCode
	deviceNames         map[uint]string            // deviceID -> deviceName
	devicePorts         map[uint]string            // deviceID -> serialPort
	adapterDevices      map[string]map[uint]bool   // adapterKey -> set of deviceIDs
	adapterStatus       map[string]string          // adapterKey -> TransportStatus
	adapterMissing      map[string]bool            // adapterKey -> isMissing
	customAdapterKeys   map[string]string          // portPath -> adapterKey for testing/explicit configuration
	activityMonitor     *ActivityMonitor
	deviceService       *service.DeviceService
	hub                 *websocket.Hub
	connManager         *ConnectionManager
	adapterPresenceHook func(adapterKey string) bool // Optional test hook
}

var (
	globalAdapterTracker     *USBAdapterTracker
	globalAdapterTrackerOnce sync.Once
	byIDRegex                = regexp.MustCompile(`(-if[0-9a-fA-F]+.*|-port[0-9]+.*)$`)
	byPathRegex              = regexp.MustCompile(`(:1\.[0-9]+.*|-port[0-9]+.*)$`)
)

// GetSharedAdapterTracker returns the global USBAdapterTracker singleton
func GetSharedAdapterTracker() *USBAdapterTracker {
	globalAdapterTrackerOnce.Do(func() {
		globalAdapterTracker = &USBAdapterTracker{
			deviceAdapters:    make(map[uint]string),
			deviceCodes:       make(map[uint]string),
			deviceNames:       make(map[uint]string),
			devicePorts:       make(map[uint]string),
			adapterDevices:    make(map[string]map[uint]bool),
			adapterStatus:     make(map[string]string),
			adapterMissing:    make(map[string]bool),
			customAdapterKeys: make(map[string]string),
		}
	})
	return globalAdapterTracker
}

// ResetSharedAdapterTracker resets tracker state (primarily for tests)
func ResetSharedAdapterTracker() {
	tracker := GetSharedAdapterTracker()
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	tracker.deviceAdapters = make(map[uint]string)
	tracker.deviceCodes = make(map[uint]string)
	tracker.deviceNames = make(map[uint]string)
	tracker.devicePorts = make(map[uint]string)
	tracker.adapterDevices = make(map[string]map[uint]bool)
	tracker.adapterStatus = make(map[string]string)
	tracker.adapterMissing = make(map[string]bool)
	tracker.customAdapterKeys = make(map[string]string)
	tracker.adapterPresenceHook = nil
}

// SetActivityMonitor attaches the ActivityMonitor
func (t *USBAdapterTracker) SetActivityMonitor(mon *ActivityMonitor) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.activityMonitor = mon
}

// SetDeviceService attaches the DeviceService
func (t *USBAdapterTracker) SetDeviceService(ds *service.DeviceService) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.deviceService = ds
}

// SetHub attaches the WebSocket hub
func (t *USBAdapterTracker) SetHub(hub *websocket.Hub) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hub = hub
}

// SetConnectionManager attaches the ConnectionManager
func (t *USBAdapterTracker) SetConnectionManager(cm *ConnectionManager) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connManager = cm
}

// SetAdapterPresenceHookForTest registers a custom presence evaluator for tests
func (t *USBAdapterTracker) SetAdapterPresenceHookForTest(hook func(adapterKey string) bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.adapterPresenceHook = hook
}

// SetPortAdapterMappingForTest forces an explicit adapterKey for a port in unit tests
func (t *USBAdapterTracker) SetPortAdapterMappingForTest(port string, adapterKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.customAdapterKeys[port] = adapterKey
}

// ResolveAdapterKey extracts the physical USB adapter identity from a port path
func (t *USBAdapterTracker) ResolveAdapterKey(portPath string) string {
	t.mu.RLock()
	if custom, exists := t.customAdapterKeys[portPath]; exists {
		t.mu.RUnlock()
		return custom
	}
	t.mu.RUnlock()

	clean := strings.TrimSpace(portPath)
	if clean == "" {
		return ""
	}

	// 1. Check custom test prefixes
	if strings.HasPrefix(clean, "COM_TEST_QUAD") || strings.HasPrefix(clean, "VIRTUAL_QUAD") {
		return "QINHENG_USB_QUAD_SERIAL_MOCK"
	}

	// 2. Linux persistent by-id symlinks (e.g. /dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00)
	if strings.HasPrefix(clean, "/dev/serial/by-id/") {
		base := filepath.Base(clean)
		adapterKey := byIDRegex.ReplaceAllString(base, "")
		if adapterKey != "" {
			return adapterKey
		}
		return base
	}

	// 3. Linux persistent by-path symlinks (e.g. /dev/serial/by-path/platform-3f980000.usb-usb-0:1.1:1.0-port0)
	if strings.HasPrefix(clean, "/dev/serial/by-path/") {
		base := filepath.Base(clean)
		adapterKey := byPathRegex.ReplaceAllString(base, "")
		if adapterKey != "" {
			return adapterKey
		}
		return base
	}

	// 4. Linux sysfs path inspection for direct device nodes (/dev/ttyACM* or /dev/ttyUSB*)
	if runtime.GOOS != "windows" && (strings.HasPrefix(clean, "/dev/ttyACM") || strings.HasPrefix(clean, "/dev/ttyUSB")) {
		baseDev := filepath.Base(clean)
		sysDevicePath := filepath.Join("/sys/class/tty", baseDev, "device")
		if realTarget, err := filepath.EvalSymlinks(sysDevicePath); err == nil {
			// In sysfs, realTarget is e.g. /sys/devices/pci.../usb1/1-1/1-1.1/1-1.1:1.0
			// The parent directory represents the USB device itself (1-1.1)
			parentDir := filepath.Dir(realTarget)
			parentName := filepath.Base(parentDir)
			if parentName != "" && parentName != "usb" && parentName != "." && parentName != "/" {
				return "usb:" + parentName
			}
		}
	}

	return clean
}

// RegisterDevice registers a logical device with its serial port and resolves its adapter
func (t *USBAdapterTracker) RegisterDevice(deviceID uint, deviceCode, deviceName, serialPort string) {
	if serialPort == "" {
		return
	}

	adapterKey := t.ResolveAdapterKey(serialPort)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Clean previous registration if port changed
	if oldAdapter, exists := t.deviceAdapters[deviceID]; exists && oldAdapter != adapterKey {
		if devs, ok := t.adapterDevices[oldAdapter]; ok {
			delete(devs, deviceID)
			if len(devs) == 0 {
				delete(t.adapterDevices, oldAdapter)
				delete(t.adapterStatus, oldAdapter)
				delete(t.adapterMissing, oldAdapter)
			}
		}
	}

	t.deviceAdapters[deviceID] = adapterKey
	t.deviceCodes[deviceID] = deviceCode
	t.deviceNames[deviceID] = deviceName
	t.devicePorts[deviceID] = serialPort

	if _, exists := t.adapterDevices[adapterKey]; !exists {
		t.adapterDevices[adapterKey] = make(map[uint]bool)
		t.adapterStatus[adapterKey] = TransportAvailable
	}
	t.adapterDevices[adapterKey][deviceID] = true

	logger.Debug("Device %s (%d) registered on USB adapter key: '%s' (port: %s)",
		deviceCode, deviceID, adapterKey, serialPort)
}

// UnregisterDevice unregisters a device from the tracker
func (t *USBAdapterTracker) UnregisterDevice(deviceID uint) {
	t.mu.Lock()
	defer t.mu.Unlock()

	adapterKey, exists := t.deviceAdapters[deviceID]
	if !exists {
		return
	}

	delete(t.deviceAdapters, deviceID)
	delete(t.deviceCodes, deviceID)
	delete(t.deviceNames, deviceID)
	delete(t.devicePorts, deviceID)

	if devs, ok := t.adapterDevices[adapterKey]; ok {
		delete(devs, deviceID)
		if len(devs) == 0 {
			delete(t.adapterDevices, adapterKey)
			delete(t.adapterStatus, adapterKey)
			delete(t.adapterMissing, adapterKey)
		}
	}
}

// IsAdapterPresent checks whether the USB adapter hardware is currently present on the system
func (t *USBAdapterTracker) IsAdapterPresent(adapterKey string) bool {
	t.mu.RLock()
	hook := t.adapterPresenceHook
	t.mu.RUnlock()

	if hook != nil {
		return hook(adapterKey)
	}

	if adapterKey == "" {
		return false
	}

	// 1. Linux sysfs USB path check (e.g. usb:1-1.1)
	if strings.HasPrefix(adapterKey, "usb:") {
		usbID := strings.TrimPrefix(adapterKey, "usb:")
		sysPath := filepath.Join("/sys/bus/usb/devices", usbID)
		if _, err := os.Stat(sysPath); err != nil {
			return false
		}
		return true
	}

	// 2. Linux by-id prefix check (e.g. usb-1a86_USB_Quad_Serial_55D5)
	if runtime.GOOS != "windows" && strings.HasPrefix(adapterKey, "usb-") {
		matches, err := filepath.Glob(filepath.Join("/dev/serial/by-id", adapterKey+"*"))
		if err != nil || len(matches) == 0 {
			return false
		}
		// Verify at least one symlink points to a valid physical device node
		for _, m := range matches {
			if target, err := filepath.EvalSymlinks(m); err == nil {
				if _, statErr := os.Stat(target); statErr == nil {
					return true
				}
			}
		}
		return false
	}

	// 3. Fallback: check if the port path itself is present
	return modbus.IsPortPhysicallyPresent(adapterKey)
}

// IsPortPhysicallyPresent checks whether the specified serial port device is physically present on the OS
func IsPortPhysicallyPresent(port string) bool {
	return modbus.IsPortPhysicallyPresent(port)
}

// OnPortMissing is triggered when a port read/write or proactive stat check discovers that a port is missing.
// It propagates the disconnect to ALL devices sharing the same USB adapter!
func (t *USBAdapterTracker) OnPortMissing(deviceID uint, portPath string, reason string) {
	adapterKey := t.ResolveAdapterKey(portPath)

	t.mu.Lock()
	t.adapterMissing[adapterKey] = true
	t.adapterStatus[adapterKey] = TransportPortMissing

	// Collect affected device IDs
	var affectedIDs []uint
	if devs, exists := t.adapterDevices[adapterKey]; exists {
		for dID := range devs {
			affectedIDs = append(affectedIDs, dID)
		}
	} else if deviceID > 0 {
		affectedIDs = append(affectedIDs, deviceID)
	}

	mon := t.activityMonitor
	ds := t.deviceService
	hub := t.hub
	cm := t.connManager

	t.mu.Unlock()

	now := time.Now()
	logger.Warn("[USBAdapterTracker] USB adapter '%s' lost (trigger device %d on %s). Propagating OFFLINE to %d sharing devices: %v",
		adapterKey, deviceID, portPath, len(affectedIDs), affectedIDs)

	for _, dID := range affectedIDs {
		t.mu.RLock()
		devCode := t.deviceCodes[dID]
		devName := t.deviceNames[dID]
		devPort := t.devicePorts[dID]
		t.mu.RUnlock()

		if devCode == "" {
			devCode = fmt.Sprintf("DEV-%d", dID)
		}
		if devPort == "" {
			devPort = portPath
		}

		// 1. Update In-Memory Polling Snapshot
		if mon != nil {
			mon.UpdateDevicePollingSnapshot(dID, func(s *DevicePollingSnapshot) {
				s.TransportStatus = TransportPortMissing
				s.CommunicationStatus = CommOffline
				s.PollingState = PollingStateStopped
				s.RecoveryState = "RECONNECTING"
				s.ConsecutiveFailures++
				s.IsFresh = false
				s.LastError = fmt.Sprintf("Serial port %s disconnected/missing (adapter %s)", devPort, adapterKey)
			})

			// Record high-visibility audit events
			mon.RecordEvent(ActivityEvent{
				Severity:    SeverityWarning,
				Category:    CategorySerial,
				Component:   "USBAdapterTracker",
				DeviceID:    dID,
				DeviceCode:  devCode,
				DeviceName:  devName,
				EventName:   "USB_PORT_MISSING",
				Description: fmt.Sprintf("USB serial port %s missing (adapter %s). Awaiting hardware reconnect.", devPort, adapterKey),
				ErrorDetail: reason,
			})
			mon.RecordEvent(ActivityEvent{
				Severity:    SeverityError,
				Category:    CategoryCommunication,
				Component:   "USBAdapterTracker",
				DeviceID:    dID,
				DeviceCode:  devCode,
				DeviceName:  devName,
				EventName:   "DEVICE_OFFLINE",
				Description: fmt.Sprintf("Device %s marked OFFLINE: USB transport lost on adapter %s", devCode, adapterKey),
				ErrorDetail: reason,
			})
		}

		// 2. Update DB status
		if ds != nil {
			_ = ds.SetOffline(dID)
		}

		// 3. WebSocket Real-time Broadcast
		if hub != nil {
			hub.Broadcast(websocket.WSMessage{
				Type:      "device.connection.changed",
				Timestamp: now,
				Data: map[string]interface{}{
					"device_id":          dID,
					"id":                 dID,
					"device_code":        devCode,
					"connection_status":  "OFFLINE",
					"transport_status":   TransportPortMissing,
					"status":             "OFFLINE",
					"reason":             "USB_PORT_MISSING",
					"adapter_key":        adapterKey,
				},
			})
		}

		// 4. Clean up broken physical transport handle
		if cm != nil {
			_ = cm.DisconnectDevice(dID)
		}
	}
}

// OnPortRecovered is called when a previously missing port or adapter is detected back on the system.
// It sets TransportStatus to AVAILABLE and triggers auto-reconnect without marking the device ONLINE yet.
func (t *USBAdapterTracker) OnPortRecovered(deviceID uint, portPath string) {
	adapterKey := t.ResolveAdapterKey(portPath)

	t.mu.Lock()
	wasMissing := t.adapterMissing[adapterKey]
	t.adapterMissing[adapterKey] = false
	t.adapterStatus[adapterKey] = TransportAvailable

	var affectedIDs []uint
	if devs, exists := t.adapterDevices[adapterKey]; exists {
		for dID := range devs {
			affectedIDs = append(affectedIDs, dID)
		}
	} else if deviceID > 0 {
		affectedIDs = append(affectedIDs, deviceID)
	}

	mon := t.activityMonitor
	hub := t.hub
	t.mu.Unlock()

	now := time.Now()
	if wasMissing {
		logger.Info("[USBAdapterTracker] USB adapter '%s' recovered! Notifying %d sharing devices: %v",
			adapterKey, len(affectedIDs), affectedIDs)
	}

	for _, dID := range affectedIDs {
		t.mu.RLock()
		devCode := t.deviceCodes[dID]
		devName := t.deviceNames[dID]
		devPort := t.devicePorts[dID]
		t.mu.RUnlock()

		if devCode == "" {
			devCode = fmt.Sprintf("DEV-%d", dID)
		}

		if mon != nil {
			mon.UpdateDevicePollingSnapshot(dID, func(s *DevicePollingSnapshot) {
				s.TransportStatus = TransportAvailable
				s.PollingState = PollingStateWaiting
				s.RecoveryState = "RECONNECTING"
				s.LastError = ""
				// CRITICAL: CommunicationStatus REMAINS OFFLINE until valid Modbus response is verified!
			})

			mon.RecordEvent(ActivityEvent{
				Severity:    SeverityInfo,
				Category:    CategorySerial,
				Component:   "USBAdapterTracker",
				DeviceID:    dID,
				DeviceCode:  devCode,
				DeviceName:  devName,
				EventName:   "SERIAL_PORT_REOPENED",
				Description: fmt.Sprintf("Transport link re-opened for %s on %s. Awaiting verified Modbus response to confirm ONLINE.", devCode, devPort),
			})
			mon.RecordEvent(ActivityEvent{
				Severity:    SeverityInfo,
				Category:    CategoryCommunication,
				Component:   "USBAdapterTracker",
				DeviceID:    dID,
				DeviceCode:  devCode,
				DeviceName:  devName,
				EventName:   "RECONNECT_STARTED",
				Description: fmt.Sprintf("Reconnection initiated for %s after USB adapter recovery", devCode),
			})
		}

		if hub != nil {
			hub.Broadcast(websocket.WSMessage{
				Type:      "device.connection.changed",
				Timestamp: now,
				Data: map[string]interface{}{
					"device_id":         dID,
					"id":                dID,
					"device_code":       devCode,
					"connection_status": "OFFLINE",
					"transport_status":  TransportAvailable,
					"status":            "OFFLINE",
					"reason":            "TRANSPORT_AVAILABLE",
				},
			})
		}
	}
}

// OnCommunicationSuccess confirms that a verified Modbus response was received.
// This is the ONLY place where a device transitions to ONLINE!
func (t *USBAdapterTracker) OnCommunicationSuccess(deviceID uint, latencyMs int) {
	t.mu.RLock()
	devCode := t.deviceCodes[deviceID]
	devName := t.deviceNames[deviceID]
	mon := t.activityMonitor
	ds := t.deviceService
	hub := t.hub
	t.mu.RUnlock()

	now := time.Now()

	if mon != nil {
		wasOffline := false
		if snap, exists := mon.GetDevicePollingSnapshot(deviceID); exists && snap != nil {
			wasOffline = snap.CommunicationStatus != CommOnline
		}

		mon.UpdateDevicePollingSnapshot(deviceID, func(s *DevicePollingSnapshot) {
			s.CommunicationStatus = CommOnline
			s.TransportStatus = TransportAvailable
			s.PollingState = PollingStateSuccess
			s.LastSuccessfulResponse = &now
			s.LastSuccessfulTelemetry = &now
			s.LastLatencyMs = latencyMs
			s.ConsecutiveFailures = 0
			s.RecoveryState = "NORMAL"
			s.LastError = ""
			s.IsFresh = true
		})

		if wasOffline {
			mon.RecordEvent(ActivityEvent{
				Severity:    SeverityInfo,
				Category:    CategoryCommunication,
				Component:   "USBAdapterTracker",
				DeviceID:    deviceID,
				DeviceCode:  devCode,
				DeviceName:  devName,
				EventName:   "COMMUNICATION_RECOVERED",
				Description: fmt.Sprintf("Device %s communication fully restored: valid Modbus response verified (%dms)", devCode, latencyMs),
				DurationMs:  latencyMs,
			})
		}
	}

	if ds != nil {
		_ = ds.SetOnline(deviceID)
		_ = ds.UpdateLastSeen(deviceID)
		_ = ds.RecordCommunicationResult(deviceID, true, latencyMs)
	}

	if hub != nil {
		hub.Broadcast(websocket.WSMessage{
			Type:      "device.connection.changed",
			Timestamp: now,
			Data: map[string]interface{}{
				"device_id":         deviceID,
				"id":                deviceID,
				"device_code":       devCode,
				"connection_status": CommOnline,
				"transport_status":  TransportAvailable,
				"status":            CommOnline,
				"reason":            "COMMUNICATION_RESTORED",
			},
		})
	}
}

// CalculateFreshnessThreshold determines the maximum duration between successful responses
// before the device is marked stale and can no longer be reported ONLINE.
func CalculateFreshnessThreshold(conn *model.DeviceConnection) time.Duration {
	if conn == nil {
		return 5 * time.Second
	}

	interval := time.Duration(conn.PollingInterval) * time.Millisecond
	if interval < 200*time.Millisecond {
		interval = 1000 * time.Millisecond
	}

	timeout := time.Duration(conn.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 1000 * time.Millisecond
	}

	retryLimit := conn.RetryCount
	if retryLimit < 0 {
		retryLimit = 0
	}
	if retryLimit > 5 {
		retryLimit = 5
	}

	// Maximum time for one poll attempt with full retries
	attemptTime := time.Duration(retryLimit+1) * timeout
	// Freshness window: 2 poll cycles + 1 full retry attempt + 1s jitter grace
	thresh := (2 * interval) + attemptTime + 1*time.Second
	if thresh < 3*time.Second {
		thresh = 3 * time.Second
	}
	return thresh
}

// EvaluateDeviceFreshness is a package-level helper to evaluate freshness on a snapshot
func EvaluateDeviceFreshness(snap *DevicePollingSnapshot) bool {
	return GetSharedAdapterTracker().EvaluateDeviceFreshness(snap, nil)
}

// EvaluateDeviceFreshness verifies that a device's ONLINE status is backed by verified, fresh telemetry
func (t *USBAdapterTracker) EvaluateDeviceFreshness(snap *DevicePollingSnapshot, conn *model.DeviceConnection) bool {
	if snap == nil {
		return false
	}

	thresh := CalculateFreshnessThreshold(conn)
	snap.FreshnessThresholdMs = thresh.Milliseconds()

	// 1. Proactive physical port presence check for serial devices
	if conn != nil && conn.Protocol == model.ProtocolModbusRTU && conn.SerialPort != "" {
		targetPort := modbus.ResolvePortAddress(conn.SerialPort)
		if targetPort == "" {
			targetPort = conn.SerialPort
		}
		if !modbus.IsPortPhysicallyPresent(targetPort) {
			snap.TransportStatus = TransportPortMissing
			snap.CommunicationStatus = CommOffline
			snap.PollingState = PollingStateStopped
			snap.IsFresh = false
			snap.LastError = fmt.Sprintf("Serial port %s missing from OS", targetPort)
			return false
		} else if snap.TransportStatus == TransportPortMissing || snap.TransportStatus == "" || snap.TransportStatus == TransportUnknown {
			snap.TransportStatus = TransportAvailable
		}
	}

	// 2. Communication status freshness evaluation
	if snap.CommunicationStatus == CommOnline {
		if snap.LastSuccessfulResponse == nil || time.Since(*snap.LastSuccessfulResponse) > thresh {
			snap.IsFresh = false
			snap.CommunicationStatus = CommOffline
			if snap.TransportStatus == TransportPortMissing {
				snap.PollingState = PollingStateStopped
			} else {
				snap.PollingState = PollingStateTimeout
			}
			return false
		}
		snap.IsFresh = true
		return true
	}

	snap.IsFresh = false
	return false
}

// EvaluateAllDevicesFreshness iterates over all device snapshots and transitions stale ONLINE devices to OFFLINE
func (t *USBAdapterTracker) EvaluateAllDevicesFreshness() {
	mon := t.activityMonitor
	if mon == nil {
		return
	}

	snapshots := mon.GetAllPollingSnapshots()
	ds := t.deviceService
	hub := t.hub

	for _, snap := range snapshots {
		var conn *model.DeviceConnection
		if ds != nil {
			if dev, err := ds.GetDeviceByID(snap.DeviceID); err == nil && dev != nil {
				conn = dev.Connection
			}
		}

		wasOnline := snap.CommunicationStatus == CommOnline
		isFresh := t.EvaluateDeviceFreshness(&snap, conn)

		if wasOnline && !isFresh {
			dID := snap.DeviceID
			devCode := snap.DeviceCode

			mon.UpdateDevicePollingSnapshot(dID, func(s *DevicePollingSnapshot) {
				s.CommunicationStatus = CommOffline
				s.IsFresh = false
				if s.TransportStatus == TransportPortMissing {
					s.PollingState = PollingStateStopped
				} else {
					s.PollingState = PollingStateTimeout
				}
			})

			mon.RecordEvent(ActivityEvent{
				Severity:    SeverityWarning,
				Category:    CategoryCommunication,
				Component:   "USBAdapterTracker",
				DeviceID:    dID,
				DeviceCode:  devCode,
				EventName:   "COMMUNICATION_TIMEOUT",
				Description: fmt.Sprintf("Device %s communication expired freshness window (%dms). Status set to OFFLINE.", devCode, snap.FreshnessThresholdMs),
			})

			if ds != nil {
				_ = ds.SetOffline(dID)
			}

			if hub != nil {
				hub.Broadcast(websocket.WSMessage{
					Type:      "device.connection.changed",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"device_id":         dID,
						"id":                dID,
						"device_code":       devCode,
						"connection_status": CommOffline,
						"transport_status":  snap.TransportStatus,
						"status":            CommOffline,
						"reason":            "FRESHNESS_EXPIRED",
					},
				})
			}
		}
	}
}

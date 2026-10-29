package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/pkg/sysinfo"

	"github.com/goburrow/serial"
)

// SerialTransport provides the I/O interface required for serial communications
type SerialTransport interface {
	io.ReadWriteCloser
	SetDeadline(t time.Time) error
}

// defaultSerialWrapper wraps a serial.Port to implement SerialTransport
type defaultSerialWrapper struct {
	port serial.Port
}

func (w *defaultSerialWrapper) Read(b []byte) (int, error) {
	return w.port.Read(b)
}

func (w *defaultSerialWrapper) Write(b []byte) (int, error) {
	return w.port.Write(b)
}

func (w *defaultSerialWrapper) Close() error {
	return w.port.Close()
}

func (w *defaultSerialWrapper) SetDeadline(t time.Time) error {
	return nil
}

// SerialOpenerFunc is a factory function for opening serial transports
type SerialOpenerFunc func(cfg *model.DeviceConnection) (SerialTransport, error)

var portPresenceHook func(port string) bool

// SetPortPresenceHookForTest registers an override for physical serial port detection in tests
func SetPortPresenceHookForTest(hook func(port string) bool) {
	portPresenceHook = hook
}

// IsPortPhysicallyPresent checks if the serial port device node/COM port is present on the OS
func IsPortPhysicallyPresent(port string) bool {
	if port == "" {
		return false
	}
	if portPresenceHook != nil {
		return portPresenceHook(port)
	}
	if runtime.GOOS == "windows" {
		available := sysinfo.GetAvailableSerialPorts()
		for _, p := range available {
			if strings.EqualFold(p, port) {
				return true
			}
		}
		return false
	}

	// POSIX / Linux: Check symlink first, then direct stat
	if realPath, err := filepath.EvalSymlinks(port); err == nil {
		if _, statErr := os.Stat(realPath); statErr == nil {
			return true
		}
		// If realPath exists as a symlink target but doesn't stat, the symlink is broken!
		return false
	}
	if _, err := os.Stat(port); err == nil {
		return true
	}

	// Check if ResolvePortAddress finds an active unowned port
	resolved := ResolvePortAddress(port)
	if resolved != "" && resolved != port {
		if _, err := os.Stat(resolved); err == nil {
			return true
		}
	}
	return false
}

func (a *ModbusRTUAdapter) isPortMissing(port string) bool {
	if portPresenceHook != nil {
		return !IsPortPhysicallyPresent(port)
	}
	if a.isMock {
		return false
	}
	return !IsPortPhysicallyPresent(port)
}

// ResolvePortAddress evaluates symlinks and performs self-healing port migration
// when USB serial hardware glitches cause ttyUSB0 -> ttyUSB1 re-enumeration
func ResolvePortAddress(configuredPort string) string {
	if configuredPort == "" {
		return ""
	}

	// 1. If it's a Linux symlink (e.g. /dev/serial/by-id/* or /dev/serial/by-path/*),
	// evaluating the symlink automatically tracks wherever udev re-pointed the device!
	if realPath, err := filepath.EvalSymlinks(configuredPort); err == nil {
		if _, statErr := os.Stat(realPath); statErr == nil {
			return realPath
		}
	}

	// 2. If configured port already exists directly as-is, use it
	if _, err := os.Stat(configuredPort); err == nil {
		return configuredPort
	}

	// 3. Self-healing migration for Linux /dev/ttyUSB* /dev/ttyACM* and Windows COM*
	if (runtime.GOOS != "windows" && strings.HasPrefix(configuredPort, "/dev/tty")) ||
		(runtime.GOOS == "windows" && strings.HasPrefix(strings.ToUpper(configuredPort), "COM")) {
		// Scan all currently available serial ports
		availablePorts := sysinfo.GetAvailableSerialPorts()
		var usbPorts []string
		for _, p := range availablePorts {
			if strings.HasPrefix(p, "/dev/ttyUSB") || strings.HasPrefix(p, "/dev/ttyACM") ||
				(runtime.GOOS == "windows" && strings.HasPrefix(strings.ToUpper(p), "COM")) {
				usbPorts = append(usbPorts, p)
			}
		}

		// Find active unowned ports (ports not currently locked by another device's shared bus)
		var unownedPorts []string
		for _, p := range usbPorts {
			if p != configuredPort && !IsPortOwnedByBus(p) {
				unownedPorts = append(unownedPorts, p)
			}
		}

		// If exactly 1 active unowned port is found, rebind to it
		if len(unownedPorts) == 1 {
			logger.Warn("USB serial port migration detected! Configured '%s' is missing, but unowned active '%s' was found. Auto-rebinding transport to '%s'",
				configuredPort, unownedPorts[0], unownedPorts[0])
			return unownedPorts[0]
		}

		if len(usbPorts) > 1 {
			logger.Warn("Configured port '%s' is offline. Available ports: %v (unowned: %v).",
				configuredPort, usbPorts, unownedPorts)
		}
	}

	return configuredPort
}

// DefaultSerialOpener opens a physical or virtual OS serial port (COMx or /dev/ttyUSBx)
func DefaultSerialOpener(cfg *model.DeviceConnection) (SerialTransport, error) {
	parity := cfg.Parity
	if parity == "" {
		parity = "N"
	}
	baud := cfg.BaudRate
	if baud <= 0 {
		baud = 9600
	}
	dataBits := cfg.DataBits
	if dataBits <= 0 {
		dataBits = 8
	}
	stopBits := cfg.StopBits
	if stopBits <= 0 {
		stopBits = 1
	}
	timeout := time.Duration(cfg.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 1000 * time.Millisecond
	}

	// Resolve target port with self-healing migration
	targetPort := ResolvePortAddress(cfg.SerialPort)
	if targetPort == "" {
		targetPort = cfg.SerialPort
	}

	sc := &serial.Config{
		Address:  targetPort,
		BaudRate: baud,
		DataBits: dataBits,
		StopBits: stopBits,
		Parity:   parity,
		Timeout:  timeout,
	}

	port, err := serial.Open(sc)
	if err != nil {
		return nil, err
	}
	return &defaultSerialWrapper{port: port}, nil
}

// SharedSerialBus manages mutual exclusion, physical transport lifecycle,
// and RS-485 inter-frame silent intervals across all devices sharing a physical serial port.
type SharedSerialBus struct {
	mu           sync.Mutex
	sem          chan struct{}
	portPath     string
	transport    SerialTransport
	refCount     int
	lastActivity time.Time
}

// Lock acquires the mutual exclusion lock for RS-485 wire operations with context cancellation support.
func (b *SharedSerialBus) Lock(ctx context.Context) error {
	select {
	case <-b.sem:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("serial bus %s lock timed out or cancelled: %w", b.portPath, ctx.Err())
	}
}

// Unlock releases the mutual exclusion lock.
func (b *SharedSerialBus) Unlock() {
	select {
	case b.sem <- struct{}{}:
	default:
	}
}

var (
	sharedBusesMu sync.Mutex
	sharedBuses   = make(map[string]*SharedSerialBus)
)

// GetOrCreateSharedBus returns the shared bus for the given port address
func GetOrCreateSharedBus(portPath string) *SharedSerialBus {
	sharedBusesMu.Lock()
	defer sharedBusesMu.Unlock()

	bus, exists := sharedBuses[portPath]
	if !exists {
		bus = &SharedSerialBus{
			portPath: portPath,
			sem:      make(chan struct{}, 1),
		}
		bus.sem <- struct{}{} // Initialize with 1 token
		sharedBuses[portPath] = bus
	}
	return bus
}

// ResetSharedBuses closes and clears all active shared buses (used in tests and teardown)
func ResetSharedBuses() {
	sharedBusesMu.Lock()
	defer sharedBusesMu.Unlock()
	for _, b := range sharedBuses {
		b.mu.Lock()
		if b.transport != nil {
			_ = b.transport.Close()
			b.transport = nil
		}
		b.mu.Unlock()
	}
	sharedBuses = make(map[string]*SharedSerialBus)
}

// IsPortOwnedByBus checks if a physical serial port currently has an active transport open in SharedSerialBus
func IsPortOwnedByBus(portPath string) bool {
	sharedBusesMu.Lock()
	defer sharedBusesMu.Unlock()
	b, exists := sharedBuses[portPath]
	if !exists || b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.transport != nil
}

// ModbusRTUAdapter implements ProtocolAdapter for RS485/RS232 Modbus RTU devices
type ModbusRTUAdapter struct {
	config    *model.DeviceConnection
	transport SerialTransport
	opener    SerialOpenerFunc
	bus       *SharedSerialBus
	mu        sync.Mutex
	status    AdapterStatus
	isMock    bool
}

// NewModbusRTUAdapter creates an RTU adapter with the standard OS serial opener
func NewModbusRTUAdapter(cfg *model.DeviceConnection) *ModbusRTUAdapter {
	return NewModbusRTUAdapterWithOpener(cfg, DefaultSerialOpener)
}

// NewModbusRTUAdapterWithOpener creates an RTU adapter with a custom transport opener (for testing)
func NewModbusRTUAdapterWithOpener(cfg *model.DeviceConnection, opener SerialOpenerFunc) *ModbusRTUAdapter {
	if cfg.BaudRate <= 0 {
		cfg.BaudRate = 9600
	}
	if cfg.DataBits <= 0 {
		cfg.DataBits = 8
	}
	if cfg.StopBits <= 0 {
		cfg.StopBits = 1
	}
	if cfg.Parity == "" {
		cfg.Parity = "N"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 1000
	}
	if cfg.SlaveID <= 0 {
		cfg.SlaveID = 1
	}

	isMock := false
	if opener != nil && reflect.ValueOf(opener).Pointer() != reflect.ValueOf(DefaultSerialOpener).Pointer() {
		isMock = true
	}

	return &ModbusRTUAdapter{
		config: cfg,
		opener: opener,
		isMock: isMock,
		status: AdapterStatus{
			State: StateDisconnected,
		},
	}
}

// Connect opens or attaches to the shared serial bus for the configured port
func (a *ModbusRTUAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.config.SerialPort == "" {
		a.status.State = StateError
		a.status.LastError = "serial_port cannot be empty"
		return fmt.Errorf("serial_port is not configured")
	}

	targetPort := ResolvePortAddress(a.config.SerialPort)
	if targetPort == "" {
		targetPort = a.config.SerialPort
	}

	// If already attached to a different bus, detach first
	if a.bus != nil && a.bus.portPath != targetPort {
		oldBus := a.bus
		oldBus.mu.Lock()
		oldBus.refCount--
		if oldBus.refCount <= 0 {
			if oldBus.transport != nil {
				_ = oldBus.transport.Close()
				oldBus.transport = nil
			}
			sharedBusesMu.Lock()
			delete(sharedBuses, oldBus.portPath)
			sharedBusesMu.Unlock()
		}
		oldBus.mu.Unlock()
		a.bus = nil
		a.transport = nil
	}

	bus := GetOrCreateSharedBus(targetPort)
	bus.mu.Lock()
	defer bus.mu.Unlock()

	// Proactive Check: Ensure target port physically exists on system before opening (skipped in mock/test mode unless hook set)
	if a.isPortMissing(targetPort) {
		a.status.State = StateError
		a.status.LastError = fmt.Sprintf("serial port offline: '%s' not found on system", targetPort)
		a.status.FailedCount++
		if bus.transport != nil {
			_ = bus.transport.Close()
			bus.transport = nil
		}
		if a.transport != nil {
			_ = a.transport.Close()
			a.transport = nil
		}
		return fmt.Errorf("serial port offline: '%s' not found on system", targetPort)
	}

	if bus.transport == nil {
		tr, err := a.opener(a.config)
		if err != nil {
			a.status.State = StateError
			a.status.LastError = fmt.Sprintf("serial open failed: %v", err)
			a.status.FailedCount++
			return fmt.Errorf("failed to open serial port %s: %w", a.config.SerialPort, err)
		}
		bus.transport = tr
	}

	if a.bus != bus {
		bus.refCount++
		a.bus = bus
	}
	a.transport = bus.transport

	now := time.Now()
	a.status.State = StateConnected
	a.status.ConnectedSince = &now
	a.status.LastError = ""

	logger.Debug("Modbus RTU serial bus ready: %s at %d baud (attached devices: %d)", targetPort, a.config.BaudRate, bus.refCount)
	return nil
}

// Disconnect detaches from the shared serial bus gracefully
func (a *ModbusRTUAdapter) Disconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.bus != nil {
		bus := a.bus
		bus.mu.Lock()
		bus.refCount--
		if bus.refCount <= 0 {
			if bus.transport != nil {
				err = bus.transport.Close()
				bus.transport = nil
			}
			sharedBusesMu.Lock()
			delete(sharedBuses, bus.portPath)
			sharedBusesMu.Unlock()
		}
		bus.mu.Unlock()
		a.bus = nil
		a.transport = nil
	} else if a.transport != nil {
		err = a.transport.Close()
		a.transport = nil
	}

	a.status.State = StateDisconnected
	a.status.ConnectedSince = nil
	return err
}

// IsConnected returns whether the serial transport is currently open
func (a *ModbusRTUAdapter) IsConnected() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.bus != nil {
		return a.bus.transport != nil && a.status.State == StateConnected
	}
	return a.transport != nil && a.status.State == StateConnected
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "timeout") || strings.Contains(s, "deadline exceeded")
}

// ReadRegisters executes Modbus FC 01, 02, 03, or 04 over serial RTU framing
func (a *ModbusRTUAdapter) ReadRegisters(ctx context.Context, req ModbusReadRequest) (*ModbusReadResponse, error) {
	// 1. Resolve and verify target port (with self-healing symlink migration)
	targetPort := ResolvePortAddress(a.config.SerialPort)
	if targetPort == "" {
		targetPort = a.config.SerialPort
	}

	a.mu.Lock()
	bus := a.bus
	// If the resolved target port differs from currently attached bus, re-bind bus
	if bus != nil && targetPort != "" && bus.portPath != targetPort {
		logger.Warn("Modbus RTU device %d serial port migrated from %s to %s",
			a.config.DeviceID, bus.portPath, targetPort)
		bus.mu.Lock()
		bus.refCount--
		if bus.refCount <= 0 {
			if bus.transport != nil {
				_ = bus.transport.Close()
				bus.transport = nil
			}
			sharedBusesMu.Lock()
			delete(sharedBuses, bus.portPath)
			sharedBusesMu.Unlock()
		}
		bus.mu.Unlock()
		a.bus = nil
		a.transport = nil
		bus = nil
	}

	if a.bus == nil && targetPort != "" {
		newBus := GetOrCreateSharedBus(targetPort)
		newBus.mu.Lock()
		newBus.refCount++
		a.bus = newBus
		a.transport = newBus.transport
		bus = newBus
		newBus.mu.Unlock()
	}
	a.mu.Unlock()

	// 2. Acquire context-aware serial bus lock to guarantee strict serialization on RS-485
	if bus != nil {
		if err := bus.Lock(ctx); err != nil {
			return nil, &ModbusTransactionError{
				Category: FailCategoryTimeout,
				Message:  fmt.Sprintf("serial bus %s lock timed out or cancelled: %v", bus.portPath, err),
				Err:      err,
			}
		}
		defer bus.Unlock()
	} else {
		return nil, &ModbusTransactionError{
			Category: FailCategoryPortNotFound,
			Message:  "no serial bus attached",
			Err:      fmt.Errorf("serial bus is not initialized"),
		}
	}

	// Ensure Modbus RTU 3.5-character silent interval between serial frames
	if bus != nil && !bus.lastActivity.IsZero() {
		since := time.Since(bus.lastActivity)
		if since < 4*time.Millisecond {
			time.Sleep(4*time.Millisecond - since)
		}
	}
	defer func() {
		if bus != nil {
			bus.lastActivity = time.Now()
		}
	}()

	start := time.Now()

	slaveID := req.SlaveID
	if slaveID == 0 {
		slaveID = byte(a.config.SlaveID)
	}
	if slaveID == 0 {
		slaveID = 1
	}

	// 1. Build Request ADU (8 bytes)
	reqPDU := make([]byte, 6)
	reqPDU[0] = slaveID
	reqPDU[1] = req.FunctionCode
	binary.BigEndian.PutUint16(reqPDU[2:4], req.StartingAddress)
	binary.BigEndian.PutUint16(reqPDU[4:6], req.Quantity)

	reqADU := AppendCRC16(reqPDU)

	// Proactive Hardware Check: Verify port physically exists before communicating
	if a.isPortMissing(targetPort) {
		a.handleSerialError(fmt.Errorf("serial port offline: '%s' physically disconnected from system", targetPort))
		return nil, &ModbusTransactionError{
			Category:    FailCategoryPortNotFound,
			Message:     fmt.Sprintf("serial port offline: '%s' physically disconnected from system", targetPort),
			RequestADU:  reqADU,
			RequestSent: false,
			Err:         fmt.Errorf("serial port offline: '%s' physically disconnected from system", targetPort),
		}
	}

	// Ensure port is open
	if bus != nil && bus.transport != nil {
		// The shared serial bus already has an active open transport (e.g. opened by device 1).
		// Reuse it directly without attempting an OS-level reopen, which causes "Access is denied" / "busy".
		a.transport = bus.transport
		now := time.Now()
		a.status.State = StateConnected
		if a.status.ConnectedSince == nil {
			a.status.ConnectedSince = &now
		}
	} else if (bus != nil && bus.transport == nil) || a.transport == nil || a.status.State != StateConnected {
		tr, err := a.opener(a.config)
		if err != nil {
			a.status.State = StateError
			a.status.LastError = err.Error()
			a.status.FailedCount++
			return nil, &ModbusTransactionError{
				Category:    FailCategoryPortOpenFailed,
				Message:     fmt.Sprintf("serial port offline: %v", err),
				RequestADU:  reqADU,
				RequestSent: false,
				Err:         err,
			}
		}
		if bus != nil {
			bus.mu.Lock()
			bus.transport = tr
			bus.mu.Unlock()
		}
		a.transport = tr
		now := time.Now()
		a.status.State = StateConnected
		a.status.ConnectedSince = &now
	}

	// Set deadline
	timeout := time.Duration(a.config.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 1000 * time.Millisecond
	}
	_ = a.transport.SetDeadline(time.Now().Add(timeout))

	// Write Request
	if _, err := a.transport.Write(reqADU); err != nil {
		a.handleSerialError(err)
		return nil, &ModbusTransactionError{
			Category:    FailCategoryTransmitFailed,
			Message:     fmt.Sprintf("Modbus RTU serial write failed: %v", err),
			RequestADU:  reqADU,
			RequestSent: false,
			Err:         err,
		}
	}

	// 2. Read Response Header (2 bytes: SlaveID and FunctionCode)
	hdr := make([]byte, 2)
	if _, err := io.ReadFull(a.transport, hdr); err != nil {
		a.handleSerialError(err)
		cat := FailCategoryTimeout
		if !isTimeoutError(err) && isFatalSerialError(err) {
			cat = FailCategoryTransmitFailed
		}
		return nil, &ModbusTransactionError{
			Category:         cat,
			Message:          fmt.Sprintf("Modbus RTU read header failed: %v", err),
			RequestADU:       reqADU,
			RequestSent:      true,
			ResponseReceived: false,
			Err:              err,
		}
	}

	respSlaveID := hdr[0]
	respFC := hdr[1]

	// Validate Slave ID
	if respSlaveID != slaveID {
		return nil, &ModbusTransactionError{
			Category:         FailCategoryFrameCorrupted,
			Message:          fmt.Sprintf("unexpected slave ID in response: expected %d, got %d", slaveID, respSlaveID),
			RequestADU:       reqADU,
			ResponseADU:      hdr,
			RequestSent:      true,
			ResponseReceived: true,
			Err:              fmt.Errorf("unexpected slave ID in response: expected %d, got %d", slaveID, respSlaveID),
		}
	}

	// Check for Exception Response (FC | 0x80)
	if respFC == (req.FunctionCode | 0x80) {
		exBuf := make([]byte, 3) // exCode + 2 CRC bytes
		if _, err := io.ReadFull(a.transport, exBuf); err != nil {
			a.handleSerialError(err)
			return nil, &ModbusTransactionError{
				Category:         FailCategoryFrameCorrupted,
				Message:          fmt.Sprintf("failed reading exception payload: %v", err),
				RequestADU:       reqADU,
				ResponseADU:      hdr,
				RequestSent:      true,
				ResponseReceived: true,
				Err:              err,
			}
		}

		fullExFrame := append(hdr, exBuf...)
		if !ValidateCRC16(fullExFrame) {
			return nil, &ModbusTransactionError{
				Category:         FailCategoryInvalidCRC,
				Message:          "invalid CRC16 on Modbus exception response",
				RequestADU:       reqADU,
				ResponseADU:      fullExFrame,
				RequestSent:      true,
				ResponseReceived: true,
				CRCPassed:        false,
				Err:              fmt.Errorf("invalid CRC16 on Modbus exception response"),
			}
		}

		exCode := exBuf[0]
		a.status.FailedCount++
		a.status.LastError = fmt.Sprintf("Modbus Exception: %s", ModbusExceptionName(exCode))
		return nil, &ModbusTransactionError{
			Category:         FailCategoryModbusException,
			Message:          fmt.Sprintf("Modbus exception response: %s", ModbusExceptionName(exCode)),
			RequestADU:       reqADU,
			ResponseADU:      fullExFrame,
			RequestSent:      true,
			ResponseReceived: true,
			ExceptionCode:    exCode,
			ExceptionMessage: ModbusExceptionName(exCode),
			CRCPassed:        true,
			Err:              fmt.Errorf("Modbus exception response: %s", ModbusExceptionName(exCode)),
		}
	}

	// Validate Function Code
	if respFC != req.FunctionCode {
		return nil, &ModbusTransactionError{
			Category:         FailCategoryFrameCorrupted,
			Message:          fmt.Sprintf("unexpected function code: expected %02X, got %02X", req.FunctionCode, respFC),
			RequestADU:       reqADU,
			ResponseADU:      hdr,
			RequestSent:      true,
			ResponseReceived: true,
			Err:              fmt.Errorf("unexpected function code: expected %02X, got %02X", req.FunctionCode, respFC),
		}
	}

	// Read Byte Count (1 byte)
	bcBuf := make([]byte, 1)
	if _, err := io.ReadFull(a.transport, bcBuf); err != nil {
		a.handleSerialError(err)
		return nil, &ModbusTransactionError{
			Category:         FailCategoryFrameCorrupted,
			Message:          fmt.Sprintf("failed reading byte count: %v", err),
			RequestADU:       reqADU,
			ResponseADU:      hdr,
			RequestSent:      true,
			ResponseReceived: true,
			Err:              err,
		}
	}
	byteCount := bcBuf[0]

	// Read Data + CRC (byteCount + 2 bytes)
	dataWithCRC := make([]byte, int(byteCount)+2)
	if _, err := io.ReadFull(a.transport, dataWithCRC); err != nil {
		a.handleSerialError(err)
		return nil, &ModbusTransactionError{
			Category:         FailCategoryFrameCorrupted,
			Message:          fmt.Sprintf("failed reading register data: %v", err),
			RequestADU:       reqADU,
			ResponseADU:      append(hdr, byteCount),
			RequestSent:      true,
			ResponseReceived: true,
			Err:              err,
		}
	}

	// Assemble full frame for CRC verification
	fullFrame := append([]byte{respSlaveID, respFC, byteCount}, dataWithCRC...)
	expCRC := CalculateCRC16(fullFrame[:len(fullFrame)-2])
	recCRC := binary.LittleEndian.Uint16(fullFrame[len(fullFrame)-2:])

	if !ValidateCRC16(fullFrame) {
		a.status.FailedCount++
		a.status.LastError = "CRC16 validation failed on received frame"
		return nil, &ModbusTransactionError{
			Category:         FailCategoryInvalidCRC,
			Message:          fmt.Sprintf("CRC16 validation failed: expected %04X, received %04X", expCRC, recCRC),
			RequestADU:       reqADU,
			ResponseADU:      fullFrame,
			RequestSent:      true,
			ResponseReceived: true,
			CRCExpected:      expCRC,
			CRCReceived:      recCRC,
			CRCPassed:        false,
			Err:              fmt.Errorf("CRC16 validation failed: frame corrupted or noise on serial line"),
		}
	}

	data := dataWithCRC[:byteCount]
	elapsed := time.Since(start)
	now := time.Now()
	a.status.State = StateConnected
	a.status.LastCommunication = &now
	a.status.LatencyMs = int(elapsed.Milliseconds())
	a.status.SuccessCount++
	a.status.LastError = ""

	return &ModbusReadResponse{
		SlaveID:          respSlaveID,
		FunctionCode:     respFC,
		ByteCount:        byteCount,
		Data:             data,
		ResponseTime:     elapsed,
		RequestADU:       reqADU,
		ResponseADU:      fullFrame,
		CRCExpected:      expCRC,
		CRCReceived:      recCRC,
		CRCPassed:        true,
	}, nil
}

// WriteRegisters executes Modbus FC 05 (Single Coil), 06 (Single Register), 15 (Multiple Coils), or 16 (Multiple Registers)
func (a *ModbusRTUAdapter) WriteRegisters(ctx context.Context, req ModbusWriteRequest) (*ModbusWriteResponse, error) {
	targetPort := ResolvePortAddress(a.config.SerialPort)
	if targetPort == "" {
		targetPort = a.config.SerialPort
	}

	a.mu.Lock()
	bus := a.bus
	if bus != nil && targetPort != "" && bus.portPath != targetPort {
		logger.Warn("Modbus RTU device %d serial port migrated from %s to %s",
			a.config.DeviceID, bus.portPath, targetPort)
		bus.mu.Lock()
		bus.refCount--
		if bus.refCount <= 0 {
			if bus.transport != nil {
				_ = bus.transport.Close()
				bus.transport = nil
			}
			sharedBusesMu.Lock()
			delete(sharedBuses, bus.portPath)
			sharedBusesMu.Unlock()
		}
		bus.mu.Unlock()
		a.bus = nil
		a.transport = nil
		bus = nil
	}

	if a.bus == nil && targetPort != "" {
		newBus := GetOrCreateSharedBus(targetPort)
		newBus.mu.Lock()
		newBus.refCount++
		a.bus = newBus
		a.transport = newBus.transport
		bus = newBus
		newBus.mu.Unlock()
	}
	a.mu.Unlock()

	// Acquire context-aware serial bus lock to guarantee strict serialization on RS-485
	if bus != nil {
		if err := bus.Lock(ctx); err != nil {
			return nil, &ModbusTransactionError{
				Category: FailCategoryTimeout,
				Message:  fmt.Sprintf("serial bus %s lock timed out or cancelled: %v", bus.portPath, err),
				Err:      err,
			}
		}
		defer bus.Unlock()
	} else {
		return nil, &ModbusTransactionError{
			Category: FailCategoryPortNotFound,
			Message:  "no serial bus attached",
			Err:      fmt.Errorf("serial bus is not initialized"),
		}
	}

	// Ensure Modbus RTU 3.5-character silent interval between serial frames
	if bus != nil && !bus.lastActivity.IsZero() {
		since := time.Since(bus.lastActivity)
		if since < 4*time.Millisecond {
			time.Sleep(4*time.Millisecond - since)
		}
	}
	defer func() {
		if bus != nil {
			bus.lastActivity = time.Now()
		}
	}()

	start := time.Now()

	slaveID := req.SlaveID
	if slaveID == 0 {
		slaveID = byte(a.config.SlaveID)
	}
	if slaveID == 0 {
		slaveID = 1
	}

	// 1. Build Request PDU
	var reqPDU []byte
	switch req.FunctionCode {
	case FunctionWriteSingleCoil, FunctionWriteSingleRegister:
		// FC 05 & FC 06: [SlaveID, FC, AddrHi, AddrLo, ValHi, ValLo] (6 bytes before CRC)
		reqPDU = make([]byte, 6)
		reqPDU[0] = slaveID
		reqPDU[1] = req.FunctionCode
		binary.BigEndian.PutUint16(reqPDU[2:4], req.Address)
		binary.BigEndian.PutUint16(reqPDU[4:6], req.Value)

	case FunctionWriteMultipleCoils: // FC 15 (0x0F)
		qty := uint16(len(req.Values))
		if qty == 0 {
			qty = 1
			if req.Value != 0 {
				req.Values = []uint16{1}
			} else {
				req.Values = []uint16{0}
			}
		}
		byteCount := byte((qty + 7) / 8)
		reqPDU = make([]byte, 7+int(byteCount))
		reqPDU[0] = slaveID
		reqPDU[1] = FunctionWriteMultipleCoils
		binary.BigEndian.PutUint16(reqPDU[2:4], req.Address)
		binary.BigEndian.PutUint16(reqPDU[4:6], qty)
		reqPDU[6] = byteCount
		for i, val := range req.Values {
			if val != 0 {
				reqPDU[7+(i/8)] |= (1 << (i % 8))
			}
		}

	case FunctionWriteMultipleRegisters: // FC 16 (0x10)
		qty := uint16(len(req.Values))
		if qty == 0 {
			qty = 1
			req.Values = []uint16{req.Value}
		}
		byteCount := byte(qty * 2)
		reqPDU = make([]byte, 7+int(byteCount))
		reqPDU[0] = slaveID
		reqPDU[1] = FunctionWriteMultipleRegisters
		binary.BigEndian.PutUint16(reqPDU[2:4], req.Address)
		binary.BigEndian.PutUint16(reqPDU[4:6], qty)
		reqPDU[6] = byteCount
		for i, v := range req.Values {
			binary.BigEndian.PutUint16(reqPDU[7+i*2:9+i*2], v)
		}

	default:
		return nil, fmt.Errorf("unsupported Modbus write function code: 0x%02X", req.FunctionCode)
	}

	reqADU := AppendCRC16(reqPDU)

	// 2. Ensure transport is open
	a.mu.Lock()
	if bus != nil && bus.transport != nil {
		a.transport = bus.transport
		a.status.State = StateConnected
		a.status.LastError = ""
	} else if (bus != nil && bus.transport == nil) || a.transport == nil || a.status.State != StateConnected {
		tr, err := a.opener(a.config)
		if err != nil {
			a.status.State = StateError
			a.status.LastError = err.Error()
			a.status.FailedCount++
			a.mu.Unlock()
			return nil, &ModbusTransactionError{
				Category: FailCategoryPortNotFound,
				Message:  fmt.Sprintf("failed to reopen serial port %s: %v", a.config.SerialPort, err),
				Err:      err,
			}
		}
		a.transport = tr
		if bus != nil {
			bus.mu.Lock()
			bus.transport = tr
			bus.mu.Unlock()
		}
		a.status.State = StateConnected
		a.status.LastError = ""
	}
	a.mu.Unlock()

	// Drain any residual bytes
	drainBuf := make([]byte, 256)
	_, _ = a.transport.Read(drainBuf)

	// Write Request ADU
	if _, err := a.transport.Write(reqADU); err != nil {
		a.handleSerialError(err)
		return nil, &ModbusTransactionError{
			Category:    FailCategoryTransmitFailed,
			Message:     fmt.Sprintf("Modbus RTU serial write failed: %v", err),
			RequestADU:  reqADU,
			RequestSent: false,
			Err:         err,
		}
	}

	// Read Response Header (2 bytes: SlaveID and FunctionCode)
	hdr := make([]byte, 2)
	if _, err := io.ReadFull(a.transport, hdr); err != nil {
		a.handleSerialError(err)
		cat := FailCategoryTimeout
		if !isTimeoutError(err) && isFatalSerialError(err) {
			cat = FailCategoryTransmitFailed
		}
		return nil, &ModbusTransactionError{
			Category:         cat,
			Message:          fmt.Sprintf("Modbus RTU read write-response header failed: %v", err),
			RequestADU:       reqADU,
			RequestSent:      true,
			ResponseReceived: false,
			Err:              err,
		}
	}

	respSlaveID := hdr[0]
	respFC := hdr[1]

	if respSlaveID != slaveID {
		return nil, &ModbusTransactionError{
			Category:         FailCategoryFrameCorrupted,
			Message:          fmt.Sprintf("unexpected slave ID in write response: expected %d, got %d", slaveID, respSlaveID),
			RequestADU:       reqADU,
			ResponseADU:      hdr,
			RequestSent:      true,
			ResponseReceived: true,
			Err:              fmt.Errorf("unexpected slave ID in write response: expected %d, got %d", slaveID, respSlaveID),
		}
	}

	// Check for Exception Response (FC | 0x80)
	if respFC == (req.FunctionCode | 0x80) {
		exBuf := make([]byte, 3) // exCode + 2 CRC bytes
		if _, err := io.ReadFull(a.transport, exBuf); err != nil {
			a.handleSerialError(err)
			return nil, &ModbusTransactionError{
				Category:         FailCategoryFrameCorrupted,
				Message:          fmt.Sprintf("failed reading exception payload: %v", err),
				RequestADU:       reqADU,
				ResponseADU:      hdr,
				RequestSent:      true,
				ResponseReceived: true,
				Err:              err,
			}
		}

		fullExFrame := append(hdr, exBuf...)
		if !ValidateCRC16(fullExFrame) {
			return nil, &ModbusTransactionError{
				Category:         FailCategoryInvalidCRC,
				Message:          "invalid CRC16 on Modbus exception response",
				RequestADU:       reqADU,
				ResponseADU:      fullExFrame,
				RequestSent:      true,
				ResponseReceived: true,
				CRCPassed:        false,
				Err:              fmt.Errorf("invalid CRC16 on Modbus exception response"),
			}
		}

		exCode := exBuf[0]
		a.status.FailedCount++
		a.status.LastError = fmt.Sprintf("Modbus Exception: %s", ModbusExceptionName(exCode))
		return nil, &ModbusTransactionError{
			Category:         FailCategoryModbusException,
			Message:          fmt.Sprintf("Modbus exception response: %s", ModbusExceptionName(exCode)),
			RequestADU:       reqADU,
			ResponseADU:      fullExFrame,
			RequestSent:      true,
			ResponseReceived: true,
			CRCPassed:        true,
			ExceptionCode:    exCode,
			ExceptionMessage: ModbusExceptionName(exCode),
			Err:              fmt.Errorf("Modbus exception response: %s", ModbusExceptionName(exCode)),
		}
	}

	if respFC != req.FunctionCode {
		return nil, &ModbusTransactionError{
			Category:         FailCategoryFrameCorrupted,
			Message:          fmt.Sprintf("unexpected function code in write response: expected 0x%02X, got 0x%02X", req.FunctionCode, respFC),
			RequestADU:       reqADU,
			ResponseADU:      hdr,
			RequestSent:      true,
			ResponseReceived: true,
			Err:              fmt.Errorf("unexpected function code in write response: expected 0x%02X, got 0x%02X", req.FunctionCode, respFC),
		}
	}

	// For FC05, FC06, FC15, FC16 standard response is 6 more bytes:
	// FC05/FC06: Address (2) + Value (2) + CRC (2)
	// FC15/FC16: Address (2) + Quantity (2) + CRC (2)
	remBuf := make([]byte, 6)
	if _, err := io.ReadFull(a.transport, remBuf); err != nil {
		a.handleSerialError(err)
		return nil, &ModbusTransactionError{
			Category:         FailCategoryFrameCorrupted,
			Message:          fmt.Sprintf("failed reading write response body: %v", err),
			RequestADU:       reqADU,
			ResponseADU:      hdr,
			RequestSent:      true,
			ResponseReceived: true,
			Err:              err,
		}
	}

	fullFrame := append(hdr, remBuf...)
	if !ValidateCRC16(fullFrame) {
		return nil, &ModbusTransactionError{
			Category:         FailCategoryInvalidCRC,
			Message:          "invalid CRC16 on Modbus write response",
			RequestADU:       reqADU,
			ResponseADU:      fullFrame,
			RequestSent:      true,
			ResponseReceived: true,
			CRCPassed:        false,
			Err:              fmt.Errorf("invalid CRC16 on Modbus write response"),
		}
	}

	respAddress := binary.BigEndian.Uint16(remBuf[0:2])
	respValue := binary.BigEndian.Uint16(remBuf[2:4])

	elapsed := time.Since(start)
	now := time.Now()
	a.status.State = StateConnected
	a.status.LastCommunication = &now
	a.status.LatencyMs = int(elapsed.Milliseconds())
	a.status.SuccessCount++
	a.status.LastError = ""

	return &ModbusWriteResponse{
		SlaveID:      respSlaveID,
		FunctionCode: respFC,
		Address:      respAddress,
		Value:        respValue,
		ResponseTime: elapsed,
		RequestADU:   reqADU,
		ResponseADU:  fullFrame,
		CRCPassed:    true,
	}, nil
}

func isFatalSerialError(err error) bool {
	if err == nil {
		return false
	}
	if isTimeoutError(err) {
		return false
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "crc16") || strings.Contains(s, "unexpected slave id") ||
		strings.Contains(s, "unexpected function code") || strings.Contains(s, "modbus exception") {
		return false
	}
	return strings.Contains(s, "bad file descriptor") ||
		strings.Contains(s, "file already closed") ||
		strings.Contains(s, "no such file or directory") ||
		strings.Contains(s, "no such device") ||
		strings.Contains(s, "device not configured") ||
		strings.Contains(s, "input/output error") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "pipe closed") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "eof") ||
		strings.Contains(s, "not functioning") ||
		strings.Contains(s, "invalid handle") ||
		strings.Contains(s, "device does not recognize") ||
		strings.Contains(s, "not connected") ||
		strings.Contains(s, "access is denied") ||
		strings.Contains(s, "permission denied") ||
		strings.Contains(s, "inappropriate ioctl") ||
		strings.Contains(s, "interrupted system call") ||
		strings.Contains(s, "bad state")
}

func (a *ModbusRTUAdapter) handleSerialError(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.status.State = StateError
	a.status.LastError = err.Error()
	a.status.FailedCount++

	targetPort := ResolvePortAddress(a.config.SerialPort)
	if targetPort == "" {
		targetPort = a.config.SerialPort
	}
	portMissing := a.isPortMissing(targetPort)

	// Release physical transport handle if hardware error OR port physically missing from OS
	if portMissing || isFatalSerialError(err) {
		logger.Warn("Modbus RTU serial connection dropped on %s (portMissing=%v, err=%v). Releasing physical transport handle.",
			a.config.SerialPort, portMissing, err)
		bus := a.bus
		if bus != nil {
			bus.mu.Lock()
			if bus.transport != nil {
				_ = bus.transport.Close()
				bus.transport = nil
			}
			bus.mu.Unlock()
		}
		if a.transport != nil {
			_ = a.transport.Close()
			a.transport = nil
		}

		// Dynamic Self-Healing Port Migration Check:
		if runtime.GOOS != "windows" && a.config != nil && (strings.HasPrefix(a.config.SerialPort, "/dev/tty") || strings.HasPrefix(a.config.SerialPort, "/dev/serial")) {
			resolved := ResolvePortAddress(a.config.SerialPort)
			if resolved != "" && resolved != a.config.SerialPort {
				logger.Warn("Dynamic port migration applied to active adapter: %s ➔ %s", a.config.SerialPort, resolved)
				a.config.SerialPort = resolved
			}
		}
	}
}

// HealthCheck executes a lightweight read
func (a *ModbusRTUAdapter) HealthCheck(ctx context.Context) error {
	_, err := a.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         byte(a.config.SlaveID),
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	return err
}

// GetStatus returns the current adapter status and metrics
func (a *ModbusRTUAdapter) GetStatus() AdapterStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

// GetConfig returns the connection configuration
func (a *ModbusRTUAdapter) GetConfig() *model.DeviceConnection {
	return a.config
}

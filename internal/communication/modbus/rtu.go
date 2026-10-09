package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	// 3. Self-healing migration for Linux /dev/ttyUSB* /dev/ttyACM*
	// If configured as /dev/ttyUSB0 and it does not exist:
	if runtime.GOOS != "windows" && strings.HasPrefix(configuredPort, "/dev/tty") {
		// Scan all currently available serial ports
		availablePorts := sysinfo.GetAvailableSerialPorts()
		var usbPorts []string
		for _, p := range availablePorts {
			if strings.HasPrefix(p, "/dev/ttyUSB") || strings.HasPrefix(p, "/dev/ttyACM") {
				usbPorts = append(usbPorts, p)
			}
		}

		// If exactly 1 active USB serial adapter is found on the system (e.g. ttyUSB1):
		// This is the classic ttyUSB0 -> ttyUSB1 re-enumeration jump!
		if len(usbPorts) == 1 && usbPorts[0] != configuredPort {
			logger.Warn("USB serial port migration detected! Configured '%s' is missing, but active '%s' was found. Auto-rebinding transport to '%s'",
				configuredPort, usbPorts[0], usbPorts[0])
			return usbPorts[0]
		}

		// If multiple USB ports exist, try to check if any /dev/serial/by-id or by-path matches
		if len(usbPorts) > 1 {
			logger.Warn("Configured port '%s' is offline. Multiple USB serial ports detected (%v). Please consider using '/dev/serial/by-id/' or '/dev/serial/by-path/' to lock device identity.",
				configuredPort, usbPorts)
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
	portPath     string
	transport    SerialTransport
	refCount     int
	lastActivity time.Time
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
		}
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

	return &ModbusRTUAdapter{
		config: cfg,
		opener: opener,
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

	bus := GetOrCreateSharedBus(targetPort)
	bus.mu.Lock()
	defer bus.mu.Unlock()

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

	bus.refCount++
	a.bus = bus
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
		a.bus.mu.Lock()
		a.bus.refCount--
		if a.bus.refCount <= 0 {
			if a.bus.transport != nil {
				err = a.bus.transport.Close()
				a.bus.transport = nil
			}
			sharedBusesMu.Lock()
			delete(sharedBuses, a.bus.portPath)
			sharedBusesMu.Unlock()
		}
		a.bus.mu.Unlock()
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
	// 1. Acquire serial bus lock to guarantee strict serialization on RS-485
	a.mu.Lock()
	bus := a.bus
	a.mu.Unlock()

	var busLock sync.Locker = &a.mu
	if bus != nil {
		busLock = &bus.mu
	}

	busLock.Lock()
	defer busLock.Unlock()

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

	// Ensure port is open
	if (bus != nil && bus.transport == nil) || a.transport == nil || a.status.State != StateConnected {
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
			bus.transport = tr
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
		strings.Contains(s, "device not configured") ||
		strings.Contains(s, "input/output error") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "connection reset")
}

func (a *ModbusRTUAdapter) handleSerialError(err error) {
	a.status.State = StateError
	a.status.LastError = err.Error()
	a.status.FailedCount++

	// Only close physical transport on true hardware/OS I/O failures (not slave timeouts)
	if isFatalSerialError(err) {
		if a.bus != nil {
			a.bus.mu.Lock()
			if a.bus.transport != nil {
				_ = a.bus.transport.Close()
				a.bus.transport = nil
			}
			a.bus.mu.Unlock()
		} else if a.transport != nil {
			_ = a.transport.Close()
			a.transport = nil
		}
		a.transport = nil

		// Dynamic Self-Healing Port Migration Check:
		if runtime.GOOS != "windows" && a.config != nil && strings.HasPrefix(a.config.SerialPort, "/dev/tty") {
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

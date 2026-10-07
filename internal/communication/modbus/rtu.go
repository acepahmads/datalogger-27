package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"

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

	sc := &serial.Config{
		Address:  cfg.SerialPort,
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

// ModbusRTUAdapter implements ProtocolAdapter for RS485/RS232 Modbus RTU devices
type ModbusRTUAdapter struct {
	config    *model.DeviceConnection
	transport SerialTransport
	opener    SerialOpenerFunc
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

// Connect opens the serial port
func (a *ModbusRTUAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.transport != nil {
		_ = a.transport.Close()
		a.transport = nil
	}

	a.status.State = StateConnecting

	if a.config.SerialPort == "" {
		a.status.State = StateError
		a.status.LastError = "serial_port cannot be empty"
		return fmt.Errorf("serial_port is not configured")
	}

	tr, err := a.opener(a.config)
	if err != nil {
		a.status.State = StateError
		a.status.LastError = fmt.Sprintf("serial open failed: %v", err)
		a.status.FailedCount++
		return fmt.Errorf("failed to open serial port %s: %w", a.config.SerialPort, err)
	}

	now := time.Now()
	a.transport = tr
	a.status.State = StateConnected
	a.status.ConnectedSince = &now
	a.status.LastError = ""

	logger.Debug("Modbus RTU port opened: %s at %d baud", a.config.SerialPort, a.config.BaudRate)
	return nil
}

// Disconnect closes the serial port
func (a *ModbusRTUAdapter) Disconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.transport != nil {
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
	return a.transport != nil && a.status.State == StateConnected
}

// ReadRegisters executes Modbus FC 01, 02, 03, or 04 over serial RTU framing
func (a *ModbusRTUAdapter) ReadRegisters(ctx context.Context, req ModbusReadRequest) (*ModbusReadResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	start := time.Now()

	// Ensure port is open
	if a.transport == nil || a.status.State != StateConnected {
		tr, err := a.opener(a.config)
		if err != nil {
			a.status.State = StateError
			a.status.LastError = err.Error()
			a.status.FailedCount++
			return nil, fmt.Errorf("serial port offline: %w", err)
		}
		a.transport = tr
		now := time.Now()
		a.status.State = StateConnected
		a.status.ConnectedSince = &now
	}

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

	// Set deadline
	timeout := time.Duration(a.config.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 1000 * time.Millisecond
	}
	_ = a.transport.SetDeadline(time.Now().Add(timeout))

	// Write Request
	if _, err := a.transport.Write(reqADU); err != nil {
		a.handleSerialError(err)
		return nil, fmt.Errorf("Modbus RTU serial write failed: %w", err)
	}

	// 2. Read Response Header (2 bytes: SlaveID and FunctionCode)
	hdr := make([]byte, 2)
	if _, err := io.ReadFull(a.transport, hdr); err != nil {
		a.handleSerialError(err)
		return nil, fmt.Errorf("Modbus RTU read header failed: %w", err)
	}

	respSlaveID := hdr[0]
	respFC := hdr[1]

	// Validate Slave ID
	if respSlaveID != slaveID {
		return nil, fmt.Errorf("unexpected slave ID in response: expected %d, got %d", slaveID, respSlaveID)
	}

	// Check for Exception Response (FC | 0x80)
	if respFC == (req.FunctionCode | 0x80) {
		exBuf := make([]byte, 3) // exCode + 2 CRC bytes
		if _, err := io.ReadFull(a.transport, exBuf); err != nil {
			a.handleSerialError(err)
			return nil, fmt.Errorf("failed reading exception payload: %w", err)
		}

		fullExFrame := append(hdr, exBuf...)
		if !ValidateCRC16(fullExFrame) {
			return nil, fmt.Errorf("invalid CRC16 on Modbus exception response")
		}

		exCode := exBuf[0]
		a.status.FailedCount++
		a.status.LastError = fmt.Sprintf("Modbus Exception: %s", ModbusExceptionName(exCode))
		return nil, fmt.Errorf("Modbus exception response: %s", ModbusExceptionName(exCode))
	}

	// Validate Function Code
	if respFC != req.FunctionCode {
		return nil, fmt.Errorf("unexpected function code: expected %02X, got %02X", req.FunctionCode, respFC)
	}

	// Read Byte Count (1 byte)
	bcBuf := make([]byte, 1)
	if _, err := io.ReadFull(a.transport, bcBuf); err != nil {
		a.handleSerialError(err)
		return nil, fmt.Errorf("failed reading byte count: %w", err)
	}
	byteCount := bcBuf[0]

	// Read Data + CRC (byteCount + 2 bytes)
	dataWithCRC := make([]byte, int(byteCount)+2)
	if _, err := io.ReadFull(a.transport, dataWithCRC); err != nil {
		a.handleSerialError(err)
		return nil, fmt.Errorf("failed reading register data: %w", err)
	}

	// Assemble full frame for CRC verification
	fullFrame := append([]byte{respSlaveID, respFC, byteCount}, dataWithCRC...)
	if !ValidateCRC16(fullFrame) {
		a.status.FailedCount++
		a.status.LastError = "CRC16 validation failed on received frame"
		return nil, fmt.Errorf("CRC16 validation failed: frame corrupted or noise on serial line")
	}

	data := dataWithCRC[:byteCount]
	elapsed := time.Since(start)
	now := time.Now()
	a.status.LastCommunication = &now
	a.status.LatencyMs = int(elapsed.Milliseconds())
	a.status.SuccessCount++
	a.status.LastError = ""

	return &ModbusReadResponse{
		SlaveID:      respSlaveID,
		FunctionCode: respFC,
		ByteCount:    byteCount,
		Data:         data,
		ResponseTime: elapsed,
	}, nil
}

func (a *ModbusRTUAdapter) handleSerialError(err error) {
	a.status.State = StateError
	a.status.LastError = err.Error()
	a.status.FailedCount++
	if a.transport != nil {
		_ = a.transport.Close()
		a.transport = nil
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

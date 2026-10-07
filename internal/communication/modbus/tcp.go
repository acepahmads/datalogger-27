package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
)

// ModbusTCPAdapter implements ProtocolAdapter for Ethernet/WiFi Modbus TCP devices
type ModbusTCPAdapter struct {
	config *model.DeviceConnection
	conn   net.Conn
	mu     sync.Mutex
	txID   uint16
	status AdapterStatus
}

// NewModbusTCPAdapter constructs a new Modbus TCP master adapter
func NewModbusTCPAdapter(cfg *model.DeviceConnection) *ModbusTCPAdapter {
	if cfg.Port <= 0 {
		cfg.Port = 502
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 1000
	}
	if cfg.SlaveID <= 0 {
		cfg.SlaveID = 1
	}

	return &ModbusTCPAdapter{
		config: cfg,
		status: AdapterStatus{
			State: StateDisconnected,
		},
	}
}

// Connect establishes the TCP connection with the configured remote host and port
func (a *ModbusTCPAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn != nil {
		_ = a.conn.Close()
		a.conn = nil
	}

	target := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)
	a.status.State = StateConnecting

	timeout := time.Duration(a.config.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 1000 * time.Millisecond
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		a.status.State = StateError
		a.status.LastError = fmt.Sprintf("TCP dial failed: %v", err)
		a.status.FailedCount++
		return fmt.Errorf("failed to dial Modbus TCP at %s: %w", target, err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
		_ = tcpConn.SetNoDelay(true)
	}

	now := time.Now()
	a.conn = conn
	a.status.State = StateConnected
	a.status.ConnectedSince = &now
	a.status.LastError = ""

	logger.Debug("Modbus TCP connected to %s", target)
	return nil
}

// Disconnect closes the active TCP connection
func (a *ModbusTCPAdapter) Disconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.conn != nil {
		err = a.conn.Close()
		a.conn = nil
	}
	a.status.State = StateDisconnected
	a.status.ConnectedSince = nil
	return err
}

// IsConnected returns whether the socket is active
func (a *ModbusTCPAdapter) IsConnected() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.conn != nil && a.status.State == StateConnected
}

// ReadRegisters executes Modbus FC 01, 02, 03, or 04 over Modbus TCP
func (a *ModbusTCPAdapter) ReadRegisters(ctx context.Context, req ModbusReadRequest) (*ModbusReadResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	start := time.Now()

	// Ensure connected
	if a.conn == nil || a.status.State != StateConnected {
		target := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)
		timeout := time.Duration(a.config.Timeout) * time.Millisecond
		var d net.Dialer
		d.Timeout = timeout
		conn, err := d.DialContext(ctx, "tcp", target)
		if err != nil {
			a.status.State = StateError
			a.status.LastError = err.Error()
			a.status.FailedCount++
			return nil, fmt.Errorf("connection offline: %w", err)
		}
		a.conn = conn
		now := time.Now()
		a.status.State = StateConnected
		a.status.ConnectedSince = &now
	}

	timeout := time.Duration(a.config.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 1000 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)
	_ = a.conn.SetDeadline(deadline)

	// Increment transaction ID
	a.txID++
	txID := a.txID

	slaveID := req.SlaveID
	if slaveID == 0 {
		slaveID = byte(a.config.SlaveID)
	}
	if slaveID == 0 {
		slaveID = 1
	}

	// Build MBAP + PDU (12 bytes)
	reqBuf := make([]byte, 12)
	binary.BigEndian.PutUint16(reqBuf[0:2], txID)
	binary.BigEndian.PutUint16(reqBuf[2:4], 0x0000)
	binary.BigEndian.PutUint16(reqBuf[4:6], 0x0006)
	reqBuf[6] = slaveID
	reqBuf[7] = req.FunctionCode
	binary.BigEndian.PutUint16(reqBuf[8:10], req.StartingAddress)
	binary.BigEndian.PutUint16(reqBuf[10:12], req.Quantity)

	// Send Request
	if _, err := a.conn.Write(reqBuf); err != nil {
		a.handleConnError(err)
		return nil, fmt.Errorf("Modbus TCP write failed: %w", err)
	}

	// Read MBAP Header (7 bytes)
	mbapHdr := make([]byte, 7)
	if _, err := io.ReadFull(a.conn, mbapHdr); err != nil {
		a.handleConnError(err)
		return nil, fmt.Errorf("Modbus TCP read MBAP header failed: %w", err)
	}

	respTxID := binary.BigEndian.Uint16(mbapHdr[0:2])
	protoID := binary.BigEndian.Uint16(mbapHdr[2:4])
	remainingLen := binary.BigEndian.Uint16(mbapHdr[4:6])
	respUnitID := mbapHdr[6]

	if respTxID != txID {
		return nil, fmt.Errorf("transaction ID mismatch: expected %d, got %d", txID, respTxID)
	}
	if protoID != 0 {
		return nil, fmt.Errorf("invalid protocol identifier: expected 0, got %d", protoID)
	}
	if remainingLen < 2 || remainingLen > 260 {
		return nil, fmt.Errorf("invalid Modbus TCP frame length: %d", remainingLen)
	}

	// Read PDU payload: remainingLen - 1 (unitID was already read)
	pduLen := int(remainingLen - 1)
	pdu := make([]byte, pduLen)
	if _, err := io.ReadFull(a.conn, pdu); err != nil {
		a.handleConnError(err)
		return nil, fmt.Errorf("Modbus TCP read PDU payload failed: %w", err)
	}

	fc := pdu[0]
	// Check for Modbus Exception (FC with MSB set)
	if fc == (req.FunctionCode | 0x80) {
		exCode := byte(0)
		if len(pdu) > 1 {
			exCode = pdu[1]
		}
		a.status.FailedCount++
		a.status.LastError = fmt.Sprintf("Modbus Exception: %s", ModbusExceptionName(exCode))
		return nil, fmt.Errorf("Modbus exception response: %s", ModbusExceptionName(exCode))
	}

	if fc != req.FunctionCode {
		return nil, fmt.Errorf("unexpected function code in response: expected %02X, got %02X", req.FunctionCode, fc)
	}

	if len(pdu) < 2 {
		return nil, fmt.Errorf("malformed Modbus TCP response: PDU length %d too short", len(pdu))
	}

	byteCount := pdu[1]
	data := pdu[2:]
	if int(byteCount) != len(data) {
		return nil, fmt.Errorf("byte count mismatch: header states %d, received %d bytes", byteCount, len(data))
	}

	elapsed := time.Since(start)
	now := time.Now()
	a.status.LastCommunication = &now
	a.status.LatencyMs = int(elapsed.Milliseconds())
	a.status.SuccessCount++
	a.status.LastError = ""

	return &ModbusReadResponse{
		SlaveID:      respUnitID,
		FunctionCode: fc,
		ByteCount:    byteCount,
		Data:         data,
		ResponseTime: elapsed,
	}, nil
}

func (a *ModbusTCPAdapter) handleConnError(err error) {
	a.status.State = StateError
	a.status.LastError = err.Error()
	a.status.FailedCount++
	if a.conn != nil {
		_ = a.conn.Close()
		a.conn = nil
	}
}

// HealthCheck executes a lightweight read to verify active transport
func (a *ModbusTCPAdapter) HealthCheck(ctx context.Context) error {
	_, err := a.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         byte(a.config.SlaveID),
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	return err
}

// GetStatus returns a snapshot of adapter health and metrics
func (a *ModbusTCPAdapter) GetStatus() AdapterStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

// GetConfig returns the connection configuration
func (a *ModbusTCPAdapter) GetConfig() *model.DeviceConnection {
	return a.config
}

package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// SimulatorMode dictates simulated behavior for testing edge cases
type SimulatorMode string

const (
	SimModeNormal      SimulatorMode = "NORMAL"
	SimModeException   SimulatorMode = "EXCEPTION"
	SimModeTimeout     SimulatorMode = "TIMEOUT"
	SimModeCRCError     SimulatorMode = "CRC_ERROR"
	SimModeMalformed   SimulatorMode = "MALFORMED"
	SimModeDrop        SimulatorMode = "DROP"
)

// MockModbusServer provides a deterministic, local in-memory/socket Modbus server for automated testing
type MockModbusServer struct {
	mu               sync.RWMutex
	holdingRegisters map[uint16]uint16
	inputRegisters   map[uint16]uint16
	coils            map[uint16]bool
	discreteInputs   map[uint16]bool
	mode             SimulatorMode
	exceptionCode    byte
	delay            time.Duration
	listener         net.Listener
	address          string
	stopChan         chan struct{}
	conns            map[net.Conn]struct{}
}

// NewMockModbusServer initializes an in-memory register bank
func NewMockModbusServer() *MockModbusServer {
	s := &MockModbusServer{
		holdingRegisters: make(map[uint16]uint16),
		inputRegisters:   make(map[uint16]uint16),
		coils:            make(map[uint16]bool),
		discreteInputs:   make(map[uint16]bool),
		mode:             SimModeNormal,
		stopChan:         make(chan struct{}),
		conns:            make(map[net.Conn]struct{}),
	}
	return s
}

// SetHoldingRegister sets a 16-bit word in simulated holding registers
func (s *MockModbusServer) SetHoldingRegister(addr uint16, val uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.holdingRegisters[addr] = val
}

// SetInputRegister sets a 16-bit word in simulated input registers
func (s *MockModbusServer) SetInputRegister(addr uint16, val uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inputRegisters[addr] = val
}

// SetCoil sets a boolean bit in simulated coils
func (s *MockModbusServer) SetCoil(addr uint16, val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.coils[addr] = val
}

// SetDiscreteInput sets a boolean bit in simulated discrete inputs
func (s *MockModbusServer) SetDiscreteInput(addr uint16, val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.discreteInputs[addr] = val
}

// SetMode configures failure simulation behavior
func (s *MockModbusServer) SetMode(mode SimulatorMode, exCode byte, delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
	s.exceptionCode = exCode
	s.delay = delay
}

// StartTCP starts a local TCP server on an OS-assigned ephemeral port
func (s *MockModbusServer) StartTCP() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	s.listener = l
	s.address = l.Addr().String()

	go s.serveTCP()
	return s.address, nil
}

// StartTCPAt starts a local TCP server on a specific host:port address
func (s *MockModbusServer) StartTCPAt(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = l
	s.address = l.Addr().String()

	go s.serveTCP()
	return nil
}

// Stop closes the server listener and active sessions
func (s *MockModbusServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}
	for c := range s.conns {
		_ = c.Close()
		delete(s.conns, c)
	}
	select {
	case <-s.stopChan:
	default:
		close(s.stopChan)
	}
}

func (s *MockModbusServer) serveTCP() {
	s.mu.RLock()
	l := s.listener
	s.mu.RUnlock()
	if l == nil {
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		if s.conns != nil {
			s.conns[conn] = struct{}{}
		}
		s.mu.Unlock()
		go s.handleTCPConn(conn)
	}
}

func (s *MockModbusServer) handleTCPConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		s.mu.Lock()
		delete(s.conns, conn)
		s.mu.Unlock()
	}()

	for {
		hdr := make([]byte, 6)
		if _, err := io.ReadFull(conn, hdr); err != nil {
			return
		}

		txID := binary.BigEndian.Uint16(hdr[0:2])
		proto := binary.BigEndian.Uint16(hdr[2:4])
		length := binary.BigEndian.Uint16(hdr[4:6])

		pdu := make([]byte, length)
		if _, err := io.ReadFull(conn, pdu); err != nil {
			return
		}

		s.mu.RLock()
		mode := s.mode
		exCode := s.exceptionCode
		delay := s.delay
		s.mu.RUnlock()

		if mode == SimModeDrop {
			return
		}

		if mode == SimModeTimeout && delay > 0 {
			time.Sleep(delay)
		}

		unitID := pdu[0]
		fc := pdu[1]
		startAddr := binary.BigEndian.Uint16(pdu[2:4])
		qty := binary.BigEndian.Uint16(pdu[4:6])

		var respPDU []byte

		if mode == SimModeException {
			// Exception response: FC | 0x80, exceptionCode
			respPDU = []byte{unitID, fc | 0x80, exCode}
		} else if mode == SimModeMalformed {
			// Truncated payload
			respPDU = []byte{unitID, fc, 0xFF}
		} else {
			respPDU = s.generatePDU(unitID, fc, startAddr, qty)
		}

		// Assemble MBAP response
		respMBAP := make([]byte, 6)
		binary.BigEndian.PutUint16(respMBAP[0:2], txID)
		binary.BigEndian.PutUint16(respMBAP[2:4], proto)
		binary.BigEndian.PutUint16(respMBAP[4:6], uint16(len(respPDU)))

		fullPacket := append(respMBAP, respPDU...)
		if _, err := conn.Write(fullPacket); err != nil {
			return
		}
	}
}

func (s *MockModbusServer) generatePDU(unitID, fc byte, startAddr, qty uint16) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch fc {
	case 0x01: // Read Coils
		byteCount := byte((qty + 7) / 8)
		data := make([]byte, byteCount)
		for i := uint16(0); i < qty; i++ {
			if s.coils[startAddr+i] {
				data[i/8] |= (1 << (i % 8))
			}
		}
		res := []byte{unitID, fc, byteCount}
		return append(res, data...)

	case 0x02: // Read Discrete Inputs
		byteCount := byte((qty + 7) / 8)
		data := make([]byte, byteCount)
		for i := uint16(0); i < qty; i++ {
			if s.discreteInputs[startAddr+i] {
				data[i/8] |= (1 << (i % 8))
			}
		}
		res := []byte{unitID, fc, byteCount}
		return append(res, data...)

	case 0x03: // Read Holding Registers
		byteCount := byte(qty * 2)
		data := make([]byte, byteCount)
		for i := uint16(0); i < qty; i++ {
			val := s.holdingRegisters[startAddr+i]
			binary.BigEndian.PutUint16(data[i*2:i*2+2], val)
		}
		res := []byte{unitID, fc, byteCount}
		return append(res, data...)

	case 0x04: // Read Input Registers
		byteCount := byte(qty * 2)
		data := make([]byte, byteCount)
		for i := uint16(0); i < qty; i++ {
			val := s.inputRegisters[startAddr+i]
			binary.BigEndian.PutUint16(data[i*2:i*2+2], val)
		}
		res := []byte{unitID, fc, byteCount}
		return append(res, data...)

	default:
		// Illegal function exception
		return []byte{unitID, fc | 0x80, 0x01}
	}
}

// MockSerialPipe simulates a full-duplex serial connection using in-memory buffers
type MockSerialPipe struct {
	server       *MockModbusServer
	readBuf      []byte
	writeBuf     []byte
	mu           sync.Mutex
	isClosed     bool
	readDeadline time.Time
}

// NewMockSerialPipe connects an RTU adapter directly to a MockModbusServer in memory
func NewMockSerialPipe(server *MockModbusServer) *MockSerialPipe {
	return &MockSerialPipe{
		server: server,
	}
}

func (p *MockSerialPipe) Read(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return 0, io.EOF
	}

	if len(p.readBuf) == 0 {
		return 0, io.EOF
	}

	n := copy(b, p.readBuf)
	p.readBuf = p.readBuf[n:]
	return n, nil
}

func (p *MockSerialPipe) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return 0, fmt.Errorf("pipe closed")
	}

	p.server.mu.RLock()
	mode := p.server.mode
	exCode := p.server.exceptionCode
	delay := p.server.delay
	p.server.mu.RUnlock()

	if mode == SimModeDrop {
		return len(b), nil
	}

	if mode == SimModeTimeout && delay > 0 {
		time.Sleep(delay)
		return len(b), nil
	}

	if len(b) < 8 {
		return len(b), nil
	}

	slaveID := b[0]
	fc := b[1]
	startAddr := binary.BigEndian.Uint16(b[2:4])
	qty := binary.BigEndian.Uint16(b[4:6])

	var respPDU []byte
	if mode == SimModeException {
		respPDU = []byte{slaveID, fc | 0x80, exCode}
	} else if mode == SimModeMalformed {
		respPDU = []byte{slaveID, fc, 0x05, 0xAA} // Byte count says 5, only 1 byte given
	} else {
		respPDU = p.server.generatePDU(slaveID, fc, startAddr, qty)
	}

	frame := AppendCRC16(respPDU)
	if mode == SimModeCRCError {
		// Invert CRC bytes
		frame[len(frame)-1] ^= 0xFF
	}

	p.readBuf = append(p.readBuf, frame...)
	return len(b), nil
}

func (p *MockSerialPipe) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isClosed = true
	return nil
}

func (p *MockSerialPipe) SetDeadline(t time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.readDeadline = t
	return nil
}

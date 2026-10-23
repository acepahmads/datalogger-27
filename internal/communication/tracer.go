package communication

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/websocket"
)

// DecodedValueItem represents a decoded, scaled, and formula-calculated parameter value from a captured frame
type DecodedValueItem struct {
	ParameterID   uint     `json:"parameter_id,omitempty"`
	ParameterCode string   `json:"parameter_code"`
	ParameterName string   `json:"parameter_name,omitempty"`
	RawValue      float64  `json:"raw_value"`
	ScaledValue   float64  `json:"scaled_value"`
	FormulaValue  *float64 `json:"formula_value,omitempty"`
	Value         float64  `json:"value"`
	Formatted     string   `json:"formatted"`
	Unit          string   `json:"unit,omitempty"`
}

// CommTraceFrame represents a captured live communication transaction frame
type CommTraceFrame struct {
	ID              int64              `json:"id"`
	Timestamp       time.Time          `json:"timestamp"`
	DeviceID        uint               `json:"device_id"`
	DeviceCode      string             `json:"device_code"`
	DeviceName      string             `json:"device_name"`
	Protocol        string             `json:"protocol"`  // MODBUS_RTU, MODBUS_TCP
	Transport       string             `json:"transport"` // /dev/ttyUSB0 (9600 8-N-1) or 192.168.1.50:502
	SlaveID         byte               `json:"slave_id"`
	FunctionCode    byte               `json:"function_code"`
	FCName          string             `json:"fc_name"`
	StartingAddress uint16             `json:"starting_address"`
	DisplayAddress  int                `json:"display_address"` // Modicon notation (e.g. 40021)
	Quantity        uint16             `json:"quantity"`
	ParameterCodes  string             `json:"parameter_codes,omitempty"`  // e.g. "PM25" or "SO2, NO2"
	ParameterNames  string             `json:"parameter_names,omitempty"`  // e.g. "Particulate Matter 2.5"
	ConvertedValues string             `json:"converted_values,omitempty"` // Formatted converted values string: e.g. "35.00 µg/m³" or "PM2.5: 35.00 µg/m³, PM10: 44.00 µg/m³"
	DecodedValues   []DecodedValueItem `json:"decoded_values,omitempty"`   // Structured decoded parameters with raw, scaled, and formula values
	TXHex           string             `json:"tx_hex"`
	TXBytes         int                `json:"tx_bytes"`
	RXHex           string             `json:"rx_hex"`
	RXBytes         int                `json:"rx_bytes"`
	LatencyMs       int                `json:"latency_ms"`
	Status          string             `json:"status"` // SUCCESS, TIMEOUT, CRC_ERROR, EXCEPTION, ERROR
	ErrorMessage    string             `json:"error_message,omitempty"`
	Summary         string             `json:"summary"`
}

// CommTraceStatus represents the operational state of the live communication sniffer
type CommTraceStatus struct {
	Active           bool      `json:"active"`
	ActivatedBy      string    `json:"activated_by,omitempty"`
	ActivatedAt      *time.Time `json:"activated_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	RemainingSeconds int       `json:"remaining_seconds"`
	TargetDeviceID   uint      `json:"target_device_id"`
	TotalFrames      int       `json:"total_frames"`
}

// CommTracer manages real-time serial/socket frame tapping and broadcasting with auto-timeout
type CommTracer struct {
	mu            sync.RWMutex
	active        bool
	activatedBy   string
	activatedAt   time.Time
	expiresAt     time.Time
	autoStopTimer *time.Timer
	targetDevice  uint // 0 = all devices
	buffer        []CommTraceFrame
	maxBuffer     int
	nextID        int64
	hub           *websocket.Hub
}

// NewCommTracer initializes the singleton communication tracer
func NewCommTracer(hub *websocket.Hub) *CommTracer {
	return &CommTracer{
		maxBuffer: 200,
		hub:       hub,
		buffer:    make([]CommTraceFrame, 0, 200),
	}
}

// IsActive returns whether live communication sniffing is currently enabled and within time bounds
func (t *CommTracer) IsActive() bool {
	if t == nil {
		return false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if !t.active {
		return false
	}
	return time.Now().Before(t.expiresAt)
}

// Start activates live communication sniffing with a strict auto-timeout
func (t *CommTracer) Start(activatedBy string, duration time.Duration, targetDevice uint) CommTraceStatus {
	t.mu.Lock()
	defer t.mu.Unlock()

	if duration <= 0 {
		duration = 10 * time.Minute
	}
	if duration > 30*time.Minute {
		duration = 30 * time.Minute
	}

	if t.autoStopTimer != nil {
		t.autoStopTimer.Stop()
	}

	now := time.Now()
	t.active = true
	t.activatedBy = activatedBy
	t.activatedAt = now
	t.expiresAt = now.Add(duration)
	t.targetDevice = targetDevice

	t.autoStopTimer = time.AfterFunc(duration, func() {
		t.Stop("system:auto-timeout")
	})

	logger.Info("[CommTracer] Live communication sniffing ACTIVATED by '%s' for %v (Target Device: %d)",
		activatedBy, duration, targetDevice)

	status := t.statusUnlocked()
	t.broadcastStatus(status)
	return status
}

// Stop deactivates live communication sniffing and stops any running timer
func (t *CommTracer) Stop(stoppedBy string) CommTraceStatus {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.autoStopTimer != nil {
		t.autoStopTimer.Stop()
		t.autoStopTimer = nil
	}

	wasActive := t.active
	t.active = false

	if wasActive {
		logger.Info("[CommTracer] Live communication sniffing STOPPED by '%s'", stoppedBy)
	}

	status := t.statusUnlocked()
	t.broadcastStatus(status)
	return status
}

// ClearBuffer clears all in-memory captured frames
func (t *CommTracer) ClearBuffer() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buffer = make([]CommTraceFrame, 0, t.maxBuffer)
}

// GetStatus returns the current live tracer status
func (t *CommTracer) GetStatus() CommTraceStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.statusUnlocked()
}

func (t *CommTracer) statusUnlocked() CommTraceStatus {
	isActive := t.active && time.Now().Before(t.expiresAt)
	remSec := 0
	if isActive {
		remSec = int(time.Until(t.expiresAt).Seconds())
		if remSec < 0 {
			remSec = 0
		}
	}

	var actAt *time.Time
	var expAt *time.Time
	if !t.activatedAt.IsZero() {
		a := t.activatedAt
		actAt = &a
	}
	if !t.expiresAt.IsZero() {
		e := t.expiresAt
		expAt = &e
	}

	return CommTraceStatus{
		Active:           isActive,
		ActivatedBy:      t.activatedBy,
		ActivatedAt:      actAt,
		ExpiresAt:        expAt,
		RemainingSeconds: remSec,
		TargetDeviceID:   t.targetDevice,
		TotalFrames:      len(t.buffer),
	}
}

// GetRecentFrames returns captured frames filtered by optional deviceID and limit.
// Frames are ordered newest first (most recent at index 0).
func (t *CommTracer) GetRecentFrames(limit int, deviceID uint) []CommTraceFrame {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if limit <= 0 || limit > len(t.buffer) {
		limit = len(t.buffer)
	}

	var filtered []CommTraceFrame
	for i := len(t.buffer) - 1; i >= 0 && len(filtered) < limit; i-- {
		frame := t.buffer[i]
		if deviceID == 0 || frame.DeviceID == deviceID {
			filtered = append(filtered, frame)
		}
	}

	return filtered
}

// Record captures a single request/response transaction
func (t *CommTracer) Record(
	device *model.Device,
	req modbus.ModbusReadRequest,
	resp *modbus.ModbusReadResponse,
	err error,
	latencyMs int,
) {
	if !t.IsActive() || device == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// If expired while waiting for lock, auto-deactivate
	if !t.active || time.Now().After(t.expiresAt) {
		t.active = false
		return
	}

	if t.targetDevice > 0 && device.ID != t.targetDevice {
		return
	}

	t.nextID++
	frameID := t.nextID

	transport := "N/A"
	protocol := "MODBUS"
	if device.Connection != nil {
		protocol = string(device.Connection.Protocol)
		if device.Connection.Protocol == model.ProtocolModbusRTU {
			baud := device.Connection.BaudRate
			if baud == 0 {
				baud = 9600
			}
			transport = fmt.Sprintf("%s (%d %d-%s-%d)",
				device.Connection.SerialPort, baud,
				device.Connection.DataBits, device.Connection.Parity, device.Connection.StopBits)
		} else if device.Connection.Protocol == model.ProtocolModbusTCP {
			host := device.Connection.Host
			if host == "" {
				host = device.Connection.Address
			}
			transport = fmt.Sprintf("%s:%d", host, device.Connection.Port)
		}
	}

	status := "SUCCESS"
	errMsg := ""
	var txBytes []byte
	var rxBytes []byte

	if resp != nil {
		txBytes = resp.RequestADU
		rxBytes = resp.ResponseADU
		if resp.ResponseTime > 0 {
			latencyMs = int(resp.ResponseTime.Milliseconds())
		}
	}

	if err != nil {
		errMsg = err.Error()
		status = "ERROR"

		if txErr, ok := err.(*modbus.ModbusTransactionError); ok {
			if len(txBytes) == 0 {
				txBytes = txErr.RequestADU
			}
			if len(rxBytes) == 0 {
				rxBytes = txErr.ResponseADU
			}
			switch txErr.Category {
			case modbus.FailCategoryTimeout:
				status = "TIMEOUT"
			case modbus.FailCategoryInvalidCRC:
				status = "CRC_ERROR"
			case modbus.FailCategoryModbusException:
				status = "EXCEPTION"
			default:
				status = "ERROR"
			}
		} else if strings.Contains(strings.ToLower(errMsg), "timeout") {
			status = "TIMEOUT"
		}
	}

	txHex := FormatHexBytes(txBytes)
	rxHex := FormatHexBytes(rxBytes)
	fcName := FunctionCodeName(req.FunctionCode)
	modiconAddr := ToModiconDisplay(req.FunctionCode, req.StartingAddress)

	// Match target parameter(s) if configured on the device or passed in request
	var matchedCodes []string
	var matchedNames []string
	var decodedItems []DecodedValueItem
	var convertedParts []string

	reqStart := req.StartingAddress
	reqEnd := req.StartingAddress + req.Quantity - 1
	seenParams := make(map[string]bool)

	if len(device.Parameters) > 0 {
		for _, p := range device.Parameters {
			if !p.Enabled {
				continue
			}
			res, err := modbus.ResolveRegisterAddress(p.RegisterType, p.RegisterAddress)
			if err != nil || res.FunctionCode != req.FunctionCode {
				continue
			}
			paramQty := uint16(modbus.RequiredRegisterCount(p.DataType))
			if paramQty == 0 {
				paramQty = 1
			}
			paramStart := res.PDUAddress
			paramEnd := res.PDUAddress + paramQty - 1

			// Check if register range overlaps
			if reqStart <= paramEnd && reqEnd >= paramStart {
				code := strings.TrimSpace(p.ParameterCode)
				if code == "" {
					code = strings.TrimSpace(p.Code)
				}
				name := strings.TrimSpace(p.ParameterName)
				if name == "" {
					name = strings.TrimSpace(p.Name)
				}

				paramKey := fmt.Sprintf("%d_%s", p.ID, code)
				if seenParams[paramKey] {
					continue
				}
				seenParams[paramKey] = true

				if code != "" {
					matchedCodes = append(matchedCodes, code)
					if name != "" && name != code {
						matchedNames = append(matchedNames, fmt.Sprintf("%s (%s)", code, name))
					} else {
						matchedNames = append(matchedNames, code)
					}
				}

				// If read succeeded and we have response register data, decode value according to setup
				if status == "SUCCESS" && resp != nil && len(resp.Data) > 0 {
					// Verify parameter registers are fully contained within request
					if paramStart >= reqStart && paramEnd <= reqEnd {
						var paramBytes []byte
						if req.FunctionCode == modbus.FunctionReadCoils || req.FunctionCode == modbus.FunctionReadDiscreteInputs {
							bitOffset := int(paramStart - reqStart)
							val, bitErr := modbus.DecodeCoilBit(resp.Data, bitOffset)
							if bitErr == nil {
								if val {
									paramBytes = []byte{0x01}
								} else {
									paramBytes = []byte{0x00}
								}
							}
						} else {
							regOffset := int(paramStart - reqStart)
							byteOffset := regOffset * 2
							byteLen := int(paramQty) * 2
							if byteOffset >= 0 && byteOffset+byteLen <= len(resp.Data) {
								paramBytes = resp.Data[byteOffset : byteOffset+byteLen]
							}
						}

						if len(paramBytes) > 0 {
							byteOrder := p.ByteOrder
							if byteOrder == "" && device.Connection != nil {
								byteOrder = device.Connection.ByteOrder
							}
							if byteOrder == "" {
								byteOrder = "ABCD"
							}

							decoded, decErr := modbus.DecodeRegisters(paramBytes, p.DataType, byteOrder)
							if decErr == nil {
								engVal := modbus.ApplyScaleAndOffset(decoded.RawValue, p.Scale, p.Offset, p.Precision)
								finalVal := engVal
								var formulaVal *float64
								if p.Formula != "" {
									formulaVal = evalParamFormula(p.Formula, engVal, decoded.RawValue)
									if formulaVal != nil {
										finalVal = *formulaVal
									}
								}

								prec := p.Precision
								if prec < 0 {
									prec = 2
								} else if prec > 6 {
									prec = 6
								}
								formattedNum := fmt.Sprintf("%.*f", prec, finalVal)
								unitStr := strings.TrimSpace(p.Unit)

								displayStr := formattedNum
								if unitStr != "" {
									displayStr = fmt.Sprintf("%s %s", formattedNum, unitStr)
								}

								decodedItems = append(decodedItems, DecodedValueItem{
									ParameterID:   p.ID,
									ParameterCode: code,
									ParameterName: name,
									RawValue:      decoded.RawValue,
									ScaledValue:   engVal,
									FormulaValue:  formulaVal,
									Value:         finalVal,
									Formatted:     formattedNum,
									Unit:          unitStr,
								})

								convertedParts = append(convertedParts, fmt.Sprintf("%s: %s", code, displayStr))
							}
						}
					}
				}
			}
		}
	}

	if len(matchedCodes) == 0 && req.ParameterCode != "" {
		matchedCodes = append(matchedCodes, req.ParameterCode)
		if req.ParameterName != "" {
			matchedNames = append(matchedNames, req.ParameterName)
		} else {
			matchedNames = append(matchedNames, req.ParameterCode)
		}
	}

	convertedValuesStr := ""
	if len(decodedItems) == 1 {
		item := decodedItems[0]
		if item.Unit != "" {
			convertedValuesStr = fmt.Sprintf("%s %s", item.Formatted, item.Unit)
		} else {
			convertedValuesStr = item.Formatted
		}
	} else if len(convertedParts) > 1 {
		convertedValuesStr = strings.Join(convertedParts, ", ")
	}

	paramCodesStr := strings.Join(matchedCodes, ", ")
	paramNamesStr := strings.Join(matchedNames, ", ")

	summary := fmt.Sprintf("%s at #%d (PDU 0x%04X), Qty %d",
		fcName, modiconAddr, req.StartingAddress, req.Quantity)
	if paramCodesStr != "" {
		summary = fmt.Sprintf("[%s] %s", paramCodesStr, summary)
	}

	frame := CommTraceFrame{
		ID:              frameID,
		Timestamp:       time.Now(),
		DeviceID:        device.ID,
		DeviceCode:      device.DeviceCode,
		DeviceName:      device.DeviceName,
		Protocol:        protocol,
		Transport:       transport,
		SlaveID:         req.SlaveID,
		FunctionCode:    req.FunctionCode,
		FCName:          fcName,
		StartingAddress: req.StartingAddress,
		DisplayAddress:  modiconAddr,
		Quantity:        req.Quantity,
		ParameterCodes:  paramCodesStr,
		ParameterNames:  paramNamesStr,
		ConvertedValues: convertedValuesStr,
		DecodedValues:   decodedItems,
		TXHex:           txHex,
		TXBytes:         len(txBytes),
		RXHex:           rxHex,
		RXBytes:         len(rxBytes),
		LatencyMs:       latencyMs,
		Status:          status,
		ErrorMessage:    errMsg,
		Summary:         summary,
	}

	// Append to ring buffer
	if len(t.buffer) >= t.maxBuffer {
		t.buffer = t.buffer[1:]
	}
	t.buffer = append(t.buffer, frame)

	// Broadcast via WebSocket
	if t.hub != nil {
		t.hub.Broadcast(websocket.WSMessage{
			Type:      "comm.live.frame",
			Timestamp: frame.Timestamp,
			Data:      frame,
		})
	}
}

func (t *CommTracer) broadcastStatus(status CommTraceStatus) {
	if t.hub != nil {
		t.hub.Broadcast(websocket.WSMessage{
			Type:      "comm.trace.status",
			Timestamp: time.Now(),
			Data:      status,
		})
	}
}

// FormatHexBytes formats a byte slice into an uppercase space-delimited hex string
func FormatHexBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	parts := make([]string, len(b))
	for i, v := range b {
		parts[i] = fmt.Sprintf("%02X", v)
	}
	return strings.Join(parts, " ")
}

// FunctionCodeName returns the human-readable standard Modbus function name
func FunctionCodeName(fc byte) string {
	switch fc {
	case 1:
		return "FC01 (Read Coils)"
	case 2:
		return "FC02 (Read Discrete Inputs)"
	case 3:
		return "FC03 (Read Holding Registers)"
	case 4:
		return "FC04 (Read Input Registers)"
	case 5:
		return "FC05 (Write Single Coil)"
	case 6:
		return "FC06 (Write Single Register)"
	case 15:
		return "FC15 (Write Multiple Coils)"
	case 16:
		return "FC16 (Write Multiple Registers)"
	default:
		return fmt.Sprintf("FC%02d", fc)
	}
}

// ToModiconDisplay converts 0-indexed PDU address to 1-indexed Modicon convention
func ToModiconDisplay(fc byte, pdu uint16) int {
	switch fc {
	case 1:
		return int(pdu) + 1
	case 2:
		return int(pdu) + 10001
	case 4:
		return int(pdu) + 30001
	case 3, 6, 16:
		return int(pdu) + 40001
	default:
		return int(pdu)
	}
}

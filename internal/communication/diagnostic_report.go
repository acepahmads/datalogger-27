package communication

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

// DiagnosticParameterRead performs an isolated, engineering-grade diagnostic read for a specific parameter.
// It traces the full path across all 4 stages: Port Verification, Serial Configuration,
// Modbus Transaction, and Data Interpretation without corrupting operational telemetry.
func (m *ConnectionManager) DiagnosticParameterRead(
	ctx context.Context,
	device *model.Device,
	param *model.Parameter,
) (*EngineeringDiagnosticReport, error) {
	now := time.Now().UTC()

	report := &EngineeringDiagnosticReport{
		DiagnosticID:         fmt.Sprintf("DIAG-%s-%d-%d", now.Format("20060102150405"), device.ID, param.ID),
		Timestamp:            now,
		DeviceID:             device.ID,
		DeviceCode:           device.DeviceCode,
		DeviceName:           device.DeviceName,
		ParameterID:          param.ID,
		ParameterCode:        param.ParameterCode,
		ParameterName:        param.ParameterName,
		OverallStatus:        "FAIL",
		Success:              false,
		FailureCategory:      FailCategoryNone,
		EvidenceCompleteness: "PARTIAL",
		Stages:               make([]DiagnosticStageEvidence, 0, 4),
	}

	if device == nil || device.Connection == nil {
		report.FailureCategory = FailCategoryRegisterConfig
		report.ErrorMessage = "device or connection configuration is missing"
		report.FirstFailedStage = "PORT_VERIFICATION"
		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "PORT_VERIFICATION",
			Status:  StageFail,
			Message: report.ErrorMessage,
		})
		report.Stages = append(report.Stages,
			DiagnosticStageEvidence{Stage: "SERIAL_CONFIGURATION", Status: StageNotTested, Message: "Pre-requisite stage failed"},
			DiagnosticStageEvidence{Stage: "MODBUS_TRANSACTION", Status: StageNotTested, Message: "Pre-requisite stage failed"},
			DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "Pre-requisite stage failed"},
		)
		report.SuggestedAction = "Configure valid connection settings for this device in Device Settings."
		report.SuggestedActionID = "CONFIG_MISSING"
		return report, nil
	}

	cfg := device.Connection
	report.Protocol = string(cfg.Protocol)
	if cfg.Protocol == model.ProtocolModbusRTU || cfg.Protocol == model.ProtocolSerial {
		report.Transport = "SERIAL_RTU"
	} else {
		report.Transport = "TCP"
	}

	// Prevent concurrent diagnostic tests on the same device
	m.diagMu.Lock()
	if m.diagActive == nil {
		m.diagActive = make(map[uint]bool)
	}
	if m.diagActive[device.ID] {
		m.diagMu.Unlock()
		report.FailureCategory = FailCategoryPortBusy
		report.ErrorMessage = fmt.Sprintf("device %s is currently executing another diagnostic test", device.DeviceCode)
		report.FirstFailedStage = "PORT_VERIFICATION"
		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "PORT_VERIFICATION",
			Status:  StageFail,
			Message: report.ErrorMessage,
		})
		report.Stages = append(report.Stages,
			DiagnosticStageEvidence{Stage: "SERIAL_CONFIGURATION", Status: StageNotTested, Message: "Device busy with concurrent test"},
			DiagnosticStageEvidence{Stage: "MODBUS_TRANSACTION", Status: StageNotTested, Message: "Device busy with concurrent test"},
			DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "Device busy with concurrent test"},
		)
		report.SuggestedAction = "Wait a few seconds for the active test to complete before running another diagnostic."
		report.SuggestedActionID = "TEST_BUSY"
		return report, nil
	}
	m.diagActive[device.ID] = true
	m.diagMu.Unlock()

	defer func() {
		m.diagMu.Lock()
		delete(m.diagActive, device.ID)
		m.diagMu.Unlock()
	}()

	// -------------------------------------------------------------
	// STAGE 1: PORT_VERIFICATION
	// -------------------------------------------------------------
	if report.Transport == "SERIAL_RTU" {
		portEv := InspectSerialPort(cfg.SerialPort)
		report.PortEvidence = portEv

		if !portEv.Exists {
			report.FailureCategory = FailCategoryPortNotFound
			report.ErrorMessage = fmt.Sprintf("serial port offline: '%s' not found on system", cfg.SerialPort)
			report.FirstFailedStage = "PORT_VERIFICATION"
			report.Stages = append(report.Stages, DiagnosticStageEvidence{
				Stage:   "PORT_VERIFICATION",
				Status:  StageFail,
				Message: report.ErrorMessage,
				Details: map[string]interface{}{
					"configured_port": cfg.SerialPort,
					"available_ports": portEv.SystemAvailablePorts,
				},
			})
			report.Stages = append(report.Stages,
				DiagnosticStageEvidence{Stage: "SERIAL_CONFIGURATION", Status: StageNotTested, Message: "Port does not exist. No serial parameters tested."},
				DiagnosticStageEvidence{Stage: "MODBUS_TRANSACTION", Status: StageNotTested, Message: "Port offline. No Modbus request was transmitted."},
				DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "No data received to decode."},
			)

			if len(portEv.SystemAvailablePorts) > 0 {
				bestPort := portEv.SystemAvailablePorts[0]
				report.SuggestedAction = fmt.Sprintf("Configured serial port '%s' does not exist. The system detected available serial port(s): %s. If your USB-RS485 adapter re-enumerated, update the device port to '%s' or use persistent '/dev/serial/by-id/'.",
					cfg.SerialPort, strings.Join(portEv.SystemAvailablePorts, ", "), bestPort)
				report.SuggestedActionID = "PORT_NOT_FOUND_PORTS_AVAILABLE"
				report.MbpollCommand = FormatMbpollCommand(cfg, param, 0x03, 0, 1, byte(cfg.SlaveID), bestPort)
			} else {
				report.SuggestedAction = fmt.Sprintf("Configured serial port '%s' does not exist and no USB serial ports were found. Check physical USB cabling, adapter power, and run 'ls -l /dev/serial*' on the Raspberry Pi.",
					cfg.SerialPort)
				report.SuggestedActionID = "PORT_NOT_FOUND_NO_PORTS"
				report.MbpollCommand = FormatMbpollCommand(cfg, param, 0x03, 0, 1, byte(cfg.SlaveID), cfg.SerialPort)
			}
			return report, nil
		}

		if portEv.IsSymlink && !portEv.SymlinkTargetExists {
			report.FailureCategory = FailCategoryBrokenSymlink
			report.ErrorMessage = fmt.Sprintf("broken serial symlink: '%s' points to non-existent target '%s'", cfg.SerialPort, portEv.SymlinkTarget)
			report.FirstFailedStage = "PORT_VERIFICATION"
			report.Stages = append(report.Stages, DiagnosticStageEvidence{
				Stage:   "PORT_VERIFICATION",
				Status:  StageFail,
				Message: report.ErrorMessage,
				Details: map[string]interface{}{
					"symlink": cfg.SerialPort,
					"target":  portEv.SymlinkTarget,
				},
			})
			report.Stages = append(report.Stages,
				DiagnosticStageEvidence{Stage: "SERIAL_CONFIGURATION", Status: StageNotTested, Message: "Symlink is broken."},
				DiagnosticStageEvidence{Stage: "MODBUS_TRANSACTION", Status: StageNotTested, Message: "Port target missing. No Modbus request transmitted."},
				DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "No data received."},
			)
			report.SuggestedAction = fmt.Sprintf("Symlink '%s' points to '%s', which has disconnected or disappeared. Verify USB connection or udev rules in /etc/udev/rules.d/.", cfg.SerialPort, portEv.SymlinkTarget)
			report.SuggestedActionID = "BROKEN_SYMLINK"
			report.MbpollCommand = FormatMbpollCommand(cfg, param, 0x03, 0, 1, byte(cfg.SlaveID), cfg.SerialPort)
			return report, nil
		}

		if !portEv.CanAccess {
			if portEv.IsBusy {
				report.FailureCategory = FailCategoryPortBusy
				report.ErrorMessage = fmt.Sprintf("serial port busy: %s", portEv.BusyReason)
				report.SuggestedAction = fmt.Sprintf("Serial port '%s' is locked by another external process (e.g. mbpoll or minicom). Terminate external tools before testing.", portEv.ResolvedPort)
				report.SuggestedActionID = "PORT_BUSY"
			} else {
				report.FailureCategory = FailCategoryPermissionDenied
				report.ErrorMessage = fmt.Sprintf("serial port permission denied: %s", portEv.PermissionError)
				report.SuggestedAction = fmt.Sprintf("Permission denied accessing '%s'. Ensure the datalogger process user is in the 'dialout' group: 'sudo usermod -a -G dialout $USER' and restart.", portEv.ResolvedPort)
				report.SuggestedActionID = "PORT_PERMISSION_DENIED"
			}
			report.FirstFailedStage = "PORT_VERIFICATION"
			report.Stages = append(report.Stages, DiagnosticStageEvidence{
				Stage:   "PORT_VERIFICATION",
				Status:  StageFail,
				Message: report.ErrorMessage,
			})
			report.Stages = append(report.Stages,
				DiagnosticStageEvidence{Stage: "SERIAL_CONFIGURATION", Status: StageNotTested, Message: "Cannot access port."},
				DiagnosticStageEvidence{Stage: "MODBUS_TRANSACTION", Status: StageNotTested, Message: "No Modbus request transmitted."},
				DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "No data received."},
			)
			report.MbpollCommand = FormatMbpollCommand(cfg, param, 0x03, 0, 1, byte(cfg.SlaveID), portEv.ResolvedPort)
			return report, nil
		}

		// Stage 1 PASS
		report.LastSuccessfulStage = "PORT_VERIFICATION"
		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "PORT_VERIFICATION",
			Status:  StagePass,
			Message: fmt.Sprintf("Serial device '%s' verified and accessible", portEv.ResolvedPort),
			Details: map[string]interface{}{
				"configured": portEv.ConfiguredPort,
				"resolved":   portEv.ResolvedPort,
				"symlink":    portEv.IsSymlink,
				"owned":      portEv.AdapterOwned,
			},
		})
	} else {
		// Modbus TCP
		if cfg.Host == "" || cfg.Port <= 0 {
			report.FailureCategory = FailCategoryRegisterConfig
			report.ErrorMessage = "host or port invalid for Modbus TCP"
			report.FirstFailedStage = "PORT_VERIFICATION"
			report.Stages = append(report.Stages, DiagnosticStageEvidence{
				Stage:   "PORT_VERIFICATION",
				Status:  StageFail,
				Message: report.ErrorMessage,
			})
			report.Stages = append(report.Stages,
				DiagnosticStageEvidence{Stage: "SERIAL_CONFIGURATION", Status: StageNotTested, Message: "TCP configuration invalid."},
				DiagnosticStageEvidence{Stage: "MODBUS_TRANSACTION", Status: StageNotTested, Message: "No request transmitted."},
				DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "No data received."},
			)
			report.SuggestedAction = "Configure valid IP host and port in device settings."
			report.SuggestedActionID = "TCP_CONFIG_INVALID"
			return report, nil
		}
		report.LastSuccessfulStage = "PORT_VERIFICATION"
		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "PORT_VERIFICATION",
			Status:  StagePass,
			Message: fmt.Sprintf("TCP endpoint configured (%s:%d)", cfg.Host, cfg.Port),
		})
	}

	// -------------------------------------------------------------
	// STAGE 2: SERIAL_CONFIGURATION & REGISTER RESOLUTION
	// -------------------------------------------------------------
	report.SerialConfig = SerialConfigEvidence{
		BaudRate:  cfg.BaudRate,
		DataBits:  cfg.DataBits,
		Parity:    cfg.Parity,
		StopBits:  cfg.StopBits,
		TimeoutMs: cfg.Timeout,
	}

	res, err := modbus.ResolveRegisterAddress(param.RegisterType, param.RegisterAddress)
	if err != nil {
		report.FailureCategory = FailCategoryRegisterConfig
		report.ErrorMessage = fmt.Sprintf("invalid register configuration: %v", err)
		report.FirstFailedStage = "SERIAL_CONFIGURATION"
		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "SERIAL_CONFIGURATION",
			Status:  StageFail,
			Message: report.ErrorMessage,
		})
		report.Stages = append(report.Stages,
			DiagnosticStageEvidence{Stage: "MODBUS_TRANSACTION", Status: StageNotTested, Message: "Invalid register mapping. No transaction executed."},
			DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "No data received."},
		)
		report.SuggestedAction = "Check parameter register address and type (e.g. Holding vs Input)."
		report.SuggestedActionID = "REGISTER_CONFIG_INVALID"
		return report, nil
	}

	qty := modbus.RequiredRegisterCount(param.DataType)
	if res.FunctionCode == FunctionReadCoils || res.FunctionCode == FunctionReadDiscreteInputs {
		qty = 1
	}

	report.LastSuccessfulStage = "SERIAL_CONFIGURATION"
	report.Stages = append(report.Stages, DiagnosticStageEvidence{
		Stage:   "SERIAL_CONFIGURATION",
		Status:  StagePass,
		Message: fmt.Sprintf("Register #%d mapped to FC %02X (PDU offset %d, %d words). Framing %d-%s-%d @ %d baud",
			param.RegisterAddress, res.FunctionCode, res.PDUAddress, qty, cfg.DataBits, cfg.Parity, cfg.StopBits, cfg.BaudRate),
		Details: map[string]interface{}{
			"function_code":   res.FunctionCode,
			"pdu_address":     res.PDUAddress,
			"register_count":  qty,
			"data_type":       param.DataType,
		},
	})

	// -------------------------------------------------------------
	// STAGE 3: MODBUS_TRANSACTION
	// -------------------------------------------------------------
	slaveID := byte(cfg.SlaveID)
	if slaveID == 0 {
		slaveID = 1
	}

	targetPort := cfg.SerialPort
	if report.PortEvidence.ResolvedPort != "" {
		targetPort = report.PortEvidence.ResolvedPort
	}
	report.MbpollCommand = FormatMbpollCommand(cfg, param, res.FunctionCode, res.PDUAddress, qty, slaveID, targetPort)

	// Pre-build request frame for diagnostic audit
	reqPDU := make([]byte, 6)
	reqPDU[0] = slaveID
	reqPDU[1] = res.FunctionCode
	binary.BigEndian.PutUint16(reqPDU[2:4], res.PDUAddress)
	binary.BigEndian.PutUint16(reqPDU[4:6], qty)
	reqADU := modbus.AppendCRC16(reqPDU)

	report.TransactionEvidence = ModbusTransactionEvidence{
		SlaveID:              slaveID,
		FunctionCode:         res.FunctionCode,
		RegisterType:         param.RegisterType,
		ConfiguredAddress:    uint16(param.RegisterAddress),
		PDUAddress:           res.PDUAddress,
		Quantity:             qty,
		AddressingConvention: "0-BASED_PDU",
		RequestFrameHex:      formatHexBytes(reqADU),
		RequestFrameBytes:    formatByteSlice(reqADU),
		CrcValidation:        "NOT_CHECKED",
	}

	// Reuse existing adapter ownership via ConnectionManager
	adapter, err := m.GetOrCreateAdapter(device)
	if err != nil {
		report.FailureCategory = FailCategoryPortOpenFailed
		report.ErrorMessage = fmt.Sprintf("failed initializing adapter for port %s: %v", targetPort, err)
		report.FirstFailedStage = "MODBUS_TRANSACTION"
		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "MODBUS_TRANSACTION",
			Status:  StageFail,
			Message: report.ErrorMessage,
		})
		report.Stages = append(report.Stages,
			DiagnosticStageEvidence{Stage: "DATA_INTERPRETATION", Status: StageNotTested, Message: "No data received."},
		)
		report.SuggestedAction = fmt.Sprintf("Failed to initialize communication adapter for '%s'. Check driver and baud settings.", targetPort)
		report.SuggestedActionID = "ADAPTER_INIT_FAILED"
		return report, nil
	}

	timeoutMs := cfg.Timeout
	if timeoutMs <= 0 {
		timeoutMs = 1200
	}
	testCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	start := time.Now()
	readReq := ModbusReadRequest{
		SlaveID:         slaveID,
		FunctionCode:    res.FunctionCode,
		StartingAddress: res.PDUAddress,
		Quantity:        qty,
	}

	resp, err := adapter.ReadRegisters(testCtx, readReq)
	elapsed := time.Since(start)
	report.ResponseTimeMs = elapsed.Milliseconds()

	if err != nil {
		report.FirstFailedStage = "MODBUS_TRANSACTION"
		var txErr *modbus.ModbusTransactionError
		if errors.As(err, &txErr) {
			report.TransactionEvidence.RequestTransmitted = txErr.RequestSent
			report.TransactionEvidence.ResponseReceived = txErr.ResponseReceived
			if len(txErr.ResponseADU) > 0 {
				report.TransactionEvidence.ResponseFrameHex = formatHexBytes(txErr.ResponseADU)
				report.TransactionEvidence.ResponseFrameBytes = formatByteSlice(txErr.ResponseADU)
			}
			report.FailureCategory = txErr.Category
			report.ErrorMessage = txErr.Error()

			switch txErr.Category {
			case FailCategoryPortOpenFailed:
				report.TransactionEvidence.RequestTransmitted = false
				report.Stages = append(report.Stages, DiagnosticStageEvidence{
					Stage:   "MODBUS_TRANSACTION",
					Status:  StageFail,
					Message: fmt.Sprintf("Serial port could not be opened: %v. No Modbus request was transmitted.", txErr.Err),
				})
				report.SuggestedAction = fmt.Sprintf("Port offline. Ensure '%s' is plugged in and recognized.", targetPort)
				report.SuggestedActionID = "PORT_OFFLINE"

			case FailCategoryTransmitFailed:
				report.TransactionEvidence.RequestTransmitted = false
				report.Stages = append(report.Stages, DiagnosticStageEvidence{
					Stage:   "MODBUS_TRANSACTION",
					Status:  StageFail,
					Message: fmt.Sprintf("Failed to transmit frame across serial line: %v", txErr.Err),
				})
				report.SuggestedAction = "Serial transmission failed. Check USB serial converter and cable integrity."
				report.SuggestedActionID = "TRANSMIT_FAILED"

			case FailCategoryTimeout:
				report.TransactionEvidence.RequestTransmitted = true
				report.TransactionEvidence.ResponseReceived = false
				report.Stages = append(report.Stages, DiagnosticStageEvidence{
					Stage:   "MODBUS_TRANSACTION",
					Status:  StageFail,
					Message: fmt.Sprintf("Modbus request (%d bytes: %s) transmitted, but no response received within %d ms", len(reqADU), report.TransactionEvidence.RequestFrameHex, timeoutMs),
				})
				report.SuggestedAction = fmt.Sprintf("Request was transmitted, but Slave ID %d did not respond within %d ms. Check slave address, RS-485 polarity (swap A and B), wiring, power supply, or baud rate.", slaveID, timeoutMs)
				report.SuggestedActionID = "TIMEOUT_NO_RESPONSE"

			case FailCategoryInvalidCRC:
				report.TransactionEvidence.RequestTransmitted = true
				report.TransactionEvidence.ResponseReceived = true
				report.TransactionEvidence.CrcValidation = "FAIL"
				report.TransactionEvidence.CrcExpectedHex = fmt.Sprintf("%04X", txErr.CRCExpected)
				report.TransactionEvidence.CrcReceivedHex = fmt.Sprintf("%04X", txErr.CRCReceived)
				report.Stages = append(report.Stages, DiagnosticStageEvidence{
					Stage:   "MODBUS_TRANSACTION",
					Status:  StageFail,
					Message: fmt.Sprintf("Response frame received (%s), but CRC16 check failed (Expected: %04X, Received: %04X)", report.TransactionEvidence.ResponseFrameHex, txErr.CRCExpected, txErr.CRCReceived),
				})
				report.SuggestedAction = "CRC error detected. Frame was corrupted by electrical noise, missing 120Ω termination resistor, ground loop, or parity mismatch."
				report.SuggestedActionID = "CRC_ERROR"

			case FailCategoryModbusException:
				report.TransactionEvidence.RequestTransmitted = true
				report.TransactionEvidence.ResponseReceived = true
				report.TransactionEvidence.CrcValidation = "PASS"
				report.TransactionEvidence.ModbusExceptionCode = txErr.ExceptionCode
				report.TransactionEvidence.ModbusExceptionName = txErr.ExceptionMessage
				report.Stages = append(report.Stages, DiagnosticStageEvidence{
					Stage:   "MODBUS_TRANSACTION",
					Status:  StageFail,
					Message: fmt.Sprintf("Physical link WORKING! Slave ID %d responded with Modbus Exception %s", slaveID, txErr.ExceptionMessage),
				})
				report.SuggestedAction = fmt.Sprintf("Physical link and slave communication are 100%% ACTIVE! However, the sensor rejected the request with Exception %s. Check if register address #%d exists or if Function Code %02X is supported.", txErr.ExceptionMessage, param.RegisterAddress, res.FunctionCode)
				report.SuggestedActionID = "MODBUS_EXCEPTION"

			default:
				report.Stages = append(report.Stages, DiagnosticStageEvidence{
					Stage:   "MODBUS_TRANSACTION",
					Status:  StageFail,
					Message: txErr.Error(),
				})
				report.SuggestedAction = fmt.Sprintf("Communication failed: %s. Verify device connection parameters.", txErr.Error())
				report.SuggestedActionID = "COMM_ERROR"
			}
		} else {
			// Generic fallback error
			errStr := err.Error()
			report.ErrorMessage = errStr
			if strings.Contains(strings.ToLower(errStr), "timeout") || errors.Is(err, context.DeadlineExceeded) {
				report.FailureCategory = FailCategoryTimeout
				report.TransactionEvidence.RequestTransmitted = true
				report.SuggestedAction = fmt.Sprintf("Timeout: No response from Slave ID %d within %d ms. Check slave address and RS485 A/B wiring.", slaveID, timeoutMs)
				report.SuggestedActionID = "TIMEOUT_NO_RESPONSE"
			} else if strings.Contains(errStr, "offline") || strings.Contains(errStr, "no such file") {
				report.FailureCategory = FailCategoryPortNotFound
				report.SuggestedAction = fmt.Sprintf("Serial port '%s' went offline. Check hardware connection.", targetPort)
				report.SuggestedActionID = "PORT_OFFLINE"
			} else {
				report.FailureCategory = FailCategoryTransmitFailed
				report.SuggestedAction = fmt.Sprintf("Communication error: %s", errStr)
				report.SuggestedActionID = "COMM_ERROR"
			}
			report.Stages = append(report.Stages, DiagnosticStageEvidence{
				Stage:   "MODBUS_TRANSACTION",
				Status:  StageFail,
				Message: report.ErrorMessage,
			})
		}

		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "DATA_INTERPRETATION",
			Status:  StageNotTested,
			Message: "Transaction failed. No data to decode.",
		})
		return report, nil
	}

	// Transaction Succeeded!
	report.TransactionEvidence.RequestTransmitted = true
	report.TransactionEvidence.ResponseReceived = true
	if len(resp.ResponseADU) > 0 {
		report.TransactionEvidence.ResponseFrameHex = formatHexBytes(resp.ResponseADU)
		report.TransactionEvidence.ResponseFrameBytes = formatByteSlice(resp.ResponseADU)
	}
	report.TransactionEvidence.CrcValidation = "PASS"
	report.TransactionEvidence.CrcExpectedHex = fmt.Sprintf("%04X", resp.CRCExpected)
	report.TransactionEvidence.CrcReceivedHex = fmt.Sprintf("%04X", resp.CRCReceived)

	report.LastSuccessfulStage = "MODBUS_TRANSACTION"
	report.Stages = append(report.Stages, DiagnosticStageEvidence{
		Stage:   "MODBUS_TRANSACTION",
		Status:  StagePass,
		Message: fmt.Sprintf("Transaction verified: received %d data bytes in %d ms with valid CRC16", len(resp.Data), report.ResponseTimeMs),
		Details: map[string]interface{}{
			"response_time_ms": report.ResponseTimeMs,
			"byte_count":       len(resp.Data),
		},
	})

	// -------------------------------------------------------------
	// STAGE 4: DATA_INTERPRETATION
	// -------------------------------------------------------------
	byteOrder := param.ByteOrder
	if byteOrder == "" && cfg.ByteOrder != "" {
		byteOrder = cfg.ByteOrder
	}
	if byteOrder == "" {
		byteOrder = "ABCD"
	}

	report.DataEvidence = DataInterpretationEvidence{
		RawBytesHex: formatHexBytes(resp.Data),
		DataType:    string(param.DataType),
		ByteOrder:   byteOrder,
		Scale:       param.Scale,
		Offset:      param.Offset,
		Unit:        param.Unit,
		Formula:     param.Formula,
	}

	for i := 0; i+1 < len(resp.Data); i += 2 {
		w := binary.BigEndian.Uint16(resp.Data[i : i+2])
		report.DataEvidence.RawRegisterWords = append(report.DataEvidence.RawRegisterWords, w)
		report.DataEvidence.RawRegisterHex = append(report.DataEvidence.RawRegisterHex, fmt.Sprintf("0x%04X", w))
	}

	decoded, err := modbus.DecodeRegisters(resp.Data, param.DataType, byteOrder)
	if err != nil {
		report.FailureCategory = FailCategoryDecodeError
		report.ErrorMessage = fmt.Sprintf("register decoding failed: %v", err)
		report.FirstFailedStage = "DATA_INTERPRETATION"
		report.Stages = append(report.Stages, DiagnosticStageEvidence{
			Stage:   "DATA_INTERPRETATION",
			Status:  StageFail,
			Message: report.ErrorMessage,
			Details: map[string]interface{}{
				"raw_bytes":  report.DataEvidence.RawBytesHex,
				"data_type":  param.DataType,
				"byte_order": byteOrder,
			},
		})
		report.SuggestedAction = fmt.Sprintf("Registers read successfully, but decoding as %s failed: %v. Try adjusting Byte Order (%s) or Data Type.", param.DataType, err, byteOrder)
		report.SuggestedActionID = "DECODE_ERROR"
		return report, nil
	}

	engVal := modbus.ApplyScaleAndOffset(decoded.RawValue, param.Scale, param.Offset, param.Precision)
	report.DataEvidence.RawValue = &decoded.RawValue
	report.DataEvidence.DecodedValue = &decoded.RawValue
	report.DataEvidence.ScaledValue = &engVal

	finalVal := engVal
	if param.Formula != "" {
		fVal := evalParamFormula(param.Formula, engVal, decoded.RawValue)
		if fVal != nil {
			report.DataEvidence.FormulaValue = fVal
			finalVal = *fVal
		}
	}
	report.DataEvidence.EngineeringValue = &finalVal

	// Stage 4 PASS
	report.LastSuccessfulStage = "DATA_INTERPRETATION"
	report.Stages = append(report.Stages, DiagnosticStageEvidence{
		Stage:   "DATA_INTERPRETATION",
		Status:  StagePass,
		Message: fmt.Sprintf("Engineering value: %.2f %s (Raw: %.2f, Scaled: %.2f)", finalVal, param.Unit, decoded.RawValue, engVal),
		Details: map[string]interface{}{
			"engineering_val": finalVal,
			"raw_val":         decoded.RawValue,
			"scaled_val":      engVal,
		},
	})

	// Overall Success!
	report.OverallStatus = "PASS"
	report.Success = true
	report.FailureCategory = FailCategoryNone
	report.EvidenceCompleteness = "FULL"
	report.SuggestedAction = "All diagnostic stages passed. Physical link, Modbus framing, and data interpretation are operating normally."
	report.SuggestedActionID = "ALL_PASSED"

	// Legacy backward compatibility fields
	report.RawValue = decoded.RawValue
	report.DecodedValue = finalVal
	report.ScaledValue = engVal
	report.Formula = param.Formula
	report.FormulaValue = report.DataEvidence.FormulaValue
	report.RawBytesHex = formatHexBytes(resp.Data)
	report.FunctionCode = res.FunctionCode
	report.RegisterAddress = res.PDUAddress
	report.RegisterCount = qty

	// Update cached parameter state for UI responsiveness
	if m.deviceService != nil {
		_ = m.deviceService.UpdateParameterCurrentValue(param.ID, finalVal, report.DataEvidence.FormulaValue, false)
		_ = m.deviceService.UpdateLastData(device.ID)
		_ = m.deviceService.SetOnline(device.ID)
		_ = m.deviceService.UpdateLastSeen(device.ID)
	}

	return report, nil
}

// FormatMbpollCommand generates an exact copy-paste terminal command for side-by-side verification with mbpoll
func FormatMbpollCommand(
	cfg *model.DeviceConnection,
	param *model.Parameter,
	fc byte,
	pduAddr, qty uint16,
	slaveID byte,
	targetPort string,
) string {
	if cfg == nil {
		return ""
	}
	port := targetPort
	if port == "" {
		port = cfg.SerialPort
	}
	if port == "" {
		port = "/dev/ttyUSB0"
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
	parity := strings.ToLower(cfg.Parity)
	switch parity {
	case "e", "even":
		parity = "even"
	case "o", "odd":
		parity = "odd"
	default:
		parity = "none"
	}

	// mbpoll type flag
	typeFlag := "3"
	dType := ""
	if param != nil {
		dType = strings.ToUpper(string(param.DataType))
	}

	switch fc {
	case FunctionReadHoldingRegisters:
		if strings.Contains(dType, "FLOAT") {
			typeFlag = "3:float"
		} else if strings.Contains(dType, "INT32") || strings.Contains(dType, "UINT32") {
			typeFlag = "3:int"
		} else {
			typeFlag = "3"
		}
	case FunctionReadInputRegisters:
		if strings.Contains(dType, "FLOAT") {
			typeFlag = "4:float"
		} else if strings.Contains(dType, "INT32") || strings.Contains(dType, "UINT32") {
			typeFlag = "4:int"
		} else {
			typeFlag = "4"
		}
	case FunctionReadCoils:
		typeFlag = "0"
	case FunctionReadDiscreteInputs:
		typeFlag = "1"
	}

	cCount := qty
	if (typeFlag == "3:float" || typeFlag == "4:float" || typeFlag == "3:int" || typeFlag == "4:int") && qty >= 2 {
		cCount = qty / 2
	}
	if cCount <= 0 {
		cCount = 1
	}

	if cfg.Protocol == model.ProtocolModbusTCP || cfg.Protocol == model.ProtocolTCP {
		h := cfg.Host
		if h == "" {
			h = "127.0.0.1"
		}
		p := cfg.Port
		if p <= 0 {
			p = 502
		}
		return fmt.Sprintf("mbpoll -m tcp -a %d -t %s -r %d -c %d -0 -1 -p %d %s",
			slaveID, typeFlag, pduAddr, cCount, p, h)
	}

	return fmt.Sprintf("mbpoll -m rtu -a %d -b %d -d %d -s %d -P %s -t %s -r %d -c %d -0 -1 %s",
		slaveID, baud, dataBits, stopBits, parity, typeFlag, pduAddr, cCount, port)
}

func formatHexBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	parts := make([]string, len(b))
	for i, v := range b {
		parts[i] = fmt.Sprintf("%02X", v)
	}
	return strings.Join(parts, " ")
}

func formatByteSlice(b []byte) []string {
	if len(b) == 0 {
		return []string{}
	}
	parts := make([]string, len(b))
	for i, v := range b {
		parts[i] = fmt.Sprintf("%02X", v)
	}
	return parts
}

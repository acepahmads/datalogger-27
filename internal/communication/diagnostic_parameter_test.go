package communication

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

// Helper to create a test device and parameter
func createTestDeviceAndParam(port string, slaveID byte, paramCode, regType string, regAddr int, dataType model.ParameterDataType) (*model.Device, *model.Parameter) {
	dev := &model.Device{
		ID:         20,
		DeviceCode: "AQMS-02",
		DeviceName: "Air Quality Station 02",
		Connection: &model.DeviceConnection{
			DeviceID:   20,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: port,
			BaudRate:   9600,
			DataBits:   8,
			StopBits:   1,
			Parity:     "N",
			Timeout:    500,
			SlaveID:    int(slaveID),
		},
	}
	param := &model.Parameter{
		ID:              101,
		DeviceID:        20,
		ParameterCode:   paramCode,
		ParameterName:   "Particulate Matter PM10",
		RegisterType:    regType,
		RegisterAddress: regAddr,
		DataType:        dataType,
		Unit:            "µg/m³",
		Scale:           1.0,
		Offset:          0.0,
		Precision:       2,
		Enabled:         true,
	}
	return dev, param
}

// 1. Missing serial device
func TestDiagnosticParameter_MissingSerialPort(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("/dev/non_existent_serial_aqms_9999", 2, "PM10", "HOLDING", 0, "FLOAT32")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" || report.Success {
		t.Fatalf("expected FAIL, got status=%s success=%v", report.OverallStatus, report.Success)
	}
	if report.FailureCategory != FailCategoryPortNotFound {
		t.Fatalf("expected FailCategoryPortNotFound, got: %s", report.FailureCategory)
	}
	if report.FirstFailedStage != "PORT_VERIFICATION" {
		t.Fatalf("expected first failed stage PORT_VERIFICATION, got: %s", report.FirstFailedStage)
	}
	if len(report.Stages) != 4 {
		t.Fatalf("expected 4 stages, got %d", len(report.Stages))
	}
	if report.Stages[0].Status != StageFail {
		t.Fatalf("expected stage 0 FAIL, got: %s", report.Stages[0].Status)
	}
	if report.Stages[1].Status != StageNotTested || report.Stages[2].Status != StageNotTested || report.Stages[3].Status != StageNotTested {
		t.Fatalf("expected downstream stages NOT_TESTED, got: %v, %v, %v",
			report.Stages[1].Status, report.Stages[2].Status, report.Stages[3].Status)
	}
	// Request must NOT be transmitted
	if report.TransactionEvidence.RequestTransmitted {
		t.Fatalf("expected request_transmitted=false when port is offline")
	}
	if report.SuggestedAction == "" {
		t.Fatalf("expected actionable suggested action")
	}
}

// 2. Broken serial symlink
func TestDiagnosticParameter_BrokenSymlink(t *testing.T) {
	defer modbus.ResetSharedBuses()

	// Create a broken symlink in a temporary directory
	tempDir, err := os.MkdirTemp("", "serial_symlink_test")
	if err != nil {
		t.Skip("skipping symlink test: cannot create temp dir")
	}
	defer os.RemoveAll(tempDir)

	symlinkPath := filepath.Join(tempDir, "by-id-test-port")
	targetPath := filepath.Join(tempDir, "non_existent_target_usb")

	err = os.Symlink(targetPath, symlinkPath)
	if err != nil {
		t.Skip("skipping symlink test: OS symlink creation not permitted")
	}

	server := modbus.NewMockModbusServer()
	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam(symlinkPath, 2, "PM10", "HOLDING", 0, "FLOAT32")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" {
		t.Fatalf("expected FAIL on broken symlink, got %s", report.OverallStatus)
	}
	if report.FailureCategory != FailCategoryBrokenSymlink {
		t.Fatalf("expected FailCategoryBrokenSymlink, got: %s", report.FailureCategory)
	}
	if !report.PortEvidence.IsSymlink {
		t.Fatalf("expected port to be detected as symlink")
	}
	if report.PortEvidence.SymlinkTargetExists {
		t.Fatalf("expected symlink target to NOT exist")
	}
	if report.TransactionEvidence.RequestTransmitted {
		t.Fatalf("request should not be transmitted on broken symlink")
	}
}

// 3. Port busy (concurrent diagnostic active)
func TestDiagnosticParameter_PortBusy(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("COM_TEST_BUSY", 1, "TEMP", "HOLDING", 0, "INT16")

	// Artificially simulate active diagnostic lock
	mgr.diagMu.Lock()
	if mgr.diagActive == nil {
		mgr.diagActive = make(map[uint]bool)
	}
	mgr.diagActive[dev.ID] = true
	mgr.diagMu.Unlock()

	ctx := context.Background()
	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" || report.FailureCategory != FailCategoryPortBusy {
		t.Fatalf("expected FailCategoryPortBusy, got: %s", report.FailureCategory)
	}

	// Release lock
	mgr.diagMu.Lock()
	delete(mgr.diagActive, dev.ID)
	mgr.diagMu.Unlock()
}

// 4. Timeout waiting for Modbus response
func TestDiagnosticParameter_Timeout(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetMode(modbus.SimModeDrop, 0, 0) // Server drops all frames (no response)
	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("COM_TEST_TIMEOUT", 2, "PM10", "HOLDING", 0, "FLOAT32")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" {
		t.Fatalf("expected FAIL on timeout, got %s", report.OverallStatus)
	}
	if report.FailureCategory != FailCategoryTimeout {
		t.Fatalf("expected FailCategoryTimeout, got: %s", report.FailureCategory)
	}
	if report.FirstFailedStage != "MODBUS_TRANSACTION" {
		t.Fatalf("expected first failed stage MODBUS_TRANSACTION, got: %s", report.FirstFailedStage)
	}
	// Request WAS transmitted across RS485!
	if !report.TransactionEvidence.RequestTransmitted {
		t.Fatalf("expected request_transmitted=true on timeout")
	}
	if report.TransactionEvidence.ResponseReceived {
		t.Fatalf("expected response_received=false on timeout")
	}
	if report.Stages[0].Status != StagePass {
		t.Fatalf("expected stage 0 PASS (port verified), got: %s", report.Stages[0].Status)
	}
	if report.Stages[1].Status != StagePass {
		t.Fatalf("expected stage 1 PASS (framing verified), got: %s", report.Stages[1].Status)
	}
	if report.Stages[2].Status != StageFail {
		t.Fatalf("expected stage 2 FAIL (transaction timeout), got: %s", report.Stages[2].Status)
	}
	if report.Stages[3].Status != StageNotTested {
		t.Fatalf("expected stage 3 NOT_TESTED, got: %s", report.Stages[3].Status)
	}
}

// 5. Valid RTU response and CRC validation
func TestDiagnosticParameter_ValidResponseAndCRC(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	// Set PM10 = 45.5 as 32-bit float in IEEE 754 Big-Endian (ABCD)
	// 45.5 in float32 = 0x42360000 -> Reg 0 = 0x4236, Reg 1 = 0x0000
	server.SetHoldingRegister(0, 0x4236)
	server.SetHoldingRegister(1, 0x0000)

	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("COM_TEST_VALID", 2, "PM10", "HOLDING", 0, "FLOAT32")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "PASS" || !report.Success {
		t.Fatalf("expected PASS, got: %s, error=%s", report.OverallStatus, report.ErrorMessage)
	}
	if report.FailureCategory != FailCategoryNone {
		t.Fatalf("expected FailCategoryNone, got: %s", report.FailureCategory)
	}
	if report.LastSuccessfulStage != "DATA_INTERPRETATION" {
		t.Fatalf("expected last stage DATA_INTERPRETATION, got: %s", report.LastSuccessfulStage)
	}

	// Verify all 4 stages passed
	for i, stage := range report.Stages {
		if stage.Status != StagePass {
			t.Fatalf("expected stage %d (%s) to PASS, got: %s", i, stage.Stage, stage.Status)
		}
	}

	// Verify transaction evidence
	if !report.TransactionEvidence.RequestTransmitted || !report.TransactionEvidence.ResponseReceived {
		t.Fatalf("expected request transmitted and response received")
	}
	if report.TransactionEvidence.CrcValidation != "PASS" {
		t.Fatalf("expected CRC PASS, got: %s", report.TransactionEvidence.CrcValidation)
	}
	if report.TransactionEvidence.RequestFrameHex == "" || report.TransactionEvidence.ResponseFrameHex == "" {
		t.Fatalf("expected hex frame bytes populated")
	}

	// Verify decoded value
	if report.DataEvidence.EngineeringValue == nil {
		t.Fatalf("expected engineering value populated")
	}
	if math.Abs(*report.DataEvidence.EngineeringValue-45.5) > 0.01 {
		t.Fatalf("expected PM10=45.5, got: %v", *report.DataEvidence.EngineeringValue)
	}
	if len(report.DataEvidence.RawRegisterWords) != 2 {
		t.Fatalf("expected 2 register words, got: %d", len(report.DataEvidence.RawRegisterWords))
	}
	if report.DataEvidence.RawRegisterWords[0] != 0x4236 || report.DataEvidence.RawRegisterWords[1] != 0x0000 {
		t.Fatalf("expected words [0x4236, 0x0000], got %v", report.DataEvidence.RawRegisterWords)
	}
}

// 6. Invalid CRC
func TestDiagnosticParameter_InvalidCRC(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetMode(modbus.SimModeCRCError, 0, 0)
	server.SetHoldingRegister(0, 0x1234)
	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("COM_TEST_CRC_FAIL", 2, "PM10", "HOLDING", 0, "INT16")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" {
		t.Fatalf("expected FAIL on invalid CRC, got: %s", report.OverallStatus)
	}
	if report.FailureCategory != FailCategoryInvalidCRC {
		t.Fatalf("expected FailCategoryInvalidCRC, got: %s", report.FailureCategory)
	}
	if report.TransactionEvidence.CrcValidation != "FAIL" {
		t.Fatalf("expected CRC validation FAIL, got: %s", report.TransactionEvidence.CrcValidation)
	}
	if !report.TransactionEvidence.RequestTransmitted || !report.TransactionEvidence.ResponseReceived {
		t.Fatalf("expected request transmitted and response received before CRC check")
	}
	if report.Stages[3].Status != StageNotTested {
		t.Fatalf("expected stage 3 NOT_TESTED when CRC fails")
	}
}

// 7. Modbus Exception Response
func TestDiagnosticParameter_ModbusException(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetMode(modbus.SimModeException, modbus.ExIllegalDataAddress, 0)
	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("COM_TEST_EX", 2, "PM10", "HOLDING", 999, "FLOAT32")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" {
		t.Fatalf("expected FAIL on exception, got: %s", report.OverallStatus)
	}
	if report.FailureCategory != FailCategoryModbusException {
		t.Fatalf("expected FailCategoryModbusException, got: %s", report.FailureCategory)
	}
	if report.TransactionEvidence.ModbusExceptionCode != modbus.ExIllegalDataAddress {
		t.Fatalf("expected exception code 0x02, got: %d", report.TransactionEvidence.ModbusExceptionCode)
	}
	// Physical link was proven WORKING!
	if report.TransactionEvidence.CrcValidation != "PASS" {
		t.Fatalf("exception frame has valid CRC, expected PASS")
	}
}

// 8. Incorrect Register Mapping
func TestDiagnosticParameter_IncorrectRegisterMapping(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	mgr := setupMockRTUManager(server)

	// Address 999999 exceeds allowable 16-bit Modbus PDU limit
	dev, param := createTestDeviceAndParam("COM_TEST_REG_MAP", 1, "INVALID", "HOLDING", 999999, "FLOAT32")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" || report.FailureCategory != FailCategoryRegisterConfig {
		t.Fatalf("expected FailCategoryRegisterConfig, got: %s", report.FailureCategory)
	}
	if report.FirstFailedStage != "SERIAL_CONFIGURATION" {
		t.Fatalf("expected first failed stage SERIAL_CONFIGURATION, got: %s", report.FirstFailedStage)
	}
}

// 9. Successful read with incorrect data decoding (1 byte coil decoded as float32)
func TestDiagnosticParameter_DecodeError(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetCoil(0, true)
	mgr := setupMockRTUManager(server)

	// Parameter configured as COIL with FLOAT32 data type (1 byte response cannot decode 4-byte float)
	dev, param := createTestDeviceAndParam("COM_TEST_DECODE_ERR", 1, "COIL_FLOAT", "COIL", 0, "FLOAT32")
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "FAIL" {
		t.Fatalf("expected FAIL on mismatched coil to float32 decode, got %s", report.OverallStatus)
	}
	if report.FailureCategory != FailCategoryDecodeError {
		t.Fatalf("expected FailCategoryDecodeError, got: %s", report.FailureCategory)
	}
	if report.FirstFailedStage != "DATA_INTERPRETATION" {
		t.Fatalf("expected first failed stage DATA_INTERPRETATION, got: %s", report.FirstFailedStage)
	}
	if report.Stages[0].Status != StagePass || report.Stages[1].Status != StagePass || report.Stages[2].Status != StagePass {
		t.Fatalf("expected stages 0, 1, 2 to pass before decode error, got: %s, %s, %s",
			report.Stages[0].Status, report.Stages[1].Status, report.Stages[2].Status)
	}
	if report.Stages[3].Status != StageFail {
		t.Fatalf("expected stage 3 to FAIL on decode error, got: %s", report.Stages[3].Status)
	}
}

// 10. Byte and word order decoding test (CDAB vs ABCD)
func TestDiagnosticParameter_ByteWordOrder(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	// Float32 123.456 in ABCD: High word 0x42F6, Low word 0xE979
	// In CDAB (Word-swapped): Reg 0 = 0xE979, Reg 1 = 0x42F6
	server.SetHoldingRegister(0, 0xE979)
	server.SetHoldingRegister(1, 0x42F6)

	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("COM_TEST_CDAB", 1, "FLOW", "HOLDING", 0, "FLOAT32")
	param.ByteOrder = "CDAB"
	ctx := context.Background()

	report, err := mgr.DiagnosticParameterRead(ctx, dev, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.OverallStatus != "PASS" {
		t.Fatalf("expected PASS on CDAB decode, got %s", report.OverallStatus)
	}
	if report.DataEvidence.EngineeringValue == nil {
		t.Fatalf("expected engineering value")
	}
	val := *report.DataEvidence.EngineeringValue
	if math.Abs(val-123.456) > 0.01 {
		t.Fatalf("expected ~123.456 for CDAB float32, got: %v", val)
	}
}

// 11. Diagnostic execution while background polling is active (shared serial bus safety)
func TestDiagnosticParameter_WhilePollingActive(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 0x1111)
	server.SetHoldingRegister(1, 0x2222)
	mgr := setupMockRTUManager(server)

	dev, param := createTestDeviceAndParam("COM_TEST_SHARED_BUS_ACTIVE", 1, "SENSOR", "HOLDING", 0, "INT16")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var pollErrors int64

	// Launch background concurrent polling transactions on the same serial port
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := mgr.ExecuteWithRetry(ctx, dev, ModbusReadRequest{
						SlaveID:         1,
						FunctionCode:    modbus.FunctionReadHoldingRegisters,
						StartingAddress: 0,
						Quantity:        1,
					})
					if err != nil && !errorsIsCanceled(err) {
						pollErrors++
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	// Execute diagnostic parameter read while background polling is running
	time.Sleep(20 * time.Millisecond)
	diagReport, err := mgr.DiagnosticParameterRead(context.Background(), dev, param)
	cancel()
	wg.Wait()

	if err != nil {
		t.Fatalf("diagnostic read failed during active polling: %v", err)
	}
	if diagReport.OverallStatus != "PASS" {
		t.Fatalf("expected diagnostic PASS while polling is active, got: %s (error: %s)",
			diagReport.OverallStatus, diagReport.ErrorMessage)
	}
	if diagReport.TransactionEvidence.CrcValidation != "PASS" {
		t.Fatalf("expected CRC PASS, got: %s", diagReport.TransactionEvidence.CrcValidation)
	}
}

// 12. mbpoll command generation check
func TestDiagnosticParameter_MbpollCommandGeneration(t *testing.T) {
	cfg := &model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: "/dev/ttyUSB1",
		BaudRate:   9600,
		DataBits:   8,
		StopBits:   1,
		Parity:     "N",
		SlaveID:    2,
	}
	param := &model.Parameter{
		DataType: "FLOAT32",
	}

	cmd := FormatMbpollCommand(cfg, param, modbus.FunctionReadHoldingRegisters, 0, 2, 2, "/dev/ttyUSB1")
	expected := "mbpoll -m rtu -a 2 -b 9600 -d 8 -s 1 -P none -t 3:float -r 0 -c 1 -0 -1 /dev/ttyUSB1"
	if cmd != expected {
		t.Errorf("mbpoll command mismatch:\nExpected: %s\nGot:      %s", expected, cmd)
	}
}

func errorsIsCanceled(err error) bool {
	if err == nil {
		return false
	}
	return runtime.GOOS != "" && (err == context.Canceled || err == context.DeadlineExceeded)
}

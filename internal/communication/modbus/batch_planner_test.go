package modbus

import (
	"strings"
	"testing"

	"datalogger/internal/model"
)

func TestBatchPlanner_ContiguousRegisters(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "METER-01",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusTCP,
			SlaveID:           1,
			BatchReadEnabled:  true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:    0,
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			ParameterCode:   "VOLT_L1",
			DataType:        model.DataTypeFloat32, // 2 registers
			RegisterType:    "INPUT_REGISTER",
			RegisterAddress: 30001, // PDU 0
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			ParameterCode:   "VOLT_L2",
			DataType:        model.DataTypeFloat32, // 2 registers
			RegisterType:    "INPUT_REGISTER",
			RegisterAddress: 30003, // PDU 2
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              3,
			ParameterCode:   "VOLT_L3",
			DataType:        model.DataTypeFloat32, // 2 registers
			RegisterType:    "INPUT_REGISTER",
			RegisterAddress: 30005, // PDU 4
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalParameters != 3 {
		t.Fatalf("expected 3 total parameters, got %d", plan.TotalParameters)
	}
	if plan.TotalModbusRequests != 1 {
		t.Fatalf("expected 1 consolidated batch request, got %d", plan.TotalModbusRequests)
	}

	job := plan.Jobs[0]
	if job.IsSingle {
		t.Errorf("expected job to be batch, got is_single=true")
	}
	if job.StartingAddress != 0 {
		t.Errorf("expected starting address 0, got %d", job.StartingAddress)
	}
	if job.Quantity != 6 {
		t.Errorf("expected quantity 6 registers (3 * 2), got %d", job.Quantity)
	}
	if job.DisplayAddress != 30001 {
		t.Errorf("expected display address 30001, got %d", job.DisplayAddress)
	}
	if job.DisplayEndAddr != 30006 {
		t.Errorf("expected display end address 30006, got %d", job.DisplayEndAddr)
	}
	if len(job.Parameters) != 3 {
		t.Fatalf("expected 3 parameters in job, got %d", len(job.Parameters))
	}

	// Verify byte offsets
	// Param 1: PDU 0, 2 regs = 4 bytes, offset 0
	if job.Parameters[0].ByteOffset != 0 || job.Parameters[0].ByteLength != 4 {
		t.Errorf("param 1 byte slice expected [0:4], got offset %d len %d",
			job.Parameters[0].ByteOffset, job.Parameters[0].ByteLength)
	}
	// Param 2: PDU 2, 2 regs = 4 bytes, offset 4
	if job.Parameters[1].ByteOffset != 4 || job.Parameters[1].ByteLength != 4 {
		t.Errorf("param 2 byte slice expected [4:8], got offset %d len %d",
			job.Parameters[1].ByteOffset, job.Parameters[1].ByteLength)
	}
	// Param 3: PDU 4, 2 regs = 4 bytes, offset 8
	if job.Parameters[2].ByteOffset != 8 || job.Parameters[2].ByteLength != 4 {
		t.Errorf("param 3 byte slice expected [8:12], got offset %d len %d",
			job.Parameters[2].ByteOffset, job.Parameters[2].ByteLength)
	}

	if plan.EstimatedReductionPercent != 66.7 {
		t.Errorf("expected ~66.7%% reduction, got %.1f%%", plan.EstimatedReductionPercent)
	}
}

func TestBatchPlanner_GapExceedingMaxRegisterGap(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "METER-01",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusTCP,
			SlaveID:           1,
			BatchReadEnabled:  true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:    0, // zero gap tolerance
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			ParameterCode:   "TEMP",
			DataType:        model.DataTypeFloat32, // 2 registers
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001, // PDU 0
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			ParameterCode:   "HUMIDITY",
			DataType:        model.DataTypeFloat32, // 2 registers
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40005, // PDU 4 (gap of 2 registers: 40003 & 40004)
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalModbusRequests != 2 {
		t.Fatalf("expected 2 separate requests due to gap, got %d", plan.TotalModbusRequests)
	}

	if plan.Jobs[0].StartingAddress != 0 || plan.Jobs[0].Quantity != 2 {
		t.Errorf("job 0 expected PDU 0 qty 2, got PDU %d qty %d", plan.Jobs[0].StartingAddress, plan.Jobs[0].Quantity)
	}
	if plan.Jobs[1].StartingAddress != 4 || plan.Jobs[1].Quantity != 2 {
		t.Errorf("job 1 expected PDU 4 qty 2, got PDU %d qty %d", plan.Jobs[1].StartingAddress, plan.Jobs[1].Quantity)
	}
}

func TestBatchPlanner_GapAllowedWhenConfigured(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "METER-01",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusTCP,
			SlaveID:           1,
			BatchReadEnabled:  true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:    2, // allows up to 2 unused registers gap
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			ParameterCode:   "TEMP",
			DataType:        model.DataTypeFloat32, // 2 registers
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001, // PDU 0..1
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			ParameterCode:   "HUMIDITY",
			DataType:        model.DataTypeFloat32, // 2 registers
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40004, // PDU 3..4 (gap = PDU 2, 1 register gap <= 2 allowed)
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalModbusRequests != 1 {
		t.Fatalf("expected 1 consolidated batch request, got %d", plan.TotalModbusRequests)
	}

	job := plan.Jobs[0]
	// Starting at PDU 0, ending at PDU 5 (PDU 0,1,2,3,4 = 5 registers)
	if job.StartingAddress != 0 || job.Quantity != 5 {
		t.Errorf("expected PDU 0 qty 5, got PDU %d qty %d", job.StartingAddress, job.Quantity)
	}

	// Param 1 offset 0, len 4 (PDU 0..1)
	if job.Parameters[0].ByteOffset != 0 || job.Parameters[0].ByteLength != 4 {
		t.Errorf("param 1 slice error: offset %d len %d", job.Parameters[0].ByteOffset, job.Parameters[0].ByteLength)
	}
	// Param 2 offset: PDU 3 * 2 = 6, len 4 (PDU 3..4)
	if job.Parameters[1].ByteOffset != 6 || job.Parameters[1].ByteLength != 4 {
		t.Errorf("param 2 slice error: expected offset 6 len 4, got offset %d len %d",
			job.Parameters[1].ByteOffset, job.Parameters[1].ByteLength)
	}
}

func TestBatchPlanner_DifferentSlaveOrFunctionCode(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "MULTI-DEV",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusRTU,
			SlaveID:           1,
			BatchReadEnabled:  true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:    0,
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			ParameterCode:   "INPUT_REG",
			DataType:        model.DataTypeFloat32,
			RegisterType:    "INPUT_REGISTER", // FC 04
			RegisterAddress: 30001,
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			ParameterCode:   "HOLDING_REG",
			DataType:        model.DataTypeFloat32,
			RegisterType:    "HOLDING_REGISTER", // FC 03
			RegisterAddress: 40001,
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              3,
			ParameterCode:   "SLAVE2_REG",
			DataType:        model.DataTypeFloat32,
			RegisterType:    "INPUT_REGISTER", // FC 04
			RegisterAddress: 30003,
			SlaveID:         2, // Slave ID 2
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalModbusRequests != 3 {
		t.Fatalf("expected 3 strictly isolated requests, got %d", plan.TotalModbusRequests)
	}

	for _, j := range plan.Jobs {
		if !j.IsSingle {
			t.Errorf("expected all 3 jobs to be single, job %s was not", j.ID)
		}
	}
}

func TestBatchPlanner_ExceedingMaxBatchRegisters(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "BIG-BLOCK",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusTCP,
			SlaveID:           1,
			BatchReadEnabled:  true,
			MaxBatchRegisters: 4, // artificially small limit: max 4 registers
			MaxRegisterGap:    0,
		},
	}

	params := []model.Parameter{
		{ID: 1, ParameterCode: "P1", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40001, SlaveID: 1, BatchMode: "AUTO", Enabled: true}, // 2 reg (PDU 0..1)
		{ID: 2, ParameterCode: "P2", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40003, SlaveID: 1, BatchMode: "AUTO", Enabled: true}, // 2 reg (PDU 2..3) -> total 4 regs!
		{ID: 3, ParameterCode: "P3", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40005, SlaveID: 1, BatchMode: "AUTO", Enabled: true}, // 2 reg (PDU 4..5) -> exceeds 4!
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalModbusRequests != 2 {
		t.Fatalf("expected 2 batch requests due to limit 4, got %d", plan.TotalModbusRequests)
	}

	// Job 1 has P1 and P2 (4 registers)
	if plan.Jobs[0].StartingAddress != 0 || plan.Jobs[0].Quantity != 4 {
		t.Errorf("job 0 expected PDU 0 qty 4, got PDU %d qty %d", plan.Jobs[0].StartingAddress, plan.Jobs[0].Quantity)
	}
	if len(plan.Jobs[0].Parameters) != 2 {
		t.Errorf("job 0 expected 2 parameters, got %d", len(plan.Jobs[0].Parameters))
	}

	// Job 2 has P3 (2 registers)
	if plan.Jobs[1].StartingAddress != 4 || plan.Jobs[1].Quantity != 2 {
		t.Errorf("job 1 expected PDU 4 qty 2, got PDU %d qty %d", plan.Jobs[1].StartingAddress, plan.Jobs[1].Quantity)
	}
	if len(plan.Jobs[1].Parameters) != 1 {
		t.Errorf("job 1 expected 1 parameter, got %d", len(plan.Jobs[1].Parameters))
	}
}

func TestBatchPlanner_SingleModeOptOut(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "METER-01",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusTCP,
			SlaveID:           1,
			BatchReadEnabled:  true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:    0,
		},
	}

	params := []model.Parameter{
		{ID: 1, ParameterCode: "AUTO_1", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40001, SlaveID: 1, BatchMode: "AUTO", Enabled: true},
		{ID: 2, ParameterCode: "OPT_OUT", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40003, SlaveID: 1, BatchMode: "SINGLE", Enabled: true}, // SINGLE!
		{ID: 3, ParameterCode: "AUTO_2", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40005, SlaveID: 1, BatchMode: "AUTO", Enabled: true},
	}

	plan := PlanDeviceBatches(dev, params)

	// Since OPT_OUT is SINGLE (PDU 2..3), AUTO_1 (PDU 0..1) and AUTO_2 (PDU 4..5) are separated by a gap of 2 registers
	// and with max_register_gap=0, AUTO_1 and AUTO_2 also cannot be merged!
	// So there will be 3 separate jobs.
	if plan.TotalModbusRequests != 3 {
		t.Fatalf("expected 3 requests (1 single opt-out + 2 separated auto), got %d", plan.TotalModbusRequests)
	}

	var foundOptOut bool
	for _, j := range plan.Jobs {
		if len(j.Parameters) == 1 && j.Parameters[0].Parameter.ParameterCode == "OPT_OUT" {
			foundOptOut = true
			if !j.IsSingle {
				t.Errorf("expected OPT_OUT job to have IsSingle=true")
			}
		}
	}
	if !foundOptOut {
		t.Errorf("OPT_OUT parameter was not found in planned jobs")
	}
}

func TestBatchPlanner_DefaultOffPreservesSingleRead(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "METER-01",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusTCP,
			SlaveID:           1,
			BatchReadEnabled:  false, // DEFAULT OFF!
			MaxBatchRegisters: 32,
			MaxRegisterGap:    0,
		},
	}

	params := []model.Parameter{
		{ID: 1, ParameterCode: "P1", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40001, SlaveID: 1, BatchMode: "AUTO", Enabled: true},
		{ID: 2, ParameterCode: "P2", DataType: model.DataTypeFloat32, RegisterType: "HOLDING_REGISTER", RegisterAddress: 40003, SlaveID: 1, BatchMode: "AUTO", Enabled: true},
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalModbusRequests != 2 {
		t.Fatalf("when BatchReadEnabled=false, expected exactly 2 single requests, got %d", plan.TotalModbusRequests)
	}
	if plan.EstimatedReductionPercent != 0.0 {
		t.Errorf("expected 0.0%% reduction, got %.1f%%", plan.EstimatedReductionPercent)
	}
}

func TestBatchPlanner_MixedDataTypesAndByteOrders(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "METER-01",
		Connection: &model.DeviceConnection{
			DeviceID:          1,
			Protocol:          model.ProtocolModbusTCP,
			SlaveID:           1,
			BatchReadEnabled:  true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:    0,
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			ParameterCode:   "VOLT_FLOAT32_CDAB",
			DataType:        model.DataTypeFloat32, // 2 regs (4 bytes)
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001, // PDU 0..1
			ByteOrder:       "CDAB",
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			ParameterCode:   "CURRENT_INT32_BADC",
			DataType:        model.DataTypeInt32, // 2 regs (4 bytes)
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40003, // PDU 2..3
			ByteOrder:       "BADC",
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              3,
			ParameterCode:   "STATUS_INT16_BA",
			DataType:        model.DataTypeInt16, // 1 reg (2 bytes)
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40005, // PDU 4
			ByteOrder:       "BA",
			SlaveID:         1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalModbusRequests != 1 {
		t.Fatalf("expected 1 consolidated batch request, got %d", plan.TotalModbusRequests)
	}

	job := plan.Jobs[0]
	if job.Quantity != 5 { // 2 + 2 + 1 = 5 registers
		t.Fatalf("expected quantity 5, got %d", job.Quantity)
	}

	// Simulated Modbus raw response of 5 registers (10 bytes)
	// Reg 0..1 (CDAB float 220.5):
	// Big-Endian IEEE754 of 220.5: 0x435C8000 -> ABCD = [0x43, 0x5C, 0x80, 0x00]
	// CDAB word-swapped: [0x80, 0x00, 0x43, 0x5C]
	// Reg 2..3 (BADC int32 100000):
	// Big-Endian int32 100000: 0x000186A0 -> ABCD = [0x00, 0x01, 0x86, 0xA0]
	// BADC byte-swapped: [0x01, 0x00, 0xA0, 0x86]
	// Reg 4 (BA int16 500):
	// Big-Endian int16 500: 0x01F4 -> AB = [0x01, 0xF4]
	// BA byte-swapped: [0xF4, 0x01]

	rawResponseData := []byte{
		0x80, 0x00, 0x43, 0x5C, // Param 1 (CDAB float 220.5)
		0x01, 0x00, 0xA0, 0x86, // Param 2 (BADC int32 100000)
		0xF4, 0x01,             // Param 3 (BA int16 500)
	}

	// 1. Decode Param 1 from slice
	sp1 := job.Parameters[0]
	slice1 := rawResponseData[sp1.ByteOffset : sp1.ByteOffset+sp1.ByteLength]
	dec1, err1 := DecodeRegisters(slice1, sp1.Parameter.DataType, sp1.Parameter.ByteOrder)
	if err1 != nil {
		t.Fatalf("decode param 1 failed: %v", err1)
	}
	if dec1.NumericValue != 220.5 {
		t.Errorf("param 1 expected 220.5, got %f", dec1.NumericValue)
	}

	// 2. Decode Param 2 from slice
	sp2 := job.Parameters[1]
	slice2 := rawResponseData[sp2.ByteOffset : sp2.ByteOffset+sp2.ByteLength]
	dec2, err2 := DecodeRegisters(slice2, sp2.Parameter.DataType, sp2.Parameter.ByteOrder)
	if err2 != nil {
		t.Fatalf("decode param 2 failed: %v", err2)
	}
	if dec2.NumericValue != 100000 {
		t.Errorf("param 2 expected 100000, got %f", dec2.NumericValue)
	}

	// 3. Decode Param 3 from slice
	sp3 := job.Parameters[2]
	slice3 := rawResponseData[sp3.ByteOffset : sp3.ByteOffset+sp3.ByteLength]
	dec3, err3 := DecodeRegisters(slice3, sp3.Parameter.DataType, sp3.Parameter.ByteOrder)
	if err3 != nil {
		t.Fatalf("decode param 3 failed: %v", err3)
	}
	if dec3.NumericValue != 500 {
		t.Errorf("param 3 expected 500, got %f", dec3.NumericValue)
	}
}

func TestBatchPlanner_InvalidAndSkippedParametersAuditing(t *testing.T) {
	dev := &model.Device{
		ID:         1,
		DeviceCode: "AUDIT-DEV-01",
		Connection: &model.DeviceConnection{
			DeviceID:         1,
			Protocol:         model.ProtocolModbusTCP,
			SlaveID:          1,
			BatchReadEnabled: true,
		},
	}

	params := []model.Parameter{
		{
			ID:            1,
			ParameterCode: "DISABLED_PARAM",
			ParameterName: "Disabled Sensor",
			Enabled:       false,
		},
		{
			ID:            2,
			ParameterCode: "CALCULATED_ISPU",
			ParameterName: "Calculated Formula Param",
			RegisterType:  "VARIABLE",
			Enabled:       true,
		},
		{
			ID:              3,
			ParameterCode:   "BAD_ADDR",
			ParameterName:   "Bad Address Param",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 999999,
			Enabled:         true,
		},
		{
			ID:              4,
			ParameterCode:   "VALID_AUTO",
			ParameterName:   "Valid Auto Param",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001,
			DataType:        model.DataTypeFloat32,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	plan := PlanDeviceBatches(dev, params)

	if plan.TotalParameters != 1 {
		t.Fatalf("expected 1 active parameter, got %d", plan.TotalParameters)
	}
	if len(plan.InvalidParameters) != 3 {
		t.Fatalf("expected 3 invalid/skipped parameters, got %d", len(plan.InvalidParameters))
	}

	// Verify reason diagnostics
	reasons := make(map[string]string)
	for _, inv := range plan.InvalidParameters {
		reasons[inv.ParameterCode] = inv.Reason
	}

	if reasons["DISABLED_PARAM"] != "Parameter is disabled" {
		t.Errorf("unexpected reason for disabled param: %s", reasons["DISABLED_PARAM"])
	}
	if reasons["CALCULATED_ISPU"] != "Virtual / Internal calculation variable (not read via Modbus)" {
		t.Errorf("unexpected reason for virtual param: %s", reasons["CALCULATED_ISPU"])
	}
	if !strings.Contains(reasons["BAD_ADDR"], "Invalid Modbus register address") {
		t.Errorf("unexpected reason for bad address: %s", reasons["BAD_ADDR"])
	}
}



package modbus

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"datalogger/internal/model"
)

type resolvedParam struct {
	Param         *model.Parameter
	SlaveID       byte
	FunctionCode  byte
	PDUAddress    uint16
	RegisterCount uint16
	RegisterType  string
	ModiconAddr   int
}

// PlanDeviceBatches calculates the optimal Modbus batch polling plan for a device and its active parameters.
// This is a deterministic, side-effect-free pure function.
func PlanDeviceBatches(device *model.Device, params []model.Parameter) BatchPlan {
	if device == nil || device.Connection == nil {
		return BatchPlan{}
	}

	cfg := device.Connection
	batchEnabled := cfg.BatchReadEnabled

	maxBatchRegs := cfg.MaxBatchRegisters
	if maxBatchRegs <= 0 {
		maxBatchRegs = 32
	}
	if maxBatchRegs > 120 {
		maxBatchRegs = 120
	}

	maxGap := cfg.MaxRegisterGap
	if maxGap < 0 {
		maxGap = 0
	}
	if maxGap > 10 {
		maxGap = 10
	}

	plan := BatchPlan{
		DeviceID:          device.ID,
		DeviceCode:        device.DeviceCode,
		BatchReadEnabled:  batchEnabled,
		MaxBatchRegisters: maxBatchRegs,
		MaxRegisterGap:    maxGap,
		Jobs:              []BatchJob{},
	}

	// 1. Resolve and filter enabled Modbus parameters
	var resolvedList []*resolvedParam
	for i := range params {
		p := &params[i]
		if !p.Enabled {
			continue
		}

		// Skip virtual / internal variables that are not read from Modbus
		normType := strings.ToUpper(strings.TrimSpace(p.RegisterType))
		if normType == "VARIABLE" || normType == "VIRTUAL" {
			continue
		}

		res, err := ResolveRegisterAddress(p.RegisterType, p.RegisterAddress)
		if err != nil {
			continue
		}

		qty := RequiredRegisterCount(p.DataType)
		if res.FunctionCode == FunctionReadCoils || res.FunctionCode == FunctionReadDiscreteInputs {
			qty = 1
		}

		slaveID := byte(p.SlaveID)
		if slaveID == 0 {
			slaveID = byte(cfg.SlaveID)
		}
		if slaveID == 0 {
			slaveID = 1
		}

		resolvedList = append(resolvedList, &resolvedParam{
			Param:         p,
			SlaveID:       slaveID,
			FunctionCode:  res.FunctionCode,
			PDUAddress:    res.PDUAddress,
			RegisterCount: qty,
			RegisterType:  p.RegisterType,
			ModiconAddr:   p.RegisterAddress,
		})
	}

	plan.TotalParameters = len(resolvedList)
	if plan.TotalParameters == 0 {
		return plan
	}

	// 2. If BatchRead is disabled, construct individual single-read jobs for each parameter
	if !batchEnabled {
		for _, rp := range resolvedList {
			job := buildJobFromChunk(device, []*resolvedParam{rp}, true)
			plan.Jobs = append(plan.Jobs, job)
		}
		plan.TotalModbusRequests = len(plan.Jobs)
		plan.EstimatedReductionPercent = 0.0
		return plan
	}

	// 3. Batch Read is enabled: Partition into SINGLE jobs and AUTO batch candidates
	var singleJobs []BatchJob
	groupedCandidates := make(map[string][]*resolvedParam)

	for _, rp := range resolvedList {
		mode := strings.ToUpper(strings.TrimSpace(rp.Param.BatchMode))
		if mode == string(BatchModeSingle) {
			singleJobs = append(singleJobs, buildJobFromChunk(device, []*resolvedParam{rp}, true))
			continue
		}

		groupKey := fmt.Sprintf("%d_%d", rp.SlaveID, rp.FunctionCode)
		groupedCandidates[groupKey] = append(groupedCandidates[groupKey], rp)
	}

	// 4. In each group, sort parameters by PDUAddress and chunk into contiguous batches
	var batchJobs []BatchJob

	// Sort group keys deterministically
	var groupKeys []string
	for k := range groupedCandidates {
		groupKeys = append(groupKeys, k)
	}
	sort.Strings(groupKeys)

	for _, k := range groupKeys {
		group := groupedCandidates[k]

		// Sort by PDUAddress ascending; tie-break with Parameter ID
		sort.SliceStable(group, func(i, j int) bool {
			if group[i].PDUAddress != group[j].PDUAddress {
				return group[i].PDUAddress < group[j].PDUAddress
			}
			return group[i].Param.ID < group[j].Param.ID
		})

		var currentChunk []*resolvedParam

		for _, rp := range group {
			if len(currentChunk) == 0 {
				currentChunk = append(currentChunk, rp)
				continue
			}

			lastEnd := chunkEnd(currentChunk)
			gap := 0
			if rp.PDUAddress > lastEnd {
				gap = int(rp.PDUAddress - lastEnd)
			}

			chunkStart := currentChunk[0].PDUAddress
			proposedEnd := lastEnd
			if rp.PDUAddress+rp.RegisterCount > proposedEnd {
				proposedEnd = rp.PDUAddress + rp.RegisterCount
			}
			proposedSpan := int(proposedEnd - chunkStart)

			if gap <= maxGap && proposedSpan <= maxBatchRegs {
				currentChunk = append(currentChunk, rp)
			} else {
				// Close current chunk and start new one
				isSingle := len(currentChunk) == 1
				batchJobs = append(batchJobs, buildJobFromChunk(device, currentChunk, isSingle))
				currentChunk = []*resolvedParam{rp}
			}
		}

		if len(currentChunk) > 0 {
			isSingle := len(currentChunk) == 1
			batchJobs = append(batchJobs, buildJobFromChunk(device, currentChunk, isSingle))
		}
	}

	// 5. Combine batch jobs and single jobs
	allJobs := append(batchJobs, singleJobs...)

	// Sort all jobs deterministically for clean execution order: SlaveID, FunctionCode, StartingAddress
	sort.SliceStable(allJobs, func(i, j int) bool {
		if allJobs[i].SlaveID != allJobs[j].SlaveID {
			return allJobs[i].SlaveID < allJobs[j].SlaveID
		}
		if allJobs[i].FunctionCode != allJobs[j].FunctionCode {
			return allJobs[i].FunctionCode < allJobs[j].FunctionCode
		}
		return allJobs[i].StartingAddress < allJobs[j].StartingAddress
	})

	// Re-assign human-readable sequential IDs
	for idx := range allJobs {
		prefix := "BATCH"
		if allJobs[idx].IsSingle {
			prefix = "SINGLE"
		}
		allJobs[idx].ID = fmt.Sprintf("%s-%s-FC%02d-S%d-ADDR%d",
			prefix, device.DeviceCode, allJobs[idx].FunctionCode, allJobs[idx].SlaveID, allJobs[idx].DisplayAddress)
	}

	plan.Jobs = allJobs
	plan.TotalModbusRequests = len(allJobs)

	if plan.TotalParameters > 0 && plan.TotalModbusRequests <= plan.TotalParameters {
		reduction := float64(plan.TotalParameters-plan.TotalModbusRequests) / float64(plan.TotalParameters) * 100.0
		plan.EstimatedReductionPercent = math.Round(reduction*10.0) / 10.0
	} else {
		plan.EstimatedReductionPercent = 0.0
	}

	return plan
}

func chunkEnd(chunk []*resolvedParam) uint16 {
	var maxEnd uint16 = 0
	for _, rp := range chunk {
		end := rp.PDUAddress + rp.RegisterCount
		if end > maxEnd {
			maxEnd = end
		}
	}
	return maxEnd
}

func buildJobFromChunk(device *model.Device, chunk []*resolvedParam, isSingle bool) BatchJob {
	first := chunk[0]
	batchStart := first.PDUAddress
	batchEnd := chunkEnd(chunk)
	qty := batchEnd - batchStart
	if qty == 0 {
		qty = 1
	}

	displayAddr := first.ModiconAddr
	if len(chunk) > 1 || displayAddr <= 0 {
		displayAddr = pduToModicon(first.FunctionCode, batchStart)
	}
	displayEnd := displayAddr + int(qty) - 1

	job := BatchJob{
		DeviceID:        device.ID,
		DeviceCode:      device.DeviceCode,
		SlaveID:         first.SlaveID,
		FunctionCode:    first.FunctionCode,
		RegisterType:    first.RegisterType,
		StartingAddress: batchStart,
		Quantity:        qty,
		DisplayAddress:  displayAddr,
		DisplayEndAddr:  displayEnd,
		IsSingle:        isSingle,
		Parameters:      make([]SlicedParameter, 0, len(chunk)),
	}

	for _, rp := range chunk {
		sp := SlicedParameter{
			Parameter:     rp.Param,
			PDUAddress:    rp.PDUAddress,
			RegisterCount: rp.RegisterCount,
		}

		if first.FunctionCode == FunctionReadCoils || first.FunctionCode == FunctionReadDiscreteInputs {
			sp.BitOffset = int(rp.PDUAddress - batchStart)
			sp.ByteOffset = sp.BitOffset / 8
			sp.ByteLength = (int(rp.RegisterCount) + 7) / 8
		} else {
			sp.ByteOffset = int(rp.PDUAddress - batchStart) * 2
			sp.ByteLength = int(rp.RegisterCount) * 2
		}

		job.Parameters = append(job.Parameters, sp)
	}

	return job
}

func pduToModicon(fc byte, pdu uint16) int {
	switch fc {
	case FunctionReadHoldingRegisters:
		return 40001 + int(pdu)
	case FunctionReadInputRegisters:
		return 30001 + int(pdu)
	case FunctionReadCoils:
		return 1 + int(pdu)
	case FunctionReadDiscreteInputs:
		return 10001 + int(pdu)
	default:
		return int(pdu)
	}
}

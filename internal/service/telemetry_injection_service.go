package service

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"datalogger/internal/aggregation"
	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/pkg/formula"
)

const (
	ConflictFillMissingAndNull = "FILL_MISSING_AND_NULL"
	ConflictOverwriteAll       = "OVERWRITE_ALL"

	RandomModeSmoothWalk = "SMOOTH_WALK"
	RandomModeUniform    = "UNIFORM"

	ValueTargetProcessed = "PROCESSED"
	ValueTargetRaw       = "RAW"

	maxInjectionTotalSamples = 60000
	maxPreviewSamples        = 60
)

// TelemetryInjectionParamInput defines the per-parameter range input for Superadmin telemetry injection
type TelemetryInjectionParamInput struct {
	ParameterID uint    `json:"parameter_id"`
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
}

// TelemetryInjectionRequest encapsulates a preview or execution request for Superadmin telemetry injection
type TelemetryInjectionRequest struct {
	DeviceID               uint                           `json:"device_id"`
	StartTime              time.Time                      `json:"start_time"`
	EndTime                time.Time                      `json:"end_time"`
	SampleStepSeconds      int                            `json:"sample_step_seconds"`
	ConflictMode           string                         `json:"conflict_mode"`
	RandomMode             string                         `json:"random_mode"`
	ValueTarget            string                         `json:"value_target"`
	RefreshOutbound        *bool                          `json:"refresh_outbound,omitempty"`
	EnqueueMissingOutbound bool                           `json:"enqueue_missing_outbound"`
	Seed                   int64                          `json:"seed,omitempty"`
	Note                   string                         `json:"note,omitempty"`
	Parameters             []TelemetryInjectionParamInput `json:"parameters"`
}

// InjectionSamplePreview represents a single generated sample row shown in dry-run preview
type InjectionSamplePreview struct {
	ParameterID    uint      `json:"parameter_id"`
	ParameterCode  string    `json:"parameter_code"`
	ParameterName  string    `json:"parameter_name"`
	Unit           string    `json:"unit"`
	ReceivedAt     time.Time `json:"received_at"`
	RawValue       float64   `json:"raw_value"`
	ProcessedValue float64   `json:"processed_value"`
	RawHex         string    `json:"raw_hex"`
}

// InjectionParameterSummary summarizes planned/executed injection metrics per parameter
type InjectionParameterSummary struct {
	ParameterID           uint    `json:"parameter_id"`
	ParameterCode         string  `json:"parameter_code"`
	ParameterName         string  `json:"parameter_name"`
	Unit                  string  `json:"unit"`
	MinInput              float64 `json:"min_input"`
	MaxInput              float64 `json:"max_input"`
	PlannedSamples        int     `json:"planned_samples"`
	SkippedSamples        int     `json:"skipped_samples"`
	ExistingValidInWindow int     `json:"existing_valid_in_window"`
	ExistingBadInWindow   int     `json:"existing_bad_in_window"`
	GeneratedMinProcessed float64 `json:"generated_min_processed"`
	GeneratedMaxProcessed float64 `json:"generated_max_processed"`
	GeneratedAvgProcessed float64 `json:"generated_avg_processed"`
	GeneratedMinRaw       float64 `json:"generated_min_raw"`
	GeneratedMaxRaw       float64 `json:"generated_max_raw"`
}

// TelemetryInjectionPreviewResponse contains dry-run metrics and sample rows prior to persistence
type TelemetryInjectionPreviewResponse struct {
	DeviceID            uint                        `json:"device_id"`
	DeviceCode          string                      `json:"device_code"`
	DeviceName          string                      `json:"device_name"`
	StartTime           time.Time                   `json:"start_time"`
	EndTime             time.Time                   `json:"end_time"`
	SampleStepSeconds   int                         `json:"sample_step_seconds"`
	ConflictMode        string                      `json:"conflict_mode"`
	RandomMode          string                      `json:"random_mode"`
	ValueTarget         string                      `json:"value_target"`
	TotalCandidateSlots int                         `json:"total_candidate_slots"`
	TotalPlannedSamples int                         `json:"total_planned_samples"`
	TotalSkippedSamples int                         `json:"total_skipped_samples"`
	ActiveAggIntervals  []int                       `json:"active_agg_intervals"`
	EstimatedAggBuckets int                         `json:"estimated_agg_buckets"`
	ParameterSummaries  []InjectionParameterSummary `json:"parameter_summaries"`
	SamplePreview       []InjectionSamplePreview    `json:"sample_preview"`
}

// TelemetryInjectionExecuteResponse returns the persisted batch and preview breakdown
type TelemetryInjectionExecuteResponse struct {
	Batch   *model.TelemetryInjectionBatch     `json:"batch"`
	Preview *TelemetryInjectionPreviewResponse `json:"preview"`
}

// TelemetryInjectionRollbackResponse returns rollback metrics
type TelemetryInjectionRollbackResponse struct {
	Batch               *model.TelemetryInjectionBatch `json:"batch"`
	DeletedRawCount     int64                          `json:"deleted_raw_count"`
	RecalculatedBuckets int                            `json:"recalculated_buckets"`
	RefreshedOutbound   int                            `json:"refreshed_outbound"`
}

// DeviceParameterInjectionHint provides parameter scaling metadata and 1-click range hints for the UI
type DeviceParameterInjectionHint struct {
	ParameterID           uint       `json:"parameter_id"`
	ParameterCode         string     `json:"parameter_code"`
	ParameterName         string     `json:"parameter_name"`
	Unit                  string     `json:"unit"`
	DataType              string     `json:"data_type"`
	Precision             int        `json:"precision"`
	Scale                 float64    `json:"scale"`
	Offset                float64    `json:"offset"`
	Formula               string     `json:"formula,omitempty"`
	CurrentProcessedValue *float64   `json:"current_processed_value,omitempty"`
	CurrentRawValue       *float64   `json:"current_raw_value,omitempty"`
	CurrentQuality        string     `json:"current_quality,omitempty"`
	CurrentReceivedAt     *time.Time `json:"current_received_at,omitempty"`
	SuggestedMin          float64    `json:"suggested_min"`
	SuggestedMax          float64    `json:"suggested_max"`
	AggregationIntervals  []int      `json:"aggregation_intervals"`
}

// DeviceInjectionHintsResponse provides full device & parameter hints for the Superadmin Telemetry Injector UI
type DeviceInjectionHintsResponse struct {
	DeviceID           uint                           `json:"device_id"`
	DeviceCode         string                         `json:"device_code"`
	DeviceName         string                         `json:"device_name"`
	Protocol           string                         `json:"protocol"`
	Timezone           string                         `json:"timezone"`
	ActiveAggIntervals []int                          `json:"active_agg_intervals"`
	Parameters         []DeviceParameterInjectionHint `json:"parameters"`
}

// TelemetryInjectionService orchestrates Superadmin-only telemetry generation, aggregation recalculation,
// outbound queue re-hydration, rollback, and strict Superadmin badge annotation.
type TelemetryInjectionService struct {
	repo        *repository.TelemetryInjectionRepository
	devRepo     *repository.DeviceRepository
	aggRepo     *repository.AggregationRepository
	aggService  *AggregationService
	outboundSvc OutboundService
}

// NewTelemetryInjectionService creates a new TelemetryInjectionService
func NewTelemetryInjectionService(
	repo *repository.TelemetryInjectionRepository,
	devRepo *repository.DeviceRepository,
	aggRepo *repository.AggregationRepository,
	aggService *AggregationService,
	outboundSvc OutboundService,
) *TelemetryInjectionService {
	return &TelemetryInjectionService{
		repo:        repo,
		devRepo:     devRepo,
		aggRepo:     aggRepo,
		aggService:  aggService,
		outboundSvc: outboundSvc,
	}
}

// GetDeviceParameterHints returns parameter metadata, active aggregation intervals, and suggested [Min, Max] ranges
func (s *TelemetryInjectionService) GetDeviceParameterHints(ctx context.Context, deviceID uint) (*DeviceInjectionHintsResponse, error) {
	if deviceID == 0 {
		return nil, fmt.Errorf("device_id wajib dipilih")
	}
	dev, err := s.devRepo.GetByID(deviceID)
	if err != nil || dev == nil {
		return nil, fmt.Errorf("device #%d tidak ditemukan", deviceID)
	}

	defs, _ := s.aggRepo.ListDefinitions(ctx, deviceID, true)
	paramIntervalsMap := make(map[uint]map[int]bool)
	deviceIntervalsSet := make(map[int]bool)
	for _, d := range defs {
		if d.IntervalSeconds > 0 {
			deviceIntervalsSet[d.IntervalSeconds] = true
			if paramIntervalsMap[d.ParameterID] == nil {
				paramIntervalsMap[d.ParameterID] = make(map[int]bool)
			}
			paramIntervalsMap[d.ParameterID][d.IntervalSeconds] = true
		}
	}

	devIntervals := sortedIntKeys(deviceIntervalsSet)
	protocol := string(dev.DeviceType)
	if dev.Connection != nil && dev.Connection.Protocol != "" {
		protocol = string(dev.Connection.Protocol)
	}
	if protocol == "" {
		protocol = "MODBUS_RTU"
	}
	tz := dev.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}

	hints := make([]DeviceParameterInjectionHint, 0, len(dev.Parameters))
	for _, p := range dev.Parameters {
		if !p.Enabled {
			continue
		}
		scale := effectiveScale(&p)
		prec := effectivePrecision(&p)

		var curProc *float64
		var curRaw *float64
		if p.CurrentProcessedValue != nil && !math.IsNaN(*p.CurrentProcessedValue) && !math.IsInf(*p.CurrentProcessedValue, 0) {
			v := roundToPrecision(*p.CurrentProcessedValue, prec)
			curProc = &v
		} else if p.CurrentValue != nil && !math.IsNaN(*p.CurrentValue) && !math.IsInf(*p.CurrentValue, 0) {
			v := roundToPrecision(*p.CurrentValue, prec)
			curProc = &v
		}
		if curProc != nil {
			rv := roundToPrecision((*curProc-p.Offset)/scale, prec+2)
			curRaw = &rv
		}

		sugMin, sugMax := suggestParameterRange(&p, curProc, prec)

		hints = append(hints, DeviceParameterInjectionHint{
			ParameterID:           p.ID,
			ParameterCode:         p.ParameterCode,
			ParameterName:         p.ParameterName,
			Unit:                  p.Unit,
			DataType:              string(p.DataType),
			Precision:             prec,
			Scale:                 scale,
			Offset:                p.Offset,
			Formula:               p.Formula,
			CurrentProcessedValue: curProc,
			CurrentRawValue:       curRaw,
			CurrentQuality:        string(p.CurrentQuality),
			CurrentReceivedAt:     p.CurrentReceivedAt,
			SuggestedMin:          sugMin,
			SuggestedMax:          sugMax,
			AggregationIntervals:  sortedIntKeys(paramIntervalsMap[p.ID]),
		})
	}

	return &DeviceInjectionHintsResponse{
		DeviceID:           dev.ID,
		DeviceCode:         dev.DeviceCode,
		DeviceName:         dev.DeviceName,
		Protocol:           protocol,
		Timezone:           tz,
		ActiveAggIntervals: devIntervals,
		Parameters:         hints,
	}, nil
}

type injectionPlan struct {
	preview        *TelemetryInjectionPreviewResponse
	rawRows        []*model.RawData
	overwriteSpec  *repository.OverwriteWindowSpec
	paramIDs       []uint
	paramsSummary  string
	configJSON     string
	normalizedStep int
	conflictMode   string
	randomMode     string
	valueTarget    string
	startUTC       time.Time
	endUTC         time.Time
}

func (s *TelemetryInjectionService) buildInjectionPlan(ctx context.Context, req TelemetryInjectionRequest) (*injectionPlan, error) {
	if req.DeviceID == 0 {
		return nil, fmt.Errorf("device_id wajib diisi")
	}
	if len(req.Parameters) == 0 {
		return nil, fmt.Errorf("pilih minimal 1 parameter untuk disuntikkan")
	}
	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		return nil, fmt.Errorf("start_time dan end_time wajib diisi")
	}

	now := time.Now().UTC()
	startUTC := req.StartTime.UTC()
	endUTC := req.EndTime.UTC()

	if startUTC.After(now.Add(5 * time.Second)) {
		return nil, fmt.Errorf("waktu mulai (start_time) tidak boleh berada di masa depan")
	}
	if endUTC.After(now) {
		endUTC = now
	}
	if endUTC.Before(startUTC) {
		return nil, fmt.Errorf("waktu selesai (end_time) harus lebih besar atau sama dengan start_time")
	}
	if endUTC.Sub(startUTC) > 7*24*time.Hour {
		return nil, fmt.Errorf("rentang waktu maksimal per satu batch suntikan adalah 7 hari")
	}

	stepSec := req.SampleStepSeconds
	if stepSec <= 0 {
		stepSec = 30
	}
	if stepSec < 5 {
		stepSec = 5
	}
	if stepSec > 3600 {
		stepSec = 3600
	}

	conflictMode := strings.ToUpper(strings.TrimSpace(req.ConflictMode))
	if conflictMode != ConflictOverwriteAll {
		conflictMode = ConflictFillMissingAndNull
	}

	randomMode := strings.ToUpper(strings.TrimSpace(req.RandomMode))
	if randomMode != RandomModeUniform {
		randomMode = RandomModeSmoothWalk
	}

	valueTarget := strings.ToUpper(strings.TrimSpace(req.ValueTarget))
	if valueTarget != ValueTargetRaw {
		valueTarget = ValueTargetProcessed
	}

	dev, err := s.devRepo.GetByID(req.DeviceID)
	if err != nil || dev == nil {
		return nil, fmt.Errorf("device #%d tidak ditemukan", req.DeviceID)
	}

	paramByID := make(map[uint]model.Parameter, len(dev.Parameters))
	for _, p := range dev.Parameters {
		paramByID[p.ID] = p
	}

	var validInputs []TelemetryInjectionParamInput
	seenParams := make(map[uint]bool)
	for _, in := range req.Parameters {
		if in.ParameterID == 0 || seenParams[in.ParameterID] {
			continue
		}
		if _, ok := paramByID[in.ParameterID]; !ok {
			return nil, fmt.Errorf("parameter #%d tidak ditemukan pada device %s", in.ParameterID, dev.DeviceCode)
		}
		if math.IsNaN(in.MinValue) || math.IsInf(in.MinValue, 0) || math.IsNaN(in.MaxValue) || math.IsInf(in.MaxValue, 0) {
			return nil, fmt.Errorf("nilai min/max tidak valid pada parameter #%d", in.ParameterID)
		}
		if in.MaxValue < in.MinValue {
			in.MinValue, in.MaxValue = in.MaxValue, in.MinValue
		}
		seenParams[in.ParameterID] = true
		validInputs = append(validInputs, in)
	}
	if len(validInputs) == 0 {
		return nil, fmt.Errorf("tidak ada parameter valid yang dipilih")
	}

	// Build target timestamp slots across [startUTC, endUTC) (or single point if startUTC == endUTC or span < step)
	stepDur := time.Duration(stepSec) * time.Second
	var slots []time.Time
	if startUTC.Equal(endUTC) || endUTC.Sub(startUTC) < stepDur {
		slots = []time.Time{startUTC}
	} else {
		for t := startUTC; t.Before(endUTC); t = t.Add(stepDur) {
			slots = append(slots, t)
		}
	}

	if len(slots)*len(validInputs) > maxInjectionTotalSamples {
		return nil, fmt.Errorf("jumlah sampel terlalu besar (%d baris > batas %d). Perbesar interval sampel (misal 60s/120s) atau perkecil rentang waktu",
			len(slots)*len(validInputs), maxInjectionTotalSamples)
	}

	paramIDs := make([]uint, 0, len(validInputs))
	for _, in := range validInputs {
		paramIDs = append(paramIDs, in.ParameterID)
	}

	// Load active aggregation definitions for bucket-aware conflict detection & estimation
	defs, _ := s.aggRepo.ListDefinitions(ctx, dev.ID, true)
	minAggByParam := make(map[uint]struct {
		intervalSec int
		timezone    string
	})
	activeIntervalsSet := make(map[int]bool)
	for _, d := range defs {
		if d.IntervalSeconds <= 0 {
			continue
		}
		if seenParams[d.ParameterID] {
			activeIntervalsSet[d.IntervalSeconds] = true
			cur, exists := minAggByParam[d.ParameterID]
			if !exists || d.IntervalSeconds < cur.intervalSec {
				minAggByParam[d.ParameterID] = struct {
					intervalSec int
					timezone    string
				}{intervalSec: d.IntervalSeconds, timezone: d.Timezone}
			}
		}
	}

	// Query existing RawData in window (expanded to cover overlapping smallest aggregation buckets)
	queryStart := startUTC
	queryEnd := endUTC
	for _, m := range minAggByParam {
		bFirst := aggregation.CalculateBucket(startUTC, m.intervalSec, m.timezone)
		if bFirst.PeriodStart.Before(queryStart) {
			queryStart = bFirst.PeriodStart
		}
		lastRef := endUTC.Add(-time.Nanosecond)
		if startUTC.Equal(endUTC) {
			lastRef = endUTC
		}
		bLast := aggregation.CalculateBucket(lastRef, m.intervalSec, m.timezone)
		if bLast.PeriodEnd.After(queryEnd) {
			queryEnd = bLast.PeriodEnd
		}
	}

	existingRows, err := s.repo.GetExistingRawInWindow(ctx, dev.ID, paramIDs, queryStart, queryEnd)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa data eksisting: %w", err)
	}

	validTimesByParam := make(map[uint][]time.Time)
	validCountInWindow := make(map[uint]int)
	badCountInWindow := make(map[uint]int)
	for _, ex := range existingRows {
		inMainWindow := !ex.ReceivedAt.Before(startUTC) && (ex.ReceivedAt.Before(endUTC) || (startUTC.Equal(endUTC) && ex.ReceivedAt.Equal(endUTC)))
		isValid := ex.Quality != model.QualityBad && !math.IsNaN(ex.ProcessedValue) && !math.IsInf(ex.ProcessedValue, 0)
		if isValid {
			if pSpec, ok := paramByID[ex.ParameterID]; ok && pSpec.ZeroValueAlarmPolicy == model.ZeroValuePolicyInvalidReading && ex.ProcessedValue == 0 {
				isValid = false
			}
		}
		if isValid {
			validTimesByParam[ex.ParameterID] = append(validTimesByParam[ex.ParameterID], ex.ReceivedAt.UTC())
			if inMainWindow {
				validCountInWindow[ex.ParameterID]++
			}
		} else if inMainWindow {
			badCountInWindow[ex.ParameterID]++
		}
	}

	sourceProtocol := "MODBUS_RTU"
	if dev.Connection != nil && dev.Connection.Protocol != "" {
		sourceProtocol = string(dev.Connection.Protocol)
	} else if dev.DeviceType != "" {
		sourceProtocol = string(dev.DeviceType)
	}

	seed := req.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))
	seqBase := uint64(time.Now().UnixNano() / 1000)

	var rawRows []*model.RawData
	var summaries []InjectionParameterSummary
	var previews []InjectionSamplePreview
	var summaryLabels []string

	totalPlanned := 0
	totalSkipped := 0

	for pIdx, in := range validInputs {
		param := paramByID[in.ParameterID]
		scale := effectiveScale(&param)
		prec := effectivePrecision(&param)
		byteOrder := param.ByteOrder
		if byteOrder == "" && dev.Connection != nil {
			byteOrder = dev.Connection.ByteOrder
		}

		validTimes := validTimesByParam[param.ID]
		aggInfo, hasAgg := minAggByParam[param.ID]

		// Initialize smooth walk state near midpoint of [MinValue, MaxValue]
		span := in.MaxValue - in.MinValue
		walkCur := in.MinValue + span*(0.35+0.30*rng.Float64())
		if span == 0 {
			walkCur = in.MinValue
		}

		plannedForParam := 0
		skippedForParam := 0
		var minProc, maxProc, sumProc float64
		var minRaw, maxRaw float64

		for sIdx, ts := range slots {
			if conflictMode == ConflictFillMissingAndNull {
				var winStart, winEnd time.Time
				if hasAgg {
					b := aggregation.CalculateBucket(ts, aggInfo.intervalSec, aggInfo.timezone)
					winStart = b.PeriodStart
					winEnd = b.PeriodEnd
				} else {
					half := stepDur / 2
					if half < time.Second {
						half = time.Second
					}
					winStart = ts.Add(-half)
					winEnd = ts.Add(half)
				}
				if hasTimestampInHalfOpenWindow(validTimes, winStart, winEnd) {
					skippedForParam++
					continue
				}
			}

			// Generate value according to RandomMode
			var chosen float64
			if span <= 0 {
				chosen = in.MinValue
			} else if randomMode == RandomModeUniform {
				chosen = in.MinValue + rng.Float64()*span
			} else {
				// SMOOTH_WALK: natural bounded sensor drift with gentle mean-reversion toward midpoint
				mid := (in.MinValue + in.MaxValue) / 2.0
				maxStep := span * 0.16
				drift := (rng.Float64()*2.0 - 1.0)*maxStep + (mid-walkCur)*0.12
				walkCur += drift
				if walkCur < in.MinValue {
					walkCur = in.MinValue + math.Abs(in.MinValue-walkCur)*0.35
					if walkCur > in.MaxValue {
						walkCur = in.MinValue
					}
				} else if walkCur > in.MaxValue {
					walkCur = in.MaxValue - math.Abs(walkCur-in.MaxValue)*0.35
					if walkCur < in.MinValue {
						walkCur = in.MaxValue
					}
				}
				chosen = walkCur
			}

			rawVal, engVal, procVal, formulaVal := computeRawAndProcessedValues(&param, chosen, valueTarget, scale, prec)
			rawBytes, rawHex := encodeModbusWireBytes(rawVal, param.DataType, byteOrder)

			seq := seqBase + uint64(pIdx*100000+sIdx+1)
			tsCopy := ts
			valNum := engVal
			rec := &model.RawData{
				RecordUUID:      fmt.Sprintf("REC-%d-%d-%d-%d", dev.ID, param.ID, seq, ts.UnixNano()),
				DeviceID:        dev.ID,
				ParameterID:     param.ID,
				Value:           engVal,
				ProcessedValue:  procVal,
				ValueNumeric:    &valNum,
				FormulaValue:    formulaVal,
				IsHeldValue:     false,
				RawValue:        rawVal,
				RawHex:          rawHex,
				RawBytes:        rawBytes,
				Quality:         model.QualityGood,
				QualityReason:   model.ReasonNone,
				Source:          sourceProtocol,
				Sequence:        seq,
				DeviceTimestamp: &tsCopy,
				ReceivedAt:      ts,
				ProcessedAt:     &tsCopy,
				StoredAt:        now,
				Timestamp:       ts,
			}
			rawRows = append(rawRows, rec)

			if plannedForParam == 0 {
				minProc, maxProc = procVal, procVal
				minRaw, maxRaw = rawVal, rawVal
			} else {
				if procVal < minProc {
					minProc = procVal
				}
				if procVal > maxProc {
					maxProc = procVal
				}
				if rawVal < minRaw {
					minRaw = rawVal
				}
				if rawVal > maxRaw {
					maxRaw = rawVal
				}
			}
			sumProc += procVal
			plannedForParam++

			if len(previews) < maxPreviewSamples {
				previews = append(previews, InjectionSamplePreview{
					ParameterID:    param.ID,
					ParameterCode:  param.ParameterCode,
					ParameterName:  param.ParameterName,
					Unit:           param.Unit,
					ReceivedAt:     ts,
					RawValue:       rawVal,
					ProcessedValue: procVal,
					RawHex:         rawHex,
				})
			}
		}

		avgProc := 0.0
		if plannedForParam > 0 {
			avgProc = roundToPrecision(sumProc/float64(plannedForParam), prec)
		}

		summaries = append(summaries, InjectionParameterSummary{
			ParameterID:           param.ID,
			ParameterCode:         param.ParameterCode,
			ParameterName:         param.ParameterName,
			Unit:                  param.Unit,
			MinInput:              in.MinValue,
			MaxInput:              in.MaxValue,
			PlannedSamples:        plannedForParam,
			SkippedSamples:        skippedForParam,
			ExistingValidInWindow: validCountInWindow[param.ID],
			ExistingBadInWindow:   badCountInWindow[param.ID],
			GeneratedMinProcessed: minProc,
			GeneratedMaxProcessed: maxProc,
			GeneratedAvgProcessed: avgProc,
			GeneratedMinRaw:       minRaw,
			GeneratedMaxRaw:       maxRaw,
		})

		summaryLabels = append(summaryLabels, fmt.Sprintf("%s[%.2f..%.2f]", param.ParameterCode, in.MinValue, in.MaxValue))
		totalPlanned += plannedForParam
		totalSkipped += skippedForParam
	}

	// Estimate affected aggregation buckets
	estimatedBuckets := 0
	for _, d := range defs {
		if !seenParams[d.ParameterID] || d.IntervalSeconds <= 0 {
			continue
		}
		spanSec := int(endUTC.Sub(startUTC).Seconds())
		if spanSec <= 0 {
			estimatedBuckets++
		} else {
			n := (spanSec + d.IntervalSeconds - 1) / d.IntervalSeconds
			if n < 1 {
				n = 1
			}
			estimatedBuckets += n
		}
	}

	var overwriteSpec *repository.OverwriteWindowSpec
	if conflictMode == ConflictOverwriteAll {
		overwriteSpec = &repository.OverwriteWindowSpec{
			DeviceID:     dev.ID,
			ParameterIDs: paramIDs,
			StartTime:    startUTC,
			EndTime:      endUTC,
			InclusiveEnd: startUTC.Equal(endUTC) || endUTC.Sub(startUTC) < stepDur,
		}
	}

	cfgBytes, _ := json.Marshal(validInputs)
	paramsSummaryStr := strings.Join(summaryLabels, ", ")
	if len(paramsSummaryStr) > 500 {
		paramsSummaryStr = paramsSummaryStr[:497] + "..."
	}

	previewResp := &TelemetryInjectionPreviewResponse{
		DeviceID:            dev.ID,
		DeviceCode:          dev.DeviceCode,
		DeviceName:          dev.DeviceName,
		StartTime:           startUTC,
		EndTime:             endUTC,
		SampleStepSeconds:   stepSec,
		ConflictMode:        conflictMode,
		RandomMode:          randomMode,
		ValueTarget:         valueTarget,
		TotalCandidateSlots: len(slots) * len(validInputs),
		TotalPlannedSamples: totalPlanned,
		TotalSkippedSamples: totalSkipped,
		ActiveAggIntervals:  sortedIntKeys(activeIntervalsSet),
		EstimatedAggBuckets: estimatedBuckets,
		ParameterSummaries:  summaries,
		SamplePreview:       previews,
	}

	return &injectionPlan{
		preview:        previewResp,
		rawRows:        rawRows,
		overwriteSpec:  overwriteSpec,
		paramIDs:       paramIDs,
		paramsSummary:  paramsSummaryStr,
		configJSON:     string(cfgBytes),
		normalizedStep: stepSec,
		conflictMode:   conflictMode,
		randomMode:     randomMode,
		valueTarget:    valueTarget,
		startUTC:       startUTC,
		endUTC:         endUTC,
	}, nil
}

// PreviewInjection generates a dry-run simulation without persisting any database changes
func (s *TelemetryInjectionService) PreviewInjection(ctx context.Context, req TelemetryInjectionRequest) (*TelemetryInjectionPreviewResponse, error) {
	plan, err := s.buildInjectionPlan(ctx, req)
	if err != nil {
		return nil, err
	}
	return plan.preview, nil
}

// ExecuteInjection persists injected RawData & Superadmin tracking records, recalculates all device aggregations
// across [StartTime, EndTime], and re-hydrates overlapping OutboundDeliveryQueue items.
func (s *TelemetryInjectionService) ExecuteInjection(
	ctx context.Context,
	req TelemetryInjectionRequest,
	username string,
) (*TelemetryInjectionExecuteResponse, error) {
	plan, err := s.buildInjectionPlan(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(plan.rawRows) == 0 && plan.overwriteSpec == nil {
		return nil, fmt.Errorf("seluruh slot pada rentang waktu ini sudah memiliki data valid (%d sampel dilewati). Gunakan mode OVERWRITE_ALL jika ingin menimpa data eksisting", plan.preview.TotalSkippedSamples)
	}

	if strings.TrimSpace(username) == "" {
		username = "superadmin"
	}
	batchCode := fmt.Sprintf("INJ-%s-%04d", time.Now().UTC().Format("20060102-150405"), time.Now().UnixNano()%10000)

	batch := &model.TelemetryInjectionBatch{
		BatchCode:            batchCode,
		DeviceID:             plan.preview.DeviceID,
		DeviceCode:           plan.preview.DeviceCode,
		DeviceName:           plan.preview.DeviceName,
		StartTime:            plan.startUTC,
		EndTime:              plan.endUTC,
		SampleStepSeconds:    plan.normalizedStep,
		ConflictMode:         plan.conflictMode,
		RandomMode:           plan.randomMode,
		ValueTarget:          plan.valueTarget,
		ParametersSummary:    plan.paramsSummary,
		ParametersConfigJSON: plan.configJSON,
		SkippedCount:         plan.preview.TotalSkippedSamples,
		Status:               "ACTIVE",
		CreatedBy:            username,
		Note:                 strings.TrimSpace(req.Note),
	}

	if _, err := s.repo.PersistInjectionBatch(ctx, batch, plan.rawRows, plan.overwriteSpec); err != nil {
		return nil, err
	}

	aggBuckets := 0
	var completedEnds []time.Time
	if s.aggService != nil {
		aggBuckets, completedEnds, _ = s.aggService.RecalculateRangeForDevice(
			ctx,
			batch.DeviceID,
			plan.paramIDs,
			plan.startUTC,
			plan.endUTC,
			false,
		)
	}

	refreshOutbound := true
	if req.RefreshOutbound != nil {
		refreshOutbound = *req.RefreshOutbound
	}
	refreshedOutbound := 0
	if refreshOutbound && s.outboundSvc != nil {
		refreshedOutbound, _ = s.outboundSvc.RefreshQueueForWindow(
			ctx,
			batch.DeviceID,
			plan.startUTC,
			plan.endUTC,
			completedEnds,
			req.EnqueueMissingOutbound,
		)
	}

	batch.AggregatedBuckets = aggBuckets
	batch.RefreshedOutbound = refreshedOutbound
	_ = s.repo.UpdateBatchSummary(ctx, batch.ID, aggBuckets, refreshedOutbound)

	return &TelemetryInjectionExecuteResponse{
		Batch:   batch,
		Preview: plan.preview,
	}, nil
}

// ListBatches returns paginated Superadmin injection batches
func (s *TelemetryInjectionService) ListBatches(ctx context.Context, deviceID uint, page, pageSize int) ([]model.TelemetryInjectionBatch, int64, error) {
	return s.repo.ListBatches(ctx, deviceID, page, pageSize)
}

// RollbackBatch deletes all RawData created by batchID, cleans up empty aggregation buckets,
// recalculates remaining aggregations, and refreshes affected outbound queue items.
func (s *TelemetryInjectionService) RollbackBatch(
	ctx context.Context,
	batchID uint,
	username string,
) (*TelemetryInjectionRollbackResponse, error) {
	if batchID == 0 {
		return nil, fmt.Errorf("batch_id tidak valid")
	}
	if strings.TrimSpace(username) == "" {
		username = "superadmin"
	}

	batch, paramIDs, deletedRaw, err := s.repo.RollbackBatch(ctx, batchID, username)
	if err != nil {
		return nil, err
	}

	recalculated := 0
	if s.aggService != nil {
		recalculated, _, _ = s.aggService.RecalculateRangeForDevice(
			ctx,
			batch.DeviceID,
			paramIDs,
			batch.StartTime,
			batch.EndTime,
			true, // cleanupEmptyBuckets = true removes synthetic-only buckets cleanly
		)
	}

	refreshedOutbound := 0
	if s.outboundSvc != nil {
		refreshedOutbound, _ = s.outboundSvc.RefreshQueueForWindow(
			ctx,
			batch.DeviceID,
			batch.StartTime,
			batch.EndTime,
			nil,
			false,
		)
	}

	return &TelemetryInjectionRollbackResponse{
		Batch:               batch,
		DeletedRawCount:     deletedRaw,
		RecalculatedBuckets: recalculated,
		RefreshedOutbound:   refreshedOutbound,
	}, nil
}

// AnnotateRawDataForSuperadmin populates transient IsInjected and InjectionBatchID strictly for SUPERADMIN responses
func (s *TelemetryInjectionService) AnnotateRawDataForSuperadmin(ctx context.Context, records []model.RawData) {
	if s == nil || s.repo == nil || len(records) == 0 {
		return
	}
	ids := make([]uint64, 0, len(records))
	for i := range records {
		if records[i].ID > 0 {
			ids = append(ids, records[i].ID)
		}
	}
	if len(ids) == 0 {
		return
	}
	injMap, err := s.repo.FindInjectedRawMap(ctx, ids)
	if err != nil || len(injMap) == 0 {
		return
	}
	for i := range records {
		if batchCode, ok := injMap[records[i].ID]; ok {
			records[i].IsInjected = true
			records[i].InjectionBatchID = batchCode
		}
	}
}

// AnnotateBucketSamplesForSuperadmin populates transient IsInjected and InjectionBatchID on bucket sample DTOs strictly for SUPERADMIN
func (s *TelemetryInjectionService) AnnotateBucketSamplesForSuperadmin(ctx context.Context, samples []BucketSampleDTO) {
	if s == nil || s.repo == nil || len(samples) == 0 {
		return
	}
	ids := make([]uint64, 0, len(samples))
	for i := range samples {
		if samples[i].ID > 0 {
			ids = append(ids, samples[i].ID)
		}
	}
	if len(ids) == 0 {
		return
	}
	injMap, err := s.repo.FindInjectedRawMap(ctx, ids)
	if err != nil || len(injMap) == 0 {
		return
	}
	for i := range samples {
		if batchCode, ok := injMap[samples[i].ID]; ok {
			samples[i].IsInjected = true
			samples[i].InjectionBatchID = batchCode
		}
	}
}

// AnnotateAggregationsForSuperadmin populates transient HasInjectedSamples on AggregationResult slices strictly for SUPERADMIN
func (s *TelemetryInjectionService) AnnotateAggregationsForSuperadmin(ctx context.Context, results []model.AggregationResult) {
	if s == nil || s.repo == nil || len(results) == 0 {
		return
	}

	devSet := make(map[uint]bool)
	paramSet := make(map[uint]bool)
	var minStart, maxEnd time.Time

	for i, r := range results {
		devSet[r.DeviceID] = true
		paramSet[r.ParameterID] = true
		if i == 0 || r.PeriodStart.Before(minStart) {
			minStart = r.PeriodStart
		}
		if i == 0 || r.PeriodEnd.After(maxEnd) {
			maxEnd = r.PeriodEnd
		}
	}

	devIDs := make([]uint, 0, len(devSet))
	for d := range devSet {
		devIDs = append(devIDs, d)
	}
	paramIDs := make([]uint, 0, len(paramSet))
	for p := range paramSet {
		paramIDs = append(paramIDs, p)
	}

	grouped, err := s.repo.FindInjectedTimestampsByDeviceParam(ctx, devIDs, paramIDs, minStart, maxEnd)
	if err != nil || len(grouped) == 0 {
		return
	}

	for i := range results {
		key := [2]uint{results[i].DeviceID, results[i].ParameterID}
		tsSlice := grouped[key]
		if len(tsSlice) > 0 && hasTimestampInHalfOpenWindow(tsSlice, results[i].PeriodStart, results[i].PeriodEnd) {
			results[i].HasInjectedSamples = true
		}
	}
}

// --- Internal Helpers ---

func effectiveScale(p *model.Parameter) float64 {
	if p == nil {
		return 1.0
	}
	if p.Scale != 0 {
		return p.Scale
	}
	if p.ScaleFactor != 0 {
		return p.ScaleFactor
	}
	return 1.0
}

func effectivePrecision(p *model.Parameter) int {
	if p == nil {
		return 2
	}
	if p.Precision >= 0 && p.Precision <= 6 {
		return p.Precision
	}
	return 2
}

func roundToPrecision(val float64, prec int) float64 {
	if prec < 0 {
		prec = 2
	}
	if prec > 8 {
		prec = 8
	}
	pow := math.Pow10(prec)
	return math.Round(val*pow) / pow
}

func suggestParameterRange(p *model.Parameter, curProc *float64, prec int) (float64, float64) {
	if curProc != nil {
		base := *curProc
		delta := math.Abs(base) * 0.03
		if delta == 0 {
			delta = 0.5
		}
		minV := roundToPrecision(base-delta, prec)
		maxV := roundToPrecision(base+delta, prec)
		if p != nil && p.MinValue != nil && minV < *p.MinValue {
			minV = *p.MinValue
		}
		if p != nil && p.MaxValue != nil && maxV > *p.MaxValue {
			maxV = *p.MaxValue
		}
		if maxV < minV {
			maxV = minV
		}
		return minV, maxV
	}

	code := ""
	if p != nil {
		code = strings.ToLower(strings.TrimSpace(p.ParameterCode))
	}
	switch {
	case strings.Contains(code, "ph"):
		return 6.85, 7.15
	case strings.Contains(code, "cod"):
		return 45.0, 55.0
	case strings.Contains(code, "bod"):
		return 18.0, 24.0
	case strings.Contains(code, "tss"):
		return 15.0, 25.0
	case strings.Contains(code, "nh3") || strings.Contains(code, "nh4") || strings.Contains(code, "amonia"):
		return 0.5, 1.5
	case strings.Contains(code, "debit") || strings.Contains(code, "flow"):
		return 0.80, 0.95
	case strings.Contains(code, "do"):
		return 5.2, 6.4
	case strings.Contains(code, "temp") || strings.Contains(code, "suhu"):
		return 27.5, 29.5
	default:
		return 10.0, 20.0
	}
}

func isIntegerDataType(dt model.ParameterDataType) bool {
	switch dt {
	case model.DataTypeInt16, model.DataTypeUInt16, model.DataTypeInt32, model.DataTypeUInt32, model.DataTypeBoolean:
		return true
	default:
		return false
	}
}

func computeRawAndProcessedValues(
	param *model.Parameter,
	chosen float64,
	valueTarget string,
	scale float64,
	prec int,
) (rawVal float64, engVal float64, procVal float64, formulaVal *float64) {
	if scale == 0 {
		scale = 1.0
	}

	if valueTarget == ValueTargetRaw {
		if isIntegerDataType(param.DataType) {
			rawVal = math.Round(chosen)
		} else {
			rawVal = roundToPrecision(chosen, prec+2)
		}
		engVal = modbus.ApplyScaleAndOffset(rawVal, scale, param.Offset, prec)
		procVal = engVal
		if trimmed := strings.TrimSpace(param.Formula); trimmed != "" {
			vars := map[string]float64{
				"x":     engVal,
				"val":   engVal,
				"value": engVal,
				"raw":   rawVal,
			}
			if fRes, err := formula.Evaluate(trimmed, vars); err == nil && !math.IsNaN(fRes) && !math.IsInf(fRes, 0) {
				roundedF := roundToPrecision(fRes, prec)
				formulaVal = &roundedF
				procVal = roundedF
			}
		}
		return rawVal, engVal, procVal, formulaVal
	}

	// Default: ValueTargetProcessed — user specifies desired ProcessedValue, back-calculate RawValue
	procVal = roundToPrecision(chosen, prec)
	engVal = procVal
	invRaw := (procVal - param.Offset) / scale
	if isIntegerDataType(param.DataType) {
		rawVal = math.Round(invRaw)
	} else {
		rawVal = roundToPrecision(invRaw, prec+3)
	}
	if strings.TrimSpace(param.Formula) != "" {
		fCopy := procVal
		formulaVal = &fCopy
	}
	return rawVal, engVal, procVal, formulaVal
}

func encodeModbusWireBytes(rawVal float64, dataType model.ParameterDataType, byteOrder string) ([]byte, string) {
	var beBytes []byte
	switch dataType {
	case model.DataTypeInt16:
		v := int16(math.Round(clampFloat(rawVal, -32768, 32767)))
		beBytes = make([]byte, 2)
		binary.BigEndian.PutUint16(beBytes, uint16(v))
	case model.DataTypeUInt16:
		v := uint16(math.Round(clampFloat(rawVal, 0, 65535)))
		beBytes = make([]byte, 2)
		binary.BigEndian.PutUint16(beBytes, v)
	case model.DataTypeBoolean:
		beBytes = make([]byte, 2)
		if rawVal >= 0.5 {
			binary.BigEndian.PutUint16(beBytes, 1)
		}
	case model.DataTypeInt32:
		v := int32(math.Round(clampFloat(rawVal, -2147483648, 2147483647)))
		beBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(beBytes, uint32(v))
	case model.DataTypeUInt32:
		v := uint32(math.Round(clampFloat(rawVal, 0, 4294967295)))
		beBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(beBytes, v)
	case model.DataTypeFloat64:
		beBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(beBytes, math.Float64bits(rawVal))
	default: // FLOAT32
		beBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(beBytes, math.Float32bits(float32(rawVal)))
	}

	wireBytes := make([]byte, len(beBytes))
	copy(wireBytes, beBytes)

	if parsed, err := modbus.ParseAndValidateByteOrder(byteOrder, dataType); err == nil && len(parsed.Permutation) == len(beBytes) {
		// DecodeRegisters computes beBytes[i] = wireBytes[perm[i]], so wireBytes[perm[i]] = beBytes[i] is the exact inverse
		inv := make([]byte, len(beBytes))
		for i, srcIdx := range parsed.Permutation {
			if srcIdx >= 0 && srcIdx < len(inv) {
				inv[srcIdx] = beBytes[i]
			}
		}
		wireBytes = inv
	}

	hexUpper := strings.ToUpper(hex.EncodeToString(wireBytes))
	var sb strings.Builder
	for i := 0; i < len(hexUpper); i += 2 {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if i+2 <= len(hexUpper) {
			sb.WriteString(hexUpper[i : i+2])
		} else {
			sb.WriteString(hexUpper[i:])
		}
	}
	return wireBytes, sb.String()
}

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func hasTimestampInHalfOpenWindow(sortedUTC []time.Time, start, end time.Time) bool {
	if len(sortedUTC) == 0 {
		return false
	}
	sUTC := start.UTC()
	eUTC := end.UTC()
	idx := sort.Search(len(sortedUTC), func(i int) bool {
		return !sortedUTC[i].Before(sUTC)
	})
	if idx >= len(sortedUTC) {
		return false
	}
	if sUTC.Equal(eUTC) {
		return sortedUTC[idx].Equal(sUTC)
	}
	return sortedUTC[idx].Before(eUTC)
}

func sortedIntKeys(m map[int]bool) []int {
	if len(m) == 0 {
		return []int{}
	}
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

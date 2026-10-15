package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"

	"gorm.io/gorm"
)

type ispuTriggerJob struct {
	DeviceID  uint
	BucketEnd time.Time
	Reason    string
}

// ISPUService coordinates regulatory ISPU calculations, data window validation, MariaDB persistence, and exports.
type ISPUService struct {
	db              *gorm.DB
	ispuRepo        *repository.ISPURepository
	systemRepo      *repository.SystemRepository
	calculator      *ISPUCalculator
	hub             *websocket.Hub
	jobChan         chan ispuTriggerJob
	jobMu           sync.Mutex
	pendingJobs     map[string]ispuTriggerJob
	stopWorker      chan struct{}
	workerWg        sync.WaitGroup
	isWorkerRunning bool
}

// NewISPUService instantiates an ISPUService.
func NewISPUService(db *gorm.DB, ispuRepo *repository.ISPURepository, systemRepo *repository.SystemRepository) *ISPUService {
	return &ISPUService{
		db:          db,
		ispuRepo:    ispuRepo,
		systemRepo:  systemRepo,
		calculator:  NewISPUCalculator(),
		jobChan:     make(chan ispuTriggerJob, 256),
		pendingJobs: make(map[string]ispuTriggerJob),
		stopWorker:  make(chan struct{}),
	}
}

// SetHub updates or sets the WebSocket hub for real-time calculation broadcasting.
func (s *ISPUService) SetHub(hub *websocket.Hub) {
	s.hub = hub
}

// AlignToAggregationInterval snaps a timestamp to the exact completion boundary of the station's aggregation setup.
// If aggregation is set to 2m (120s), boundaries will be 15:36:00, 15:38:00, 15:40:00, 15:42:00 (strictly even minutes).
func (s *ISPUService) AlignToAggregationInterval(t time.Time, intervalSec int, loc *time.Location) time.Time {
	if intervalSec <= 0 {
		intervalSec = 120 // Default 2 minutes
	}
	localT := t.In(loc)

	// If timestamp ends at :59 (closed interval end like 15:39:59), add 1 second to snap to boundary (15:40:00)
	if localT.Second() == 59 {
		localT = localT.Add(1 * time.Second)
	}

	if 86400%intervalSec == 0 {
		dayStart := time.Date(localT.Year(), localT.Month(), localT.Day(), 0, 0, 0, 0, loc)
		secIntoDay := int(localT.Sub(dayStart).Seconds())
		bucketIndex := secIntoDay / intervalSec
		return dayStart.Add(time.Duration(bucketIndex*intervalSec) * time.Second)
	}

	bucketIndex := localT.Unix() / int64(intervalSec)
	return time.Unix(bucketIndex*int64(intervalSec), 0).In(loc)
}

// StartWorker starts the bounded background calculation worker.
func (s *ISPUService) StartWorker(ctx context.Context) {
	s.jobMu.Lock()
	defer s.jobMu.Unlock()
	if s.isWorkerRunning {
		return
	}
	s.isWorkerRunning = true
	s.workerWg.Add(1)
	go s.workerLoop(ctx)
	logger.Info("ISPU calculation worker initialized with bounded queue")
}

// StopWorker gracefully stops the calculation worker.
func (s *ISPUService) StopWorker() {
	s.jobMu.Lock()
	if !s.isWorkerRunning {
		s.jobMu.Unlock()
		return
	}
	s.isWorkerRunning = false
	close(s.stopWorker)
	s.jobMu.Unlock()
	s.workerWg.Wait()
	logger.Info("ISPU calculation worker stopped cleanly")
}

// EnqueueCalculation enqueues an asynchronous calculation job with debouncing and coalescing.
func (s *ISPUService) EnqueueCalculation(deviceID uint, bucketEnd time.Time, reason string) {
	if deviceID == 0 || bucketEnd.IsZero() {
		return
	}
	if reason == "" {
		reason = "AGGREGATION_COMPLETED"
	}

	job := ispuTriggerJob{
		DeviceID:  deviceID,
		BucketEnd: bucketEnd,
		Reason:    reason,
	}

	select {
	case s.jobChan <- job:
	default:
		logger.Warn("ISPU calculation job queue is full, processing immediately in background for station %d bucket %v", deviceID, bucketEnd)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			_, _ = s.CalculateStationWindowWithReason(ctx, deviceID, bucketEnd, reason)
		}()
	}
}

// HandleAggregationCompleted evaluates whether a completed aggregation bucket is relevant to an enabled ISPU mapping,
// and if so, enqueues the station window calculation.
func (s *ISPUService) HandleAggregationCompleted(ctx context.Context, event AggregationCompletedEvent) {
	if event.DeviceID == 0 || event.ParameterID == 0 {
		return
	}

	mappings, err := s.ispuRepo.GetMappingsByDeviceID(ctx, event.DeviceID)
	if err != nil {
		logger.Warn("Failed to lookup ISPU mappings for station %d: %v", event.DeviceID, err)
		return
	}

	isMapped := false
	for _, m := range mappings {
		if m.Enabled && m.ParameterID == event.ParameterID {
			isMapped = true
			break
		}
	}

	if !isMapped {
		return // Not an ISPU pollutant (e.g. ambient temperature, humidity, wind)
	}

	s.EnqueueCalculation(event.DeviceID, event.PeriodEnd, "AGGREGATION_COMPLETED")
}

// workerLoop coordinates job debouncing/coalescing and invokes calculation.
func (s *ISPUService) workerLoop(ctx context.Context) {
	defer s.workerWg.Done()
	logger.Info("ISPU background calculation loop started")

	coalesceDelay := 300 * time.Millisecond
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	scheduledJobs := make(map[string]struct {
		job      ispuTriggerJob
		deadline time.Time
	})

	for {
		select {
		case <-s.stopWorker:
			return
		case <-ctx.Done():
			return
		case job, ok := <-s.jobChan:
			if !ok {
				return
			}
			key := fmt.Sprintf("%d_%d", job.DeviceID, job.BucketEnd.Unix())
			s.jobMu.Lock()
			existing, exists := scheduledJobs[key]
			if exists {
				if job.Reason == "MANUAL_RECALCULATE" || job.Reason == "LATE_DATA" {
					existing.job.Reason = job.Reason
				}
				scheduledJobs[key] = existing
			} else {
				scheduledJobs[key] = struct {
					job      ispuTriggerJob
					deadline time.Time
				}{
					job:      job,
					deadline: time.Now().Add(coalesceDelay),
				}
			}
			s.jobMu.Unlock()

		case <-ticker.C:
			now := time.Now()
			var dueJobs []ispuTriggerJob

			s.jobMu.Lock()
			for k, v := range scheduledJobs {
				if now.After(v.deadline) {
					dueJobs = append(dueJobs, v.job)
					delete(scheduledJobs, k)
				}
			}
			s.jobMu.Unlock()

			for _, job := range dueJobs {
				calcCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				_, err := s.CalculateStationWindowWithReason(calcCtx, job.DeviceID, job.BucketEnd, job.Reason)
				if err != nil {
					logger.Warn("Failed to execute ISPU calculation for station %d bucket %v: %v", job.DeviceID, job.BucketEnd, err)
				}
				cancel()
			}
		}
	}
}

// ReconcilePendingCalculations reconciles completed aggregation periods on startup or restart,
// ensuring any completed aggregation bucket that lacks an ISPU record is calculated.
func (s *ISPUService) ReconcilePendingCalculations(ctx context.Context) (int, error) {
	mappings, err := s.ispuRepo.GetAllActiveMappings(ctx)
	if err != nil {
		return 0, err
	}
	if len(mappings) == 0 {
		return 0, nil
	}

	deviceMap := make(map[uint]bool)
	for _, m := range mappings {
		deviceMap[m.DeviceID] = true
	}

	totalReconciled := 0
	for deviceID := range deviceMap {
		var dev model.Device
		if err := s.db.WithContext(ctx).First(&dev, deviceID).Error; err != nil {
			continue
		}
		tzName := dev.Timezone
		if tzName == "" {
			tzName = "Asia/Jakarta"
		}
		loc, err := time.LoadLocation(tzName)
		if err != nil {
			loc = time.FixedZone("WIB", 7*3600)
		}

		var aggDef model.AggregationDefinition
		intervalSec := 120
		if err := s.db.WithContext(ctx).
			Where("device_id = ? AND enabled = ?", deviceID, true).
			Order("interval_seconds ASC").
			First(&aggDef).Error; err == nil && aggDef.IntervalSeconds > 0 {
			intervalSec = aggDef.IntervalSeconds
		}

		twoHoursAgo := time.Now().UTC().Add(-2 * time.Hour)
		var completedPeriods []time.Time
		s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Distinct("period_end").
			Where("device_id = ? AND period_end >= ? AND quality != ?", deviceID, twoHoursAgo, model.QualityBad).
			Order("period_end ASC").
			Pluck("period_end", &completedPeriods)

		for _, pEnd := range completedPeriods {
			alignedEnd := s.AlignToAggregationInterval(pEnd, intervalSec, loc)

			var count int64
			s.db.WithContext(ctx).Model(&model.ISPUCalculationResult{}).
				Where("device_id = ? AND period_end = ?", deviceID, alignedEnd.UTC()).
				Count(&count)

			if count < 2 {
				logger.Info("Reconciling missing ISPU calculation for station %d period_end %v", deviceID, alignedEnd)
				_, err := s.CalculateStationWindowWithReason(ctx, deviceID, alignedEnd, "STARTUP_RECOVERY")
				if err == nil {
					totalReconciled++
				}
			}
		}
	}

	logger.Info("ISPU restart recovery reconciliation completed: %d missing calculation periods recovered", totalReconciled)
	return totalReconciled, nil
}

// CalculateStationWindow performs a 24-hour regulatory ISPU evaluation for a station and persists the result.
func (s *ISPUService) CalculateStationWindow(ctx context.Context, deviceID uint, targetEnd time.Time) (*model.ISPUCalculationResult, error) {
	return s.CalculateStationWindowWithReason(ctx, deviceID, targetEnd, "AGGREGATION_COMPLETED")
}

// CalculateStationWindowWithReason performs a 24-hour regulatory ISPU evaluation with explicit reason and persists the result.
func (s *ISPUService) CalculateStationWindowWithReason(ctx context.Context, deviceID uint, targetEnd time.Time, reason string) (*model.ISPUCalculationResult, error) {
	var dev model.Device
	if err := s.db.WithContext(ctx).First(&dev, deviceID).Error; err != nil {
		return nil, fmt.Errorf("station/device %d not found: %w", deviceID, err)
	}

	tzName := dev.Timezone
	if tzName == "" {
		tzName = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	// 1. Determine active aggregation definitions and interval for this device
	var intervalSec int = 120
	aggFunction := "AVG"
	var aggDefs []model.AggregationDefinition
	if err := s.db.WithContext(ctx).
		Where("device_id = ? AND enabled = ?", deviceID, true).
		Order("interval_seconds ASC").
		Find(&aggDefs).Error; err == nil && len(aggDefs) > 0 {
		intervalSec = aggDefs[0].IntervalSeconds
		if aggDefs[0].Function != "" {
			aggFunction = string(aggDefs[0].Function)
		}
	}
	if intervalSec <= 0 {
		intervalSec = 120
	}

	// 2. Anchor targetEnd in station local time
	refTime := targetEnd.In(loc)
	if targetEnd.IsZero() {
		var latestAgg model.AggregationResult
		if err := s.db.WithContext(ctx).
			Where("device_id = ?", deviceID).
			Order("period_end DESC").
			First(&latestAgg).Error; err == nil && !latestAgg.PeriodEnd.IsZero() {
			refTime = latestAgg.PeriodEnd.In(loc)
		} else {
			refTime = time.Now().In(loc)
		}
	}

	// 3. Align periodEnd to the exact aggregation interval boundary
	periodEnd := s.AlignToAggregationInterval(refTime, intervalSec, loc)
	regPeriodStart := periodEnd.Add(-24 * time.Hour)
	opPeriodStart := periodEnd.Add(-time.Duration(intervalSec) * time.Second)
	sourceBucketIdent := periodEnd.In(loc).Format("20060102150405")
	if reason == "" {
		reason = "AGGREGATION_COMPLETED"
	}

	// Fetch active parameter mappings for station
	mappings, err := s.ispuRepo.GetMappingsByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parameter mappings: %w", err)
	}

	var activeMappings []model.ISPUParameterMapping
	for _, m := range mappings {
		if m.Enabled {
			activeMappings = append(activeMappings, m)
		}
	}

	if len(activeMappings) == 0 {
		return nil, fmt.Errorf("no active ISPU parameter mappings configured for station %s", dev.DeviceCode)
	}

	var opParamOutputs []ParameterCalculationOutput
	var regParamOutputs []ParameterCalculationOutput

	for _, m := range activeMappings {
		// Look up specific definition for this parameter if available
		paramInterval := intervalSec
		paramFunc := aggFunction
		hasParamDef := false
		for _, ad := range aggDefs {
			if ad.ParameterID == m.ParameterID {
				paramInterval = ad.IntervalSeconds
				paramFunc = string(ad.Function)
				hasParamDef = true
				break
			}
		}

		// --- Mode A: Operational (Aggregation-driven for latest bucket) ---
		var opAgg model.AggregationResult
		hasOpAgg := false
		alignBuffer := time.Duration(intervalSec) * time.Second
		if err := s.db.WithContext(ctx).
			Where("device_id = ? AND parameter_id = ? AND period_end >= ? AND period_end <= ? AND quality != ?",
				deviceID, m.ParameterID, opPeriodStart.Add(-alignBuffer).UTC(), periodEnd.Add(alignBuffer).UTC(), model.QualityBad).
			Order("period_end DESC").
			First(&opAgg).Error; err == nil && opAgg.ID > 0 {
			hasOpAgg = true
		}

		var opOut ParameterCalculationOutput
		if hasOpAgg {
			val := 0.0
			if opAgg.AvgValue != nil {
				val = *opAgg.AvgValue
			} else if opAgg.Value != nil {
				val = *opAgg.Value
			}
			opOut = s.calculator.CalculateParameter(ParameterCalculationInput{
				PollutantCode:    m.PollutantCode,
				ParameterID:      m.ParameterID,
				SourceUnit:       m.SourceUnit,
				RawConcentration: val,
				ConversionFactor: m.ConversionFactor,
				ConversionRule:   m.ConversionRule,
				ValidSamples:     1,
				ExpectedSamples:  1,
				MinCompleteness:  0,
			})
			opOut.SourceAggregationCount = 1
			opOut.SourceLatestIdentifier = opAgg.Identifier
			if opOut.SourceLatestIdentifier == "" {
				opOut.SourceLatestIdentifier = opAgg.PeriodEnd.In(loc).Format("20060102150405")
			}
			opOut.SourceAggregationInterval = paramInterval
			opOut.SourceAggregationFunction = paramFunc
		} else {
			opOut = ParameterCalculationOutput{
				PollutantCode:             m.PollutantCode,
				ParameterID:               m.ParameterID,
				SourceUnit:                m.SourceUnit,
				Category:                  model.CategoryUnavailable,
				ExpectedSamples:           1,
				ValidSamples:              0,
				IsEligible:                false,
				SourceAggregationInterval: paramInterval,
				SourceAggregationFunction: paramFunc,
				ErrorMessage:              "no aggregation record for current bucket",
			}
		}
		opParamOutputs = append(opParamOutputs, opOut)

		// --- Mode B: Regulatory (24-hour evaluation under Permen LHK 14/2020) ---
		var aggStats []struct {
			Val   float64
			Rec   time.Time
			Ident string
		}
		s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Select("COALESCE(avg_value, value) AS val, period_end AS rec, identifier AS ident").
			Where("device_id = ? AND parameter_id = ? AND period_end >= ? AND period_start <= ? AND quality != ?",
				deviceID, m.ParameterID, regPeriodStart.UTC(), periodEnd.UTC(), model.QualityBad).
			Order("period_end ASC").
			Find(&aggStats)

		var regOut ParameterCalculationOutput
		if len(aggStats) > 0 {
			hoursWithData := make(map[int]bool)
			var totalSum float64

			for _, as := range aggStats {
				hoursDiff := int(periodEnd.Sub(as.Rec).Hours())
				if hoursDiff >= 0 && hoursDiff < 24 {
					hoursWithData[hoursDiff] = true
					totalSum += as.Val
				}
			}

			latest := aggStats[len(aggStats)-1]
			latestAggIdent := latest.Ident
			if latestAggIdent == "" {
				latestAggIdent = latest.Rec.In(loc).Format("20060102150405")
			}

			regInput := ParameterCalculationInput{
				PollutantCode:    m.PollutantCode,
				ParameterID:      m.ParameterID,
				SourceUnit:       m.SourceUnit,
				RawConcentration: totalSum / float64(len(aggStats)),
				ConversionFactor: m.ConversionFactor,
				ConversionRule:   m.ConversionRule,
				ValidSamples:     len(hoursWithData),
				ExpectedSamples:  24,
				MinCompleteness:  m.MinCompleteness,
			}
			regOut = s.calculator.CalculateParameter(regInput)
			regOut.SourceAggregationCount = len(aggStats)
			regOut.SourceLatestIdentifier = latestAggIdent
			regOut.SourceAggregationInterval = paramInterval
			regOut.SourceAggregationFunction = paramFunc

			if hasOpAgg {
				regOut.AggregationConcentration = opOut.NormalizedConcentration
				regOut.AggregationISPU = opOut.CalculatedISPU
				regOut.AggregationCategory = opOut.Category
			} else {
				regOut.AggregationCategory = model.CategoryUnavailable
			}
		} else if !hasParamDef && len(aggDefs) == 0 {
			// Fallback to RawData ONLY when NO aggregation definitions exist at all
			var rawStats []struct {
				Val float64
				Rec time.Time
			}
			s.db.WithContext(ctx).Model(&model.RawData{}).
				Select("processed_value AS val, received_at AS rec").
				Where("device_id = ? AND parameter_id = ? AND received_at >= ? AND received_at <= ? AND quality != ?",
					deviceID, m.ParameterID, regPeriodStart.UTC(), periodEnd.UTC(), model.QualityBad).
				Order("received_at ASC").
				Find(&rawStats)

			if len(rawStats) > 0 {
				hoursWithData := make(map[int]bool)
				var totalSum float64

				for _, rs := range rawStats {
					hoursDiff := int(periodEnd.Sub(rs.Rec).Hours())
					if hoursDiff >= 0 && hoursDiff < 24 {
						hoursWithData[hoursDiff] = true
						totalSum += rs.Val
					}
				}

				regInput := ParameterCalculationInput{
					PollutantCode:    m.PollutantCode,
					ParameterID:      m.ParameterID,
					SourceUnit:       m.SourceUnit,
					RawConcentration: totalSum / float64(len(rawStats)),
					ConversionFactor: m.ConversionFactor,
					ConversionRule:   m.ConversionRule,
					ValidSamples:     len(hoursWithData),
					ExpectedSamples:  24,
					MinCompleteness:  m.MinCompleteness,
				}
				regOut = s.calculator.CalculateParameter(regInput)
				regOut.SourceAggregationCount = len(rawStats)
				regOut.SourceLatestIdentifier = rawStats[len(rawStats)-1].Rec.In(loc).Format("20060102150405")
				regOut.SourceAggregationInterval = paramInterval
				regOut.SourceAggregationFunction = paramFunc
				if hasOpAgg {
					regOut.AggregationConcentration = opOut.NormalizedConcentration
					regOut.AggregationISPU = opOut.CalculatedISPU
					regOut.AggregationCategory = opOut.Category
				}
			} else {
				regOut = ParameterCalculationOutput{
					PollutantCode:             m.PollutantCode,
					ParameterID:               m.ParameterID,
					SourceUnit:                m.SourceUnit,
					Category:                  model.CategoryUnavailable,
					ExpectedSamples:           24,
					ValidSamples:              0,
					IsEligible:                false,
					SourceAggregationInterval: paramInterval,
					SourceAggregationFunction: paramFunc,
				}
			}
		} else {
			regOut = ParameterCalculationOutput{
				PollutantCode:             m.PollutantCode,
				ParameterID:               m.ParameterID,
				SourceUnit:                m.SourceUnit,
				Category:                  model.CategoryUnavailable,
				ExpectedSamples:           24,
				ValidSamples:              0,
				IsEligible:                false,
				SourceAggregationInterval: paramInterval,
				SourceAggregationFunction: paramFunc,
			}
		}
		regParamOutputs = append(regParamOutputs, regOut)
	}

	// 1. Persist Mode A (Operational ISPU — Aggregation Driven)
	var maxOpISPU float64 = -1
	maxOpCat := model.CategoryUnavailable
	var maxOpDominant model.ISPUParameter
	validOpCount := 0

	for _, po := range opParamOutputs {
		if po.IsEligible && po.ValidSamples > 0 {
			validOpCount++
			if po.CalculatedISPU > maxOpISPU {
				maxOpISPU = po.CalculatedISPU
				maxOpCat = po.Category
				maxOpDominant = po.PollutantCode
			}
		}
	}
	if maxOpISPU < 0 {
		maxOpISPU = 0
		maxOpCat = model.CategoryUnavailable
		maxOpDominant = "-"
	}
	opCompleteness := 0.0
	opStatus := model.ISPUStatusInvalid
	if len(activeMappings) > 0 {
		opCompleteness = (float64(validOpCount) / float64(len(activeMappings))) * 100.0
	}
	if validOpCount > 0 {
		opStatus = model.ISPUStatusValid
	}

	opCalcRes := &model.ISPUCalculationResult{
		DeviceID:                   deviceID,
		CalculationMode:            model.CalculationModeOperational,
		PeriodStart:                opPeriodStart.UTC(),
		PeriodEnd:                  periodEnd.UTC(),
		CalculatedAt:               time.Now().UTC(),
		OverallISPU:                maxOpISPU,
		OverallCategory:            maxOpCat,
		DominantPollutant:          maxOpDominant,
		CompletenessPercent:        opCompleteness,
		CalculationStatus:          opStatus,
		Method:                     "PERMEN_LHK_14_2020",
		Timezone:                   tzName,
		AggregationISPU:            maxOpISPU,
		AggregationCategory:        maxOpCat,
		AggregationDominant:        maxOpDominant,
		AggregationIntervalSeconds: intervalSec,
		AggregationFunction:        aggFunction,
		SourceBucketIdentifier:     sourceBucketIdent,
		TriggerReason:              reason,
		Notes:                      fmt.Sprintf("Operational ISPU evaluated for %ds aggregation bucket", intervalSec),
	}
	for _, po := range opParamOutputs {
		opCalcRes.ParameterResults = append(opCalcRes.ParameterResults, model.ISPUParameterResult{
			PollutantCode:             po.PollutantCode,
			ParameterID:               po.ParameterID,
			Concentration:             po.RawConcentration,
			SourceUnit:                po.SourceUnit,
			NormalizedConcentration:   po.NormalizedConcentration,
			NormalizedUnit:            po.NormalizedUnit,
			CalculatedISPU:            po.CalculatedISPU,
			Category:                  po.Category,
			AggregationConcentration:  po.NormalizedConcentration,
			AggregationISPU:           po.CalculatedISPU,
			AggregationCategory:       po.Category,
			Ia:                        po.Ia,
			Ib:                        po.Ib,
			Xa:                        po.Xa,
			Xb:                        po.Xb,
			CompletenessPercent:       po.CompletenessPercent,
			ValidSamples:              po.ValidSamples,
			TotalExpectedSamples:      po.ExpectedSamples,
			Formula:                   po.Formula,
			SourceAggregationCount:    po.SourceAggregationCount,
			SourceLatestIdentifier:    po.SourceLatestIdentifier,
			SourceAggregationInterval: po.SourceAggregationInterval,
			SourceAggregationFunction: po.SourceAggregationFunction,
		})
	}
	if err := s.ispuRepo.UpsertCalculationResult(ctx, opCalcRes); err != nil {
		logger.Warn("Failed to persist operational ISPU result for station %d: %v", deviceID, err)
	}

	// 2. Persist Mode B (Regulatory ISPU — Required 24-hour Measurement Window)
	minComp := 75.0
	if len(activeMappings) > 0 && activeMappings[0].MinCompleteness > 0 {
		minComp = activeMappings[0].MinCompleteness
	}
	regOverall := s.calculator.CalculateOverall(regParamOutputs, minComp)

	tiedStr := ""
	if len(regOverall.TiedPollutants) > 0 {
		for i, p := range regOverall.TiedPollutants {
			if i > 0 {
				tiedStr += ", "
			}
			tiedStr += string(p)
		}
	}

	regCalcRes := &model.ISPUCalculationResult{
		DeviceID:                   deviceID,
		CalculationMode:            model.CalculationModeRegulatory,
		PeriodStart:                regPeriodStart.UTC(),
		PeriodEnd:                  periodEnd.UTC(),
		CalculatedAt:               time.Now().UTC(),
		OverallISPU:                regOverall.OverallISPU,
		OverallCategory:            regOverall.Category,
		DominantPollutant:          regOverall.DominantPollutant,
		TiedPollutants:             tiedStr,
		CompletenessPercent:        regOverall.CompletenessPercent,
		CalculationStatus:          regOverall.CalculationStatus,
		Method:                     "PERMEN_LHK_14_2020",
		Timezone:                   tzName,
		AggregationISPU:            maxOpISPU,
		AggregationCategory:        maxOpCat,
		AggregationDominant:        maxOpDominant,
		AggregationIntervalSeconds: intervalSec,
		AggregationFunction:        aggFunction,
		SourceBucketIdentifier:     sourceBucketIdent,
		TriggerReason:              reason,
		Notes:                      regOverall.Notes,
	}

	for _, po := range regParamOutputs {
		regCalcRes.ParameterResults = append(regCalcRes.ParameterResults, model.ISPUParameterResult{
			PollutantCode:             po.PollutantCode,
			ParameterID:               po.ParameterID,
			Concentration:             po.RawConcentration,
			SourceUnit:                po.SourceUnit,
			NormalizedConcentration:   po.NormalizedConcentration,
			NormalizedUnit:            po.NormalizedUnit,
			CalculatedISPU:            po.CalculatedISPU,
			Category:                  po.Category,
			AggregationConcentration:  po.AggregationConcentration,
			AggregationISPU:           po.AggregationISPU,
			AggregationCategory:       po.AggregationCategory,
			Ia:                        po.Ia,
			Ib:                        po.Ib,
			Xa:                        po.Xa,
			Xb:                        po.Xb,
			CompletenessPercent:       po.CompletenessPercent,
			ValidSamples:              po.ValidSamples,
			TotalExpectedSamples:      po.ExpectedSamples,
			Formula:                   po.Formula,
			SourceAggregationCount:    po.SourceAggregationCount,
			SourceLatestIdentifier:    po.SourceLatestIdentifier,
			SourceAggregationInterval: po.SourceAggregationInterval,
			SourceAggregationFunction: po.SourceAggregationFunction,
		})
	}

	// Persist idempotently
	if err := s.ispuRepo.UpsertCalculationResult(ctx, regCalcRes); err != nil {
		return nil, fmt.Errorf("failed to persist regulatory ISPU calculation result: %w", err)
	}

	logger.Info("ISPU calculated for station %d [%s] (reason=%s): Operational=%.1f (%s, Dominant=%s), Regulatory=%.1f (%s, Dominant=%s, Status=%s)",
		deviceID, dev.DeviceCode, reason, opCalcRes.OverallISPU, opCalcRes.OverallCategory, opCalcRes.DominantPollutant,
		regCalcRes.OverallISPU, regCalcRes.OverallCategory, regCalcRes.DominantPollutant, regCalcRes.CalculationStatus)

	// Broadcast real-time ISPU calculation over WebSocket
	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "ispu.calculated",
			Timestamp: time.Now().UTC(),
			Data:      regCalcRes,
		})
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "ispu.operational_calculated",
			Timestamp: time.Now().UTC(),
			Data:      opCalcRes,
		})
	}

	return regCalcRes, nil
}

// BackfillStationHourlyTrend retrospectively calculates and persists 24-hour rolling ISPU points
// at 1-hour increments for the specified lookback duration (e.g., 24 hours).
func (s *ISPUService) BackfillStationHourlyTrend(ctx context.Context, deviceID uint, lookbackHours int) (int, error) {
	if lookbackHours <= 0 {
		lookbackHours = 24
	}
	if lookbackHours > 720 { // Cap at 30 days
		lookbackHours = 720
	}

	var dev model.Device
	if err := s.db.WithContext(ctx).First(&dev, deviceID).Error; err != nil {
		return 0, fmt.Errorf("station/device %d not found: %w", deviceID, err)
	}

	tzName := dev.Timezone
	if tzName == "" {
		tzName = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil || loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	now := time.Now().In(loc)
	currentHour := now.Truncate(time.Hour)

	processed := 0
	// 1. Step hourly backwards from current hour down to lookbackHours
	for i := lookbackHours; i >= 0; i-- {
		targetHour := currentHour.Add(-time.Duration(i) * time.Hour)
		if _, err := s.CalculateStationWindow(ctx, deviceID, targetHour); err == nil {
			processed++
		}
	}

	// 2. Fetch active aggregation interval for this station (e.g. 120s = 2m even minutes)
	var intervalSec int = 120
	var aggDef model.AggregationDefinition
	if err := s.db.WithContext(ctx).
		Where("device_id = ? AND enabled = ?", deviceID, true).
		Order("interval_seconds ASC").
		First(&aggDef).Error; err == nil && aggDef.IntervalSeconds > 0 {
		intervalSec = aggDef.IntervalSeconds
	}

	// 3. Backfill recent aggregation intervals (last 2 hours) so high-frequency table is populated with even minutes
	if intervalSec > 0 && intervalSec < 3600 {
		recentStart := now.Add(-2 * time.Hour)
		alignedStart := s.AlignToAggregationInterval(recentStart, intervalSec, loc)
		alignedNow := s.AlignToAggregationInterval(now, intervalSec, loc)
		for t := alignedStart; !t.After(alignedNow); t = t.Add(time.Duration(intervalSec) * time.Second) {
			if _, err := s.CalculateStationWindow(ctx, deviceID, t); err == nil {
				processed++
			}
		}
	}

	// 4. Clean up any unaligned historical ISPU records (e.g. odd minutes when interval is 2m even)
	var allRes []model.ISPUCalculationResult
	if err := s.db.WithContext(ctx).Where("device_id = ?", deviceID).Select("id, period_end").Find(&allRes).Error; err == nil {
		var idsToDelete []uint
		for _, r := range allRes {
			aligned := s.AlignToAggregationInterval(r.PeriodEnd, intervalSec, loc)
			if !aligned.Equal(r.PeriodEnd.In(loc)) {
				idsToDelete = append(idsToDelete, r.ID)
			}
		}
		if len(idsToDelete) > 0 {
			_ = s.db.WithContext(ctx).Where("id IN (?)", idsToDelete).Delete(&model.ISPUCalculationResult{}).Error
			_ = s.db.WithContext(ctx).Where("calculation_result_id IN (?)", idsToDelete).Delete(&model.ISPUParameterResult{}).Error
			logger.Info("Purged %d unaligned historical ISPU records for station %d", len(idsToDelete), deviceID)
		}
	}

	logger.Info("ISPU trend backfill completed for station %d [%s]: %d windows generated",
		deviceID, dev.DeviceCode, processed)

	return processed, nil
}

// GetLatestSummary retrieves the latest ISPU summary for a station.
// It strictly reads the persisted calculation result from MariaDB.
// Mode can be REGULATORY or OPERATIONAL (defaults to REGULATORY).
// It NEVER executes calculations on read.
func (s *ISPUService) GetLatestSummary(ctx context.Context, deviceID uint, modes ...string) (*model.ISPUCalculationResult, error) {
	mode := string(model.CalculationModeRegulatory)
	if len(modes) > 0 && modes[0] != "" {
		mode = modes[0]
	}
	return s.ispuRepo.GetLatestByDeviceID(ctx, deviceID, mode)
}

// GetTrend retrieves historical calculations for chart rendering across selected time ranges ("24h", "7d", "30d") and mode.
// It strictly reads persisted records from MariaDB without triggering calculations on read.
func (s *ISPUService) GetTrend(ctx context.Context, deviceID uint, timeRange string, modes ...string) ([]model.ISPUCalculationResult, error) {
	mode := string(model.CalculationModeRegulatory)
	if len(modes) > 0 && modes[0] != "" {
		mode = modes[0]
	}
	now := time.Now()
	var startTime time.Time

	switch timeRange {
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	default: // "24h"
		startTime = now.Add(-24 * time.Hour)
	}

	return s.ispuRepo.GetTrend(ctx, deviceID, startTime, now, mode)
}

// GetHistory retrieves paginated historical records.
func (s *ISPUService) GetHistory(ctx context.Context, filter repository.ISPUHistoryFilter) ([]model.ISPUCalculationResult, int64, error) {
	return s.ispuRepo.GetHistory(ctx, filter)
}

// GetCalculationDetail retrieves calculation result with all parameter details.
func (s *ISPUService) GetCalculationDetail(ctx context.Context, id uint) (*model.ISPUCalculationResult, error) {
	return s.ispuRepo.GetByID(ctx, id)
}

// Recalculate recalculates ISPU for station and audit logs the action.
func (s *ISPUService) Recalculate(ctx context.Context, deviceID uint, targetTime *time.Time, actor, ipAddress string) (*model.ISPUCalculationResult, error) {
	var t time.Time
	if targetTime != nil && !targetTime.IsZero() {
		t = *targetTime
	}

	res, err := s.CalculateStationWindowWithReason(ctx, deviceID, t, "MANUAL_RECALCULATE")
	if err != nil {
		return nil, err
	}

	// Audit trail
	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  actor,
			Action:    "ISPU_RECALCULATE",
			Resource:  fmt.Sprintf("station_%d", deviceID),
			IPAddress: ipAddress,
			Details: fmt.Sprintf("Manual ISPU recalculation triggered for period ending %s. Result: %.1f (%s, %s)",
				res.PeriodEnd.Format(time.RFC3339), res.OverallISPU, res.OverallCategory, res.DominantPollutant),
			CreatedAt: time.Now().UTC(),
		})
	}

	return s.ispuRepo.GetByID(ctx, res.ID)
}

// ExportCSV exports persisted historical ISPU records into CSV.
func (s *ISPUService) ExportCSV(ctx context.Context, filter repository.ISPUHistoryFilter) ([]byte, error) {
	filter.Page = 1
	filter.PageSize = 2000 // Fetch up to 2000 records for export
	records, _, err := s.ispuRepo.GetHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Metadata headers
	_ = w.Write([]string{"# INDEKS STANDAR PENCEMAR UDARA (ISPU) - PERMEN LHK NO. 14 TAHUN 2020"})
	_ = w.Write([]string{"# Export Generated At", time.Now().Format("2006-01-02 15:04:05 MST")})
	_ = w.Write([]string{"# Total Records", strconv.Itoa(len(records))})
	_ = w.Write([]string{""})

	// CSV Columns
	headers := []string{
		"ID", "Station Code", "Station Name", "Period Start (UTC)", "Period End (UTC)", "Calculated At (UTC)",
		"Overall ISPU", "Category", "Dominant Pollutant", "Tied Pollutants", "Completeness (%)", "Status", "Method", "Notes",
		"PM2.5 ISPU", "PM2.5 Conc (µg/m³)", "PM10 ISPU", "PM10 Conc (µg/m³)",
		"SO2 ISPU", "SO2 Conc (µg/m³)", "CO ISPU", "CO Conc (µg/m³)",
		"O3 ISPU", "O3 Conc (µg/m³)", "NO2 ISPU", "NO2 Conc (µg/m³)", "HC ISPU", "HC Conc (µg/m³)",
	}
	_ = w.Write(headers)

	for _, r := range records {
		devCode := ""
		devName := ""
		if r.Device != nil {
			devCode = r.Device.DeviceCode
			devName = r.Device.DeviceName
		}

		// Map pollutant parameters
		paramMap := make(map[model.ISPUParameter]model.ISPUParameterResult)
		for _, pr := range r.ParameterResults {
			paramMap[pr.PollutantCode] = pr
		}

		getParamStr := func(p model.ISPUParameter) (string, string) {
			if res, ok := paramMap[p]; ok {
				return fmt.Sprintf("%.1f", res.CalculatedISPU), fmt.Sprintf("%.2f", res.NormalizedConcentration)
			}
			return "-", "-"
		}

		pm25Ispu, pm25Conc := getParamStr(model.PollutantPM25)
		pm10Ispu, pm10Conc := getParamStr(model.PollutantPM10)
		so2Ispu, so2Conc := getParamStr(model.PollutantSO2)
		coIspu, coConc := getParamStr(model.PollutantCO)
		o3Ispu, o3Conc := getParamStr(model.PollutantO3)
		no2Ispu, no2Conc := getParamStr(model.PollutantNO2)
		hcIspu, hcConc := getParamStr(model.PollutantHC)

		row := []string{
			strconv.FormatUint(uint64(r.ID), 10),
			devCode,
			devName,
			r.PeriodStart.Format("2006-01-02 15:04:05"),
			r.PeriodEnd.Format("2006-01-02 15:04:05"),
			r.CalculatedAt.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.1f", r.OverallISPU),
			string(r.OverallCategory),
			string(r.DominantPollutant),
			r.TiedPollutants,
			fmt.Sprintf("%.1f", r.CompletenessPercent),
			string(r.CalculationStatus),
			r.Method,
			r.Notes,
			pm25Ispu, pm25Conc,
			pm10Ispu, pm10Conc,
			so2Ispu, so2Conc,
			coIspu, coConc,
			o3Ispu, o3Conc,
			no2Ispu, no2Conc,
			hcIspu, hcConc,
		}
		_ = w.Write(row)
	}

	w.Flush()
	return buf.Bytes(), nil
}

// ExportExcel generates CSV with UTF-8 BOM which natively opens in Microsoft Excel with formatted tables.
func (s *ISPUService) ExportExcel(ctx context.Context, filter repository.ISPUHistoryFilter) ([]byte, error) {
	csvBytes, err := s.ExportCSV(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Prepend UTF-8 BOM
	bom := []byte{0xEF, 0xBB, 0xBF}
	excelData := append(bom, csvBytes...)
	return excelData, nil
}

// GetMappings retrieves station parameter mappings.
func (s *ISPUService) GetMappings(ctx context.Context, deviceID uint) ([]model.ISPUParameterMapping, error) {
	return s.ispuRepo.GetMappingsByDeviceID(ctx, deviceID)
}

// SaveMapping creates or updates an ISPU parameter mapping with audit logging.
func (s *ISPUService) SaveMapping(ctx context.Context, mapping *model.ISPUParameterMapping, actor, ipAddress string) error {
	action := "ISPU_MAPPING_CREATE"
	if mapping.ID > 0 {
		action = "ISPU_MAPPING_UPDATE"
	}

	if err := s.ispuRepo.SaveMapping(ctx, mapping); err != nil {
		return err
	}

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  actor,
			Action:    action,
			Resource:  fmt.Sprintf("station_%d_pollutant_%s", mapping.DeviceID, mapping.PollutantCode),
			IPAddress: ipAddress,
			Details: fmt.Sprintf("ISPU mapping saved: Station=%d, Param=%d, Pollutant=%s, Unit=%s, Factor=%.3f, Enabled=%v",
				mapping.DeviceID, mapping.ParameterID, mapping.PollutantCode, mapping.SourceUnit, mapping.ConversionFactor, mapping.Enabled),
			CreatedAt: time.Now().UTC(),
		})
	}

	return nil
}

// DeleteMapping removes a mapping with audit logging.
func (s *ISPUService) DeleteMapping(ctx context.Context, id uint, actor, ipAddress string) error {
	m, err := s.ispuRepo.GetMappingByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.ispuRepo.DeleteMapping(ctx, id); err != nil {
		return err
	}

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  actor,
			Action:    "ISPU_MAPPING_DELETE",
			Resource:  fmt.Sprintf("mapping_%d", id),
			IPAddress: ipAddress,
			Details:   fmt.Sprintf("ISPU mapping deleted: Station=%d, Pollutant=%s", m.DeviceID, m.PollutantCode),
			CreatedAt: time.Now().UTC(),
		})
	}

	return nil
}

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"sort"
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
	s.EnqueueCalculation(0, event.PeriodEnd, "AGGREGATION_COMPLETED")
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
			isHourly := alignedEnd.In(loc).Minute() == 0 && alignedEnd.In(loc).Second() == 0

			var opCount int64
			s.db.WithContext(ctx).Model(&model.ISPUCalculationResult{}).
				Where("device_id = ? AND period_end = ? AND calculation_mode = ?", deviceID, alignedEnd.UTC(), model.CalculationModeOperational).
				Count(&opCount)

			var regCount int64
			if isHourly {
				s.db.WithContext(ctx).Model(&model.ISPUCalculationResult{}).
					Where("device_id = ? AND period_end = ? AND calculation_mode = ?", deviceID, alignedEnd.UTC(), model.CalculationModeRegulatory).
					Count(&regCount)
			}

			// Operational ISPU is required for every aggregation interval.
			// Regulatory ISPU is strictly required only on hourly boundaries.
			needsRecovery := opCount == 0 || (isHourly && regCount == 0)

			if needsRecovery {
				recoveryReason := "BACKFILL_OPERATIONAL"
				if isHourly {
					recoveryReason = "STARTUP_RECOVERY"
				}
				logger.Info("Reconciling missing ISPU calculation for station %d period_end %v (isHourly=%v, op=%d, reg=%d, reason=%s)",
					deviceID, alignedEnd, isHourly, opCount, regCount, recoveryReason)
				_, err := s.CalculateStationWindowWithReason(ctx, deviceID, alignedEnd, recoveryReason)
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
	if deviceID > 0 {
		if err := s.db.WithContext(ctx).First(&dev, deviceID).Error; err != nil {
			return nil, fmt.Errorf("station/device %d not found: %w", deviceID, err)
		}
	} else {
		// Unified station: resolve timezone and parameters from active devices
		if err := s.db.WithContext(ctx).Where("enabled = true AND status = 'ACTIVE'").First(&dev).Error; err != nil {
			_ = s.db.WithContext(ctx).First(&dev)
		}
		if dev.DeviceCode == "" {
			dev.DeviceCode = "AQMS"
			dev.DeviceName = "Air Quality Monitoring Station"
		}
	}

	tzName := dev.Timezone
	if tzName == "" {
		tzName = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	// 1. Determine active aggregation definitions and interval for this device/station
	var intervalSec int = 120
	aggFunction := "AVG"
	var aggDefs []model.AggregationDefinition
	aggQuery := s.db.WithContext(ctx).Where("enabled = ?", true)
	if deviceID > 0 {
		aggQuery = aggQuery.Where("device_id = ?", deviceID)
	}
	if err := aggQuery.Order("interval_seconds ASC").Find(&aggDefs).Error; err == nil && len(aggDefs) > 0 {
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
		aggQ := s.db.WithContext(ctx)
		if deviceID > 0 {
			aggQ = aggQ.Where("device_id = ?", deviceID)
		}
		if err := aggQ.Order("period_end DESC").First(&latestAgg).Error; err == nil && !latestAgg.PeriodEnd.IsZero() {
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

	// Fetch active parameter mappings for station (cross-device if deviceID == 0)
	mappings, err := s.GetMappings(ctx, deviceID)
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

		if !hasOpAgg {
			// Fallback: Query raw_data for the bucket if no pre-aggregated result exists yet
			var rawAvg *float64
			var rawCount int64
			s.db.WithContext(ctx).Model(&model.RawData{}).
				Where("device_id = ? AND parameter_id = ? AND timestamp >= ? AND timestamp <= ? AND quality != ?",
					deviceID, m.ParameterID, opPeriodStart.UTC(), periodEnd.UTC(), model.QualityBad).
				Select("AVG(value)").
				Row().Scan(&rawAvg)

			s.db.WithContext(ctx).Model(&model.RawData{}).
				Where("device_id = ? AND parameter_id = ? AND timestamp >= ? AND timestamp <= ? AND quality != ?",
					deviceID, m.ParameterID, opPeriodStart.UTC(), periodEnd.UTC(), model.QualityBad).
				Count(&rawCount)

			if rawCount > 0 && rawAvg != nil {
				hasOpAgg = true
				opAgg = model.AggregationResult{
					DeviceID:        deviceID,
					ParameterID:     m.ParameterID,
					PeriodStart:     opPeriodStart.UTC(),
					PeriodEnd:       periodEnd.UTC(),
					IntervalSeconds: intervalSec,
					Value:           rawAvg,
					Quality:         model.QualityGood,
					Identifier:      sourceBucketIdent,
				}
			}
		}

		var opOut ParameterCalculationOutput
		if hasOpAgg {
			val := 0.0
			if opAgg.AvgValue != nil {
				val = *opAgg.AvgValue
			} else if opAgg.Value != nil {
				val = *opAgg.Value
			}

			// POS 3: Operational Gatekeeper check (guard against unphysical float overflow)
			if val < 0 || math.IsNaN(val) || math.IsInf(val, 0) || val > 1e8 {
				hasOpAgg = false
			}
		}

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

			recTime := opAgg.PeriodEnd
			opOut.PeakConcentration = opOut.NormalizedConcentration
			opOut.MinConcentration = opOut.NormalizedConcentration
			opOut.PeakTimestamp = &recTime
			opOut.PeakBucketIdentifier = opOut.SourceLatestIdentifier

			var maxBpXa float64 = 45000
			if bps, ok := OfficialBreakpoints[m.PollutantCode]; ok && len(bps) > 0 {
				maxBpXa = bps[len(bps)-1].Xa
			}
			isAnomaly := opOut.NormalizedConcentration > maxBpXa
			var anomalyReason string
			if isAnomaly {
				anomalyReason = fmt.Sprintf("Bucket concentration %.1f %s exceeds regulatory breakpoint maximum (%.1f %s)",
					opOut.NormalizedConcentration, opOut.NormalizedUnit, maxBpXa, opOut.NormalizedUnit)
			}
			opAudit := model.ISPUAuditDetails{
				PollutantCode:        m.PollutantCode,
				ParameterID:          m.ParameterID,
				PeakConcentration:    opOut.NormalizedConcentration,
				PeakTimestamp:        &recTime,
				PeakBucketIdentifier: opOut.SourceLatestIdentifier,
				MinConcentration:     opOut.NormalizedConcentration,
				MinTimestamp:         &recTime,
				AvgConcentration:     opOut.NormalizedConcentration,
				IsAnomaly:            isAnomaly,
				AnomalyReason:        anomalyReason,
				RegulatoryXa:         maxBpXa,
				SourceUnit:           m.SourceUnit,
				NormalizedUnit:       opOut.NormalizedUnit,
				TotalBuckets:         1,
				TopSpikes: []model.ISPUOutlierBucket{
					{
						Timestamp:     recTime,
						Identifier:    opOut.SourceLatestIdentifier,
						RawValue:      val,
						NormalizedVal: opOut.NormalizedConcentration,
					},
				},
			}
			if b, err := json.Marshal(opAudit); err == nil {
				opOut.DetailsJSON = string(b)
			}
		} else {
			opOut = ParameterCalculationOutput{
				PollutantCode:             m.PollutantCode,
				ParameterID:               m.ParameterID,
				SourceUnit:                m.SourceUnit,
				Category:                  model.CategoryUnavailable,
				ExpectedSamples:           1,
				ValidSamples:              0,
				IsEligible:                false,
				Formula:                   "Tidak ada data agregasi pada interval ini",
				SourceAggregationInterval: paramInterval,
				SourceAggregationFunction: paramFunc,
				ErrorMessage:              "no aggregation record for current bucket",
			}
		}
		opParamOutputs = append(opParamOutputs, opOut)

		// --- Mode B: Regulatory (24-hour evaluation under Permen LHK 14/2020) ---
		// Determine parameter-specific evaluation window anchored by DailyStartHour
		paramRegStart := regPeriodStart
		paramRegEnd := periodEnd

		if m.DailyStartHour >= 0 && m.DailyStartHour <= 23 {
			refLocal := periodEnd.In(loc)
			anchorTime := time.Date(refLocal.Year(), refLocal.Month(), refLocal.Day(), m.DailyStartHour, 0, 0, 0, loc)
			if refLocal.Before(anchorTime) {
				anchorTime = anchorTime.Add(-24 * time.Hour)
			}
			// Evaluasi siklus 24 jam dengan acuan daily start hour
			if refLocal.Equal(anchorTime) {
				paramRegStart = anchorTime.Add(-24 * time.Hour)
				paramRegEnd = anchorTime
			} else {
				paramRegStart = anchorTime.Add(-24 * time.Hour)
				if periodEnd.Sub(paramRegStart) > 24*time.Hour {
					paramRegStart = periodEnd.Add(-24 * time.Hour)
				}
				paramRegEnd = periodEnd
			}
		}

		var aggStats []struct {
			Val   float64
			Rec   time.Time
			Ident string
		}
		s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Select("COALESCE(avg_value, value) AS val, period_end AS rec, identifier AS ident").
			Where("device_id = ? AND parameter_id = ? AND period_end >= ? AND period_start <= ? AND quality != ?",
				deviceID, m.ParameterID, paramRegStart.UTC(), paramRegEnd.UTC(), model.QualityBad).
			Order("period_end ASC").
			Find(&aggStats)

		if len(aggStats) == 0 {
			// Fallback: Check raw_data hourly averages across 24h window
			var rawHourly []struct {
				Val   float64
				Rec   time.Time
				Ident string
			}
			s.db.WithContext(ctx).Model(&model.RawData{}).
				Select("AVG(value) AS val, MAX(timestamp) AS rec, DATE_FORMAT(timestamp, '%Y%m%d%H0000') AS ident").
				Where("device_id = ? AND parameter_id = ? AND timestamp >= ? AND timestamp <= ? AND quality != ?",
					deviceID, m.ParameterID, paramRegStart.UTC(), paramRegEnd.UTC(), model.QualityBad).
				Group("DATE(timestamp), HOUR(timestamp)").
				Find(&rawHourly)

			if len(rawHourly) > 0 {
				aggStats = rawHourly
			}
		}

		var regOut ParameterCalculationOutput
		if len(aggStats) > 0 {
			var maxBpXa float64 = 45000
			if bps, ok := OfficialBreakpoints[m.PollutantCode]; ok && len(bps) > 0 {
				maxBpXa = bps[len(bps)-1].Xa
			}

			// POS 3: Regulatory ISPU Gatekeeper Plausibility Ceiling
			// Atmospheric concentrations cannot plausibly exceed 20x the highest regulatory breakpoint or 1e8 µg/m³.
			// Any bucket concentration above this ceiling represents unphysical Modbus bit-overflow / transmission corruption
			// and is quarantined/excluded from the 24-hour SUM and average.
			maxPlausibleLimit := maxBpXa * 20.0
			if maxPlausibleLimit < 100000.0 {
				maxPlausibleLimit = 100000.0
			}
			if maxPlausibleLimit > 1e8 {
				maxPlausibleLimit = 1e8
			}

			hoursWithData := make(map[int]bool)
			var totalSum float64
			usableBucketCount := 0
			excludedOutlierCount := 0
			allBuckets := make([]model.ISPUOutlierBucket, 0, len(aggStats))

			for _, as := range aggStats {
				normVal, _, _ := s.calculator.NormalizeConcentration(as.Val, m.SourceUnit, m.ConversionRule, m.ConversionFactor)
				ident := as.Ident
				if ident == "" {
					ident = as.Rec.In(loc).Format("20060102150405")
				}

				isExcluded := false
				excludeReason := ""
				if as.Val < 0 || normVal > maxPlausibleLimit || math.IsNaN(as.Val) || math.IsInf(as.Val, 0) {
					isExcluded = true
					excludeReason = fmt.Sprintf("Implausible value (%.2e %s) exceeds regulatory plausibility ceiling (%.1f)", as.Val, m.SourceUnit, maxPlausibleLimit)
					excludedOutlierCount++
				}

				hoursDiff := int(paramRegEnd.Sub(as.Rec).Hours())
				if !isExcluded {
					if hoursDiff >= 0 && hoursDiff < 24 {
						hoursWithData[hoursDiff] = true
						totalSum += as.Val
						usableBucketCount++
					}
				}

				allBuckets = append(allBuckets, model.ISPUOutlierBucket{
					Timestamp:     as.Rec,
					Identifier:    ident,
					RawValue:      as.Val,
					NormalizedVal: normVal,
					IsExcluded:    isExcluded,
					ExcludeReason: excludeReason,
				})
			}

			rawAvg := 0.0
			if usableBucketCount > 0 {
				rawAvg = totalSum / float64(usableBucketCount)
			} else if len(aggStats) > 0 {
				rawAvg = totalSum / float64(len(aggStats))
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
				RawConcentration: rawAvg,
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

			if len(allBuckets) > 0 {
				sort.Slice(allBuckets, func(i, j int) bool {
					return allBuckets[i].NormalizedVal > allBuckets[j].NormalizedVal
				})

				peakConc := allBuckets[0].NormalizedVal
				tPeak := allBuckets[0].Timestamp
				peakTime := &tPeak
				peakIdent := allBuckets[0].Identifier
				minConc := allBuckets[len(allBuckets)-1].NormalizedVal
				tMin := allBuckets[len(allBuckets)-1].Timestamp
				minTime := &tMin

				topCount := 10
				if len(allBuckets) < topCount {
					topCount = len(allBuckets)
				}
				topSpikes := make([]model.ISPUOutlierBucket, topCount)
				copy(topSpikes, allBuckets[:topCount])

				isAnomaly := false
				var anomalyReason string
				if peakConc > maxBpXa {
					isAnomaly = true
					anomalyReason = fmt.Sprintf("Peak concentration %.1f %s at %s exceeds regulatory breakpoint maximum (%.1f %s)",
						peakConc, regOut.NormalizedUnit, peakTime.In(loc).Format("2006-01-02 15:04:05"), maxBpXa, regOut.NormalizedUnit)
				} else if regOut.NormalizedConcentration > 0 && peakConc > regOut.NormalizedConcentration*10 {
					isAnomaly = true
					anomalyReason = fmt.Sprintf("Extreme spike detected: Peak %.1f %s is %.1fx higher than 24h average (%.1f %s)",
						peakConc, regOut.NormalizedUnit, peakConc/regOut.NormalizedConcentration, regOut.NormalizedConcentration, regOut.NormalizedUnit)
				}

				regOut.PeakConcentration = peakConc
				regOut.PeakTimestamp = peakTime
				regOut.PeakBucketIdentifier = peakIdent
				regOut.MinConcentration = minConc

				auditDetails := model.ISPUAuditDetails{
					PollutantCode:         m.PollutantCode,
					ParameterID:           m.ParameterID,
					PeakConcentration:     peakConc,
					PeakTimestamp:         peakTime,
					PeakBucketIdentifier:  peakIdent,
					MinConcentration:      minConc,
					MinTimestamp:          minTime,
					AvgConcentration:      regOut.NormalizedConcentration,
					IsAnomaly:             isAnomaly || excludedOutlierCount > 0,
					AnomalyReason:         anomalyReason,
					RegulatoryXa:          maxBpXa,
					SourceUnit:            m.SourceUnit,
					NormalizedUnit:        regOut.NormalizedUnit,
					TotalBuckets:          len(aggStats),
					ExcludedOutliersCount: excludedOutlierCount,
					TopSpikes:             topSpikes,
				}
				if b, err := json.Marshal(auditDetails); err == nil {
					regOut.DetailsJSON = string(b)
				}
			}

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
				allBuckets := make([]model.ISPUOutlierBucket, 0, len(rawStats))

				for _, rs := range rawStats {
					hoursDiff := int(periodEnd.Sub(rs.Rec).Hours())
					if hoursDiff >= 0 && hoursDiff < 24 {
						hoursWithData[hoursDiff] = true
						totalSum += rs.Val
					}
					normVal, _, _ := s.calculator.NormalizeConcentration(rs.Val, m.SourceUnit, m.ConversionRule, m.ConversionFactor)
					ident := rs.Rec.In(loc).Format("20060102150405")
					allBuckets = append(allBuckets, model.ISPUOutlierBucket{
						Timestamp:     rs.Rec,
						Identifier:    ident,
						RawValue:      rs.Val,
						NormalizedVal: normVal,
					})
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

				if len(allBuckets) > 0 {
					sort.Slice(allBuckets, func(i, j int) bool {
						return allBuckets[i].NormalizedVal > allBuckets[j].NormalizedVal
					})

					peakConc := allBuckets[0].NormalizedVal
					tPeak := allBuckets[0].Timestamp
					peakTime := &tPeak
					peakIdent := allBuckets[0].Identifier
					minConc := allBuckets[len(allBuckets)-1].NormalizedVal
					tMin := allBuckets[len(allBuckets)-1].Timestamp
					minTime := &tMin

					topCount := 10
					if len(allBuckets) < topCount {
						topCount = len(allBuckets)
					}
					topSpikes := make([]model.ISPUOutlierBucket, topCount)
					copy(topSpikes, allBuckets[:topCount])

					var maxBpXa float64 = 45000
					if bps, ok := OfficialBreakpoints[m.PollutantCode]; ok && len(bps) > 0 {
						maxBpXa = bps[len(bps)-1].Xa
					}

					isAnomaly := false
					var anomalyReason string
					if peakConc > maxBpXa {
						isAnomaly = true
						anomalyReason = fmt.Sprintf("Peak concentration %.1f %s at %s exceeds regulatory breakpoint maximum (%.1f %s)",
							peakConc, regOut.NormalizedUnit, peakTime.In(loc).Format("2006-01-02 15:04:05"), maxBpXa, regOut.NormalizedUnit)
					} else if regOut.NormalizedConcentration > 0 && peakConc > regOut.NormalizedConcentration*10 {
						isAnomaly = true
						anomalyReason = fmt.Sprintf("Extreme spike detected: Peak %.1f %s is %.1fx higher than 24h average (%.1f %s)",
							peakConc, regOut.NormalizedUnit, peakConc/regOut.NormalizedConcentration, regOut.NormalizedConcentration, regOut.NormalizedUnit)
					}

					regOut.PeakConcentration = peakConc
					regOut.PeakTimestamp = peakTime
					regOut.PeakBucketIdentifier = peakIdent
					regOut.MinConcentration = minConc

					auditDetails := model.ISPUAuditDetails{
						PollutantCode:        m.PollutantCode,
						ParameterID:          m.ParameterID,
						PeakConcentration:    peakConc,
						PeakTimestamp:        peakTime,
						PeakBucketIdentifier: peakIdent,
						MinConcentration:     minConc,
						MinTimestamp:         minTime,
						AvgConcentration:     regOut.NormalizedConcentration,
						IsAnomaly:            isAnomaly,
						AnomalyReason:        anomalyReason,
						RegulatoryXa:         maxBpXa,
						SourceUnit:           m.SourceUnit,
						NormalizedUnit:       regOut.NormalizedUnit,
						TotalBuckets:         len(rawStats),
						TopSpikes:            topSpikes,
					}
					if b, err := json.Marshal(auditDetails); err == nil {
						regOut.DetailsJSON = string(b)
					}
				}

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
			DetailsJSON:               po.DetailsJSON,
			SourceAggregationCount:    po.SourceAggregationCount,
			SourceLatestIdentifier:    po.SourceLatestIdentifier,
			SourceAggregationInterval: po.SourceAggregationInterval,
			SourceAggregationFunction: po.SourceAggregationFunction,
			PeakConcentration:         po.PeakConcentration,
			PeakTimestamp:             po.PeakTimestamp,
			PeakBucketIdentifier:      po.PeakBucketIdentifier,
			MinConcentration:          po.MinConcentration,
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
			DetailsJSON:               po.DetailsJSON,
			SourceAggregationCount:    po.SourceAggregationCount,
			SourceLatestIdentifier:    po.SourceLatestIdentifier,
			SourceAggregationInterval: po.SourceAggregationInterval,
			SourceAggregationFunction: po.SourceAggregationFunction,
			PeakConcentration:         po.PeakConcentration,
			PeakTimestamp:             po.PeakTimestamp,
			PeakBucketIdentifier:      po.PeakBucketIdentifier,
			MinConcentration:          po.MinConcentration,
		})
	}

	// Persist idempotently strictly on documented regulatory cadence or explicit recalculation
	shouldPersistRegulatory := true
	if reason == "BACKFILL_OPERATIONAL" {
		shouldPersistRegulatory = false
	} else if reason == "AGGREGATION_COMPLETED" || reason == "STARTUP_RECOVERY" {
		isHourly := periodEnd.In(loc).Minute() == 0 && periodEnd.In(loc).Second() == 0
		if !isHourly {
			if reason == "STARTUP_RECOVERY" {
				shouldPersistRegulatory = false
			} else {
				// Check if an existing regulatory record already exists for this exact period (idempotent update)
				var existingRecordCount int64
				_ = s.db.WithContext(ctx).Model(&model.ISPUCalculationResult{}).
					Where("device_id = ? AND period_end = ? AND calculation_mode = ?", deviceID, periodEnd.UTC(), model.CalculationModeRegulatory).
					Count(&existingRecordCount)

				if existingRecordCount == 0 {
					// Allow initial seed record if station has 0 regulatory records across all history
					var existingRegCount int64
					_ = s.db.WithContext(ctx).Model(&model.ISPUCalculationResult{}).
						Where("device_id = ? AND calculation_mode = ?", deviceID, model.CalculationModeRegulatory).
						Count(&existingRegCount)
					if existingRegCount > 0 {
						shouldPersistRegulatory = false
					}
				}
			}
		}
	}

	if shouldPersistRegulatory {
		if err := s.ispuRepo.UpsertCalculationResult(ctx, regCalcRes); err != nil {
			return nil, fmt.Errorf("failed to persist regulatory ISPU calculation result: %w", err)
		}
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
	if deviceID > 0 {
		if err := s.db.WithContext(ctx).First(&dev, deviceID).Error; err != nil {
			return 0, fmt.Errorf("station/device %d not found: %w", deviceID, err)
		}
	} else {
		_ = s.db.WithContext(ctx).Where("enabled = ?", true).First(&dev).Error
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
		if _, err := s.CalculateStationWindowWithReason(ctx, deviceID, targetHour, "BACKFILL"); err == nil {
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
			if _, err := s.CalculateStationWindowWithReason(ctx, deviceID, t, "BACKFILL_OPERATIONAL"); err == nil {
				processed++
			}
		}
	}

	// 4. Clean up any unaligned historical ISPU records (e.g. odd minutes when interval is 2m even)
	// and any sub-hourly unaligned REGULATORY records (where minute != 0 and trigger_reason = 'AGGREGATION_COMPLETED')
	var allRes []model.ISPUCalculationResult
	allQuery := s.db.WithContext(ctx)
	if deviceID > 0 {
		allQuery = allQuery.Where("device_id = ?", deviceID)
	}
	if err := allQuery.Select("id, calculation_mode, period_end, trigger_reason").Find(&allRes).Error; err == nil {
		var idsToDelete []uint
		for _, r := range allRes {
			aligned := s.AlignToAggregationInterval(r.PeriodEnd, intervalSec, loc)
			if !aligned.Equal(r.PeriodEnd.In(loc)) {
				idsToDelete = append(idsToDelete, r.ID)
			} else if r.CalculationMode == model.CalculationModeRegulatory && (r.PeriodEnd.In(loc).Minute() != 0 || r.PeriodEnd.In(loc).Second() != 0) {
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

// GetCalculationDetail retrieves calculation result with all parameter details, enriching missing audit tracking.
func (s *ISPUService) GetCalculationDetail(ctx context.Context, id uint) (*model.ISPUCalculationResult, error) {
	res, err := s.ispuRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.enrichCalculationAuditTracking(ctx, res)
	return res, nil
}

// enrichCalculationAuditTracking verifies that all parameter results have peak tracking and details populated.
// If missing (e.g. historical calculation), it calculates them on-the-fly from aggregation buckets and saves back.
func (s *ISPUService) enrichCalculationAuditTracking(ctx context.Context, res *model.ISPUCalculationResult) {
	if res == nil || len(res.ParameterResults) == 0 {
		return
	}

	loc, err := time.LoadLocation(res.Timezone)
	if err != nil || loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	for i := range res.ParameterResults {
		pr := &res.ParameterResults[i]
		if pr.PeakTimestamp != nil && pr.DetailsJSON != "" {
			continue
		}

		var mapping model.ISPUParameterMapping
		s.db.WithContext(ctx).Where("device_id = ? AND parameter_id = ?", res.DeviceID, pr.ParameterID).First(&mapping)

		srcUnit := pr.SourceUnit
		convFactor := 1.0
		convRule := "DIRECT"
		if mapping.ID > 0 {
			srcUnit = mapping.SourceUnit
			convFactor = mapping.ConversionFactor
			convRule = mapping.ConversionRule
		}

		var aggStats []struct {
			Val   float64
			Rec   time.Time
			Ident string
		}

		s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Select("COALESCE(avg_value, value) AS val, period_end AS rec, identifier AS ident").
			Where("device_id = ? AND parameter_id = ? AND period_end >= ? AND period_start <= ? AND quality != ?",
				res.DeviceID, pr.ParameterID, res.PeriodStart.UTC(), res.PeriodEnd.UTC(), model.QualityBad).
			Order("period_end ASC").
			Find(&aggStats)

		if len(aggStats) == 0 {
			var rawHourly []struct {
				Val   float64
				Rec   time.Time
				Ident string
			}
			s.db.WithContext(ctx).Model(&model.RawData{}).
				Select("AVG(value) AS val, MAX(timestamp) AS rec, DATE_FORMAT(timestamp, '%Y%m%d%H0000') AS ident").
				Where("device_id = ? AND parameter_id = ? AND timestamp >= ? AND timestamp <= ? AND quality != ?",
					res.DeviceID, pr.ParameterID, res.PeriodStart.UTC(), res.PeriodEnd.UTC(), model.QualityBad).
				Group("DATE(timestamp), HOUR(timestamp)").
				Find(&rawHourly)
			if len(rawHourly) > 0 {
				aggStats = rawHourly
			}
		}

		if len(aggStats) == 0 {
			continue
		}

		allBuckets := make([]model.ISPUOutlierBucket, 0, len(aggStats))
		var maxBpXa float64 = 45000
		if bps, ok := OfficialBreakpoints[pr.PollutantCode]; ok && len(bps) > 0 {
			maxBpXa = bps[len(bps)-1].Xa
		}
		maxPlausibleLimit := maxBpXa * 20.0
		if maxPlausibleLimit < 100000.0 {
			maxPlausibleLimit = 100000.0
		}
		if maxPlausibleLimit > 1e8 {
			maxPlausibleLimit = 1e8
		}

		excludedCount := 0
		for _, as := range aggStats {
			normVal, _, _ := s.calculator.NormalizeConcentration(as.Val, srcUnit, convRule, convFactor)
			ident := as.Ident
			if ident == "" {
				ident = as.Rec.In(loc).Format("20060102150405")
			}
			isExcluded := false
			excludeReason := ""
			if as.Val < 0 || normVal > maxPlausibleLimit || math.IsNaN(as.Val) || math.IsInf(as.Val, 0) {
				isExcluded = true
				excludeReason = fmt.Sprintf("Implausible value (%.2e) exceeds regulatory physical ceiling (%.1f)", as.Val, maxPlausibleLimit)
				excludedCount++
			}
			allBuckets = append(allBuckets, model.ISPUOutlierBucket{
				Timestamp:     as.Rec,
				Identifier:    ident,
				RawValue:      as.Val,
				NormalizedVal: normVal,
				IsExcluded:    isExcluded,
				ExcludeReason: excludeReason,
			})
		}

		sort.Slice(allBuckets, func(a, b int) bool {
			return allBuckets[a].NormalizedVal > allBuckets[b].NormalizedVal
		})

		peakConc := allBuckets[0].NormalizedVal
		tPeak := allBuckets[0].Timestamp
		peakTime := &tPeak
		peakIdent := allBuckets[0].Identifier
		minConc := allBuckets[len(allBuckets)-1].NormalizedVal
		tMin := allBuckets[len(allBuckets)-1].Timestamp
		minTime := &tMin

		topCount := 10
		if len(allBuckets) < topCount {
			topCount = len(allBuckets)
		}
		topSpikes := make([]model.ISPUOutlierBucket, topCount)
		copy(topSpikes, allBuckets[:topCount])

		isAnomaly := false
		var anomalyReason string
		if peakConc > maxBpXa {
			isAnomaly = true
			anomalyReason = fmt.Sprintf("Peak concentration %.1f %s at %s exceeds regulatory breakpoint maximum (%.1f %s)",
				peakConc, pr.NormalizedUnit, peakTime.In(loc).Format("2006-01-02 15:04:05"), maxBpXa, pr.NormalizedUnit)
		} else if pr.NormalizedConcentration > 0 && peakConc > pr.NormalizedConcentration*10 {
			isAnomaly = true
			anomalyReason = fmt.Sprintf("Extreme spike detected: Peak %.1f %s is %.1fx higher than window average (%.1f %s)",
				peakConc, pr.NormalizedUnit, peakConc/pr.NormalizedConcentration, pr.NormalizedConcentration, pr.NormalizedUnit)
		}

		pr.PeakConcentration = peakConc
		pr.PeakTimestamp = peakTime
		pr.PeakBucketIdentifier = peakIdent
		pr.MinConcentration = minConc

		auditDetails := model.ISPUAuditDetails{
			CalculationResultID:   res.ID,
			PollutantCode:         pr.PollutantCode,
			ParameterID:           pr.ParameterID,
			PeakConcentration:     peakConc,
			PeakTimestamp:         peakTime,
			PeakBucketIdentifier:  peakIdent,
			MinConcentration:      minConc,
			MinTimestamp:          minTime,
			AvgConcentration:      pr.NormalizedConcentration,
			IsAnomaly:             isAnomaly || excludedCount > 0,
			AnomalyReason:         anomalyReason,
			RegulatoryXa:          maxBpXa,
			SourceUnit:            srcUnit,
			NormalizedUnit:        pr.NormalizedUnit,
			TotalBuckets:          len(aggStats),
			ExcludedOutliersCount: excludedCount,
			TopSpikes:             topSpikes,
		}

		if b, err := json.Marshal(auditDetails); err == nil {
			pr.DetailsJSON = string(b)
		}

		_ = s.db.WithContext(ctx).Model(&model.ISPUParameterResult{}).
			Where("id = ?", pr.ID).
			Updates(map[string]interface{}{
				"peak_concentration":     pr.PeakConcentration,
				"peak_timestamp":         pr.PeakTimestamp,
				"peak_bucket_identifier": pr.PeakBucketIdentifier,
				"min_concentration":      pr.MinConcentration,
				"details_json":           pr.DetailsJSON,
			})
	}
}

// GetAuditTracking returns outlier details, peak timing, and top spikes for an ISPU calculation and pollutant.
func (s *ISPUService) GetAuditTracking(ctx context.Context, calcResultID uint, pollutantCode string) (*model.ISPUAuditDetails, error) {
	res, err := s.ispuRepo.GetByID(ctx, calcResultID)
	if err != nil {
		return nil, fmt.Errorf("ISPU calculation record %d not found: %w", calcResultID, err)
	}

	s.enrichCalculationAuditTracking(ctx, res)

	var targetPR *model.ISPUParameterResult
	if pollutantCode != "" {
		for i := range res.ParameterResults {
			if string(res.ParameterResults[i].PollutantCode) == pollutantCode {
				targetPR = &res.ParameterResults[i]
				break
			}
		}
	}

	if targetPR == nil && len(res.ParameterResults) > 0 {
		for i := range res.ParameterResults {
			if res.ParameterResults[i].PollutantCode == res.DominantPollutant {
				targetPR = &res.ParameterResults[i]
				break
			}
		}
		if targetPR == nil {
			targetPR = &res.ParameterResults[0]
		}
	}

	if targetPR == nil {
		return nil, fmt.Errorf("no parameter results found for calculation %d", calcResultID)
	}

	var audit model.ISPUAuditDetails
	if targetPR.DetailsJSON != "" {
		_ = json.Unmarshal([]byte(targetPR.DetailsJSON), &audit)
	}

	audit.CalculationResultID = res.ID
	audit.PollutantCode = targetPR.PollutantCode
	audit.ParameterID = targetPR.ParameterID
	if audit.PeakConcentration == 0 {
		audit.PeakConcentration = targetPR.PeakConcentration
	}
	if audit.PeakTimestamp == nil {
		audit.PeakTimestamp = targetPR.PeakTimestamp
	}
	if audit.PeakBucketIdentifier == "" {
		audit.PeakBucketIdentifier = targetPR.PeakBucketIdentifier
	}
	if audit.MinConcentration == 0 {
		audit.MinConcentration = targetPR.MinConcentration
	}
	if audit.AvgConcentration == 0 {
		audit.AvgConcentration = targetPR.NormalizedConcentration
	}
	if audit.NormalizedUnit == "" {
		audit.NormalizedUnit = targetPR.NormalizedUnit
	}
	if audit.SourceUnit == "" {
		audit.SourceUnit = targetPR.SourceUnit
	}
	if audit.RegulatoryXa == 0 {
		if bps, ok := OfficialBreakpoints[targetPR.PollutantCode]; ok && len(bps) > 0 {
			audit.RegulatoryXa = bps[len(bps)-1].Xa
		}
	}

	return &audit, nil
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

// GetMappings retrieves station parameter mappings with auto-detection of missing pollutants from all enabled devices.
func (s *ISPUService) GetMappings(ctx context.Context, deviceID uint) ([]model.ISPUParameterMapping, error) {
	mappings, err := s.ispuRepo.GetMappingsByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Auto-detect missing official pollutants from all enabled devices
	mappedPollutants := make(map[model.ISPUParameter]bool)
	for _, m := range mappings {
		mappedPollutants[m.PollutantCode] = true
	}

	var allParams []model.Parameter
	pQuery := s.db.WithContext(ctx).
		Preload("Device").
		Where("enabled = true AND device_id IN (SELECT id FROM devices WHERE enabled = true AND status = 'ACTIVE')")
	if deviceID > 0 {
		pQuery = pQuery.Where("device_id = ?", deviceID)
	}
	if err := pQuery.Find(&allParams).Error; err == nil {
		for _, p := range allParams {
			pollutant, unit, conv := DetectOfficialPollutant(p.ParameterCode, p.ParameterName, p.Unit)
			if pollutant == "" || mappedPollutants[pollutant] {
				continue
			}

			// Automatically create mapping for detected pollutant
			autoMapping := model.ISPUParameterMapping{
				DeviceID:         p.DeviceID,
				ParameterID:      p.ID,
				PollutantCode:    pollutant,
				SourceUnit:       unit,
				NormalizedUnit:   "µg/m³",
				ConversionFactor: conv,
				ConversionRule:   "DIRECT",
				MinCompleteness:  75.0,
				DailyStartHour:   0,
				Enabled:          true,
			}
			if err := s.db.WithContext(ctx).Create(&autoMapping).Error; err == nil {
				paramCopy := p
				autoMapping.Parameter = &paramCopy
				var dev model.Device
				if s.db.WithContext(ctx).First(&dev, p.DeviceID).Error == nil {
					autoMapping.Device = &dev
				}
				mappings = append(mappings, autoMapping)
				mappedPollutants[pollutant] = true
			}
		}
	}

	return mappings, nil
}

// SaveMapping creates or updates an ISPU parameter mapping with audit logging.
func (s *ISPUService) SaveMapping(ctx context.Context, mapping *model.ISPUParameterMapping, actor, ipAddress string) error {
	action := "ISPU_MAPPING_CREATE"
	if mapping.ID > 0 {
		action = "ISPU_MAPPING_UPDATE"
	}

	// Resolve DeviceID from ParameterID if missing
	if mapping.DeviceID == 0 && mapping.ParameterID > 0 {
		var param model.Parameter
		if err := s.db.WithContext(ctx).First(&param, mapping.ParameterID).Error; err == nil {
			mapping.DeviceID = param.DeviceID
		}
	}

	// Single global mapping per pollutant: if an existing mapping for this pollutant exists, update it
	if mapping.ID == 0 {
		var existing model.ISPUParameterMapping
		if err := s.db.WithContext(ctx).Where("pollutant_code = ?", mapping.PollutantCode).First(&existing).Error; err == nil {
			mapping.ID = existing.ID
			action = "ISPU_MAPPING_UPDATE"
		}
	}

	if err := s.ispuRepo.SaveMapping(ctx, mapping); err != nil {
		return err
	}

	// Ensure 120s aggregation definition exists for this parameter
	var aggDefCount int64
	s.db.WithContext(ctx).Model(&model.AggregationDefinition{}).
		Where("device_id = ? AND parameter_id = ?", mapping.DeviceID, mapping.ParameterID).
		Count(&aggDefCount)
	if aggDefCount == 0 {
		var p model.Parameter
		if err := s.db.WithContext(ctx).First(&p, mapping.ParameterID).Error; err == nil {
			_ = s.db.WithContext(ctx).Create(&model.AggregationDefinition{
				Name:               fmt.Sprintf("ISPU Rollup 2m - %s", p.ParameterName),
				Code:               fmt.Sprintf("AGG_%d_%s_120S", mapping.DeviceID, mapping.PollutantCode),
				SourceType:         model.SourceCustomerProcessed,
				DeviceID:           mapping.DeviceID,
				ParameterID:        mapping.ParameterID,
				Function:           model.FunctionAvg,
				IntervalSeconds:    120,
				Timezone:           "Asia/Jakarta",
				Enabled:            true,
				QualityPolicy:      "STANDARD",
				IdentifierFormat:   "YYYYMMDDHHmmss",
				DestinationType:    "INTERNAL",
				GracePeriodSeconds: 5,
			}).Error
		}
	}

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  actor,
			Action:    action,
			Resource:  fmt.Sprintf("station_%d_pollutant_%s", mapping.DeviceID, mapping.PollutantCode),
			IPAddress: ipAddress,
			Details: fmt.Sprintf("ISPU mapping saved: Station=%d, Param=%d, Pollutant=%s, Unit=%s, Factor=%.3f, DailyStartHour=%d, Enabled=%v",
				mapping.DeviceID, mapping.ParameterID, mapping.PollutantCode, mapping.SourceUnit, mapping.ConversionFactor, mapping.DailyStartHour, mapping.Enabled),
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

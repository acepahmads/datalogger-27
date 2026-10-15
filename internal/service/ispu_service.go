package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"

	"gorm.io/gorm"
)

// ISPUService coordinates regulatory ISPU calculations, data window validation, MariaDB persistence, and exports.
type ISPUService struct {
	db         *gorm.DB
	ispuRepo   *repository.ISPURepository
	systemRepo *repository.SystemRepository
	calculator *ISPUCalculator
	hub        *websocket.Hub
}

// NewISPUService instantiates an ISPUService.
func NewISPUService(db *gorm.DB, ispuRepo *repository.ISPURepository, systemRepo *repository.SystemRepository) *ISPUService {
	return &ISPUService{
		db:         db,
		ispuRepo:   ispuRepo,
		systemRepo: systemRepo,
		calculator: NewISPUCalculator(),
	}
}

// SetHub updates or sets the WebSocket hub for real-time calculation broadcasting.
func (s *ISPUService) SetHub(hub *websocket.Hub) {
	s.hub = hub
}

// CalculateStationWindow performs a 24-hour regulatory ISPU evaluation for a station and persists the result.
func (s *ISPUService) CalculateStationWindow(ctx context.Context, deviceID uint, targetEnd time.Time) (*model.ISPUCalculationResult, error) {
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

	// Anchor periodEnd in station local time
	periodEnd := targetEnd.In(loc)
	if periodEnd.IsZero() {
		periodEnd = time.Now().In(loc)
	}
	periodStart := periodEnd.Add(-24 * time.Hour)

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

	var paramInputs []ParameterCalculationInput
	var paramOutputs []ParameterCalculationOutput

	for _, m := range activeMappings {
		input := ParameterCalculationInput{
			PollutantCode:    m.PollutantCode,
			ParameterID:      m.ParameterID,
			SourceUnit:       m.SourceUnit,
			ConversionFactor: m.ConversionFactor,
			ConversionRule:   m.ConversionRule,
			ExpectedSamples:  24, // 24 hourly periods in regulatory 24h window
			MinCompleteness:  m.MinCompleteness,
		}

		// 1. Query AggregationResult (per 2-min / configured interval rollups) within the 24-hour window excluding BAD quality
		var aggStats []struct {
			Val float64
			Rec time.Time
		}
		s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Select("COALESCE(avg_value, value) AS val, period_end AS rec").
			Where("device_id = ? AND parameter_id = ? AND period_end >= ? AND period_start <= ? AND quality != ?",
				deviceID, m.ParameterID, periodStart.UTC(), periodEnd.UTC(), model.QualityBad).
			Order("period_end ASC").
			Find(&aggStats)

		var latestAggVal float64
		hasLatestAgg := false

		if len(aggStats) > 0 {
			// Bucket into 24 1-hour slots to evaluate temporal coverage (Permen LHK 14/2020: min 75% hours)
			hoursWithData := make(map[int]bool)
			var totalSum float64

			for _, as := range aggStats {
				hoursDiff := int(periodEnd.Sub(as.Rec).Hours())
				if hoursDiff >= 0 && hoursDiff < 24 {
					hoursWithData[hoursDiff] = true
					totalSum += as.Val
				}
			}

			input.ValidSamples = len(hoursWithData)
			input.RawConcentration = totalSum / float64(len(aggStats))

			// The latest bucket represents the instantaneous aggregation value (e.g. latest 2-minute rollup)
			latestAggVal = aggStats[len(aggStats)-1].Val
			hasLatestAgg = true
		} else {
			// Fallback to RawData if no aggregation results exist yet (e.g. freshly started station or unit tests)
			var rawStats []struct {
				Val float64
				Rec time.Time
			}
			s.db.WithContext(ctx).Model(&model.RawData{}).
				Select("processed_value AS val, received_at AS rec").
				Where("device_id = ? AND parameter_id = ? AND received_at >= ? AND received_at <= ? AND quality != ?",
					deviceID, m.ParameterID, periodStart.UTC(), periodEnd.UTC(), model.QualityBad).
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

				input.ValidSamples = len(hoursWithData)
				input.RawConcentration = totalSum / float64(len(rawStats))
				latestAggVal = rawStats[len(rawStats)-1].Val
				hasLatestAgg = true
			} else {
				input.ValidSamples = 0
				input.RawConcentration = 0
			}
		}

		paramInputs = append(paramInputs, input)

		// Calculate 24h regulatory ISPU
		out := s.calculator.CalculateParameter(input)

		// Calculate instantaneous per-aggregation ISPU for this parameter
		if hasLatestAgg {
			normAgg, _, _ := s.calculator.NormalizeConcentration(latestAggVal, m.SourceUnit, m.ConversionRule, m.ConversionFactor)
			aggOut := s.calculator.CalculateParameter(ParameterCalculationInput{
				PollutantCode:    m.PollutantCode,
				ParameterID:      m.ParameterID,
				SourceUnit:       m.SourceUnit,
				RawConcentration: latestAggVal,
				ConversionFactor: m.ConversionFactor,
				ConversionRule:   m.ConversionRule,
				ValidSamples:     1,
				ExpectedSamples:  1,
				MinCompleteness:  0,
			})
			out.AggregationConcentration = normAgg
			out.AggregationISPU = aggOut.CalculatedISPU
			out.AggregationCategory = aggOut.Category
		} else {
			out.AggregationCategory = model.CategoryUnavailable
		}

		paramOutputs = append(paramOutputs, out)
	}

	// Calculate overall station ISPU and dominant pollutant (24-hour regulatory)
	minComp := 75.0
	if len(activeMappings) > 0 && activeMappings[0].MinCompleteness > 0 {
		minComp = activeMappings[0].MinCompleteness
	}
	overall := s.calculator.CalculateOverall(paramOutputs, minComp)

	// Determine overall station instantaneous per-aggregation ISPU (highest among pollutants)
	var maxAggISPU float64 = -1
	maxAggCat := model.CategoryUnavailable
	var maxAggDominant model.ISPUParameter

	for _, po := range paramOutputs {
		if po.AggregationISPU > maxAggISPU {
			maxAggISPU = po.AggregationISPU
			maxAggCat = po.AggregationCategory
			maxAggDominant = po.PollutantCode
		}
	}
	if maxAggISPU < 0 {
		maxAggISPU = 0
	}

	tiedStr := ""
	if len(overall.TiedPollutants) > 0 {
		for i, p := range overall.TiedPollutants {
			if i > 0 {
				tiedStr += ", "
			}
			tiedStr += string(p)
		}
	}

	// Construct DB Result entity
	calcRes := &model.ISPUCalculationResult{
		DeviceID:            deviceID,
		PeriodStart:         periodStart.UTC(),
		PeriodEnd:           periodEnd.UTC(),
		CalculatedAt:        time.Now().UTC(),
		OverallISPU:         overall.OverallISPU,
		OverallCategory:     overall.Category,
		DominantPollutant:   overall.DominantPollutant,
		TiedPollutants:      tiedStr,
		CompletenessPercent: overall.CompletenessPercent,
		CalculationStatus:   overall.CalculationStatus,
		Method:              "PERMEN_LHK_14_2020",
		Timezone:            tzName,
		AggregationISPU:     maxAggISPU,
		AggregationCategory: maxAggCat,
		AggregationDominant: maxAggDominant,
		Notes:               overall.Notes,
	}

	for _, po := range paramOutputs {
		calcRes.ParameterResults = append(calcRes.ParameterResults, model.ISPUParameterResult{
			PollutantCode:            po.PollutantCode,
			ParameterID:              po.ParameterID,
			Concentration:            po.RawConcentration,
			SourceUnit:               po.SourceUnit,
			NormalizedConcentration:  po.NormalizedConcentration,
			NormalizedUnit:           po.NormalizedUnit,
			CalculatedISPU:           po.CalculatedISPU,
			Category:                 po.Category,
			AggregationConcentration: po.AggregationConcentration,
			AggregationISPU:          po.AggregationISPU,
			AggregationCategory:      po.AggregationCategory,
			Ia:                       po.Ia,
			Ib:                       po.Ib,
			Xa:                       po.Xa,
			Xb:                       po.Xb,
			CompletenessPercent:      po.CompletenessPercent,
			ValidSamples:             po.ValidSamples,
			TotalExpectedSamples:     po.ExpectedSamples,
			Formula:                  po.Formula,
		})
	}

	// Persist idempotently
	if err := s.ispuRepo.UpsertCalculationResult(ctx, calcRes); err != nil {
		return nil, fmt.Errorf("failed to persist ISPU calculation result: %w", err)
	}

	logger.Info("ISPU calculated for station %d [%s]: ISPU=%.1f (%s, Dominant=%s, Status=%s)",
		deviceID, dev.DeviceCode, calcRes.OverallISPU, calcRes.OverallCategory, calcRes.DominantPollutant, calcRes.CalculationStatus)

	// Broadcast real-time ISPU calculation over WebSocket
	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "ispu.calculated",
			Timestamp: time.Now().UTC(),
			Data:      calcRes,
		})
	}

	return calcRes, nil
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
	// Step hourly backwards from current hour down to lookbackHours
	for i := lookbackHours; i >= 0; i-- {
		targetHour := currentHour.Add(-time.Duration(i) * time.Hour)
		if _, err := s.CalculateStationWindow(ctx, deviceID, targetHour); err == nil {
			processed++
		}
	}

	logger.Info("ISPU trend backfill completed for station %d [%s]: %d/%d hourly windows generated",
		deviceID, dev.DeviceCode, processed, lookbackHours+1)

	return processed, nil
}

// GetLatestSummary retrieves the latest ISPU summary for a station, calculating on-demand if missing or stale (>1h).
func (s *ISPUService) GetLatestSummary(ctx context.Context, deviceID uint) (*model.ISPUCalculationResult, error) {
	res, err := s.ispuRepo.GetLatestByDeviceID(ctx, deviceID)
	if err == nil && res != nil {
		// If calculation is less than 30 minutes old, return it
		if time.Since(res.CalculatedAt) < 30*time.Minute {
			return res, nil
		}
	}

	// Trigger calculation for now
	calculated, calcErr := s.CalculateStationWindow(ctx, deviceID, time.Now())
	if calcErr == nil && calculated != nil {
		return s.ispuRepo.GetByID(ctx, calculated.ID)
	}

	// Fallback to previous record if calculation fails
	if res != nil {
		return res, nil
	}

	return nil, calcErr
}

// GetTrend retrieves historical calculations for chart rendering across selected time ranges ("24h", "7d", "30d").
func (s *ISPUService) GetTrend(ctx context.Context, deviceID uint, timeRange string) ([]model.ISPUCalculationResult, error) {
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

	trend, err := s.ispuRepo.GetTrend(ctx, deviceID, startTime, now)
	if err != nil {
		return nil, err
	}

	// If fewer than 3 records in trend (e.g. freshly initialized station), backfill hourly trend
	if len(trend) < 3 {
		lookback := 24
		if timeRange == "7d" {
			lookback = 168
		} else if timeRange == "30d" {
			lookback = 720
		}
		_, _ = s.BackfillStationHourlyTrend(ctx, deviceID, lookback)
		if refreshed, err := s.ispuRepo.GetTrend(ctx, deviceID, startTime, now); err == nil && len(refreshed) > 0 {
			trend = refreshed
		} else if len(trend) == 0 {
			if res, err := s.CalculateStationWindow(ctx, deviceID, now); err == nil && res != nil {
				trend = append(trend, *res)
			}
		}
	}

	return trend, nil
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
	t := time.Now()
	if targetTime != nil && !targetTime.IsZero() {
		t = *targetTime
	}

	res, err := s.CalculateStationWindow(ctx, deviceID, t)
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

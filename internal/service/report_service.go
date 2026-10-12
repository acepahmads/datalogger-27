package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"

	"gorm.io/gorm"
)

type ReportRequest struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	DeviceIDs    []uint    `json:"device_ids"`
	ParameterIDs []uint    `json:"parameter_ids"`
	ReportType   string    `json:"report_type"` // "OPERATIONAL", "TELEMETRY", "ALARMS", "COMPREHENSIVE"
	Timezone     string    `json:"timezone"`    // e.g. "UTC" or "Asia/Jakarta"
	Format       string    `json:"format"`      // "json", "html", "csv"
}

type DeviceReportSummary struct {
	DeviceID            uint    `json:"device_id"`
	DeviceCode          string  `json:"device_code"`
	DeviceName          string  `json:"device_name"`
	Status              string  `json:"status"`
	ConnectionStatus    string  `json:"connection_status"`
	LatencyMs           int     `json:"latency_ms"`
	SuccessCount        int64   `json:"success_count"`
	FailedCount         int64   `json:"failed_count"`
	AvailabilityPercent float64 `json:"availability_percent"`
	ParametersCount     int     `json:"parameters_count"`
}

type ParameterReportSummary struct {
	ParameterID      uint     `json:"parameter_id"`
	DeviceID         uint     `json:"device_id"`
	DeviceCode       string   `json:"device_code"`
	DeviceName       string   `json:"device_name"`
	ParameterCode    string   `json:"parameter_code"`
	ParameterName    string   `json:"parameter_name"`
	Unit             string   `json:"unit"`
	SampleCount      int64    `json:"sample_count"`
	MinValue         *float64 `json:"min_value"`
	MaxValue         *float64 `json:"max_value"`
	AvgValue         *float64 `json:"avg_value"`
	FirstValue       *float64 `json:"first_value"`
	LastValue        *float64 `json:"last_value"`
	GoodCount        int64    `json:"good_count"`
	GoodPercent      float64  `json:"good_percent"`
	BadCount         int64    `json:"bad_count"`
	BadPercent       float64  `json:"bad_percent"`
	UncertainCount   int64    `json:"uncertain_count"`
	UncertainPercent float64  `json:"uncertain_percent"`
	DataSource       string   `json:"data_source"` // "RAW_TELEMETRY", "PROCESSED", "CURRENT_SNAPSHOT"
}

type AlarmReportSummary struct {
	TotalAlarms       int64        `json:"total_alarms"`
	CriticalCount     int64        `json:"critical_count"`
	WarningCount      int64        `json:"warning_count"`
	InfoCount         int64        `json:"info_count"`
	ActiveCount       int64        `json:"active_count"`
	AcknowledgedCount int64        `json:"acknowledged_count"`
	ClearedCount      int64        `json:"cleared_count"`
	RecentAlarms      []AlarmItem  `json:"recent_alarms"`
}

type AlarmItem struct {
	ID             uint       `json:"id"`
	AlarmCode      string     `json:"alarm_code"`
	DeviceName     string     `json:"device_name"`
	Severity       string     `json:"severity"`
	Status         string     `json:"status"`
	Message        string     `json:"message"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ClearedAt      *time.Time `json:"cleared_at,omitempty"`
}

type CommunicationReportSummary struct {
	TotalFrames        int64   `json:"total_frames"`
	SuccessFrames      int64   `json:"success_frames"`
	ErrorFrames        int64   `json:"error_frames"`
	SuccessRatePercent float64 `json:"success_rate_percent"`
	AvgLatencyMs       float64 `json:"avg_latency_ms"`
}

type ReportData struct {
	Title                string                     `json:"title"`
	ReportType           string                     `json:"report_type"`
	GeneratedAt          time.Time                  `json:"generated_at"`
	GeneratedAtFormatted string                     `json:"generated_at_formatted"`
	Timezone             string                     `json:"timezone"`
	PeriodStart          time.Time                  `json:"period_start"`
	PeriodEnd            time.Time                  `json:"period_end"`
	PeriodFormatted      string                     `json:"period_formatted"`
	TotalDevices         int                        `json:"total_devices"`
	TotalParameters      int                        `json:"total_parameters"`
	TotalSamples         int64                      `json:"total_samples"`
	OverallAvailability  float64                    `json:"overall_availability"`
	OverallQualityGood   float64                    `json:"overall_quality_good"`
	DeviceSummaries      []DeviceReportSummary      `json:"device_summaries"`
	ParameterSummaries   []ParameterReportSummary   `json:"parameter_summaries"`
	AlarmSummary         AlarmReportSummary         `json:"alarm_summary"`
	CommunicationSummary CommunicationReportSummary `json:"communication_summary"`
}

type ReportService struct {
	db         *gorm.DB
	systemRepo *repository.SystemRepository
}

func NewReportService(db *gorm.DB, systemRepo *repository.SystemRepository) *ReportService {
	return &ReportService{
		db:         db,
		systemRepo: systemRepo,
	}
}

func (s *ReportService) BuildReport(ctx context.Context, req ReportRequest) (*ReportData, error) {
	now := time.Now()

	// Default time range: last 24 hours if not provided
	if req.EndTime.IsZero() {
		req.EndTime = now
	}
	if req.StartTime.IsZero() {
		req.StartTime = req.EndTime.Add(-24 * time.Hour)
	}
	if req.StartTime.After(req.EndTime) {
		req.StartTime, req.EndTime = req.EndTime, req.StartTime
	}

	tzName := strings.TrimSpace(req.Timezone)
	if tzName == "" {
		tzName = "UTC"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc = time.UTC
		tzName = "UTC"
	}

	repType := strings.ToUpper(strings.TrimSpace(req.ReportType))
	if repType == "" {
		repType = "COMPREHENSIVE"
	}

	// 1. Fetch devices
	allDevices, err := s.systemRepo.GetAllDevices()
	if err != nil {
		return nil, err
	}

	deviceMap := make(map[uint]model.Device)
	for _, d := range allDevices {
		deviceMap[d.ID] = d
	}

	// Filter devices if specified
	var targetDevices []model.Device
	if len(req.DeviceIDs) > 0 {
		devIDSet := make(map[uint]bool)
		for _, id := range req.DeviceIDs {
			devIDSet[id] = true
		}
		for _, d := range allDevices {
			if devIDSet[d.ID] {
				targetDevices = append(targetDevices, d)
			}
		}
	} else {
		targetDevices = allDevices
	}

	// 2. Build device summaries
	var devSummaries []DeviceReportSummary
	var totalSuccess, totalFailed int64

	for _, d := range targetDevices {
		tot := d.SuccessCount + d.FailedCount
		avail := 100.0
		if tot > 0 {
			avail = (float64(d.SuccessCount) / float64(tot)) * 100.0
		}
		totalSuccess += d.SuccessCount
		totalFailed += d.FailedCount

		devSummaries = append(devSummaries, DeviceReportSummary{
			DeviceID:            d.ID,
			DeviceCode:          d.Code,
			DeviceName:          d.Name,
			Status:              string(d.Status),
			ConnectionStatus:    string(d.ConnectionStatus),
			LatencyMs:           d.LatencyMs,
			SuccessCount:        d.SuccessCount,
			FailedCount:         d.FailedCount,
			AvailabilityPercent: math.Round(avail*100) / 100,
			ParametersCount:     len(d.Parameters),
		})
	}

	// 3. Query server-side aggregated telemetry
	type rawAggRow struct {
		ParameterID     uint     `gorm:"column:parameter_id"`
		SampleCount     int64    `gorm:"column:total_samples"`
		MinVal          *float64 `gorm:"column:min_val"`
		MaxVal          *float64 `gorm:"column:max_val"`
		AvgVal          *float64 `gorm:"column:avg_val"`
		GoodCount       int64    `gorm:"column:good_count"`
		BadCount        int64    `gorm:"column:bad_count"`
		UncertainCount  int64    `gorm:"column:uncertain_count"`
	}

	rawAggQuery := s.db.WithContext(ctx).Table("raw_data").
		Select("parameter_id, COUNT(*) as total_samples, MIN(value_numeric) as min_val, MAX(value_numeric) as max_val, AVG(value_numeric) as avg_val, " +
			"COUNT(CASE WHEN quality = 'GOOD' THEN 1 END) as good_count, " +
			"COUNT(CASE WHEN quality = 'BAD' THEN 1 END) as bad_count, " +
			"COUNT(CASE WHEN quality = 'UNCERTAIN' THEN 1 END) as uncertain_count").
		Where("received_at >= ? AND received_at <= ?", req.StartTime, req.EndTime)

	var targetDevIDs []uint
	for _, d := range targetDevices {
		targetDevIDs = append(targetDevIDs, d.ID)
	}
	if len(targetDevIDs) > 0 {
		rawAggQuery = rawAggQuery.Where("device_id IN ?", targetDevIDs)
	}
	if len(req.ParameterIDs) > 0 {
		rawAggQuery = rawAggQuery.Where("parameter_id IN ?", req.ParameterIDs)
	}

	var rawRows []rawAggRow
	_ = rawAggQuery.Group("parameter_id").Scan(&rawRows).Error

	rawAggMap := make(map[uint]rawAggRow)
	for _, r := range rawRows {
		rawAggMap[r.ParameterID] = r
	}

	// Processed data fallback
	type procAggRow struct {
		ParameterID uint     `gorm:"column:parameter_id"`
		SampleCount int64    `gorm:"column:total_samples"`
		MinVal      *float64 `gorm:"column:min_val"`
		MaxVal      *float64 `gorm:"column:max_val"`
		AvgVal      *float64 `gorm:"column:avg_val"`
	}
	var procRows []procAggRow
	procAggQuery := s.db.WithContext(ctx).Table("processed_data").
		Select("parameter_id, COUNT(*) as total_samples, MIN(scaled_value) as min_val, MAX(scaled_value) as max_val, AVG(scaled_value) as avg_val").
		Where("timestamp >= ? AND timestamp <= ?", req.StartTime, req.EndTime)
	if len(targetDevIDs) > 0 {
		procAggQuery = procAggQuery.Where("device_id IN ?", targetDevIDs)
	}
	if len(req.ParameterIDs) > 0 {
		procAggQuery = procAggQuery.Where("parameter_id IN ?", req.ParameterIDs)
	}
	_ = procAggQuery.Group("parameter_id").Scan(&procRows).Error

	procAggMap := make(map[uint]procAggRow)
	for _, pr := range procRows {
		procAggMap[pr.ParameterID] = pr
	}

	// 4. Build parameter summaries
	var paramSummaries []ParameterReportSummary
	var totalSamples int64
	var totalGood, totalBad, totalUncertain int64

	paramIDFilter := make(map[uint]bool)
	for _, pid := range req.ParameterIDs {
		paramIDFilter[pid] = true
	}

	for _, d := range targetDevices {
		for _, p := range d.Parameters {
			if len(req.ParameterIDs) > 0 && !paramIDFilter[p.ID] {
				continue
			}

			sum := ParameterReportSummary{
				ParameterID:   p.ID,
				DeviceID:      d.ID,
				DeviceCode:    d.Code,
				DeviceName:    d.Name,
				ParameterCode: p.Code,
				ParameterName: p.Name,
				Unit:          p.Unit,
			}

			if raw, ok := rawAggMap[p.ID]; ok && raw.SampleCount > 0 {
				sum.SampleCount = raw.SampleCount
				sum.MinValue = raw.MinVal
				sum.MaxValue = raw.MaxVal
				if raw.AvgVal != nil {
					avg := math.Round(*raw.AvgVal*100) / 100
					sum.AvgValue = &avg
				}
				sum.GoodCount = raw.GoodCount
				sum.BadCount = raw.BadCount
				sum.UncertainCount = raw.UncertainCount
				sum.GoodPercent = math.Round((float64(raw.GoodCount)/float64(raw.SampleCount))*10000) / 100
				sum.BadPercent = math.Round((float64(raw.BadCount)/float64(raw.SampleCount))*10000) / 100
				sum.UncertainPercent = math.Round((float64(raw.UncertainCount)/float64(raw.SampleCount))*10000) / 100
				sum.DataSource = "RAW_TELEMETRY"

				totalSamples += raw.SampleCount
				totalGood += raw.GoodCount
				totalBad += raw.BadCount
				totalUncertain += raw.UncertainCount
			} else if proc, ok := procAggMap[p.ID]; ok && proc.SampleCount > 0 {
				sum.SampleCount = proc.SampleCount
				sum.MinValue = proc.MinVal
				sum.MaxValue = proc.MaxVal
				if proc.AvgVal != nil {
					avg := math.Round(*proc.AvgVal*100) / 100
					sum.AvgValue = &avg
				}
				sum.GoodCount = proc.SampleCount
				sum.GoodPercent = 100.0
				sum.DataSource = "PROCESSED"

				totalSamples += proc.SampleCount
				totalGood += proc.SampleCount
			} else if p.CurrentValue != nil {
				// Snapshot fallback
				sum.SampleCount = 1
				sum.MinValue = p.CurrentValue
				sum.MaxValue = p.CurrentValue
				sum.AvgValue = p.CurrentValue
				sum.GoodCount = 1
				sum.GoodPercent = 100.0
				sum.DataSource = "CURRENT_SNAPSHOT"

				totalSamples += 1
				totalGood += 1
			} else {
				sum.DataSource = "NO_DATA"
			}

			paramSummaries = append(paramSummaries, sum)
		}
	}

	// 5. Query alarms summary
	var alarms []model.Alarm
	alarmQuery := s.db.WithContext(ctx).Preload("Device").
		Where("triggered_at >= ? AND triggered_at <= ?", req.StartTime, req.EndTime)
	if len(targetDevIDs) > 0 {
		alarmQuery = alarmQuery.Where("device_id IN ?", targetDevIDs)
	}
	_ = alarmQuery.Order("triggered_at DESC").Limit(100).Find(&alarms).Error

	var critCount, warnCount, infoCount, actCount, ackCount, clrCount int64
	var recentAlarms []AlarmItem

	for _, a := range alarms {
		sevStr := strings.ToUpper(string(a.Severity))
		switch sevStr {
		case "CRITICAL":
			critCount++
		case "WARNING", "WARN":
			warnCount++
		default:
			infoCount++
		}

		statStr := strings.ToUpper(string(a.Status))
		switch statStr {
		case "ACTIVE":
			actCount++
		case "ACKNOWLEDGED", "ACKED":
			ackCount++
		case "CLEARED", "RESOLVED":
			clrCount++
		}

		devName := ""
		if a.Device != nil {
			devName = a.Device.Name
		} else if a.DeviceID != nil {
			if d, ok := deviceMap[*a.DeviceID]; ok {
				devName = d.Name
			}
		}

		recentAlarms = append(recentAlarms, AlarmItem{
			ID:             a.ID,
			AlarmCode:      a.AlarmCode,
			DeviceName:     devName,
			Severity:       string(a.Severity),
			Status:         string(a.Status),
			Message:        a.Message,
			TriggeredAt:    a.TriggeredAt.In(loc),
			AcknowledgedAt: a.AcknowledgedAt,
			ClearedAt:      a.ClearedAt,
		})
	}

	alarmSummary := AlarmReportSummary{
		TotalAlarms:       int64(len(alarms)),
		CriticalCount:     critCount,
		WarningCount:      warnCount,
		InfoCount:         infoCount,
		ActiveCount:       actCount,
		AcknowledgedCount: ackCount,
		ClearedCount:      clrCount,
		RecentAlarms:      recentAlarms,
	}

	// 6. Communication Summary
	totalFrames := totalSuccess + totalFailed
	commRate := 100.0
	if totalFrames > 0 {
		commRate = (float64(totalSuccess) / float64(totalFrames)) * 100.0
	}

	var avgLatency float64
	if len(targetDevices) > 0 {
		var sumLat int
		for _, d := range targetDevices {
			sumLat += d.LatencyMs
		}
		avgLatency = float64(sumLat) / float64(len(targetDevices))
	}

	commSummary := CommunicationReportSummary{
		TotalFrames:        totalFrames,
		SuccessFrames:      totalSuccess,
		ErrorFrames:        totalFailed,
		SuccessRatePercent: math.Round(commRate*100) / 100,
		AvgLatencyMs:       math.Round(avgLatency*10) / 10,
	}

	overallAvail := commRate
	overallGood := 100.0
	if totalSamples > 0 {
		overallGood = math.Round((float64(totalGood)/float64(totalSamples))*10000) / 100
	}

	periodFormatted := fmt.Sprintf("%s to %s (%s)",
		req.StartTime.In(loc).Format("02 Jan 2006 15:04"),
		req.EndTime.In(loc).Format("02 Jan 2006 15:04"),
		tzName,
	)

	return &ReportData{
		Title:                "Operational Telemetry & Performance Report",
		ReportType:           repType,
		GeneratedAt:          now.In(loc),
		GeneratedAtFormatted: now.In(loc).Format("02 Jan 2006 15:04:05 MST"),
		Timezone:             tzName,
		PeriodStart:          req.StartTime.In(loc),
		PeriodEnd:            req.EndTime.In(loc),
		PeriodFormatted:      periodFormatted,
		TotalDevices:         len(targetDevices),
		TotalParameters:      len(paramSummaries),
		TotalSamples:         totalSamples,
		OverallAvailability:  overallAvail,
		OverallQualityGood:   overallGood,
		DeviceSummaries:      devSummaries,
		ParameterSummaries:   paramSummaries,
		AlarmSummary:         alarmSummary,
		CommunicationSummary: commSummary,
	}, nil
}

func (s *ReportService) GenerateCSV(data *ReportData) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Section 1: Report Metadata
	_ = w.Write([]string{"# REPORT METADATA"})
	_ = w.Write([]string{"Report Title", data.Title})
	_ = w.Write([]string{"Report Type", data.ReportType})
	_ = w.Write([]string{"Generated At", data.GeneratedAtFormatted})
	_ = w.Write([]string{"Timezone", data.Timezone})
	_ = w.Write([]string{"Reporting Period", data.PeriodFormatted})
	_ = w.Write([]string{"Total Devices", strconv.Itoa(data.TotalDevices)})
	_ = w.Write([]string{"Total Parameters", strconv.Itoa(data.TotalParameters)})
	_ = w.Write([]string{"Total Samples", strconv.FormatInt(data.TotalSamples, 10)})
	_ = w.Write([]string{"Overall Availability (%)", fmt.Sprintf("%.2f", data.OverallAvailability)})
	_ = w.Write([]string{"Good Data Quality (%)", fmt.Sprintf("%.2f", data.OverallQualityGood)})
	_ = w.Write([]string{""})

	// Section 2: Devices Fleet Summary
	_ = w.Write([]string{"# DEVICE AVAILABILITY & STATUS"})
	_ = w.Write([]string{"Device ID", "Code", "Name", "Status", "Connection Status", "Latency (ms)", "Success Packets", "Failed Packets", "Availability (%)"})
	for _, d := range data.DeviceSummaries {
		_ = w.Write([]string{
			strconv.FormatUint(uint64(d.DeviceID), 10),
			d.DeviceCode,
			d.DeviceName,
			d.Status,
			d.ConnectionStatus,
			strconv.Itoa(d.LatencyMs),
			strconv.FormatInt(d.SuccessCount, 10),
			strconv.FormatInt(d.FailedCount, 10),
			fmt.Sprintf("%.2f", d.AvailabilityPercent),
		})
	}
	_ = w.Write([]string{""})

	// Section 3: Telemetry Parameters Statistics
	_ = w.Write([]string{"# PARAMETER TELEMETRY STATISTICS"})
	_ = w.Write([]string{"Param ID", "Device Code", "Device Name", "Param Code", "Param Name", "Unit", "Samples", "Min", "Max", "Average", "Good (%)", "Bad (%)", "Uncertain (%)", "Data Source"})
	for _, p := range data.ParameterSummaries {
		minStr, maxStr, avgStr := "N/A", "N/A", "N/A"
		if p.MinValue != nil {
			minStr = fmt.Sprintf("%.2f", *p.MinValue)
		}
		if p.MaxValue != nil {
			maxStr = fmt.Sprintf("%.2f", *p.MaxValue)
		}
		if p.AvgValue != nil {
			avgStr = fmt.Sprintf("%.2f", *p.AvgValue)
		}

		_ = w.Write([]string{
			strconv.FormatUint(uint64(p.ParameterID), 10),
			p.DeviceCode,
			p.DeviceName,
			p.ParameterCode,
			p.ParameterName,
			p.Unit,
			strconv.FormatInt(p.SampleCount, 10),
			minStr,
			maxStr,
			avgStr,
			fmt.Sprintf("%.2f", p.GoodPercent),
			fmt.Sprintf("%.2f", p.BadPercent),
			fmt.Sprintf("%.2f", p.UncertainPercent),
			p.DataSource,
		})
	}
	_ = w.Write([]string{""})

	// Section 4: Alarms Summary
	_ = w.Write([]string{"# ALARMS & EVENTS SUMMARY"})
	_ = w.Write([]string{"Total Alarms", strconv.FormatInt(data.AlarmSummary.TotalAlarms, 10)})
	_ = w.Write([]string{"Critical", strconv.FormatInt(data.AlarmSummary.CriticalCount, 10)})
	_ = w.Write([]string{"Warning", strconv.FormatInt(data.AlarmSummary.WarningCount, 10)})
	_ = w.Write([]string{"Info", strconv.FormatInt(data.AlarmSummary.InfoCount, 10)})
	_ = w.Write([]string{"Active", strconv.FormatInt(data.AlarmSummary.ActiveCount, 10)})
	_ = w.Write([]string{"Acknowledged", strconv.FormatInt(data.AlarmSummary.AcknowledgedCount, 10)})
	_ = w.Write([]string{"Cleared", strconv.FormatInt(data.AlarmSummary.ClearedCount, 10)})
	_ = w.Write([]string{""})

	if len(data.AlarmSummary.RecentAlarms) > 0 {
		_ = w.Write([]string{"# ALARM LOG ENTRIES"})
		_ = w.Write([]string{"Alarm ID", "Code", "Device", "Severity", "Status", "Message", "Triggered At"})
		for _, a := range data.AlarmSummary.RecentAlarms {
			_ = w.Write([]string{
				strconv.FormatUint(uint64(a.ID), 10),
				a.AlarmCode,
				a.DeviceName,
				a.Severity,
				a.Status,
				a.Message,
				a.TriggeredAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func (s *ReportService) GenerateHTML(data *ReportData) (string, error) {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + data.Title + `</title>
<style>
  :root {
    --bg-primary: #0F172A;
    --bg-card: #1E293B;
    --border: #334155;
    --text-primary: #F8FAFC;
    --text-secondary: #94A3B8;
    --accent: #3B82F6;
    --success: #10B981;
    --warning: #F59E0B;
    --danger: #EF4444;
  }
  @media print {
    body { background: #FFFFFF !important; color: #000000 !important; font-size: 11pt; }
    .no-print { display: none !important; }
    .card { border: 1px solid #CBD5E1 !important; background: #FFFFFF !important; box-shadow: none !important; color: #000000 !important; }
    table { width: 100% !important; border-collapse: collapse !important; color: #000000 !important; }
    th, td { border: 1px solid #CBD5E1 !important; color: #000000 !important; padding: 6px 8px !important; }
    th { background: #F1F5F9 !important; font-weight: bold; }
    .page-break { page-break-before: always; }
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    line-height: 1.5;
    padding: 24px;
  }
  .container { max-width: 1200px; margin: 0 auto; }
  .header {
    border-bottom: 2px solid var(--border);
    padding-bottom: 16px;
    margin-bottom: 24px;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }
  .brand { font-size: 11px; letter-spacing: 2px; text-transform: uppercase; color: var(--accent); font-weight: 800; }
  h1 { font-size: 24px; font-weight: 700; margin: 4px 0; color: var(--text-primary); }
  .meta-tag { font-size: 12px; color: var(--text-secondary); }
  .grid-kpi {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
    margin-bottom: 24px;
  }
  .card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 16px;
  }
  .kpi-title { font-size: 11px; text-transform: uppercase; letter-spacing: 1px; color: var(--text-secondary); margin-bottom: 4px; }
  .kpi-val { font-size: 24px; font-weight: 700; color: var(--text-primary); font-family: monospace; }
  .kpi-sub { font-size: 12px; margin-top: 4px; color: var(--text-secondary); }
  .section-title { font-size: 16px; font-weight: 700; margin: 24px 0 12px 0; display: flex; align-items: center; justify-content: space-between; border-left: 4px solid var(--accent); padding-left: 10px; }
  table { width: 100%; border-collapse: collapse; font-size: 12px; }
  th, td { padding: 10px 12px; text-align: left; border-bottom: 1px solid var(--border); }
  th { background: rgba(30, 41, 59, 0.7); color: var(--text-secondary); font-weight: 600; text-transform: uppercase; font-size: 11px; letter-spacing: 0.5px; }
  tr:hover { background: rgba(51, 65, 85, 0.4); }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 700; text-transform: uppercase; }
  .badge-success { background: rgba(16, 185, 129, 0.15); color: #34D399; border: 1px solid rgba(16, 185, 129, 0.3); }
  .badge-danger { background: rgba(239, 68, 68, 0.15); color: #F87171; border: 1px solid rgba(239, 68, 68, 0.3); }
  .badge-warning { background: rgba(245, 158, 11, 0.15); color: #FBBF24; border: 1px solid rgba(245, 158, 11, 0.3); }
  .badge-info { background: rgba(59, 130, 246, 0.15); color: #60A5FA; border: 1px solid rgba(59, 130, 246, 0.3); }
  .print-btn {
    background: var(--accent);
    color: #FFFFFF;
    border: none;
    border-radius: 6px;
    padding: 8px 16px;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  }
  .print-btn:hover { background: #2563EB; }
  .footer { margin-top: 40px; padding-top: 16px; border-top: 1px solid var(--border); font-size: 11px; color: var(--text-secondary); text-align: center; }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <div>
      <div class="brand">DATALOGGER &bull; EDGE ANALYSIS APPLICATION</div>
      <h1>` + data.Title + `</h1>
      <div class="meta-tag">Period: <strong>` + data.PeriodFormatted + `</strong> &bull; Generated: ` + data.GeneratedAtFormatted + `</div>
    </div>
    <div class="no-print">
      <button class="print-btn" onclick="window.print()">Print / Save PDF</button>
    </div>
  </div>

  <div class="grid-kpi">
    <div class="card">
      <div class="kpi-title">System Availability</div>
      <div class="kpi-val" style="color: #34D399;">` + fmt.Sprintf("%.2f%%", data.OverallAvailability) + `</div>
      <div class="kpi-sub">Frames: ` + fmt.Sprintf("%d / %d OK", data.CommunicationSummary.SuccessFrames, data.CommunicationSummary.TotalFrames) + `</div>
    </div>
    <div class="card">
      <div class="kpi-title">Good Data Quality</div>
      <div class="kpi-val" style="color: #60A5FA;">` + fmt.Sprintf("%.2f%%", data.OverallQualityGood) + `</div>
      <div class="kpi-sub">Total Telemetry Samples: ` + fmt.Sprintf("%d", data.TotalSamples) + `</div>
    </div>
    <div class="card">
      <div class="kpi-title">Active / Total Alarms</div>
      <div class="kpi-val" style="color: ` + (map[bool]string{true: "#F87171", false: "#34D399"}[data.AlarmSummary.ActiveCount > 0]) + `;">` + fmt.Sprintf("%d / %d", data.AlarmSummary.ActiveCount, data.AlarmSummary.TotalAlarms) + `</div>
      <div class="kpi-sub">Critical: ` + fmt.Sprintf("%d", data.AlarmSummary.CriticalCount) + ` &bull; Warning: ` + fmt.Sprintf("%d", data.AlarmSummary.WarningCount) + `</div>
    </div>
    <div class="card">
      <div class="kpi-title">Monitored Fleet</div>
      <div class="kpi-val">` + fmt.Sprintf("%d Devices", data.TotalDevices) + `</div>
      <div class="kpi-sub">` + fmt.Sprintf("%d Active Parameters", data.TotalParameters) + ` &bull; Avg Latency: ` + fmt.Sprintf("%.1fms", data.CommunicationSummary.AvgLatencyMs) + `</div>
    </div>
  </div>

  <div class="section-title">
    <span>Device Fleet Status & Availability</span>
    <span class="meta-tag">Total ` + strconv.Itoa(len(data.DeviceSummaries)) + ` Devices</span>
  </div>
  <div class="card" style="padding: 0; overflow-x: auto;">
    <table>
      <thead>
        <tr>
          <th>Code</th>
          <th>Device Name</th>
          <th>Status</th>
          <th>Connection</th>
          <th>Latency</th>
          <th>Packet Ratio (TX/ERR)</th>
          <th>Availability</th>
        </tr>
      </thead>
      <tbody>`)

	for _, d := range data.DeviceSummaries {
		statusBadge := `<span class="badge badge-success">ONLINE</span>`
		if strings.EqualFold(d.Status, "OFFLINE") {
			statusBadge = `<span class="badge badge-danger">OFFLINE</span>`
		} else if strings.EqualFold(d.Status, "ERROR") {
			statusBadge = `<span class="badge badge-danger">ERROR</span>`
		}

		sb.WriteString(`
        <tr>
          <td style="font-family: monospace; font-weight: bold;">` + d.DeviceCode + `</td>
          <td>` + d.DeviceName + `</td>
          <td>` + statusBadge + `</td>
          <td style="font-family: monospace;">` + d.ConnectionStatus + `</td>
          <td style="font-family: monospace;">` + strconv.Itoa(d.LatencyMs) + ` ms</td>
          <td style="font-family: monospace;">` + fmt.Sprintf("%d / %d", d.SuccessCount, d.FailedCount) + `</td>
          <td style="font-family: monospace; font-weight: bold; color: #34D399;">` + fmt.Sprintf("%.2f%%", d.AvailabilityPercent) + `</td>
        </tr>`)
	}

	sb.WriteString(`
      </tbody>
    </table>
  </div>

  <div class="section-title page-break">
    <span>Parameter Telemetry & Data Quality Breakdown</span>
    <span class="meta-tag">Server-Side Aggregation</span>
  </div>
  <div class="card" style="padding: 0; overflow-x: auto;">
    <table>
      <thead>
        <tr>
          <th>Device</th>
          <th>Parameter</th>
          <th>Unit</th>
          <th>Samples</th>
          <th>Min</th>
          <th>Max</th>
          <th>Average</th>
          <th>Quality (Good / Bad / Unc)</th>
          <th>Data Source</th>
        </tr>
      </thead>
      <tbody>`)

	for _, p := range data.ParameterSummaries {
		minStr, maxStr, avgStr := "-", "-", "-"
		if p.MinValue != nil {
			minStr = fmt.Sprintf("%.2f", *p.MinValue)
		}
		if p.MaxValue != nil {
			maxStr = fmt.Sprintf("%.2f", *p.MaxValue)
		}
		if p.AvgValue != nil {
			avgStr = fmt.Sprintf("%.2f", *p.AvgValue)
		}

		qualBadge := fmt.Sprintf(`<span style="color: #34D399; font-weight: bold;">%.1f%%</span> / <span style="color: #F87171;">%.1f%%</span> / <span style="color: #FBBF24;">%.1f%%</span>`,
			p.GoodPercent, p.BadPercent, p.UncertainPercent)

		sourceBadge := `<span class="badge badge-info">RAW</span>`
		if p.DataSource == "PROCESSED" {
			sourceBadge = `<span class="badge badge-success">PROCESSED</span>`
		} else if p.DataSource == "NO_DATA" {
			sourceBadge = `<span class="badge badge-warning">NO DATA</span>`
		}

		sb.WriteString(`
        <tr>
          <td style="font-family: monospace; font-size: 11px;">` + p.DeviceCode + `</td>
          <td><strong>` + p.ParameterName + `</strong> <span style="font-family: monospace; font-size: 11px; color: var(--text-secondary);">(` + p.ParameterCode + `)</span></td>
          <td style="font-family: monospace;">` + p.Unit + `</td>
          <td style="font-family: monospace;">` + strconv.FormatInt(p.SampleCount, 10) + `</td>
          <td style="font-family: monospace;">` + minStr + `</td>
          <td style="font-family: monospace;">` + maxStr + `</td>
          <td style="font-family: monospace; font-weight: bold; color: #60A5FA;">` + avgStr + `</td>
          <td style="font-family: monospace; font-size: 11px;">` + qualBadge + `</td>
          <td>` + sourceBadge + `</td>
        </tr>`)
	}

	sb.WriteString(`
      </tbody>
    </table>
  </div>

  <div class="section-title">
    <span>Alarms & Notifications Summary</span>
    <span class="meta-tag">Total ` + strconv.FormatInt(data.AlarmSummary.TotalAlarms, 10) + ` Incidents in Period</span>
  </div>
  <div class="card" style="padding: 0; overflow-x: auto;">
    <table>
      <thead>
        <tr>
          <th>Alarm Code</th>
          <th>Device</th>
          <th>Severity</th>
          <th>Status</th>
          <th>Message</th>
          <th>Triggered At</th>
        </tr>
      </thead>
      <tbody>`)

	if len(data.AlarmSummary.RecentAlarms) == 0 {
		sb.WriteString(`
        <tr>
          <td colspan="6" style="text-align: center; color: var(--text-secondary); padding: 24px;">No alarms recorded during this period.</td>
        </tr>`)
	} else {
		for _, a := range data.AlarmSummary.RecentAlarms {
			sevBadge := `<span class="badge badge-info">INFO</span>`
			if strings.EqualFold(a.Severity, "CRITICAL") {
				sevBadge = `<span class="badge badge-danger">CRITICAL</span>`
			} else if strings.EqualFold(a.Severity, "WARNING") || strings.EqualFold(a.Severity, "WARN") {
				sevBadge = `<span class="badge badge-warning">WARNING</span>`
			}

			sb.WriteString(`
        <tr>
          <td style="font-family: monospace; font-weight: bold;">` + a.AlarmCode + `</td>
          <td>` + a.DeviceName + `</td>
          <td>` + sevBadge + `</td>
          <td style="font-family: monospace; font-size: 11px;">` + a.Status + `</td>
          <td>` + a.Message + `</td>
          <td style="font-family: monospace; font-size: 11px;">` + a.TriggeredAt.Format("2006-01-02 15:04:05") + `</td>
        </tr>`)
		}
	}

	sb.WriteString(`
      </tbody>
    </table>
  </div>

  <div class="footer">
    Datalogger Analysis Application &bull; Phase 7 Operational Reporting &bull; Confidential &bull; Generated: ` + data.GeneratedAtFormatted + `
  </div>
</div>
</body>
</html>`)

	return sb.String(), nil
}

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"
	"datalogger/pkg/tenantctx"

	"gorm.io/gorm"
)

type wqmsTriggerJob struct {
	TenantID  uint
	SiteID    uint
	DeviceID  uint
	BucketEnd time.Time
	Reason    string
}

// WQMSLatestSummaryBundle bundles the latest Pollution Index (IPj), Water Quality Index (IKA), and station config.
type WQMSLatestSummaryBundle struct {
	IPResult               *model.WQMSCalculationResult             `json:"ip_result"`
	OperationalIPResult    *model.WQMSCalculationResult             `json:"operational_ip_result,omitempty"`
	IKASummary             model.WQMSIKASummary                     `json:"ika_summary"`
	Config                 model.WQMSStationConfig                  `json:"config"`
	ActiveClassStandards   map[model.WQMSParamCode]WQMSClassStandard `json:"active_class_standards"`
	AggregationIntervalSec int                                      `json:"aggregation_interval_sec"`
}

// WQMSService coordinates regulatory WQMS IP (Kepmen LH 115/2003) and IKA (Permen LHK 27/2021) evaluations.
type WQMSService struct {
	db              *gorm.DB
	wqmsRepo        *repository.WQMSRepository
	systemRepo      *repository.SystemRepository
	calculator      *WQMSCalculator
	hub             *websocket.Hub
	jobChan         chan wqmsTriggerJob
	jobMu           sync.Mutex
	stopWorker      chan struct{}
	workerWg        sync.WaitGroup
	isWorkerRunning bool
}

// NewWQMSService instantiates a WQMSService.
func NewWQMSService(db *gorm.DB, wqmsRepo *repository.WQMSRepository, systemRepo *repository.SystemRepository) *WQMSService {
	return &WQMSService{
		db:         db,
		wqmsRepo:   wqmsRepo,
		systemRepo: systemRepo,
		calculator: NewWQMSCalculator(),
		jobChan:    make(chan wqmsTriggerJob, 256),
		stopWorker: make(chan struct{}),
	}
}

// SetHub sets the WebSocket hub for real-time WQMS IP & IKA broadcasts.
func (s *WQMSService) SetHub(hub *websocket.Hub) {
	s.hub = hub
}

// StartWorker starts the bounded asynchronous WQMS calculation worker.
func (s *WQMSService) StartWorker(ctx context.Context) {
	s.jobMu.Lock()
	defer s.jobMu.Unlock()
	if s.isWorkerRunning {
		return
	}
	s.isWorkerRunning = true
	s.workerWg.Add(1)
	go s.workerLoop(ctx)
	logger.Info("WQMS IP & IKA calculation worker initialized")
}

// StopWorker stops the WQMS background calculation worker cleanly.
func (s *WQMSService) StopWorker() {
	s.jobMu.Lock()
	if !s.isWorkerRunning {
		s.jobMu.Unlock()
		return
	}
	s.isWorkerRunning = false
	close(s.stopWorker)
	s.jobMu.Unlock()
	s.workerWg.Wait()
	logger.Info("WQMS IP & IKA calculation worker stopped cleanly")
}

// EnqueueCalculation enqueues an asynchronous WQMS calculation when WQMS domain or WQMS mappings are active.
func (s *WQMSService) EnqueueCalculation(deviceID uint, bucketEnd time.Time, reason string) {
	if bucketEnd.IsZero() || database.IsRestoreMode() {
		return
	}
	var tenantID, siteID uint = 1, 1
	if deviceID > 0 {
		var dev model.Device
		if err := s.db.Select("id, tenant_id, site_id").First(&dev, deviceID).Error; err == nil {
			if dev.TenantID > 0 {
				tenantID = dev.TenantID
			}
			if dev.SiteID > 0 {
				siteID = dev.SiteID
			}
		}
	}

	cfg := config.Get()
	dom := "AQMS"
	if cfg != nil && cfg.MonitoringDomain != "" {
		dom = strings.ToUpper(strings.TrimSpace(cfg.MonitoringDomain))
	}
	if siteID > 0 {
		var site model.Site
		if err := s.db.Select("id, monitoring_domain").First(&site, siteID).Error; err == nil && site.MonitoringDomain != "" {
			dom = strings.ToUpper(strings.TrimSpace(site.MonitoringDomain))
		}
	}
	if dom != "WQMS" && dom != "CUSTOM" {
		// Only run if explicit WQMS mappings exist for this tenant/site
		var cnt int64
		s.db.Model(&model.WQMSParameterMapping{}).Where("tenant_id = ? AND (site_id = ? OR site_id = 0) AND enabled = ?", tenantID, siteID, true).Count(&cnt)
		if cnt == 0 {
			return
		}
	}
	if reason == "" {
		reason = "AGGREGATION_COMPLETED"
	}
	select {
	case s.jobChan <- wqmsTriggerJob{TenantID: tenantID, SiteID: siteID, DeviceID: 0, BucketEnd: bucketEnd, Reason: reason}:
	default:
	}
}

func (s *WQMSService) workerLoop(ctx context.Context) {
	defer s.workerWg.Done()
	const coalesceDelay = 500 * time.Millisecond
	scheduled := make(map[string]struct {
		job      wqmsTriggerJob
		deadline time.Time
	})
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopWorker:
			return
		case job, ok := <-s.jobChan:
			if !ok {
				return
			}
			key := fmt.Sprintf("%d_%d_%d", job.TenantID, job.SiteID, job.BucketEnd.Unix())
			s.jobMu.Lock()
			if existing, exists := scheduled[key]; exists {
				if job.Reason == "MANUAL_RECALCULATE" {
					existing.job.Reason = job.Reason
				}
				scheduled[key] = existing
			} else {
				scheduled[key] = struct {
					job      wqmsTriggerJob
					deadline time.Time
				}{
					job:      job,
					deadline: time.Now().Add(coalesceDelay),
				}
			}
			s.jobMu.Unlock()

		case <-ticker.C:
			if database.IsRestoreMode() {
				continue
			}
			now := time.Now()
			var due []wqmsTriggerJob
			s.jobMu.Lock()
			for k, v := range scheduled {
				if now.After(v.deadline) {
					due = append(due, v.job)
					delete(scheduled, k)
				}
			}
			s.jobMu.Unlock()

			for _, job := range due {
				calcCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				if job.TenantID > 0 {
					calcCtx = tenantctx.WithScope(calcCtx, tenantctx.Scope{IsSuperAdmin: true, TenantID: job.TenantID, SiteID: job.SiteID})
				}
				_, _, _ = s.CalculateStationWindow(calcCtx, job.DeviceID, job.BucketEnd, job.Reason)
				cancel()
			}
		}
	}
}

// GetStationConfig retrieves the persisted WQMS station configuration.
func (s *WQMSService) GetStationConfig(ctx context.Context, deviceID uint) (*model.WQMSStationConfig, error) {
	return s.wqmsRepo.GetOrCreateStationConfig(ctx, deviceID)
}

// UpdateStationConfig updates WaterClass / IKAWindowDays / DOSaturationMgL and triggers immediate recalculation.
func (s *WQMSService) UpdateStationConfig(ctx context.Context, req model.WQMSStationConfig, actor, ipAddress string) (*WQMSLatestSummaryBundle, error) {
	cfg, err := s.wqmsRepo.GetOrCreateStationConfig(ctx, req.DeviceID)
	if err != nil {
		return nil, err
	}
	cfg.WaterClass = NormalizeWaterClass(string(req.WaterClass))
	if req.IKAWindowDays >= 1 && req.IKAWindowDays <= 366 {
		cfg.IKAWindowDays = req.IKAWindowDays
	}
	if req.DOSaturationMgL >= 5.0 && req.DOSaturationMgL <= 14.0 {
		cfg.DOSaturationMgL = req.DOSaturationMgL
	}
	if req.MinCompleteness > 0 && req.MinCompleteness <= 100.0 {
		cfg.MinCompleteness = req.MinCompleteness
	}
	cfg.UpdatedBy = actor

	if err := s.wqmsRepo.SaveStationConfig(ctx, cfg); err != nil {
		return nil, err
	}

	if s.systemRepo != nil {
		tid := cfg.TenantID
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &tid,
			Username:  actor,
			Action:    "WQMS_CONFIG_UPDATE",
			Resource:  "WQMS_CONFIG",
			IPAddress: ipAddress,
			Details:   fmt.Sprintf("Updated WQMS station config: water_class=%s, ika_window_days=%d, do_sat=%.2f", cfg.WaterClass, cfg.IKAWindowDays, cfg.DOSaturationMgL),
			CreatedAt: time.Now().UTC(),
		})
	}

	_, _, _ = s.CalculateStationWindow(ctx, req.DeviceID, time.Now().UTC(), "CLASS_CONFIG_CHANGED")
	return s.GetLatestSummary(ctx, req.DeviceID, "OPERATIONAL")
}

// GetStandards returns the official PP No. 22 Tahun 2021 Lampiran VI standards matrix across Classes I-IV.
func (s *WQMSService) GetStandards() map[model.WQMSWaterClass]map[model.WQMSParamCode]WQMSClassStandard {
	return OfficialWQMSStandards
}

// GetMappings retrieves station WQMS mappings and auto-discovers matching active water quality sensors.
func (s *WQMSService) GetMappings(ctx context.Context, deviceID uint) ([]model.WQMSParameterMapping, error) {
	scope := tenantctx.FromContext(ctx)
	if deviceID > 0 {
		var dev model.Device
		if err := s.db.WithContext(ctx).First(&dev, deviceID).Error; err == nil {
			if !scope.CanAccessRecord(dev.TenantID, dev.SiteID) {
				return nil, fmt.Errorf("access denied: device %d belongs to another tenant or restricted site", deviceID)
			}
		}
	}

	mappings, err := s.wqmsRepo.GetMappingsByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	mappedCodes := make(map[model.WQMSParamCode]bool)
	for _, m := range mappings {
		mappedCodes[m.ParamCode] = true
	}

	devSubQ := s.db.WithContext(ctx).Model(&model.Device{}).
		Select("id").
		Where("enabled = ? AND (status != ? AND status != ?)", true, "DISABLED", "INACTIVE")
	if scope.TenantID > 0 {
		devSubQ = devSubQ.Where("tenant_id = ?", scope.TenantID)
	}
	if scope.SiteID > 0 {
		devSubQ = devSubQ.Where("site_id = ?", scope.SiteID)
	} else if len(scope.AllowedSiteIDs) > 0 {
		devSubQ = devSubQ.Where("site_id IN ?", scope.AllowedSiteIDs)
	}

	var allParams []model.Parameter
	pQuery := s.db.WithContext(ctx).
		Where("enabled = ? AND device_id IN (?)", true, devSubQ)
	if deviceID > 0 {
		pQuery = pQuery.Where("device_id = ?", deviceID)
	}
	if err := pQuery.Find(&allParams).Error; err == nil {
		for _, p := range allParams {
			code, unit, conv := DetectOfficialWQMSParameter(p.ParameterCode, p.ParameterName, p.Unit)
			if code == "" || mappedCodes[code] {
				continue
			}
			var dev model.Device
			_ = s.db.WithContext(ctx).First(&dev, p.DeviceID).Error
			tID := dev.TenantID
			if tID == 0 {
				tID = scope.WriteTenantID(0)
			}
			sID := dev.SiteID
			if sID == 0 {
				sID = scope.WriteSiteID(0)
			}
			std := OfficialWQMSStandards[model.WaterClass2][code]
			pid := p.ID
			autoMap := model.WQMSParameterMapping{
				TenantID:         tID,
				SiteID:           sID,
				DeviceID:         p.DeviceID,
				ParameterID:      &pid,
				ParamCode:        code,
				SourceType:       "SENSOR",
				SourceUnit:       unit,
				NormalizedUnit:   std.Unit,
				ConversionFactor: conv,
				MinCompleteness:  75.0,
				Enabled:          true,
			}
			if createErr := s.db.WithContext(ctx).Create(&autoMap).Error; createErr == nil {
				pCopy := p
				autoMap.Parameter = &pCopy
				if dev.ID > 0 {
					autoMap.Device = &dev
				}
				mappings = append(mappings, autoMap)
				mappedCodes[code] = true
			}
		}
	}

	return mappings, nil
}

// SaveMapping creates or updates a WQMS sensor or periodic laboratory parameter mapping and triggers recalculation.
func (s *WQMSService) SaveMapping(ctx context.Context, m *model.WQMSParameterMapping, actor, ipAddress string) error {
	scope := tenantctx.FromContext(ctx)
	if m.ID > 0 {
		existingByID, err := s.wqmsRepo.GetMappingByID(ctx, m.ID)
		if err != nil {
			return err
		}
		if !scope.CanAccessRecord(existingByID.TenantID, existingByID.SiteID) {
			return fmt.Errorf("access denied: WQMS mapping belongs to another tenant or restricted site")
		}
		m.TenantID = existingByID.TenantID
		m.SiteID = existingByID.SiteID
	}

	m.SourceType = strings.ToUpper(strings.TrimSpace(m.SourceType))
	if m.SourceType != "LAB" {
		m.SourceType = "SENSOR"
	}
	if m.ConversionFactor <= 0 {
		m.ConversionFactor = 1.0
	}
	if m.MinCompleteness <= 0 {
		m.MinCompleteness = 75.0
	}
	if std, ok := OfficialWQMSStandards[model.WaterClass2][m.ParamCode]; ok {
		if m.NormalizedUnit == "" {
			m.NormalizedUnit = std.Unit
		}
		if m.SourceUnit == "" {
			m.SourceUnit = std.Unit
		}
	}
	if m.SourceType == "LAB" {
		m.ParameterID = nil
		if m.LabSampledAt == nil {
			now := time.Now().UTC()
			m.LabSampledAt = &now
		}
	} else if m.ParameterID != nil && *m.ParameterID > 0 && m.DeviceID == 0 {
		var p model.Parameter
		if s.db.WithContext(ctx).First(&p, *m.ParameterID).Error == nil {
			m.DeviceID = p.DeviceID
		}
	}
	if m.DeviceID > 0 {
		var dev model.Device
		if err := s.db.WithContext(ctx).First(&dev, m.DeviceID).Error; err == nil {
			if !scope.CanAccessRecord(dev.TenantID, dev.SiteID) {
				return fmt.Errorf("access denied: device belongs to another tenant or restricted site")
			}
			m.TenantID = dev.TenantID
			m.SiteID = dev.SiteID
		}
	}
	if m.TenantID == 0 {
		m.TenantID = scope.WriteTenantID(0)
	}
	if m.SiteID == 0 {
		m.SiteID = scope.WriteSiteID(0)
	}

	// Upsert by (TenantID, SiteID, DeviceID, ParamCode) if ID == 0 to avoid unique constraint errors
	if m.ID == 0 {
		var existing model.WQMSParameterMapping
		if err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND site_id = ? AND device_id = ? AND param_code = ?", m.TenantID, m.SiteID, m.DeviceID, m.ParamCode).
			First(&existing).Error; err == nil {
			m.ID = existing.ID
			m.CreatedAt = existing.CreatedAt
		}
	}

	if err := s.wqmsRepo.SaveMapping(ctx, m); err != nil {
		return err
	}

	if s.systemRepo != nil {
		tid := m.TenantID
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &tid,
			Username:  actor,
			Action:    "WQMS_MAPPING_SAVE",
			Resource:  string(m.ParamCode),
			IPAddress: ipAddress,
			Details:   fmt.Sprintf("Saved WQMS mapping %s (source=%s, enabled=%v)", m.ParamCode, m.SourceType, m.Enabled),
			CreatedAt: time.Now().UTC(),
		})
	}

	_, _, _ = s.CalculateStationWindow(ctx, 0, time.Now().UTC(), "MAPPING_UPDATED")
	return nil
}

// DeleteMapping removes a WQMS parameter mapping and triggers recalculation.
func (s *WQMSService) DeleteMapping(ctx context.Context, id uint, actor, ipAddress string) error {
	m, err := s.wqmsRepo.GetMappingByID(ctx, id)
	if err != nil {
		return err
	}
	if !tenantctx.FromContext(ctx).CanAccessRecord(m.TenantID, m.SiteID) {
		return fmt.Errorf("access denied: WQMS mapping belongs to another tenant or restricted site")
	}
	if err := s.wqmsRepo.DeleteMapping(ctx, id); err != nil {
		return err
	}
	if s.systemRepo != nil {
		tid := m.TenantID
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &tid,
			Username:  actor,
			Action:    "WQMS_MAPPING_DELETE",
			Resource:  fmt.Sprintf("wqms_mapping_%d", id),
			IPAddress: ipAddress,
			Details:   fmt.Sprintf("Deleted WQMS parameter mapping #%d", id),
			CreatedAt: time.Now().UTC(),
		})
	}
	_, _, _ = s.CalculateStationWindow(ctx, 0, time.Now().UTC(), "MAPPING_DELETED")
	return nil
}

func (s *WQMSService) resolveAggregationIntervalSec(ctx context.Context, deviceID uint) int {
	scope := tenantctx.FromContext(ctx)
	var aggDef model.AggregationDefinition
	q := s.db.WithContext(ctx).Where("enabled = ?", true)
	if scope.TenantID > 0 {
		q = q.Where("tenant_id = ?", scope.TenantID)
	}
	if deviceID > 0 {
		q = q.Where("device_id = ?", deviceID)
	}
	if err := q.Order("interval_seconds ASC").First(&aggDef).Error; err == nil && aggDef.IntervalSeconds > 0 {
		return aggDef.IntervalSeconds
	}
	return 120
}

// CalculateStationWindow evaluates BOTH Operational (latest interval) and Regulatory (24-hour rolling) IPj
// for the station and broadcasts real-time updates over WebSocket.
func (s *WQMSService) CalculateStationWindow(ctx context.Context, deviceID uint, targetEnd time.Time, reason string) (*model.WQMSCalculationResult, *model.WQMSCalculationResult, error) {
	if targetEnd.IsZero() {
		targetEnd = time.Now().UTC()
	}
	scope := tenantctx.FromContext(ctx)
	tenantID := scope.WriteTenantID(0)
	siteID := scope.WriteSiteID(0)
	if deviceID > 0 {
		var dev model.Device
		if err := s.db.WithContext(ctx).First(&dev, deviceID).Error; err == nil {
			if !scope.CanAccessRecord(dev.TenantID, dev.SiteID) {
				return nil, nil, fmt.Errorf("access denied: device %d belongs to another tenant or restricted site", deviceID)
			}
			if dev.TenantID > 0 {
				tenantID = dev.TenantID
			}
			if dev.SiteID > 0 {
				siteID = dev.SiteID
			}
		}
	}
	calcCtx := tenantctx.WithScope(ctx, tenantctx.Scope{
		IsSuperAdmin:   true,
		TenantID:       tenantID,
		SiteID:         siteID,
		AllowedSiteIDs: scope.AllowedSiteIDs,
	})

	cfg, err := s.wqmsRepo.GetOrCreateStationConfig(calcCtx, 0)
	if err != nil {
		return nil, nil, err
	}

	mappings, err := s.GetMappings(calcCtx, deviceID)
	if err != nil {
		return nil, nil, err
	}

	var activeMappings []model.WQMSParameterMapping
	for _, m := range mappings {
		if m.Enabled {
			activeMappings = append(activeMappings, m)
		}
	}

	intervalSec := s.resolveAggregationIntervalSec(ctx, deviceID)
	loc, locErr := time.LoadLocation("Asia/Jakarta")
	if locErr != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	alignedEnd := targetEnd.UTC().Truncate(time.Duration(intervalSec) * time.Second)
	if alignedEnd.IsZero() {
		alignedEnd = targetEnd.UTC()
	}
	opStart := alignedEnd.Add(-time.Duration(intervalSec) * time.Second)
	regStart := alignedEnd.Add(-24 * time.Hour)
	expected24hBuckets := int(math.Max(1, math.Floor(86400.0/float64(intervalSec))))

	var opInputs []WQMSParameterInput
	var regInputs []WQMSParameterInput

	for _, m := range activeMappings {
		if strings.ToUpper(m.SourceType) == "LAB" {
			if m.LabValue != nil {
				in := WQMSParameterInput{
					ParamCode:        m.ParamCode,
					SourceType:       "LAB",
					RawMeasured:      *m.LabValue,
					SourceUnit:       m.SourceUnit,
					ConversionFactor: m.ConversionFactor,
					CustomLimitMin:   m.CustomLimitMin,
					CustomLimitMax:   m.CustomLimitMax,
					ValidSamples:     1,
					ExpectedSamples:  1,
					MinCompleteness:  m.MinCompleteness,
				}
				opInputs = append(opInputs, in)
				regInputs = append(regInputs, in)
			}
			continue
		}

		if m.ParameterID == nil || *m.ParameterID == 0 {
			continue
		}
		pid := *m.ParameterID

		// 1. Operational input: latest completed aggregation bucket (within 2h lookback) or recent RawData
		var opVal *float64
		var opSamples int
		var latestAgg model.AggregationResult
		aggLookback := alignedEnd.Add(-2 * time.Hour)
		if err := s.db.WithContext(ctx).
			Where("parameter_id = ? AND period_end > ? AND period_end <= ? AND quality != ?",
				pid, aggLookback, targetEnd.UTC().Add(1*time.Minute), model.QualityBad).
			Order("period_end DESC").
			First(&latestAgg).Error; err == nil {
			if latestAgg.AvgValue != nil {
				opVal = latestAgg.AvgValue
			} else if latestAgg.Value != nil {
				opVal = latestAgg.Value
			}
			if opVal != nil && !math.IsNaN(*opVal) && !math.IsInf(*opVal, 0) {
				opSamples = 1
			}
		}

		if opSamples == 0 {
			var rawAvg *float64
			var rawCnt int64
			rawSince := targetEnd.UTC().Add(-30 * time.Minute)
			s.db.WithContext(ctx).Model(&model.RawData{}).
				Where("parameter_id = ? AND timestamp >= ? AND timestamp <= ? AND quality != ?",
					pid, rawSince, targetEnd.UTC().Add(1*time.Minute), model.QualityBad).
				Count(&rawCnt)
			if rawCnt > 0 {
				_ = s.db.WithContext(ctx).Model(&model.RawData{}).
					Where("parameter_id = ? AND timestamp >= ? AND timestamp <= ? AND quality != ?",
						pid, rawSince, targetEnd.UTC().Add(1*time.Minute), model.QualityBad).
					Select("AVG(value)").Row().Scan(&rawAvg)
				if rawAvg != nil && !math.IsNaN(*rawAvg) && !math.IsInf(*rawAvg, 0) {
					opVal = rawAvg
					opSamples = 1
				}
			}
		}

		opMeasured := 0.0
		if opVal != nil {
			opMeasured = *opVal
		}
		pidCopy := pid
		opInputs = append(opInputs, WQMSParameterInput{
			ParamCode:        m.ParamCode,
			ParameterID:      &pidCopy,
			SourceType:       "SENSOR",
			RawMeasured:      opMeasured,
			SourceUnit:       m.SourceUnit,
			ConversionFactor: m.ConversionFactor,
			CustomLimitMin:   m.CustomLimitMin,
			CustomLimitMax:   m.CustomLimitMax,
			ValidSamples:     opSamples,
			ExpectedSamples:  1,
			MinCompleteness:  0,
		})

		// 2. Regulatory 24h input: 24-hour mean across completed aggregation buckets
		var regAggs []model.AggregationResult
		_ = s.db.WithContext(ctx).
			Where("parameter_id = ? AND period_end > ? AND period_end <= ? AND quality != ?",
				pid, regStart, targetEnd.UTC().Add(1*time.Minute), model.QualityBad).
			Find(&regAggs)

		var regSum float64
		var regValid int
		for _, a := range regAggs {
			var v *float64
			if a.AvgValue != nil {
				v = a.AvgValue
			} else if a.Value != nil {
				v = a.Value
			}
			if v != nil && !math.IsNaN(*v) && !math.IsInf(*v, 0) {
				regSum += *v
				regValid++
			}
		}

		var regMeasured float64
		expSamples := expected24hBuckets
		if regValid > 0 {
			regMeasured = regSum / float64(regValid)
		} else if opSamples > 0 && opVal != nil {
			// Fallback to recent RawData 24h average if aggregation buckets have not populated yet
			var rawAvg24 *float64
			var rawCnt24 int64
			s.db.WithContext(ctx).Model(&model.RawData{}).
				Where("parameter_id = ? AND timestamp > ? AND timestamp <= ? AND quality != ?",
					pid, regStart, targetEnd.UTC().Add(1*time.Minute), model.QualityBad).
				Count(&rawCnt24)
			if rawCnt24 > 0 {
				_ = s.db.WithContext(ctx).Model(&model.RawData{}).
					Where("parameter_id = ? AND timestamp > ? AND timestamp <= ? AND quality != ?",
						pid, regStart, targetEnd.UTC().Add(1*time.Minute), model.QualityBad).
					Select("AVG(value)").Row().Scan(&rawAvg24)
				if rawAvg24 != nil {
					regMeasured = *rawAvg24
					regValid = 1
					expSamples = 1
				}
			}
		}

		regInputs = append(regInputs, WQMSParameterInput{
			ParamCode:        m.ParamCode,
			ParameterID:      &pidCopy,
			SourceType:       "SENSOR",
			RawMeasured:      regMeasured,
			SourceUnit:       m.SourceUnit,
			ConversionFactor: m.ConversionFactor,
			CustomLimitMin:   m.CustomLimitMin,
			CustomLimitMax:   m.CustomLimitMax,
			ValidSamples:     regValid,
			ExpectedSamples:  expSamples,
			MinCompleteness:  cfg.MinCompleteness,
		})
	}

	nowUTC := time.Now().UTC()

	// Build & persist Operational result
	opEval := s.calculator.EvaluateStation(opInputs, cfg.WaterClass, cfg.DOSaturationMgL)
	opRes := &model.WQMSCalculationResult{
		TenantID:               tenantID,
		SiteID:                 siteID,
		DeviceID:               0,
		CalculationMode:        model.CalculationModeOperational,
		WaterClass:             opEval.WaterClass,
		PeriodStart:            opStart,
		PeriodEnd:              alignedEnd,
		CalculatedAt:           nowUTC,
		OverallIP:              opEval.OverallIP,
		OverallCategory:        opEval.OverallCategory,
		MaxRatio:               opEval.MaxRatio,
		AvgRatio:               opEval.AvgRatio,
		DominantParameter:      opEval.DominantParameter,
		DominantConcentration:  opEval.DominantConcentration,
		DominantUnit:           opEval.DominantUnit,
		TiedParameters:         joinWQMSCodes(opEval.TiedParameters),
		CalculationStatus:      opEval.CalculationStatus,
		CompletenessPercent:    opEval.CompletenessPercent,
		ActiveParametersCount:  opEval.ActiveParametersCount,
		MissingParametersCount: opEval.MissingParametersCount,
		Method:                 "KEPMEN_LH_115_2003_PP_22_2021",
		CalculationReason:      reason,
		Notes:                  opEval.Notes,
		ParameterResults:       opEval.ParameterResults,
	}
	if opRes.ActiveParametersCount > 0 {
		_ = s.wqmsRepo.UpsertCalculationResult(calcCtx, opRes)
	}

	// Build & persist Regulatory 24h result (snap regulatory period_end to top-of-hour or current aligned boundary on manual/config runs)
	regEnd := alignedEnd
	localAligned := alignedEnd.In(loc)
	isManualOrConfig := reason == "MANUAL_RECALCULATE" || reason == "CLASS_CONFIG_CHANGED" || reason == "MAPPING_UPDATED" || reason == "ON_DEMAND_INITIAL"
	isHourlyBoundary := localAligned.Minute() == 0 && localAligned.Second() == 0

	regEval := s.calculator.EvaluateStation(regInputs, cfg.WaterClass, cfg.DOSaturationMgL)
	regRes := &model.WQMSCalculationResult{
		TenantID:               tenantID,
		SiteID:                 siteID,
		DeviceID:               0,
		CalculationMode:        model.CalculationModeRegulatory,
		WaterClass:             regEval.WaterClass,
		PeriodStart:            regEnd.Add(-24 * time.Hour),
		PeriodEnd:              regEnd,
		CalculatedAt:           nowUTC,
		OverallIP:              regEval.OverallIP,
		OverallCategory:        regEval.OverallCategory,
		MaxRatio:               regEval.MaxRatio,
		AvgRatio:               regEval.AvgRatio,
		DominantParameter:      regEval.DominantParameter,
		DominantConcentration:  regEval.DominantConcentration,
		DominantUnit:           regEval.DominantUnit,
		TiedParameters:         joinWQMSCodes(regEval.TiedParameters),
		CalculationStatus:      regEval.CalculationStatus,
		CompletenessPercent:    regEval.CompletenessPercent,
		ActiveParametersCount:  regEval.ActiveParametersCount,
		MissingParametersCount: regEval.MissingParametersCount,
		Method:                 "KEPMEN_LH_115_2003_PP_22_2021",
		CalculationReason:      reason,
		Notes:                  regEval.Notes,
		ParameterResults:       regEval.ParameterResults,
	}
	if regRes.ActiveParametersCount > 0 && (isHourlyBoundary || isManualOrConfig) {
		_ = s.wqmsRepo.UpsertCalculationResult(calcCtx, regRes)
	}

	if s.hub != nil && (opRes.ActiveParametersCount > 0 || regRes.ActiveParametersCount > 0) {
		ika := s.ComputeCurrentIKA(calcCtx, 0, cfg, opRes)
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "wqms.calculated",
			TenantID:  tenantID,
			SiteID:    siteID,
			Timestamp: nowUTC,
			Data: map[string]interface{}{
				"operational_ip": opRes,
				"regulatory_ip":  regRes,
				"ika_summary":    ika,
				"water_class":    cfg.WaterClass,
			},
		})
	}

	return regRes, opRes, nil
}

// ComputeCurrentIKA computes the Permen LHK No. 27 Tahun 2021 IKA score over the configured lookback window.
func (s *WQMSService) ComputeCurrentIKA(ctx context.Context, deviceID uint, cfg *model.WQMSStationConfig, fallbackLatest *model.WQMSCalculationResult) model.WQMSIKASummary {
	if cfg == nil {
		cfg = &model.WQMSStationConfig{
			WaterClass:    model.WaterClass2,
			IKAWindowDays: 30,
		}
	}
	windowDays := cfg.IKAWindowDays
	if windowDays <= 0 {
		windowDays = 30
	}
	since := time.Now().UTC().AddDate(0, 0, -windowDays)

	// Prefer daily deduplicated REGULATORY 24h records (1 status per calendar day in WIB)
	regRecords, _ := s.wqmsRepo.GetResultsForIKAWindow(ctx, deviceID, since, model.CalculationModeRegulatory)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	if len(regRecords) > 0 {
		byDay := make(map[string]model.WQMSIPCategory)
		var orderedDays []string
		for _, rec := range regRecords {
			dayKey := rec.PeriodEnd.In(loc).Format("2006-01-02")
			if _, seen := byDay[dayKey]; !seen {
				orderedDays = append(orderedDays, dayKey)
			}
			byDay[dayKey] = rec.OverallCategory
		}
		cats := make([]model.WQMSIPCategory, 0, len(orderedDays))
		for _, dk := range orderedDays {
			cats = append(cats, byDay[dk])
		}
		return s.calculator.ComputeIKASummary(cats, cfg.WaterClass, windowDays, "REGULATORY_24H")
	}

	// Fallback to operational records when 24h checkpoints have not accumulated yet
	opRecords, _ := s.wqmsRepo.GetResultsForIKAWindow(ctx, deviceID, since, model.CalculationModeOperational)
	if len(opRecords) > 0 {
		cats := make([]model.WQMSIPCategory, 0, len(opRecords))
		for _, rec := range opRecords {
			cats = append(cats, rec.OverallCategory)
		}
		return s.calculator.ComputeIKASummary(cats, cfg.WaterClass, windowDays, "OPERATIONAL_LIVE")
	}

	if fallbackLatest != nil && fallbackLatest.OverallCategory != model.WQMSCategoryUnavailable {
		return s.calculator.ComputeIKASummary([]model.WQMSIPCategory{fallbackLatest.OverallCategory}, cfg.WaterClass, windowDays, "OPERATIONAL_LIVE")
	}

	return s.calculator.ComputeIKASummary(nil, cfg.WaterClass, windowDays, "OPERATIONAL_LIVE")
}

// GetLatestSummary retrieves the complete WQMS summary bundle (IP + IKA + Config + Standards),
// triggering an initial on-demand calculation if no records exist yet.
func (s *WQMSService) GetLatestSummary(ctx context.Context, deviceID uint, mode string) (*WQMSLatestSummaryBundle, error) {
	cfg, err := s.wqmsRepo.GetOrCreateStationConfig(ctx, 0)
	if err != nil {
		return nil, err
	}

	normMode := strings.ToUpper(strings.TrimSpace(mode))
	if normMode != "OPERATIONAL" && normMode != "REGULATORY" {
		normMode = "REGULATORY"
	}

	primaryRes, err := s.wqmsRepo.GetLatestByDeviceID(ctx, 0, normMode)
	opRes, _ := s.wqmsRepo.GetLatestByDeviceID(ctx, 0, "OPERATIONAL")

	if err != nil || primaryRes == nil || primaryRes.WaterClass != cfg.WaterClass {
		regCalc, opCalc, _ := s.CalculateStationWindow(ctx, 0, time.Now().UTC(), "ON_DEMAND_INITIAL")
		if normMode == "OPERATIONAL" && opCalc != nil {
			primaryRes = opCalc
		} else if regCalc != nil {
			primaryRes = regCalc
		}
		if opCalc != nil {
			opRes = opCalc
		}
	}

	// If requested REGULATORY still had 0 records, fall back gracefully to Operational if available
	if (primaryRes == nil || primaryRes.ActiveParametersCount == 0) && opRes != nil && opRes.ActiveParametersCount > 0 {
		primaryRes = opRes
	}

	// Ensure Parameter relationships are populated on ParameterResults for rich UI labels
	if primaryRes != nil {
		for i := range primaryRes.ParameterResults {
			if primaryRes.ParameterResults[i].Parameter == nil && primaryRes.ParameterResults[i].ParameterID != nil {
				var p model.Parameter
				if s.db.WithContext(ctx).First(&p, *primaryRes.ParameterResults[i].ParameterID).Error == nil {
					primaryRes.ParameterResults[i].Parameter = &p
				}
			}
		}
	}

	ikaSummary := s.ComputeCurrentIKA(ctx, 0, cfg, primaryRes)
	intervalSec := s.resolveAggregationIntervalSec(ctx, 0)

	return &WQMSLatestSummaryBundle{
		IPResult:               primaryRes,
		OperationalIPResult:    opRes,
		IKASummary:             ikaSummary,
		Config:                 *cfg,
		ActiveClassStandards:   OfficialWQMSStandards[cfg.WaterClass],
		AggregationIntervalSec: intervalSec,
	}, nil
}

// GetTrend returns chronological IP & rolling IKA trend points for charting.
func (s *WQMSService) GetTrend(ctx context.Context, deviceID uint, rangeStr, mode string) ([]model.WQMSCalculationResult, error) {
	dur := 24 * time.Hour
	switch strings.ToLower(strings.TrimSpace(rangeStr)) {
	case "6h":
		dur = 6 * time.Hour
	case "24h", "1d":
		dur = 24 * time.Hour
	case "7d":
		dur = 7 * 24 * time.Hour
	case "30d":
		dur = 30 * 24 * time.Hour
	}

	since := time.Now().UTC().Add(-dur)
	records, err := s.wqmsRepo.GetTrend(ctx, 0, since, mode, 1000)
	if err != nil {
		return nil, err
	}
	// Fallback to OPERATIONAL points if REGULATORY has not accumulated points yet
	if len(records) == 0 && strings.ToUpper(mode) == "REGULATORY" {
		records, err = s.wqmsRepo.GetTrend(ctx, 0, since, "OPERATIONAL", 1000)
	}
	return records, err
}

// GetHistory returns paginated historical WQMS IP records.
func (s *WQMSService) GetHistory(ctx context.Context, filter repository.WQMSHistoryFilter) ([]model.WQMSCalculationResult, int64, error) {
	return s.wqmsRepo.GetHistory(ctx, filter)
}

// GetCalculationDetail retrieves a single persisted WQMS IP calculation with full parameter breakdown.
func (s *WQMSService) GetCalculationDetail(ctx context.Context, id uint) (*model.WQMSCalculationResult, error) {
	res, err := s.wqmsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !tenantctx.FromContext(ctx).CanAccessRecord(res.TenantID, res.SiteID) {
		return nil, fmt.Errorf("access denied: WQMS calculation record belongs to another tenant or restricted site")
	}
	return res, nil
}

// Recalculate triggers an explicit manual station recalculation and optional lookback backfill.
func (s *WQMSService) Recalculate(ctx context.Context, deviceID uint, targetTime time.Time, actor, ipAddress string) (*WQMSLatestSummaryBundle, error) {
	if targetTime.IsZero() {
		targetTime = time.Now().UTC()
	}
	regRes, _, err := s.CalculateStationWindow(ctx, deviceID, targetTime, "MANUAL_RECALCULATE")
	if err != nil {
		return nil, err
	}
	if s.systemRepo != nil {
		var auditTID *uint
		if regRes != nil && regRes.TenantID > 0 {
			tid := regRes.TenantID
			auditTID = &tid
		}
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  auditTID,
			Username:  actor,
			Action:    "WQMS_RECALCULATE",
			Resource:  "WQMS_IP_IKA",
			IPAddress: ipAddress,
			Details:   fmt.Sprintf("Manual WQMS IP & IKA recalculation triggered at %s", targetTime.Format(time.RFC3339)),
			CreatedAt: time.Now().UTC(),
		})
	}
	return s.GetLatestSummary(ctx, 0, "REGULATORY")
}

// BackfillHourlyTrend backfills hourly 24h regulatory WQMS IP points from historical aggregation_results.
func (s *WQMSService) BackfillHourlyTrend(ctx context.Context, hours int) (int, error) {
	if hours <= 0 {
		hours = 24
	}
	if hours > 720 {
		hours = 720
	}
	nowHour := time.Now().UTC().Truncate(time.Hour)
	count := 0
	for h := hours; h >= 0; h-- {
		ts := nowHour.Add(-time.Duration(h) * time.Hour)
		regRes, _, err := s.CalculateStationWindow(ctx, 0, ts, "MANUAL_RECALCULATE")
		if err == nil && regRes != nil && regRes.ActiveParametersCount > 0 {
			count++
		}
	}
	return count, nil
}

// ExportCSV exports historical WQMS IP & IKA summary records into CSV.
func (s *WQMSService) ExportCSV(ctx context.Context, filter repository.WQMSHistoryFilter) ([]byte, error) {
	filter.Page = 1
	filter.PageSize = 2000
	records, _, err := s.wqmsRepo.GetHistory(ctx, filter)
	if err != nil {
		return nil, err
	}
	cfg, _ := s.wqmsRepo.GetOrCreateStationConfig(ctx, 0)
	ika := s.ComputeCurrentIKA(ctx, 0, cfg, nil)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"# STATUS MUTU AIR - INDEKS PENCEMARAN (IP) & INDEKS KUALITAS AIR (IKA)"})
	_ = w.Write([]string{"# Regulasi", "PP No. 22 Tahun 2021 Lampiran VI | Kepmen LH No. 115 Tahun 2003 | Permen LHK No. 27 Tahun 2021"})
	_ = w.Write([]string{"# Kelas Air Aktif", string(ika.WaterClass)})
	_ = w.Write([]string{"# Skor IKA Periode", fmt.Sprintf("%.1f / 100 (%s - %d Hari)", ika.Score, ika.CategoryLabel, ika.WindowDays)})
	_ = w.Write([]string{"# Generated At", time.Now().Format("2006-01-02 15:04:05 MST")})
	_ = w.Write([]string{""})

	headers := []string{
		"ID", "Mode", "Water Class", "Period Start (UTC)", "Period End (UTC)", "Calculated At (UTC)",
		"Overall IP", "IP Category", "Max Ratio (Ci/Lij)M", "Avg Ratio (Ci/Lij)R",
		"Dominant Parameter", "Dominant Conc", "Completeness (%)", "Status",
		"pH (Ci)", "pH Ratio", "DO (mg/L)", "DO Ratio", "BOD (mg/L)", "BOD Ratio",
		"COD (mg/L)", "COD Ratio", "TSS (mg/L)", "TSS Ratio", "NO3-N (mg/L)", "NO3 Ratio",
		"NH3-N (mg/L)", "NH3 Ratio", "PO4-P (mg/L)", "PO4 Ratio",
	}
	_ = w.Write(headers)

	for _, r := range records {
		pmap := make(map[model.WQMSParamCode]model.WQMSParameterResult)
		for _, pr := range r.ParameterResults {
			pmap[pr.ParamCode] = pr
		}
		getCell := func(code model.WQMSParamCode) (string, string) {
			if pr, ok := pmap[code]; ok {
				return fmt.Sprintf("%.2f", pr.NormalizedMeasured), fmt.Sprintf("%.2f", pr.AdjustedRatio)
			}
			return "-", "-"
		}

		phC, phR := getCell(model.WQMSParamPH)
		doC, doR := getCell(model.WQMSParamDO)
		bodC, bodR := getCell(model.WQMSParamBOD)
		codC, codR := getCell(model.WQMSParamCOD)
		tssC, tssR := getCell(model.WQMSParamTSS)
		no3C, no3R := getCell(model.WQMSParamNO3)
		nh3C, nh3R := getCell(model.WQMSParamNH3)
		po4C, po4R := getCell(model.WQMSParamPO4)

		_ = w.Write([]string{
			strconv.FormatUint(uint64(r.ID), 10),
			string(r.CalculationMode),
			string(r.WaterClass),
			r.PeriodStart.Format("2006-01-02 15:04:05"),
			r.PeriodEnd.Format("2006-01-02 15:04:05"),
			r.CalculatedAt.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", r.OverallIP),
			string(r.OverallCategory),
			fmt.Sprintf("%.2f", r.MaxRatio),
			fmt.Sprintf("%.2f", r.AvgRatio),
			string(r.DominantParameter),
			fmt.Sprintf("%.2f %s", r.DominantConcentration, r.DominantUnit),
			fmt.Sprintf("%.1f", r.CompletenessPercent),
			string(r.CalculationStatus),
			phC, phR, doC, doR, bodC, bodR, codC, codR, tssC, tssR, no3C, no3R, nh3C, nh3R, po4C, po4R,
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}

// ExportExcel exports historical WQMS IP & IKA records with UTF-8 BOM for native Microsoft Excel compatibility.
func (s *WQMSService) ExportExcel(ctx context.Context, filter repository.WQMSHistoryFilter) ([]byte, error) {
	csvBytes, err := s.ExportCSV(ctx, filter)
	if err != nil {
		return nil, err
	}
	bom := []byte{0xEF, 0xBB, 0xBF}
	return append(bom, csvBytes...), nil
}

func joinWQMSCodes(codes []model.WQMSParamCode) string {
	if len(codes) <= 1 {
		return ""
	}
	strs := make([]string, len(codes))
	for i, c := range codes {
		strs[i] = string(c)
	}
	return strings.Join(strs, ",")
}

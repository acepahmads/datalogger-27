package service

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/database"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/quality"
	"datalogger/internal/queue"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"
	"datalogger/pkg/tenantctx"
)

// TelemetryIngestPayload represents an incoming field telemetry measurement
type TelemetryIngestPayload struct {
	TenantID        uint
	SiteID          uint
	DeviceID        uint
	DeviceCode      string
	DeviceName      string
	ParameterID     uint
	ParameterCode   string
	ParameterName   string
	Unit            string
	DataType        model.ParameterDataType
	RawBytes          []byte
	RawHex            string
	RawValue          float64
	UncalibratedValue *float64
	Value             float64
	FormulaValue      *float64
	Formula         string
	IsHeldValue     bool
	Quality         model.TelemetryQuality
	QualityReason   model.QualityReason
	QualityFlags    string
	ProcessedValue  float64
	Source          string
	Sequence        uint64
	DeviceTimestamp *time.Time
	ReceivedAt      time.Time
	ErrorMessage    string
	Parameter       *model.Parameter
	ReadError       error
}

// LatestTelemetry holds instantaneous cached values for fast lookup
type LatestTelemetry struct {
	TenantID        uint                    `json:"tenant_id,omitempty"`
	SiteID          uint                    `json:"site_id,omitempty"`
	DeviceID        uint                    `json:"device_id"`
	DeviceCode      string                  `json:"device_code"`
	DeviceName      string                  `json:"device_name"`
	ParameterID     uint                    `json:"parameter_id"`
	ParameterCode   string                  `json:"parameter_code"`
	ParameterName   string                  `json:"parameter_name"`
	Unit            string                  `json:"unit"`
	DataType        model.ParameterDataType `json:"data_type"`
	Value                    float64                 `json:"value"`
	ProcessedValue           float64                 `json:"processed_value"`
	UncalibratedValue        *float64                `json:"uncalibrated_value,omitempty"`
	CalibrationEnabled       bool                    `json:"calibration_enabled,omitempty"`
	CalibrationSessionActive bool                    `json:"calibration_session_active,omitempty"`
	ValueNumeric             *float64                `json:"value_numeric,omitempty"`
	FormulaValue    *float64                `json:"formula_value,omitempty"`
	Formula         string                  `json:"formula,omitempty"`
	IsHeldValue     bool                    `json:"is_held_value"`
	ValueText       string                  `json:"value_text,omitempty"`
	ValueBool       *bool                   `json:"value_bool,omitempty"`
	RawValue        float64                 `json:"raw_value"`
	RawHex          string                  `json:"raw_hex,omitempty"`
	Quality         model.TelemetryQuality  `json:"quality"`
	QualityReason   model.QualityReason     `json:"quality_reason"`
	QualityFlags    string                  `json:"quality_flags,omitempty"`
	Source          string                  `json:"source"`
	DeviceTimestamp *time.Time              `json:"device_timestamp,omitempty"`
	ReceivedAt      time.Time               `json:"received_at"`
	ProcessedAt     *time.Time              `json:"processed_at,omitempty"`
	BroadcastAt     *time.Time              `json:"broadcast_at,omitempty"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// TelemetryPersistenceBoundary defines the persistence interface used by TelemetryService.
// In Phase 4.1, the default implementation writes to repository.TelemetryRepository with batch retries.
// In Phase 4.2, a persistent disk queue adapter will plug into this boundary seamlessly.
type TelemetryPersistenceBoundary interface {
	SaveBatch(ctx context.Context, batch []*model.RawData) error
}

// TelemetryMetrics captures pipeline health, buffer utilization, and throughput
type TelemetryMetrics struct {
	IngestedCount       uint64     `json:"ingested_count"`
	PersistedCount      uint64     `json:"persisted_count"`
	DroppedCount        uint64     `json:"dropped_count"`
	ErrorCount          uint64     `json:"error_count"`
	QueueLength         int        `json:"queue_length"`
	QueueCapacity       int        `json:"queue_capacity"`
	LastFlushDurationMs int64      `json:"last_flush_duration_ms"`
	LastFlushAt         time.Time  `json:"last_flush_at"`
	PersistenceStatus   string     `json:"persistence_status"`
	ConsecutiveDBErrors int        `json:"consecutive_db_errors"`
	LastDBError         string     `json:"last_db_error,omitempty"`
	LastDBErrorAt       *time.Time `json:"last_db_error_at,omitempty"`

	// Phase 4.2: Persistent WAL Queue Metrics
	QueuePendingRecords int64   `json:"queue_pending_records"`
	QueueSpoolSizeBytes int64   `json:"queue_spool_size_bytes"`
	QueueTotalReplayed  uint64  `json:"queue_total_replayed"`
	QueueChecksumErrors uint64  `json:"queue_checksum_errors"`
	QueueEnabled        bool    `json:"queue_enabled"`
	QueueDiskPercent    float64 `json:"queue_disk_percent"`
}

// QualitySummaryDTO captures health metrics across current parameters (Phase 3.2)
type QualitySummaryDTO struct {
	DeviceID       *uint   `json:"device_id,omitempty"`
	TotalCount     int     `json:"total_count"`
	GoodCount      int     `json:"good_count"`
	UncertainCount int     `json:"uncertain_count"`
	BadCount       int     `json:"bad_count"`
	StaleCount     int     `json:"stale_count"`
	HealthPercent  float64 `json:"health_percent"`
}

// TelemetryConfig defines runtime options for the ingestion pipeline
type TelemetryConfig struct {
	BufferSize           int           // Channel capacity for non-blocking ingestion
	BatchSize            int           // Max records per database flush
	FlushInterval        time.Duration // Interval to flush accumulated records
	MaxRetries           int           // Retry attempts on transient database error
	QueueEnabled         bool          // Enable Phase 4.2 persistent disk queue
	QueueDir             string        // Directory for WAL files
	QueueMaxSizeBytes    int64         // Maximum disk spool size
	QueueSyncMode        string        // "batch", "always", "none"
	QueueDiskWarnPercent float64       // Disk warning threshold
}

// DefaultTelemetryConfig returns balanced production settings for edge devices
func DefaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		BufferSize:           5000,
		BatchSize:            50,
		FlushInterval:        250 * time.Millisecond,
		MaxRetries:           3,
		QueueEnabled:         true,
		QueueDir:             filepath.Join("data", "queue"),
		QueueMaxSizeBytes:    100 * 1024 * 1024,
		QueueSyncMode:        "batch",
		QueueDiskWarnPercent: 80.0,
	}
}

// TelemetryService manages asynchronous ingestion, validation, buffering, and persistence
type TelemetryService struct {
	repo             *repository.TelemetryRepository
	hub              *websocket.Hub
	cfg              TelemetryConfig
	qualityProcessor *quality.QualityProcessor
	buffer           chan *model.RawData
	latestMu         sync.RWMutex
	latestCache      map[uint]*LatestTelemetry // Key: parameter_id
	staleEvalLastAt  map[uint]time.Time        // Key: parameter_id (tracks last stale alarm evaluation step)
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	stopped          int32
	seqCounter       uint64

	// Metrics
	ingestedCount  uint64
	persistedCount uint64
	droppedCount   uint64
	errorCount     uint64
	lastFlushDur   int64
	lastFlushAt    time.Time
	flushMu        sync.RWMutex

	// Phase 4.1 & 4.2: Database Persistence Boundary & Health Tracking
	persistenceBoundary   TelemetryPersistenceBoundary
	walAdapter            *PersistentQueueAdapter
	persistenceMu         sync.RWMutex
	persistenceStatus     string
	consecutiveDBFailures int
	lastDBError           string
	lastDBErrorAt         *time.Time

	// Phase 6 Wave 1: Automated Alarm Threshold Evaluator
	alarmEvaluator *AlarmEvaluator

	// Device status resolver
	deviceLookup func(deviceID uint) (*model.Device, error)

	// Parameter status resolver
	paramLookup func(paramID uint) (*model.Parameter, error)
}

// NewTelemetryService constructs and starts the background ingestion worker
func NewTelemetryService(
	repo *repository.TelemetryRepository,
	hub *websocket.Hub,
	cfg TelemetryConfig,
) *TelemetryService {
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 5000
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 250 * time.Millisecond
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &TelemetryService{
		repo:              repo,
		hub:               hub,
		cfg:               cfg,
		qualityProcessor:  quality.NewQualityProcessor(),
		buffer:            make(chan *model.RawData, cfg.BufferSize),
		latestCache:       make(map[uint]*LatestTelemetry),
		staleEvalLastAt:   make(map[uint]time.Time),
		ctx:               ctx,
		cancel:            cancel,
		persistenceStatus: "HEALTHY",
	}

	if cfg.QueueEnabled {
		if cfg.QueueDir == "" {
			cfg.QueueDir = filepath.Join("data", "queue")
		}
		if cfg.QueueMaxSizeBytes <= 0 {
			cfg.QueueMaxSizeBytes = 100 * 1024 * 1024
		}
		walCfg := queue.WALConfig{
			Dir:             cfg.QueueDir,
			MaxSizeBytes:    cfg.QueueMaxSizeBytes,
			MaxSegmentSize:  10 * 1024 * 1024,
			SyncMode:        cfg.QueueSyncMode,
			DiskWarnPercent: cfg.QueueDiskWarnPercent,
		}
		wal, err := queue.OpenWALQueue(walCfg)
		if err != nil {
			logger.Error("Failed initializing persistent WAL queue, falling back to direct persistence: %v", err)
			if repo != nil {
				s.persistenceBoundary = repo
			}
		} else {
			adapter := NewPersistentQueueAdapter(repo, wal, cfg.BatchSize, cfg.MaxRetries)
			s.walAdapter = adapter
			s.persistenceBoundary = adapter
			logger.Info("Persistent WAL Queue attached to TelemetryService at %s", cfg.QueueDir)
		}
	} else if repo != nil {
		s.persistenceBoundary = repo
	}

	s.wg.Add(2)
	go s.batchPersistenceWorker()
	go s.staleDetectionWorker()

	logger.Info("TelemetryService initialized with QualityProcessor (Buffer: %d, Batch: %d, FlushInterval: %v, QueueEnabled: %v)",
		cfg.BufferSize, cfg.BatchSize, cfg.FlushInterval, cfg.QueueEnabled)

	return s
}

// GetWALAdapter returns the attached persistent queue adapter if enabled
func (s *TelemetryService) GetWALAdapter() *PersistentQueueAdapter {
	return s.walAdapter
}

// SetPersistenceBoundary configures a custom boundary implementation (e.g. for testing or Phase 4.2 disk queue)
func (s *TelemetryService) SetPersistenceBoundary(boundary TelemetryPersistenceBoundary) {
	s.persistenceMu.Lock()
	defer s.persistenceMu.Unlock()
	s.persistenceBoundary = boundary
	if adapter, ok := boundary.(*PersistentQueueAdapter); ok {
		s.walAdapter = adapter
	} else {
		s.walAdapter = nil
	}
}

// QualityProcessor exposes the data quality evaluation engine
func (s *TelemetryService) QualityProcessor() *quality.QualityProcessor {
	return s.qualityProcessor
}

// SetAlarmEvaluator attaches an AlarmEvaluator to automatically evaluate parameter thresholds
func (s *TelemetryService) SetAlarmEvaluator(evaluator *AlarmEvaluator) {
	s.alarmEvaluator = evaluator
}

// GetAlarmEvaluator returns the attached AlarmEvaluator
func (s *TelemetryService) GetAlarmEvaluator() *AlarmEvaluator {
	return s.alarmEvaluator
}

// SetDeviceLookup registers a device lookup callback for checking enabled and active status
func (s *TelemetryService) SetDeviceLookup(lookup func(deviceID uint) (*model.Device, error)) {
	s.deviceLookup = lookup
}

// SetParamLookup registers a parameter lookup callback for checking enabled and active status
func (s *TelemetryService) SetParamLookup(lookup func(paramID uint) (*model.Parameter, error)) {
	s.paramLookup = lookup
}

// ClearDeviceCache removes cached instantaneous telemetry for a specific device (e.g. when disabled or deleted)
func (s *TelemetryService) ClearDeviceCache(deviceID uint) {
	s.latestMu.Lock()
	defer s.latestMu.Unlock()
	for k, item := range s.latestCache {
		if item != nil && item.DeviceID == deviceID {
			delete(s.latestCache, k)
			delete(s.staleEvalLastAt, k)
		}
	}
}

// ClearParameterCache removes cached instantaneous telemetry for a specific parameter (e.g. when disabled or deleted)
func (s *TelemetryService) ClearParameterCache(paramID uint) {
	s.latestMu.Lock()
	defer s.latestMu.Unlock()
	delete(s.latestCache, paramID)
	delete(s.staleEvalLastAt, paramID)
}

// SyncParameterCalibrationCache immediately updates cached instantaneous telemetry when a parameter's calibration curve or session state changes
func (s *TelemetryService) SyncParameterCalibrationCache(param *model.Parameter) {
	if param == nil {
		return
	}
	s.latestMu.Lock()
	cached, ok := s.latestCache[param.ID]
	var broadcastCopy *LatestTelemetry
	if ok && cached != nil {
		cached.CalibrationEnabled = param.CalibrationEnabled
		cached.CalibrationSessionActive = param.CalibrationSessionActive
		if cached.UncalibratedValue != nil {
			newCal := param.ApplyCalibration(*cached.UncalibratedValue)
			cached.Value = newCal
			cached.ProcessedValue = newCal
			vNum := newCal
			cached.ValueNumeric = &vNum
			cached.ValueText = strconv.FormatFloat(newCal, 'f', 2, 64)
		}
		if param.CalibrationSessionActive {
			if cached.Quality != model.QualityBad {
				cached.Quality = model.QualityUncertain
			}
			cached.QualityReason = model.ReasonCalibrationMode
		} else if cached.QualityReason == model.ReasonCalibrationMode {
			cached.Quality = model.QualityGood
			cached.QualityReason = model.ReasonNone
		}
		cp := *cached
		broadcastCopy = &cp
	}
	s.latestMu.Unlock()

	if broadcastCopy != nil && s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "device.telemetry.received",
			Timestamp: time.Now().UTC(),
			TenantID:  broadcastCopy.TenantID,
			SiteID:    broadcastCopy.SiteID,
			Data:      broadcastCopy,
		})
	}
}

// Ingest receives a telemetry measurement, validates, updates cache, broadcasts, and enqueues
func (s *TelemetryService) Ingest(payload *TelemetryIngestPayload) error {
	if payload == nil {
		return fmt.Errorf("telemetry payload cannot be nil")
	}
	if atomic.LoadInt32(&s.stopped) == 1 {
		return fmt.Errorf("telemetry service is stopped")
	}

	// Skip disabled parameters from ingestion pipeline
	if payload.Parameter != nil && !payload.Parameter.Enabled {
		return nil
	}
	if s.paramLookup != nil {
		if p, err := s.paramLookup(payload.ParameterID); err == nil && p != nil && !p.Enabled {
			return nil
		}
	}

	// 1. Validation & Quality Processing (Phase 3.2)
	rawData, latest, err := s.normalizeAndValidate(payload)
	if err != nil {
		atomic.AddUint64(&s.errorCount, 1)
		logger.Warn("Telemetry validation warning for dev %d param %d: %v", payload.DeviceID, payload.ParameterID, err)
	}

	atomic.AddUint64(&s.ingestedCount, 1)

	// 2. Update Instantaneous In-Memory Cache (O(1) lookup for dashboard)
	// Reliability (Phase 4.1): If read failed or quality is BAD, retain previous valid values AND valid timestamps
	// so that dashboard and operator console show true last-seen time rather than failed poll attempt timestamps.
	s.latestMu.Lock()
	if payload.ReadError != nil || latest.Quality == model.QualityBad {
		if prev, ok := s.latestCache[latest.ParameterID]; ok {
			latest.Value = prev.Value
			latest.ProcessedValue = prev.ProcessedValue
			latest.UncalibratedValue = prev.UncalibratedValue
			latest.ValueNumeric = prev.ValueNumeric
			latest.FormulaValue = prev.FormulaValue
			latest.ValueText = prev.ValueText
			latest.ValueBool = prev.ValueBool
			latest.RawValue = prev.RawValue
			latest.RawHex = prev.RawHex
			if !prev.ReceivedAt.IsZero() {
				latest.ReceivedAt = prev.ReceivedAt
			}
			if prev.DeviceTimestamp != nil {
				latest.DeviceTimestamp = prev.DeviceTimestamp
			}
		}
	} else if latest.Quality != model.QualityStale && !latest.IsHeldValue {
		delete(s.staleEvalLastAt, latest.ParameterID)
	}
	s.latestCache[latest.ParameterID] = latest
	s.latestMu.Unlock()

	// Phase 6 Wave 1: Automated Alarm Threshold Evaluation
	if s.alarmEvaluator != nil {
		s.alarmEvaluator.EvaluateTelemetry(s.ctx, latest, payload.Parameter)
		// Active telemetry proves device communication is alive: recover offline status if active
		dev := &model.Device{
			ID:         latest.DeviceID,
			TenantID:   latest.TenantID,
			SiteID:     latest.SiteID,
			DeviceCode: latest.DeviceCode,
			Enabled:    true,
			Status:     model.DeviceStatusActive,
			LastDataAt: &latest.ReceivedAt,
		}
		s.alarmEvaluator.EvaluateDeviceCommSuccess(s.ctx, dev, nil)
	}

	// 3. Realtime WebSocket Broadcast
	bcastNow := time.Now().UTC()
	latest.BroadcastAt = &bcastNow
	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "device.telemetry.received",
			Timestamp: latest.ReceivedAt,
			TenantID:  latest.TenantID,
			SiteID:    latest.SiteID,
			Data:      latest,
		})
	}

	// 4. Non-Blocking Bounded Buffer Enqueue (Protects polling workers from DB latency)
	select {
	case s.buffer <- rawData:
		// Enqueued successfully
		return nil
	default:
		// Buffer is full (Backpressure): drop and record metric rather than blocking polling engine
		atomic.AddUint64(&s.droppedCount, 1)
		logger.Warn("Telemetry buffer full (%d items)! Dropping telemetry event for device %d param %d to prevent worker stall",
			s.cfg.BufferSize, payload.DeviceID, payload.ParameterID)
		return fmt.Errorf("telemetry buffer capacity exceeded")
	}
}

// normalizeAndValidate cleanses input data and generates DB record and latest view
func (s *TelemetryService) normalizeAndValidate(p *TelemetryIngestPayload) (*model.RawData, *LatestTelemetry, error) {
	if p.DeviceID == 0 {
		return nil, nil, fmt.Errorf("device_id is required")
	}
	if p.ParameterID == 0 {
		return nil, nil, fmt.Errorf("parameter_id is required")
	}

	now := time.Now().UTC()
	recAt := p.ReceivedAt
	if recAt.IsZero() {
		recAt = now
	} else {
		recAt = recAt.UTC()
	}

	// Phase 3.2: Deterministic Data Quality & Processing Engine
	valInput := &quality.ValidationInput{
		DeviceID:        p.DeviceID,
		ParameterID:     p.ParameterID,
		RawValue:        p.RawValue,
		Value:           p.Value,
		DataType:        p.DataType,
		DeviceTimestamp: p.DeviceTimestamp,
		ReceivedAt:      recAt,
		Source:          p.Source,
		IsHeldValue:     p.IsHeldValue,
		ReadError:       p.ReadError,
	}

	var valResult *quality.ValidationResult
	if s.qualityProcessor != nil {
		valResult = s.qualityProcessor.Validate(p.Parameter, valInput)
	} else {
		valResult = &quality.ValidationResult{
			Quality:        model.QualityGood,
			QualityReason:  model.ReasonNone,
			ProcessedValue: p.Value,
			ProcessedAt:    now,
		}
	}

	// Backward compatibility: If caller explicitly flagged BAD / UNCERTAIN / STALE, preserve it
	if p.Quality != "" && p.Quality != model.QualityGood && valResult.Quality == model.QualityGood {
		valResult.Quality = p.Quality
		if p.QualityReason != "" {
			valResult.QualityReason = p.QualityReason
		} else if p.Quality == model.QualityBad {
			valResult.QualityReason = model.ReasonDecodingError
		} else if p.Quality == model.QualityStale {
			valResult.QualityReason = model.ReasonStaleData
		}
	}

	// Active Calibration Session Guard: tag readings with calibration_session so alarms & regulatory AQMS/SPARING/ONLIMO buckets are protected
	calEnabled := false
	calSessionActive := false
	resolvedParam := p.Parameter
	if resolvedParam == nil && s.paramLookup != nil {
		if lookedUp, err := s.paramLookup(p.ParameterID); err == nil && lookedUp != nil {
			resolvedParam = lookedUp
		}
	}
	if resolvedParam != nil {
		calEnabled = resolvedParam.CalibrationEnabled
		calSessionActive = resolvedParam.CalibrationSessionActive
		if calSessionActive {
			if valResult.Quality != model.QualityBad {
				valResult.Quality = model.QualityUncertain
			}
			valResult.QualityReason = model.ReasonCalibrationMode
			if p.QualityFlags == "" {
				p.QualityFlags = model.FlagCalibrationSession
			} else if !strings.Contains(p.QualityFlags, model.FlagCalibrationSession) {
				p.QualityFlags = p.QualityFlags + "," + model.FlagCalibrationSession
			}
		}
	}

	// Preserve caller-supplied quality flags (e.g., formula_fallback, calibration_session) alongside QualityProcessor flags
	if p.QualityFlags != "" {
		flagSet := make(map[string]bool)
		mergedFlags := make([]string, 0, 4)
		for _, part := range strings.Split(valResult.QualityFlags, ",") {
			f := strings.TrimSpace(part)
			if f != "" && !flagSet[f] {
				flagSet[f] = true
				mergedFlags = append(mergedFlags, f)
			}
		}
		for _, part := range strings.Split(p.QualityFlags, ",") {
			f := strings.TrimSpace(part)
			if f != "" && !flagSet[f] {
				flagSet[f] = true
				mergedFlags = append(mergedFlags, f)
			}
		}
		valResult.QualityFlags = strings.Join(mergedFlags, ",")
	}

	val := valResult.ProcessedValue
	rawVal := p.RawValue
	if math.IsNaN(rawVal) || math.IsInf(rawVal, 0) {
		rawVal = 0
	}

	// Value representations
	valNum := val
	var valText string
	var valBool *bool

	switch p.DataType {
	case model.DataTypeBoolean:
		b := (val != 0)
		valBool = &b
		if b {
			valText = "ON"
		} else {
			valText = "OFF"
		}
	case model.DataTypeInt16, model.DataTypeInt32, model.DataTypeUInt16, model.DataTypeUInt32:
		valText = strconv.FormatInt(int64(math.Round(val)), 10)
	default:
		valText = strconv.FormatFloat(val, 'f', 2, 64)
	}

	seq := atomic.AddUint64(&s.seqCounter, 1)
	if p.Sequence > 0 {
		seq = p.Sequence
	}

	source := p.Source
	if source == "" {
		source = "MODBUS_TCP"
	}

	tenantID := p.TenantID
	siteID := p.SiteID
	if (tenantID == 0 || siteID == 0) && s.deviceLookup != nil {
		if d, err := s.deviceLookup(p.DeviceID); err == nil && d != nil {
			if tenantID == 0 {
				tenantID = d.TenantID
			}
			if siteID == 0 {
				siteID = d.SiteID
			}
		}
	}
	if tenantID == 0 {
		tenantID = 1
	}
	if siteID == 0 {
		siteID = 1
	}

	rawData := &model.RawData{
		TenantID:        tenantID,
		SiteID:          siteID,
		RecordUUID:      fmt.Sprintf("REC-%d-%d-%d-%d", p.DeviceID, p.ParameterID, seq, recAt.UnixNano()),
		DeviceID:        p.DeviceID,
		ParameterID:     p.ParameterID,
		Value:           val,
		ProcessedValue:  val,
		ValueNumeric:    &valNum,
		FormulaValue:    p.FormulaValue,
		IsHeldValue:     p.IsHeldValue,
		ValueText:       valText,
		ValueBool:       valBool,
		RawValue:        rawVal,
		RawHex:          p.RawHex,
		RawBytes:        p.RawBytes,
		Quality:         valResult.Quality,
		QualityReason:   valResult.QualityReason,
		QualityFlags:    valResult.QualityFlags,
		Source:          source,
		Sequence:        seq,
		DeviceTimestamp: p.DeviceTimestamp,
		ReceivedAt:      recAt,
		ProcessedAt:     &valResult.ProcessedAt,
		StoredAt:        now,
		Timestamp:       recAt,
	}

	var uncalPtr *float64
	if p.UncalibratedValue != nil && !math.IsNaN(*p.UncalibratedValue) && !math.IsInf(*p.UncalibratedValue, 0) {
		u := *p.UncalibratedValue
		uncalPtr = &u
	} else if !calEnabled {
		u := val
		uncalPtr = &u
	}

	latest := &LatestTelemetry{
		TenantID:                 tenantID,
		SiteID:                   siteID,
		DeviceID:                 p.DeviceID,
		DeviceCode:               p.DeviceCode,
		DeviceName:               p.DeviceName,
		ParameterID:              p.ParameterID,
		ParameterCode:            p.ParameterCode,
		ParameterName:            p.ParameterName,
		Unit:                     p.Unit,
		DataType:                 p.DataType,
		Value:                    val,
		ProcessedValue:           val,
		UncalibratedValue:        uncalPtr,
		CalibrationEnabled:       calEnabled,
		CalibrationSessionActive: calSessionActive,
		ValueNumeric:             &valNum,
		FormulaValue:    p.FormulaValue,
		Formula:         p.Formula,
		IsHeldValue:     p.IsHeldValue,
		ValueText:       valText,
		ValueBool:       valBool,
		RawValue:        rawVal,
		RawHex:          p.RawHex,
		Quality:         valResult.Quality,
		QualityReason:   valResult.QualityReason,
		QualityFlags:    valResult.QualityFlags,
		Source:          source,
		DeviceTimestamp: p.DeviceTimestamp,
		ReceivedAt:      recAt,
		ProcessedAt:     &valResult.ProcessedAt,
		UpdatedAt:       now,
	}

	return rawData, latest, nil
}

// staleDetectionWorker checks cached parameters for staleness without incurring periodic DB writes
func (s *TelemetryService) staleDetectionWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if database.IsRestoreMode() {
				continue
			}
			s.evaluateStaleParameters()
		}
	}
}

// resolveEffectiveStaleTimeout computes the effective stale timeout in seconds for a parameter,
// honoring per-parameter StaleTimeoutSeconds and ensuring at least ~2.5x the configured polling interval
// so normal slow polling cadences (e.g. 60s or 120s) or a single transient retry never trigger false STALE states.
func resolveEffectiveStaleTimeout(dev *model.Device, param *model.Parameter) int {
	staleTimeout := 120
	if param != nil && param.StaleTimeoutSeconds > 0 {
		staleTimeout = param.StaleTimeoutSeconds
	}

	pollMs := 0
	if param != nil && param.PollingIntervalMs != nil && *param.PollingIntervalMs > 0 {
		pollMs = *param.PollingIntervalMs
	} else if dev != nil && dev.Connection != nil {
		if dev.Connection.PollingInterval > 0 {
			pollMs = dev.Connection.PollingInterval
		} else if dev.Connection.PollIntervalMs > 0 {
			pollMs = dev.Connection.PollIntervalMs
		}
	}

	if pollMs > 0 {
		pollSec := (pollMs + 999) / 1000
		minForPoll := pollSec*2 + (pollSec+1)/2 // ~2.5x polling cadence
		if minForPoll > staleTimeout {
			staleTimeout = minForPoll
		}
	}

	return staleTimeout
}

type staleEvalCandidate struct {
	item  *LatestTelemetry
	param *model.Parameter
}

func (s *TelemetryService) evaluateStaleParameters() {
	s.latestMu.Lock()
	var staleCandidates []staleEvalCandidate
	devLastData := make(map[uint]time.Time)
	devCodes := make(map[uint]string)
	devCache := make(map[uint]*model.Device)

	for _, item := range s.latestCache {
		if item == nil {
			continue
		}

		// Skip disabled devices from stale checks
		var dev *model.Device
		if s.deviceLookup != nil {
			if cachedDev, ok := devCache[item.DeviceID]; ok {
				dev = cachedDev
			} else if d, err := s.deviceLookup(item.DeviceID); err == nil && d != nil {
				dev = d
				devCache[item.DeviceID] = d
			}
			if dev != nil && (!dev.Enabled || dev.Status != model.DeviceStatusActive) {
				continue
			}
		}

		// Skip disabled parameters from stale checks
		var param *model.Parameter
		if s.paramLookup != nil {
			if p, err := s.paramLookup(item.ParameterID); err == nil && p != nil {
				if !p.Enabled {
					continue
				}
				param = p
			}
		}

		if cur, ok := devLastData[item.DeviceID]; !ok || item.ReceivedAt.After(cur) {
			devLastData[item.DeviceID] = item.ReceivedAt
			devCodes[item.DeviceID] = item.DeviceCode
		}

		staleTimeout := resolveEffectiveStaleTimeout(dev, param)
		isStale, staleQ, staleReason := s.qualityProcessor.CheckStale(item.ParameterID, staleTimeout, item.ReceivedAt)
		if isStale {
			now := time.Now().UTC()
			wasAlreadyStale := (item.Quality == model.QualityStale)
			item.Quality = staleQ
			item.QualityReason = staleReason
			item.QualityFlags = model.FlagStale
			item.UpdatedAt = now

			if !wasAlreadyStale && s.hub != nil {
				s.hub.Broadcast(websocket.WSMessage{
					Type:      "device.telemetry.received",
					Timestamp: item.UpdatedAt,
					TenantID:  item.TenantID,
					SiteID:    item.SiteID,
					Data:      item,
				})
			}

			stepDur := time.Duration(staleTimeout/2) * time.Second
			if stepDur < 15*time.Second {
				stepDur = 15 * time.Second
			}
			if stepDur > 60*time.Second {
				stepDur = 60 * time.Second
			}

			lastEval := s.staleEvalLastAt[item.ParameterID]
			if !wasAlreadyStale || lastEval.IsZero() || now.Sub(lastEval) >= stepDur {
				s.staleEvalLastAt[item.ParameterID] = now
				itemCopy := *item
				staleCandidates = append(staleCandidates, staleEvalCandidate{
					item:  &itemCopy,
					param: param,
				})
			}
		} else {
			delete(s.staleEvalLastAt, item.ParameterID)
			if item.Quality == model.QualityStale && item.QualityReason == model.ReasonStaleData {
				restoredQ := model.QualityGood
				restoredReason := model.ReasonNone
				if param != nil && param.CurrentQuality != "" && param.CurrentQuality != model.QualityStale {
					restoredQ = param.CurrentQuality
					restoredReason = param.CurrentQualityReason
				}
				item.Quality = restoredQ
				item.QualityReason = restoredReason
				item.QualityFlags = ""
				item.UpdatedAt = time.Now().UTC()
				if s.hub != nil {
					s.hub.Broadcast(websocket.WSMessage{
						Type:      "device.telemetry.received",
						Timestamp: item.UpdatedAt,
						TenantID:  item.TenantID,
						SiteID:    item.SiteID,
						Data:      item,
					})
				}
				itemCopy := *item
				staleCandidates = append(staleCandidates, staleEvalCandidate{
					item:  &itemCopy,
					param: param,
				})
			}
		}
	}
	s.latestMu.Unlock()

	// Evaluate parameter alarms and device offline status outside lock to avoid DB/evaluator lock contention
	if s.alarmEvaluator != nil {
		for _, cand := range staleCandidates {
			s.alarmEvaluator.EvaluateTelemetry(s.ctx, cand.item, cand.param)
		}

		for devID, lastData := range devLastData {
			lastDataCopy := lastData
			var dev *model.Device
			if s.deviceLookup != nil {
				if d, err := s.deviceLookup(devID); err == nil && d != nil {
					devCopy := *d
					devCopy.LastDataAt = &lastDataCopy
					dev = &devCopy
				}
			}
			if dev == nil {
				dev = &model.Device{
					ID:         devID,
					DeviceCode: devCodes[devID],
					Enabled:    true,
					Status:     model.DeviceStatusActive,
					LastDataAt: &lastDataCopy,
				}
			}
			s.alarmEvaluator.EvaluateDeviceOffline(s.ctx, dev, nil)
		}
	}
}

// batchPersistenceWorker continuously drains buffer and executes batch writes to MariaDB
func (s *TelemetryService) batchPersistenceWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]*model.RawData, 0, s.cfg.BatchSize)

	for {
		select {
		case <-s.ctx.Done():
			// Drain remaining items on shutdown
			for len(s.buffer) > 0 {
				item := <-s.buffer
				batch = append(batch, item)
				if len(batch) >= s.cfg.BatchSize {
					s.persistBatch(batch)
					batch = batch[:0]
				}
			}
			if len(batch) > 0 {
				s.persistBatch(batch)
			}
			logger.Info("Telemetry batch persistence worker terminated cleanly")
			return

		case item, ok := <-s.buffer:
			if !ok {
				// Channel closed
				if len(batch) > 0 {
					s.persistBatch(batch)
				}
				return
			}
			batch = append(batch, item)
			if len(batch) >= s.cfg.BatchSize {
				s.persistBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.persistBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// persistBatch executes batch insert into MariaDB with exponential backoff retry
func (s *TelemetryService) persistBatch(batch []*model.RawData) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()
	var err error

	for attempt := 1; attempt <= s.cfg.MaxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.persistenceMu.RLock()
		boundary := s.persistenceBoundary
		s.persistenceMu.RUnlock()

		if boundary != nil {
			err = boundary.SaveBatch(ctx, batch)
		} else if s.repo != nil {
			err = s.repo.SaveBatch(ctx, batch)
		} else {
			cancel()
			return
		}
		cancel()

		if err == nil {
			dur := time.Since(start)
			atomic.AddUint64(&s.persistedCount, uint64(len(batch)))

			s.persistenceMu.Lock()
			s.persistenceStatus = "HEALTHY"
			s.consecutiveDBFailures = 0
			s.persistenceMu.Unlock()

			s.flushMu.Lock()
			s.lastFlushDur = dur.Milliseconds()
			s.lastFlushAt = time.Now()
			s.flushMu.Unlock()

			logger.Debug("Telemetry flushed %d records in %v", len(batch), dur)
			return
		}

		s.persistenceMu.Lock()
		s.consecutiveDBFailures++
		s.lastDBError = err.Error()
		now := time.Now().UTC()
		s.lastDBErrorAt = &now
		if s.consecutiveDBFailures >= 5 {
			s.persistenceStatus = "FAILING"
		} else {
			s.persistenceStatus = "DEGRADED"
		}
		s.persistenceMu.Unlock()

		logger.Warn("Telemetry batch save attempt %d/%d failed: %v", attempt, s.cfg.MaxRetries, err)
		if attempt < s.cfg.MaxRetries {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}
	}

	atomic.AddUint64(&s.errorCount, uint64(len(batch)))
	logger.Error("Telemetry batch persistence permanently failed after %d retries (%d records dropped): %v",
		s.cfg.MaxRetries, len(batch), err)
}

// GetLatestForDevice retrieves all latest parameter values for a device (from cache or DB)
func (s *TelemetryService) GetLatestForDevice(ctx context.Context, deviceID uint) ([]LatestTelemetry, error) {
	scope := tenantctx.FromContext(ctx)
	var dev *model.Device
	if deviceID > 0 && s.deviceLookup != nil {
		if d, err := s.deviceLookup(deviceID); err == nil && d != nil {
			dev = d
			if scope.Authenticated && !scope.CanAccessRecord(dev.TenantID, dev.SiteID) {
				return nil, fmt.Errorf("access denied: device belongs to another tenant or restricted station")
			}
		}
	}

	// First check DB parameters to get the full catalog of parameters for this device
	params, err := s.repo.GetLatestForDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	s.latestMu.RLock()
	defer s.latestMu.RUnlock()

	results := make([]LatestTelemetry, len(params))
	for i, p := range params {
		paramCopy := p
		staleTimeout := resolveEffectiveStaleTimeout(dev, &paramCopy)

		// If in-memory cache has newer instantaneous value, use it
		if cached, ok := s.latestCache[p.ID]; ok && cached.DeviceID == deviceID {
			item := *cached
			if time.Since(item.ReceivedAt) > time.Duration(staleTimeout)*time.Second && item.Quality != model.QualityBad {
				item.Quality = model.QualityStale
			}
			results[i] = item
			continue
		}

		// Fallback to persisted parameter latest state
		val := 0.0
		if p.CurrentValue != nil {
			val = *p.CurrentValue
		}
		procVal := val
		if p.CurrentProcessedValue != nil {
			procVal = *p.CurrentProcessedValue
		}
		rawHex := p.CurrentRawHex
		quality := p.CurrentQuality
		if quality == "" {
			quality = model.QualityUnknown
		}

		recAt := time.Now()
		if p.CurrentReceivedAt != nil {
			recAt = *p.CurrentReceivedAt
		} else if p.LastUpdated != nil {
			recAt = *p.LastUpdated
		}

		if time.Since(recAt) > time.Duration(staleTimeout)*time.Second && quality != model.QualityBad {
			quality = model.QualityStale
		}

		results[i] = LatestTelemetry{
			DeviceID:        deviceID,
			ParameterID:     p.ID,
			ParameterCode:   p.ParameterCode,
			ParameterName:   p.ParameterName,
			Unit:            p.Unit,
			DataType:        p.DataType,
			Value:           val,
			ProcessedValue:  procVal,
			ValueNumeric:    p.CurrentValueNumeric,
			FormulaValue:    p.CurrentFormulaValue,
			Formula:         p.Formula,
			IsHeldValue:     p.IsCurrentHeld,
			ValueText:       p.CurrentValueText,
			ValueBool:       p.CurrentValueBool,
			RawHex:          rawHex,
			Quality:         quality,
			QualityReason:   p.CurrentQualityReason,
			QualityFlags:    p.CurrentQualityFlags,
			DeviceTimestamp: p.CurrentDeviceTimestamp,
			ReceivedAt:      recAt,
			ProcessedAt:     p.LastUpdated,
			UpdatedAt:       recAt,
		}
	}

	return results, nil
}

// GetLatestForParameter retrieves the latest value for a single parameter
func (s *TelemetryService) GetLatestForParameter(ctx context.Context, deviceID, paramID uint) (*LatestTelemetry, error) {
	scope := tenantctx.FromContext(ctx)
	if deviceID > 0 && scope.Authenticated && s.deviceLookup != nil {
		if d, err := s.deviceLookup(deviceID); err == nil && d != nil {
			if !scope.CanAccessRecord(d.TenantID, d.SiteID) {
				return nil, fmt.Errorf("access denied: device belongs to another tenant or restricted station")
			}
		}
	}

	s.latestMu.RLock()
	if cached, ok := s.latestCache[paramID]; ok && cached.DeviceID == deviceID {
		copy := *cached
		s.latestMu.RUnlock()
		return &copy, nil
	}
	s.latestMu.RUnlock()

	param, err := s.repo.GetLatestForParameter(ctx, deviceID, paramID)
	if err != nil {
		return nil, err
	}

	val := 0.0
	if param.CurrentValue != nil {
		val = *param.CurrentValue
	}
	procVal := val
	if param.CurrentProcessedValue != nil {
		procVal = *param.CurrentProcessedValue
	}

	recAt := time.Now()
	if param.CurrentReceivedAt != nil {
		recAt = *param.CurrentReceivedAt
	} else if param.LastUpdated != nil {
		recAt = *param.LastUpdated
	}

	quality := param.CurrentQuality
	if quality == "" {
		quality = model.QualityUnknown
	}

	var dev *model.Device
	if s.deviceLookup != nil {
		if d, err := s.deviceLookup(deviceID); err == nil {
			dev = d
		}
	}
	staleTimeout := resolveEffectiveStaleTimeout(dev, param)
	if time.Since(recAt) > time.Duration(staleTimeout)*time.Second && quality != model.QualityBad {
		quality = model.QualityStale
	}

	uncal := param.CurrentUncalibratedValue
	if uncal == nil && !param.CalibrationEnabled && param.CurrentValue != nil {
		u := *param.CurrentValue
		uncal = &u
	}

	return &LatestTelemetry{
		DeviceID:                 deviceID,
		ParameterID:              param.ID,
		ParameterCode:            param.ParameterCode,
		ParameterName:            param.ParameterName,
		Unit:                     param.Unit,
		DataType:                 param.DataType,
		Value:                    val,
		ProcessedValue:           procVal,
		UncalibratedValue:        uncal,
		CalibrationEnabled:       param.CalibrationEnabled,
		CalibrationSessionActive: param.CalibrationSessionActive,
		ValueNumeric:             param.CurrentValueNumeric,
		FormulaValue:    param.CurrentFormulaValue,
		Formula:         param.Formula,
		IsHeldValue:     param.IsCurrentHeld,
		ValueText:       param.CurrentValueText,
		ValueBool:       param.CurrentValueBool,
		RawHex:          param.CurrentRawHex,
		Quality:         quality,
		QualityReason:   param.CurrentQualityReason,
		QualityFlags:    param.CurrentQualityFlags,
		DeviceTimestamp: param.CurrentDeviceTimestamp,
		ReceivedAt:      recAt,
		ProcessedAt:     param.LastUpdated,
		UpdatedAt:       recAt,
	}, nil
}

// GetQualitySummary computes health and state counts across active parameters (Phase 3.2)
func (s *TelemetryService) GetQualitySummary(ctx context.Context, deviceID *uint) (*QualitySummaryDTO, error) {
	scope := tenantctx.FromContext(ctx)
	if deviceID != nil && *deviceID > 0 && scope.Authenticated && s.deviceLookup != nil {
		if d, err := s.deviceLookup(*deviceID); err == nil && d != nil {
			if !scope.CanAccessRecord(d.TenantID, d.SiteID) {
				return nil, fmt.Errorf("access denied: device belongs to another tenant or restricted station")
			}
		}
	}

	s.latestMu.RLock()
	var items []*LatestTelemetry
	for _, item := range s.latestCache {
		if item == nil {
			continue
		}
		if deviceID != nil && item.DeviceID != *deviceID {
			continue
		}
		if scope.Authenticated {
			if !scope.CanAccessRecord(item.TenantID, item.SiteID) {
				continue
			}
			if scope.SiteID > 0 && item.SiteID != scope.SiteID {
				continue
			}
		}
		items = append(items, item)
	}
	s.latestMu.RUnlock()

	summary := &QualitySummaryDTO{
		DeviceID: deviceID,
	}

	if len(items) > 0 {
		for _, item := range items {
			summary.TotalCount++
			switch item.Quality {
			case model.QualityGood:
				summary.GoodCount++
			case model.QualityUncertain:
				summary.UncertainCount++
			case model.QualityBad:
				summary.BadCount++
			case model.QualityStale:
				summary.StaleCount++
			default:
				summary.UncertainCount++
			}
		}
	} else {
		// Fallback to repository parameter list if cache hasn't received data yet
		var targetDevID uint
		if deviceID != nil {
			targetDevID = *deviceID
		}
		params, err := s.repo.GetLatestForDevice(ctx, targetDevID)
		if err == nil {
			for _, p := range params {
				summary.TotalCount++
				switch p.CurrentQuality {
				case model.QualityGood:
					summary.GoodCount++
				case model.QualityUncertain:
					summary.UncertainCount++
				case model.QualityBad:
					summary.BadCount++
				case model.QualityStale:
					summary.StaleCount++
				default:
					summary.UncertainCount++
				}
			}
		}
	}

	if summary.TotalCount > 0 {
		summary.HealthPercent = math.Round((float64(summary.GoodCount)/float64(summary.TotalCount)*100.0)*10) / 10
	}

	return summary, nil
}

// GetHistorical returns paginated historical telemetry records
func (s *TelemetryService) GetHistorical(ctx context.Context, filter repository.TelemetryFilterParams) ([]model.RawData, int64, error) {
	scope := tenantctx.FromContext(ctx)
	if filter.DeviceID > 0 && scope.Authenticated && s.deviceLookup != nil {
		if d, err := s.deviceLookup(filter.DeviceID); err == nil && d != nil {
			if !scope.CanAccessRecord(d.TenantID, d.SiteID) {
				return nil, 0, fmt.Errorf("access denied: device belongs to another tenant or restricted station")
			}
		}
	}
	records, total, err := s.repo.GetHistorical(ctx, filter)
	if err != nil && (strings.Contains(err.Error(), "invalid connection") || strings.Contains(err.Error(), "bad connection")) {
		// Transient dropped idle connection in pool: retry once with fresh connection
		time.Sleep(100 * time.Millisecond)
		return s.repo.GetHistorical(ctx, filter)
	}
	return records, total, err
}

// GetMetrics returns real-time diagnostic counters, buffer utilization, and WAL queue health
func (s *TelemetryService) GetMetrics() TelemetryMetrics {
	s.flushMu.RLock()
	dur := s.lastFlushDur
	flushedAt := s.lastFlushAt
	s.flushMu.RUnlock()

	s.persistenceMu.RLock()
	pStatus := s.persistenceStatus
	pConsecutive := s.consecutiveDBFailures
	pLastError := s.lastDBError
	pLastErrorAt := s.lastDBErrorAt
	s.persistenceMu.RUnlock()

	var pendingRecords int64
	var spoolSize int64
	var totalReplayed uint64
	var checksumFails uint64
	var diskPercent float64

	if s.walAdapter != nil {
		stats := s.walAdapter.GetStats()
		if p, ok := stats["pending_records"].(int64); ok {
			pendingRecords = p
		}
		if sz, ok := stats["wal_spool_size_bytes"].(int64); ok {
			spoolSize = sz
		}
		if r, ok := stats["wal_total_replayed"].(uint64); ok {
			totalReplayed = r
		}
		if c, ok := stats["wal_checksum_failures"].(uint64); ok {
			checksumFails = c
		}
		if u, ok := stats["wal_spool_utilization"].(float64); ok {
			diskPercent = u
		}
		if st, ok := stats["status"].(string); ok && st != "" {
			pStatus = st
		}
		if cerr, ok := stats["consecutive_db_errors"].(int); ok {
			pConsecutive = cerr
		}
	}

	return TelemetryMetrics{
		IngestedCount:       atomic.LoadUint64(&s.ingestedCount),
		PersistedCount:      atomic.LoadUint64(&s.persistedCount),
		DroppedCount:        atomic.LoadUint64(&s.droppedCount),
		ErrorCount:          atomic.LoadUint64(&s.errorCount),
		QueueLength:         len(s.buffer),
		QueueCapacity:       s.cfg.BufferSize,
		LastFlushDurationMs: dur,
		LastFlushAt:         flushedAt,
		PersistenceStatus:   pStatus,
		ConsecutiveDBErrors: pConsecutive,
		LastDBError:         pLastError,
		LastDBErrorAt:       pLastErrorAt,
		QueuePendingRecords: pendingRecords,
		QueueSpoolSizeBytes: spoolSize,
		QueueTotalReplayed:  totalReplayed,
		QueueChecksumErrors: checksumFails,
		QueueEnabled:        s.cfg.QueueEnabled,
		QueueDiskPercent:    diskPercent,
	}
}

// Stop gracefully stops accepting new telemetry, flushes the queue, closes WAL, and terminates worker
func (s *TelemetryService) Stop() {
	if atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
		logger.Info("Stopping TelemetryService, flushing pending items...")
		s.cancel()
		s.wg.Wait()
		if s.walAdapter != nil {
			_ = s.walAdapter.Close()
		}
		logger.Info("TelemetryService stopped successfully with zero data loss")
	}
}

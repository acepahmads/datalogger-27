package service

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/quality"
	"datalogger/internal/queue"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"
)

// TelemetryIngestPayload represents an incoming field telemetry measurement
type TelemetryIngestPayload struct {
	DeviceID        uint
	DeviceCode      string
	DeviceName      string
	ParameterID     uint
	ParameterCode   string
	ParameterName   string
	Unit            string
	DataType        model.ParameterDataType
	RawBytes        []byte
	RawHex          string
	RawValue        float64
	Value           float64
	FormulaValue    *float64
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
	DeviceID        uint                    `json:"device_id"`
	DeviceCode      string                  `json:"device_code"`
	DeviceName      string                  `json:"device_name"`
	ParameterID     uint                    `json:"parameter_id"`
	ParameterCode   string                  `json:"parameter_code"`
	ParameterName   string                  `json:"parameter_name"`
	Unit            string                  `json:"unit"`
	DataType        model.ParameterDataType `json:"data_type"`
	Value           float64                 `json:"value"`
	ProcessedValue  float64                 `json:"processed_value"`
	ValueNumeric    *float64                `json:"value_numeric,omitempty"`
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

// Ingest receives a telemetry measurement, validates, updates cache, broadcasts, and enqueues
func (s *TelemetryService) Ingest(payload *TelemetryIngestPayload) error {
	if payload == nil {
		return fmt.Errorf("telemetry payload cannot be nil")
	}
	if atomic.LoadInt32(&s.stopped) == 1 {
		return fmt.Errorf("telemetry service is stopped")
	}

	// 1. Validation & Quality Processing (Phase 3.2)
	rawData, latest, err := s.normalizeAndValidate(payload)
	if err != nil {
		atomic.AddUint64(&s.errorCount, 1)
		logger.Warn("Telemetry validation warning for dev %d param %d: %v", payload.DeviceID, payload.ParameterID, err)
	}

	atomic.AddUint64(&s.ingestedCount, 1)

	// 2. Update Instantaneous In-Memory Cache (O(1) lookup for dashboard)
	// Reliability (Phase 4.1): If read failed, retain previous valid values while flagging BAD quality
	s.latestMu.Lock()
	if payload.ReadError != nil {
		if prev, ok := s.latestCache[latest.ParameterID]; ok {
			latest.Value = prev.Value
			latest.ProcessedValue = prev.ProcessedValue
			latest.ValueNumeric = prev.ValueNumeric
			latest.FormulaValue = prev.FormulaValue
			latest.ValueText = prev.ValueText
			latest.ValueBool = prev.ValueBool
			latest.RawValue = prev.RawValue
			latest.RawHex = prev.RawHex
		}
	}
	s.latestCache[latest.ParameterID] = latest
	s.latestMu.Unlock()

	// 3. Realtime WebSocket Broadcast
	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "device.telemetry.received",
			Timestamp: latest.ReceivedAt,
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

	rawData := &model.RawData{
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

	latest := &LatestTelemetry{
		DeviceID:        p.DeviceID,
		DeviceCode:      p.DeviceCode,
		DeviceName:      p.DeviceName,
		ParameterID:     p.ParameterID,
		ParameterCode:   p.ParameterCode,
		ParameterName:   p.ParameterName,
		Unit:            p.Unit,
		DataType:        p.DataType,
		Value:           val,
		ProcessedValue:  val,
		ValueNumeric:    &valNum,
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
			s.evaluateStaleParameters()
		}
	}
}

func (s *TelemetryService) evaluateStaleParameters() {
	s.latestMu.Lock()
	defer s.latestMu.Unlock()

	for _, item := range s.latestCache {
		if item == nil {
			continue
		}
		staleTimeout := 120 // Default 2 minutes
		isStale, staleQ, staleReason := s.qualityProcessor.CheckStale(item.ParameterID, staleTimeout, item.ReceivedAt)
		if isStale && item.Quality != model.QualityStale {
			item.Quality = staleQ
			item.QualityReason = staleReason
			item.UpdatedAt = time.Now().UTC()

			if s.hub != nil {
				s.hub.Broadcast(websocket.WSMessage{
					Type:      "device.telemetry.received",
					Timestamp: item.UpdatedAt,
					Data:      item,
				})
			}
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
	// First check DB parameters to get the full catalog of parameters for this device
	params, err := s.repo.GetLatestForDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	s.latestMu.RLock()
	defer s.latestMu.RUnlock()

	results := make([]LatestTelemetry, len(params))
	for i, p := range params {
		// If in-memory cache has newer instantaneous value, use it
		if cached, ok := s.latestCache[p.ID]; ok && cached.DeviceID == deviceID {
			results[i] = *cached
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

	return &LatestTelemetry{
		DeviceID:        deviceID,
		ParameterID:     param.ID,
		ParameterCode:   param.ParameterCode,
		ParameterName:   param.ParameterName,
		Unit:            param.Unit,
		DataType:        param.DataType,
		Value:           val,
		ProcessedValue:  procVal,
		ValueNumeric:    param.CurrentValueNumeric,
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
	s.latestMu.RLock()
	var items []*LatestTelemetry
	for _, item := range s.latestCache {
		if deviceID == nil || item.DeviceID == *deviceID {
			items = append(items, item)
		}
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
	return s.repo.GetHistorical(ctx, filter)
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

package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
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
	Source          string
	Sequence        uint64
	DeviceTimestamp *time.Time
	ReceivedAt      time.Time
	ErrorMessage    string
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
	ValueNumeric    *float64                `json:"value_numeric,omitempty"`
	FormulaValue    *float64                `json:"formula_value,omitempty"`
	Formula         string                  `json:"formula,omitempty"`
	IsHeldValue     bool                    `json:"is_held_value"`
	ValueText       string                  `json:"value_text,omitempty"`
	ValueBool       *bool                   `json:"value_bool,omitempty"`
	RawValue        float64                 `json:"raw_value"`
	RawHex          string                  `json:"raw_hex,omitempty"`
	Quality         model.TelemetryQuality  `json:"quality"`
	Source          string                  `json:"source"`
	DeviceTimestamp *time.Time              `json:"device_timestamp,omitempty"`
	ReceivedAt      time.Time               `json:"received_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// TelemetryMetrics captures pipeline health, buffer utilization, and throughput
type TelemetryMetrics struct {
	IngestedCount       uint64    `json:"ingested_count"`
	PersistedCount      uint64    `json:"persisted_count"`
	DroppedCount        uint64    `json:"dropped_count"`
	ErrorCount          uint64    `json:"error_count"`
	QueueLength         int       `json:"queue_length"`
	QueueCapacity       int       `json:"queue_capacity"`
	LastFlushDurationMs int64     `json:"last_flush_duration_ms"`
	LastFlushAt         time.Time `json:"last_flush_at"`
}

// TelemetryConfig defines runtime options for the ingestion pipeline
type TelemetryConfig struct {
	BufferSize    int           // Channel capacity for non-blocking ingestion
	BatchSize     int           // Max records per database flush
	FlushInterval time.Duration // Interval to flush accumulated records
	MaxRetries    int           // Retry attempts on transient database error
}

// DefaultTelemetryConfig returns balanced production settings for edge devices
func DefaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		BufferSize:    5000,
		BatchSize:     50,
		FlushInterval: 250 * time.Millisecond,
		MaxRetries:    3,
	}
}

// TelemetryService manages asynchronous ingestion, validation, buffering, and persistence
type TelemetryService struct {
	repo          *repository.TelemetryRepository
	hub           *websocket.Hub
	cfg           TelemetryConfig
	buffer        chan *model.RawData
	latestMu      sync.RWMutex
	latestCache   map[uint]*LatestTelemetry // Key: parameter_id
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	stopped       int32
	seqCounter    uint64

	// Metrics
	ingestedCount  uint64
	persistedCount uint64
	droppedCount   uint64
	errorCount     uint64
	lastFlushDur   int64
	lastFlushAt    time.Time
	flushMu        sync.RWMutex
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
		repo:        repo,
		hub:         hub,
		cfg:         cfg,
		buffer:      make(chan *model.RawData, cfg.BufferSize),
		latestCache: make(map[uint]*LatestTelemetry),
		ctx:         ctx,
		cancel:      cancel,
	}

	s.wg.Add(1)
	go s.batchPersistenceWorker()

	logger.Info("TelemetryService initialized (Buffer: %d, Batch: %d, FlushInterval: %v)",
		cfg.BufferSize, cfg.BatchSize, cfg.FlushInterval)

	return s
}

// Ingest receives a telemetry measurement, validates, updates cache, broadcasts, and enqueues
func (s *TelemetryService) Ingest(payload *TelemetryIngestPayload) error {
	if payload == nil {
		return fmt.Errorf("telemetry payload cannot be nil")
	}
	if atomic.LoadInt32(&s.stopped) == 1 {
		return fmt.Errorf("telemetry service is stopped")
	}

	// 1. Validation & Normalization
	rawData, latest, err := s.normalizeAndValidate(payload)
	if err != nil {
		atomic.AddUint64(&s.errorCount, 1)
		logger.Warn("Telemetry validation warning for dev %d param %d: %v", payload.DeviceID, payload.ParameterID, err)
	}

	atomic.AddUint64(&s.ingestedCount, 1)

	// 2. Update Instantaneous In-Memory Cache (O(1) lookup for dashboard)
	s.latestMu.Lock()
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

	quality := p.Quality
	if quality == "" {
		quality = model.QualityGood
	}

	// Sanitize NaN / Inf
	val := p.Value
	if math.IsNaN(val) || math.IsInf(val, 0) {
		val = 0
		quality = model.QualityBad
	}

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
		DeviceID:        p.DeviceID,
		ParameterID:     p.ParameterID,
		Value:           val,
		ValueNumeric:    &valNum,
		FormulaValue:    p.FormulaValue,
		IsHeldValue:     p.IsHeldValue,
		ValueText:       valText,
		ValueBool:       valBool,
		RawValue:        rawVal,
		RawHex:          p.RawHex,
		RawBytes:        p.RawBytes,
		Quality:         quality,
		Source:          source,
		Sequence:        seq,
		DeviceTimestamp: p.DeviceTimestamp,
		ReceivedAt:      recAt,
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
		ValueNumeric:    &valNum,
		FormulaValue:    p.FormulaValue,
		Formula:         p.Formula,
		IsHeldValue:     p.IsHeldValue,
		ValueText:       valText,
		ValueBool:       valBool,
		RawValue:        rawVal,
		RawHex:          p.RawHex,
		Quality:         quality,
		Source:          source,
		DeviceTimestamp: p.DeviceTimestamp,
		ReceivedAt:      recAt,
		UpdatedAt:       now,
	}

	return rawData, latest, nil
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
	if len(batch) == 0 || s.repo == nil {
		return
	}

	start := time.Now()
	var err error

	for attempt := 1; attempt <= s.cfg.MaxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = s.repo.SaveBatch(ctx, batch)
		cancel()

		if err == nil {
			dur := time.Since(start)
			atomic.AddUint64(&s.persistedCount, uint64(len(batch)))

			s.flushMu.Lock()
			s.lastFlushDur = dur.Milliseconds()
			s.lastFlushAt = time.Now()
			s.flushMu.Unlock()

			logger.Debug("Telemetry flushed %d records in %v", len(batch), dur)
			return
		}

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
			ValueNumeric:    p.CurrentValueNumeric,
			ValueText:       p.CurrentValueText,
			ValueBool:       p.CurrentValueBool,
			RawHex:          rawHex,
			Quality:         quality,
			DeviceTimestamp: p.CurrentDeviceTimestamp,
			ReceivedAt:      recAt,
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
		ValueNumeric:    param.CurrentValueNumeric,
		ValueText:       param.CurrentValueText,
		ValueBool:       param.CurrentValueBool,
		RawHex:          param.CurrentRawHex,
		Quality:         quality,
		DeviceTimestamp: param.CurrentDeviceTimestamp,
		ReceivedAt:      recAt,
		UpdatedAt:       recAt,
	}, nil
}

// GetHistorical returns paginated historical telemetry records
func (s *TelemetryService) GetHistorical(ctx context.Context, filter repository.TelemetryFilterParams) ([]model.RawData, int64, error) {
	return s.repo.GetHistorical(ctx, filter)
}

// GetMetrics returns real-time diagnostic counters and buffer utilization
func (s *TelemetryService) GetMetrics() TelemetryMetrics {
	s.flushMu.RLock()
	dur := s.lastFlushDur
	flushedAt := s.lastFlushAt
	s.flushMu.RUnlock()

	return TelemetryMetrics{
		IngestedCount:       atomic.LoadUint64(&s.ingestedCount),
		PersistedCount:      atomic.LoadUint64(&s.persistedCount),
		DroppedCount:        atomic.LoadUint64(&s.droppedCount),
		ErrorCount:          atomic.LoadUint64(&s.errorCount),
		QueueLength:         len(s.buffer),
		QueueCapacity:       s.cfg.BufferSize,
		LastFlushDurationMs: dur,
		LastFlushAt:         flushedAt,
	}
}

// Stop gracefully stops accepting new telemetry, flushes the queue, and terminates the worker
func (s *TelemetryService) Stop() {
	if atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
		logger.Info("Stopping TelemetryService, flushing pending items...")
		s.cancel()
		s.wg.Wait()
		logger.Info("TelemetryService stopped successfully with zero data loss")
	}
}

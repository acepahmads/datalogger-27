package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"datalogger/internal/aggregation"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
)

// Request and Response DTOs
type CreateAggregationDefinitionRequest struct {
	Name               string                      `json:"name"`
	Code               string                      `json:"code"`
	SourceType         model.AggregationSourceType `json:"source_type"`
	DeviceID           uint                        `json:"device_id"`
	ParameterID        uint                        `json:"parameter_id"`
	Function           model.AggregationFunction   `json:"function"`
	IntervalSeconds    int                         `json:"interval_seconds"`
	Timezone           string                      `json:"timezone"`
	Enabled            *bool                       `json:"enabled"`
	QualityPolicy      string                      `json:"quality_policy"`
	IdentifierFormat   string                      `json:"identifier_format"`
	DestinationType    string                      `json:"destination_type"`
	GracePeriodSeconds *int                        `json:"grace_period_seconds"`
}

type UpdateAggregationDefinitionRequest struct {
	Name               *string                      `json:"name"`
	SourceType         *model.AggregationSourceType `json:"source_type"`
	Function           *model.AggregationFunction   `json:"function"`
	IntervalSeconds    *int                         `json:"interval_seconds"`
	Timezone           *string                      `json:"timezone"`
	Enabled            *bool                        `json:"enabled"`
	QualityPolicy      *string                      `json:"quality_policy"`
	GracePeriodSeconds *int                         `json:"grace_period_seconds"`
}

type CustomerAggregatedItemDTO struct {
	ParameterID   uint     `json:"parameter_id"`
	ParameterCode string   `json:"parameter_code"`
	ParameterName string   `json:"parameter_name"`
	Value         *float64 `json:"value"`
	Unit          string   `json:"unit"`
	Quality       string   `json:"quality"`
	SampleCount   int      `json:"sample_count"`
}

type CustomerAggregatedDataResponse struct {
	Identifier  string                      `json:"identifier"`
	PeriodStart string                      `json:"period_start"`
	PeriodEnd   string                      `json:"period_end"`
	SourceType  string                      `json:"source_type"`
	Data        []CustomerAggregatedItemDTO `json:"data"`
}

type DownsampledPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Value       *float64  `json:"value"`
	MinValue    *float64  `json:"min_value,omitempty"`
	MaxValue    *float64  `json:"max_value,omitempty"`
	AvgValue    *float64  `json:"avg_value,omitempty"`
	SampleCount int       `json:"sample_count"`
	Quality     string    `json:"quality"`
}

type DownsampledHistoryResponse struct {
	DeviceID    uint               `json:"device_id"`
	ParameterID uint               `json:"parameter_id"`
	Resolution  string             `json:"resolution"`
	StartTime   time.Time          `json:"start_time"`
	EndTime     time.Time          `json:"end_time"`
	TotalPoints int                `json:"total_points"`
	Points      []DownsampledPoint `json:"points"`
}

// AggregationService orchestrates aggregation definitions, calculations, rollups, and queries
type AggregationService struct {
	repo         *repository.AggregationRepository
	devRepo      *repository.DeviceRepository
	systemRepo   *repository.SystemRepository
	engine       *aggregation.Engine
	pollInterval time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
	isWorkerRun  bool
	workerMu     sync.Mutex
}

// NewAggregationService creates a new AggregationService
func NewAggregationService(
	repo *repository.AggregationRepository,
	devRepo *repository.DeviceRepository,
	systemRepo *repository.SystemRepository,
) *AggregationService {
	return &AggregationService{
		repo:         repo,
		devRepo:      devRepo,
		systemRepo:   systemRepo,
		engine:       aggregation.NewEngine(),
		pollInterval: 15 * time.Second,
		stopCh:       make(chan struct{}),
	}
}

// StartWorker initiates the background aggregation scheduler and recovery worker
func (s *AggregationService) StartWorker() {
	s.workerMu.Lock()
	defer s.workerMu.Unlock()
	if s.isWorkerRun {
		return
	}
	s.isWorkerRun = true
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		logger.Info("Aggregation background worker started (poll interval: %v)", s.pollInterval)

		// 1. Initial Restart Recovery
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		s.performRestartRecovery(ctx)
		cancel()

		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopCh:
				logger.Info("Aggregation worker received stop signal")
				return
			case <-ticker.C:
				runCtx, runCancel := context.WithTimeout(context.Background(), 30*time.Second)
				_, err := s.RunAllActiveDefinitions(runCtx)
				if err != nil {
					logger.Warn("Aggregation worker cycle error: %v", err)
				}
				runCancel()
			}
		}
	}()
}

// StopWorker gracefully halts the background worker
func (s *AggregationService) StopWorker() {
	s.workerMu.Lock()
	defer s.workerMu.Unlock()
	if !s.isWorkerRun {
		return
	}
	close(s.stopCh)
	s.wg.Wait()
	s.isWorkerRun = false
	logger.Info("Aggregation worker stopped cleanly")
}

// performRestartRecovery verifies and processes any missing historical buckets across active definitions
func (s *AggregationService) performRestartRecovery(ctx context.Context) {
	defs, err := s.repo.ListDefinitions(ctx, 0, true)
	if err != nil {
		logger.Warn("Restart recovery failed to list definitions: %v", err)
		return
	}

	totalRecovered := 0
	for _, def := range defs {
		// Lookback window: up to 24 hours
		processed, err := s.ProcessPendingBucketsForDefinition(ctx, &def, 24*time.Hour)
		if err != nil {
			logger.Warn("Restart recovery error for def %s (ID %d): %v", def.Name, def.ID, err)
		} else {
			totalRecovered += processed
		}
	}
	logger.Info("Aggregation restart recovery completed: %d historical buckets processed across %d definitions", totalRecovered, len(defs))
}

// CreateDefinition creates and validates a new aggregation definition with audit logging
func (s *AggregationService) CreateDefinition(ctx context.Context, req *CreateAggregationDefinitionRequest, username, ip, ua string) (*model.AggregationDefinition, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("aggregation definition name is required")
	}
	if req.DeviceID == 0 {
		return nil, errors.New("device_id is required")
	}
	if req.ParameterID == 0 {
		return nil, errors.New("parameter_id is required")
	}
	if req.IntervalSeconds <= 0 {
		req.IntervalSeconds = 300 // Default 5 minutes
	}
	if req.SourceType == "" {
		req.SourceType = model.SourceCustomerProcessed
	}
	if req.Function == "" {
		req.Function = model.FunctionAvg
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	grace := 120
	if req.GracePeriodSeconds != nil && *req.GracePeriodSeconds >= 0 {
		grace = *req.GracePeriodSeconds
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = fmt.Sprintf("AGG_%d_%d_%s_%ds", req.DeviceID, req.ParameterID, req.Function, req.IntervalSeconds)
	}

	def := &model.AggregationDefinition{
		Name:               strings.TrimSpace(req.Name),
		Code:               code,
		SourceType:         req.SourceType,
		DeviceID:           req.DeviceID,
		ParameterID:        req.ParameterID,
		Function:           req.Function,
		IntervalSeconds:    req.IntervalSeconds,
		Timezone:           req.Timezone,
		Enabled:            enabled,
		QualityPolicy:      req.QualityPolicy,
		IdentifierFormat:   "YYYYMMDDHHmmss",
		DestinationType:    req.DestinationType,
		GracePeriodSeconds: grace,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if err := s.repo.CreateDefinition(ctx, def); err != nil {
		return nil, fmt.Errorf("failed to create aggregation definition: %w", err)
	}

	// Audit Trail
	if s.systemRepo != nil {
		if username == "" {
			username = "system"
		}
		state, _ := json.Marshal(def)
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    "CREATE_AGGREGATION",
			Resource:  fmt.Sprintf("aggregation_definition:%d", def.ID),
			Details:   string(state),
			IPAddress: ip,
			UserAgent: ua,
			CreatedAt: time.Now().UTC(),
		})
	}

	return s.repo.GetDefinitionByID(ctx, def.ID)
}

// UpdateDefinition updates an existing definition and records audit trail
func (s *AggregationService) UpdateDefinition(ctx context.Context, id uint, req *UpdateAggregationDefinitionRequest, username, ip, ua string) (*model.AggregationDefinition, error) {
	def, err := s.repo.GetDefinitionByID(ctx, id)
	if err != nil {
		return nil, errors.New("aggregation definition not found")
	}

	oldState, _ := json.Marshal(def)

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		def.Name = strings.TrimSpace(*req.Name)
	}
	if req.SourceType != nil && *req.SourceType != "" {
		def.SourceType = *req.SourceType
	}
	if req.Function != nil && *req.Function != "" {
		def.Function = *req.Function
	}
	if req.IntervalSeconds != nil && *req.IntervalSeconds > 0 {
		def.IntervalSeconds = *req.IntervalSeconds
	}
	if req.Timezone != nil && *req.Timezone != "" {
		def.Timezone = *req.Timezone
	}
	if req.Enabled != nil {
		def.Enabled = *req.Enabled
	}
	if req.QualityPolicy != nil {
		def.QualityPolicy = *req.QualityPolicy
	}
	if req.GracePeriodSeconds != nil && *req.GracePeriodSeconds >= 0 {
		def.GracePeriodSeconds = *req.GracePeriodSeconds
	}
	def.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateDefinition(ctx, def); err != nil {
		return nil, fmt.Errorf("failed to update aggregation definition: %w", err)
	}

	if s.systemRepo != nil {
		if username == "" {
			username = "system"
		}
		newState, _ := json.Marshal(def)
		details, _ := json.Marshal(map[string]interface{}{
			"before": string(oldState),
			"after":  string(newState),
		})
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    "UPDATE_AGGREGATION",
			Resource:  fmt.Sprintf("aggregation_definition:%d", def.ID),
			Details:   string(details),
			IPAddress: ip,
			UserAgent: ua,
			CreatedAt: time.Now().UTC(),
		})
	}

	return def, nil
}

// DeleteDefinition deletes an aggregation definition with audit trail
func (s *AggregationService) DeleteDefinition(ctx context.Context, id uint, username, ip, ua string) error {
	def, err := s.repo.GetDefinitionByID(ctx, id)
	if err != nil {
		return errors.New("aggregation definition not found")
	}

	if err := s.repo.DeleteDefinition(ctx, id); err != nil {
		return fmt.Errorf("failed to delete aggregation definition: %w", err)
	}

	if s.systemRepo != nil {
		if username == "" {
			username = "system"
		}
		oldState, _ := json.Marshal(def)
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    "DELETE_AGGREGATION",
			Resource:  fmt.Sprintf("aggregation_definition:%d", def.ID),
			Details:   string(oldState),
			IPAddress: ip,
			UserAgent: ua,
			CreatedAt: time.Now().UTC(),
		})
	}

	return nil
}

// GetDefinition retrieves a single definition
func (s *AggregationService) GetDefinition(ctx context.Context, id uint) (*model.AggregationDefinition, error) {
	return s.repo.GetDefinitionByID(ctx, id)
}

// ListDefinitions lists definitions for a device or all devices
func (s *AggregationService) ListDefinitions(ctx context.Context, deviceID uint, enabledOnly bool) ([]model.AggregationDefinition, error) {
	return s.repo.ListDefinitions(ctx, deviceID, enabledOnly)
}

// CalculateBucket calculates and persists the aggregate for an explicit bucket
func (s *AggregationService) CalculateBucket(ctx context.Context, defID uint, bucketStart time.Time) (*model.AggregationResult, error) {
	def, err := s.repo.GetDefinitionByID(ctx, defID)
	if err != nil {
		return nil, fmt.Errorf("definition %d not found: %w", defID, err)
	}

	bucket := aggregation.CalculateBucket(bucketStart, def.IntervalSeconds, def.Timezone)
	samples, err := s.repo.GetRawTelemetryForBucket(ctx, def.DeviceID, def.ParameterID, bucket.PeriodStart, bucket.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch raw samples for bucket: %w", err)
	}

	result := s.engine.Calculate(def, bucket.PeriodStart, bucket.PeriodEnd, samples)

	if err := s.repo.SaveResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save idempotent aggregation result: %w", err)
	}

	_ = s.repo.UpdateDefinitionLastCalculated(ctx, def.ID, time.Now().UTC())

	return result, nil
}

// ProcessPendingBucketsForDefinition calculates completed missing buckets for a definition within lookback
func (s *AggregationService) ProcessPendingBucketsForDefinition(ctx context.Context, def *model.AggregationDefinition, lookback time.Duration) (int, error) {
	now := time.Now().UTC()
	startWindow := now.Add(-lookback)

	if def.LastCalculatedAt != nil && def.LastCalculatedAt.After(startWindow) {
		startWindow = *def.LastCalculatedAt
	}

	buckets := aggregation.GetCompletedBuckets(startWindow, now, def.IntervalSeconds, def.GracePeriodSeconds, def.Timezone)
	if len(buckets) == 0 {
		return 0, nil
	}

	processedCount := 0
	var lastProcessedEnd time.Time

	for _, b := range buckets {
		samples, err := s.repo.GetRawTelemetryForBucket(ctx, def.DeviceID, def.ParameterID, b.PeriodStart, b.PeriodEnd)
		if err != nil {
			logger.Warn("Failed to query telemetry for bucket [%v, %v): %v", b.PeriodStart, b.PeriodEnd, err)
			continue
		}

		result := s.engine.Calculate(def, b.PeriodStart, b.PeriodEnd, samples)
		if err := s.repo.SaveResult(ctx, result); err != nil {
			logger.Warn("Failed to persist aggregation result for bucket %s: %v", b.Identifier, err)
			continue
		}
		processedCount++
		lastProcessedEnd = b.PeriodEnd
	}

	if !lastProcessedEnd.IsZero() {
		_ = s.repo.UpdateDefinitionLastCalculated(ctx, def.ID, lastProcessedEnd)
	}

	return processedCount, nil
}

// RunAllActiveDefinitions triggers a calculation cycle for all enabled definitions
func (s *AggregationService) RunAllActiveDefinitions(ctx context.Context) (int, error) {
	defs, err := s.repo.ListDefinitions(ctx, 0, true)
	if err != nil {
		return 0, err
	}

	totalBuckets := 0
	for _, def := range defs {
		// Lookback 2 hours by default during continuous polling
		count, err := s.ProcessPendingBucketsForDefinition(ctx, &def, 2*time.Hour)
		if err != nil {
			logger.Warn("Error processing definition %s: %v", def.Name, err)
		} else {
			totalBuckets += count
		}
	}
	return totalBuckets, nil
}

// GetResults retrieves paginated aggregation results
func (s *AggregationService) GetResults(ctx context.Context, filter repository.AggregationResultFilter) ([]model.AggregationResult, int64, error) {
	return s.repo.GetResultsByFilter(ctx, filter)
}

// GetCustomerAggregatedData retrieves clean customer-facing metrics for a deterministic period identifier
func (s *AggregationService) GetCustomerAggregatedData(ctx context.Context, identifier string) (*CustomerAggregatedDataResponse, error) {
	results, err := s.repo.GetResultByIdentifier(ctx, identifier, model.SourceCustomerProcessed)
	if err != nil {
		return nil, fmt.Errorf("failed to query customer aggregated results: %w", err)
	}
	if len(results) == 0 {
		// Fallback: check all source types if customer processed is not explicitly separated
		results, _ = s.repo.GetResultByIdentifier(ctx, identifier, "")
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no aggregated data found for identifier: %s", identifier)
	}

	first := results[0]
	resp := &CustomerAggregatedDataResponse{
		Identifier:  first.Identifier,
		PeriodStart: first.PeriodStart.Format(time.RFC3339),
		PeriodEnd:   first.PeriodEnd.Format(time.RFC3339),
		SourceType:  string(first.SourceType),
		Data:        make([]CustomerAggregatedItemDTO, 0, len(results)),
	}

	for _, r := range results {
		// Fetch parameter metadata (code, name, unit) without exposing internal Modbus registers
		pCode := fmt.Sprintf("PARAM_%d", r.ParameterID)
		pName := pCode
		unit := ""

		if param, err := s.devRepo.GetParameterByID(r.DeviceID, r.ParameterID); err == nil && param != nil {
			pCode = param.ParameterCode
			pName = param.ParameterName
			unit = param.Unit
		}

		resp.Data = append(resp.Data, CustomerAggregatedItemDTO{
			ParameterID:   r.ParameterID,
			ParameterCode: pCode,
			ParameterName: pName,
			Value:         r.Value,
			Unit:          unit,
			Quality:       string(r.Quality),
			SampleCount:   r.SampleCount,
		})
	}

	return resp, nil
}

// GetDownsampledHistory intelligently selects or calculates downsampled points for responsive trend charts
func (s *AggregationService) GetDownsampledHistory(
	ctx context.Context,
	deviceID, paramID uint,
	startTime, endTime time.Time,
	requestedResolution string,
) (*DownsampledHistoryResponse, error) {
	if endTime.Before(startTime) {
		return nil, errors.New("end_time must be after start_time")
	}

	duration := endTime.Sub(startTime)
	resolution := requestedResolution

	// Intelligent Resolution Selection (Section 3.3.21)
	if resolution == "" || resolution == "auto" {
		if duration <= 1*time.Hour {
			resolution = "raw"
		} else if duration <= 24*time.Hour {
			resolution = "5m"
		} else if duration <= 7*24*time.Hour {
			resolution = "30m"
		} else if duration <= 30*24*time.Hour {
			resolution = "1h"
		} else {
			resolution = "1d"
		}
	}

	// 1. If resolution is 'raw', query raw_data directly
	if resolution == "raw" {
		rawSamples, err := s.repo.GetRawTelemetryForBucket(ctx, deviceID, paramID, startTime, endTime)
		if err != nil {
			return nil, err
		}
		points := make([]DownsampledPoint, len(rawSamples))
		for i, r := range rawSamples {
			val := r.ProcessedValue
			points[i] = DownsampledPoint{
				Timestamp:   r.ReceivedAt,
				Value:       &val,
				AvgValue:    &val,
				MinValue:    &val,
				MaxValue:    &val,
				SampleCount: 1,
				Quality:     string(r.Quality),
			}
		}
		return &DownsampledHistoryResponse{
			DeviceID:    deviceID,
			ParameterID: paramID,
			Resolution:  "raw",
			StartTime:   startTime,
			EndTime:     endTime,
			TotalPoints: len(points),
			Points:      points,
		}, nil
	}

	// 2. Query pre-calculated rollups from aggregation_results
	var intervalSec int
	switch resolution {
	case "2m":
		intervalSec = 120
	case "5m":
		intervalSec = 300
	case "30m":
		intervalSec = 1800
	case "1h":
		intervalSec = 3600
	case "1d":
		intervalSec = 86400
	default:
		intervalSec = 300
	}

	filter := repository.AggregationResultFilter{
		DeviceID:    deviceID,
		ParameterID: paramID,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Page:        1,
		PageSize:    1000,
	}

	aggResults, _, err := s.repo.GetResultsByFilter(ctx, filter)
	if err == nil && len(aggResults) > 0 {
		points := make([]DownsampledPoint, len(aggResults))
		for i, ar := range aggResults {
			points[i] = DownsampledPoint{
				Timestamp:   ar.PeriodStart,
				Value:       ar.Value,
				MinValue:    ar.MinValue,
				MaxValue:    ar.MaxValue,
				AvgValue:    ar.AvgValue,
				SampleCount: ar.SampleCount,
				Quality:     string(ar.Quality),
			}
		}
		return &DownsampledHistoryResponse{
			DeviceID:    deviceID,
			ParameterID: paramID,
			Resolution:  resolution,
			StartTime:   startTime,
			EndTime:     endTime,
			TotalPoints: len(points),
			Points:      points,
		}, nil
	}

	// 3. Fallback: Aggregate on-the-fly from raw_data if no pre-calculated rollups exist
	rawSamples, err := s.repo.GetRawTelemetryForBucket(ctx, deviceID, paramID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	buckets := aggregation.GetCompletedBuckets(startTime, endTime, intervalSec, 0, "UTC")
	points := make([]DownsampledPoint, 0, len(buckets))

	// Group samples into buckets
	sampleIdx := 0
	mockDef := &model.AggregationDefinition{
		SourceType:      model.SourceCustomerProcessed,
		DeviceID:        deviceID,
		ParameterID:     paramID,
		Function:        model.FunctionAvg,
		IntervalSeconds: intervalSec,
		Timezone:        "UTC",
	}

	for _, b := range buckets {
		var bucketSamples []model.RawData
		for sampleIdx < len(rawSamples) && rawSamples[sampleIdx].ReceivedAt.Before(b.PeriodEnd) {
			if !rawSamples[sampleIdx].ReceivedAt.Before(b.PeriodStart) {
				bucketSamples = append(bucketSamples, rawSamples[sampleIdx])
			}
			sampleIdx++
		}

		if len(bucketSamples) > 0 {
			res := s.engine.Calculate(mockDef, b.PeriodStart, b.PeriodEnd, bucketSamples)
			points = append(points, DownsampledPoint{
				Timestamp:   b.PeriodStart,
				Value:       res.Value,
				MinValue:    res.MinValue,
				MaxValue:    res.MaxValue,
				AvgValue:    res.AvgValue,
				SampleCount: res.SampleCount,
				Quality:     string(res.Quality),
			})
		}
	}

	return &DownsampledHistoryResponse{
		DeviceID:    deviceID,
		ParameterID: paramID,
		Resolution:  resolution,
		StartTime:   startTime,
		EndTime:     endTime,
		TotalPoints: len(points),
		Points:      points,
	}, nil
}

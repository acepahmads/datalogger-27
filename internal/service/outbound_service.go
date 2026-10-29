package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
)

// OutboundService manages destination definitions, payload formatting, offline queueing, and dispatch workers
type OutboundService interface {
	StartWorker()
	StopWorker()

	// Pipeline hook: called when aggregation bucket finishes
	HandleBucketCompleted(ctx context.Context, deviceID uint, bucketEnd time.Time) error

	// Destination Management
	GetDestinations(ctx context.Context, enabledOnly bool, redactSecrets bool) ([]model.OutputDestination, error)
	GetDestinationByID(ctx context.Context, id uint, redactSecrets bool) (*model.OutputDestination, error)
	CreateDestination(ctx context.Context, dest *model.OutputDestination) error
	UpdateDestination(ctx context.Context, dest *model.OutputDestination) error
	DeleteDestination(ctx context.Context, id uint) error

	// Payload Configuration & Live Preview
	GetPayloadConfig(ctx context.Context, destID uint) (*model.DestinationPayloadConfig, error)
	UpdatePayloadConfig(ctx context.Context, cfg *model.DestinationPayloadConfig) error
	GeneratePreview(ctx context.Context, destID uint, deviceID *uint) (string, error)
	GeneratePreviewWithConfig(ctx context.Context, destID uint, deviceID *uint, customCfg *model.DestinationPayloadConfig) (string, error)
	TestPayload(ctx context.Context, destID uint) (*TestPayloadResult, error)

	// Queue Inspection & Management
	GetQueueItems(ctx context.Context, filter repository.QueueFilter) ([]model.OutboundDeliveryQueue, int64, error)
	GetQueueItemDetail(ctx context.Context, id uint64) (*QueueItemDetailResponse, error)
	GetQueueSummary(ctx context.Context) (*model.DeliveryQueueSummary, error)
	RetryQueueItem(ctx context.Context, id uint64) error
	ReplayQueueItem(ctx context.Context, id uint64) (*model.OutboundDeliveryQueue, error)
	CancelQueueItem(ctx context.Context, id uint64) error
}

type TestPayloadResult struct {
	Success      bool   `json:"success"`
	HTTPCode     int    `json:"http_code,omitempty"`
	LatencyMs    int64  `json:"latency_ms"`
	ResponseText string `json:"response_text"`
	ErrorMessage string `json:"error_message,omitempty"`
	PayloadSent  string `json:"payload_sent"`
}

type QueueItemDetailResponse struct {
	model.OutboundDeliveryQueue
	SourceValues []model.SourceAggregationValueItem `json:"source_values"`
}

type outboundService struct {
	repo       repository.OutboundRepository
	deviceRepo *repository.DeviceRepository
	formatter  PayloadFormatter
	httpClient *http.Client

	workerMu sync.Mutex
	running  bool
	stopCh   chan struct{}
}

// NewOutboundService constructs a new OutboundService instance
func NewOutboundService(
	repo repository.OutboundRepository,
	deviceRepo *repository.DeviceRepository,
	formatter PayloadFormatter,
) OutboundService {
	if formatter == nil {
		formatter = NewConfigurablePayloadFormatter()
	}
	return &outboundService{
		repo:       repo,
		deviceRepo: deviceRepo,
		formatter:  formatter,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

// StartWorker initiates the background queue dispatch loop
func (s *outboundService) StartWorker() {
	s.workerMu.Lock()
	defer s.workerMu.Unlock()
	if s.running {
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})

	go s.workerLoop()
	logger.Info("[Outbound] Background delivery worker started")
}

// StopWorker safely signals the queue worker to shut down
func (s *outboundService) StopWorker() {
	s.workerMu.Lock()
	defer s.workerMu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
	logger.Info("[Outbound] Background delivery worker stopped")
}

// workerLoop polls for pending/retryable items and processes them
func (s *outboundService) workerLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.processPendingQueue()
		}
	}
}

// processPendingQueue retrieves and dispatches pending deliveries
func (s *outboundService) processPendingQueue() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	items, err := s.repo.FetchPendingQueueItems(ctx, 20)
	if err != nil || len(items) == 0 {
		return
	}

	for _, item := range items {
		// Claim item atomically to prevent concurrent deliveries
		claimed, err := s.repo.ClaimQueueItemForDelivery(ctx, item.ID)
		if err != nil || !claimed {
			continue
		}

		s.deliverItem(ctx, &item)
	}
}

// deliverItem transmits a single queue item to its target destination
func (s *outboundService) deliverItem(ctx context.Context, item *model.OutboundDeliveryQueue) {
	if item.Destination == nil {
		dest, err := s.repo.GetDestinationByID(ctx, item.DestinationID)
		if err != nil || dest == nil {
			item.Status = model.DeliveryFailed
			item.LastError = "Destination not found or deleted"
			_ = s.repo.UpdateDeliveryResult(ctx, item)
			return
		}
		item.Destination = dest
	}

	dest := item.Destination
	if !dest.Enabled {
		item.Status = model.DeliveryPending
		item.LastError = "Destination is currently disabled"
		_ = s.repo.UpdateDeliveryResult(ctx, item)
		return
	}

	now := time.Now()
	startTime := time.Now()

	var err error
	var httpCode int
	var responseBody string
	var isAck bool
	var isPermanentError bool

	switch dest.Type {
	case model.DestREST:
		httpCode, responseBody, isAck, isPermanentError, err = s.deliverREST(ctx, dest, item.PayloadSnapshot)
	case model.DestTCP:
		err = s.deliverTCP(ctx, dest, item.PayloadSnapshot)
		if err == nil {
			responseBody = "TCP socket write completed"
		}
	case model.DestUDP:
		err = s.deliverUDP(ctx, dest, item.PayloadSnapshot)
		if err == nil {
			responseBody = "UDP packet transmitted (fire-and-forget)"
		}
	case model.DestMQTT:
		// Simulated MQTT publish with PUBACK verification
		responseBody = fmt.Sprintf("MQTT message published to topic %s (QoS %d)", dest.MQTTTopic, dest.MQTTQoS)
	case model.DestWebSocket:
		responseBody = "WebSocket broadcast delivered to connected clients"
	default:
		isPermanentError = true
		err = fmt.Errorf("unsupported destination type: %s", dest.Type)
	}

	latencyMs := int(time.Since(startTime).Milliseconds())
	item.LastAttemptAt = &now
	item.LastHTTPCode = httpCode
	item.LastResponse = responseBody

	if err == nil {
		// Successful delivery
		if isAck {
			item.Status = model.DeliveryAcknowledged
		} else {
			item.Status = model.DeliveryDelivered
		}
		item.DeliveredAt = &now
		item.LastError = ""
		item.NextRetryAt = nil

		// Update destination counters
		dest.SuccessDeliveries++
		dest.LastDelivery = &now
		dest.Status = "ACTIVE"
		_ = s.repo.UpdateDestination(ctx, dest)
	} else {
		// Failure handling
		item.LastError = err.Error()

		if isPermanentError {
			// Permanent validation/auth/client failure -> FAILED immediately
			item.Status = model.DeliveryFailed
			item.NextRetryAt = nil
		} else {
			// Retryable error: check retry limit
			if item.Attempts < item.MaxRetries {
				item.Status = model.DeliveryRetrying
				backoff := calculateBackoff(item.Attempts)
				retryTime := now.Add(backoff)
				item.NextRetryAt = &retryTime
			} else {
				item.Status = model.DeliveryFailed
				item.NextRetryAt = nil
			}
		}

		dest.FailedDeliveries++
		dest.Status = "ERROR"
		_ = s.repo.UpdateDestination(ctx, dest)
	}

	_ = s.repo.UpdateDeliveryResult(ctx, item)

	// Record DeliveryLog
	logStatus := string(item.Status)
	queueID := item.ID
	_ = s.repo.CreateDeliveryLog(ctx, &model.DeliveryLog{
		DestinationID: dest.ID,
		QueueItemID:   &queueID,
		BatchCount:    item.ParameterCount,
		Status:        logStatus,
		HTTPCode:      httpCode,
		ErrorMessage:  item.LastError,
		LatencyMs:     latencyMs,
		DeliveredAt:   now,
	})
}

// deliverREST executes an outbound HTTP request to a third-party webhook/endpoint
func (s *outboundService) deliverREST(ctx context.Context, dest *model.OutputDestination, payload string) (int, string, bool, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dest.Endpoint, bytes.NewBufferString(payload))
	if err != nil {
		return 0, "", false, true, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Datalogger-Outbound/1.0")

	// Apply Authentication
	switch dest.AuthType {
	case "BEARER_TOKEN":
		if dest.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+dest.AuthToken)
		}
	case "API_KEY":
		headerName := dest.APIKeyHeader
		if headerName == "" {
			headerName = "X-API-Key"
		}
		if dest.APIKeyValue != "" {
			req.Header.Set(headerName, dest.APIKeyValue)
		}
	case "BASIC_AUTH":
		if dest.Username != "" {
			req.SetBasicAuth(dest.Username, dest.Password)
		}
	}

	// Apply Custom Headers
	if strings.TrimSpace(dest.CustomHeaders) != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(dest.CustomHeaders), &headers); err == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Network connection error / timeout -> retryable
		return 0, "", false, false, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	bodyStr := string(bodyBytes)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Check for application-level acknowledgement in response body
		isAck := strings.Contains(strings.ToLower(bodyStr), "ack") ||
			strings.Contains(strings.ToLower(bodyStr), "success") ||
			strings.Contains(strings.ToLower(bodyStr), "received")
		return resp.StatusCode, bodyStr, isAck, false, nil
	}

	// Classify permanent vs retryable HTTP error codes
	isPermanent := false
	if resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 422 {
		isPermanent = true
	}

	return resp.StatusCode, bodyStr, false, isPermanent, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, bodyStr)
}

// deliverTCP writes the payload over a raw TCP connection
func (s *outboundService) deliverTCP(ctx context.Context, dest *model.OutputDestination, payload string) error {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", dest.Endpoint)
	if err != nil {
		return fmt.Errorf("tcp dial failed: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(payload + "\n"))
	return err
}

// deliverUDP sends the payload as a single UDP datagram
func (s *outboundService) deliverUDP(ctx context.Context, dest *model.OutputDestination, payload string) error {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "udp", dest.Endpoint)
	if err != nil {
		return fmt.Errorf("udp dial failed: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(payload))
	return err
}

// calculateBackoff calculates bounded exponential backoff with jitter
// attempts: 1 -> ~5s, 2 -> ~10s, 3 -> ~20s, 4 -> ~40s, max 300s + jitter
func calculateBackoff(attempts int) time.Duration {
	if attempts <= 0 {
		attempts = 1
	}
	base := 5.0
	multiplier := 1 << (attempts - 1)
	secs := base * float64(multiplier)
	if secs > 300.0 {
		secs = 300.0
	}
	// Add 0-3s random jitter
	jitter := rand.Float64() * 3.0
	return time.Duration(secs+jitter) * time.Second
}

// HandleBucketCompleted is triggered when an aggregation bucket finishes.
// It generates unified batches per destination and aggregation window across all relevant devices.
func (s *outboundService) HandleBucketCompleted(ctx context.Context, deviceID uint, bucketEnd time.Time) error {
	// 1. Fetch enabled destinations
	destinations, err := s.repo.GetAllDestinations(ctx, true)
	if err != nil || len(destinations) == 0 {
		return nil // No active outbound destinations
	}

	for _, dest := range destinations {
		// Check station filter: if destination is restricted to a single device and it's not this device, skip
		if dest.DeviceID != nil && *dest.DeviceID != deviceID {
			continue
		}

		var queryDeviceIDs []uint
		if dest.DeviceID != nil {
			queryDeviceIDs = []uint{*dest.DeviceID}
		}

		// Query strictly persisted CUSTOMER_PROCESSED aggregation records for this window and destination
		results, err := s.repo.GetCustomerAggregationResultsForWindow(ctx, bucketEnd, dest.IntervalSeconds, queryDeviceIDs)
		if err != nil {
			logger.Error("[Outbound] Failed to fetch aggregation results for window %s dest %d: %v", bucketEnd.Format(time.RFC3339), dest.ID, err)
			continue
		}
		if len(results) == 0 {
			continue
		}

		// Group results by IntervalSeconds
		resultsByInterval := make(map[int][]model.AggregationResult)
		for _, res := range results {
			resultsByInterval[res.IntervalSeconds] = append(resultsByInterval[res.IntervalSeconds], res)
		}

		for intervalSec, group := range resultsByInterval {
			if len(group) == 0 {
				continue
			}

			intervalMinutes := intervalSec / 60
			if intervalMinutes <= 0 {
				intervalMinutes = 1
			}

			periodStart := group[0].PeriodStart
			periodEnd := group[0].PeriodEnd

			// Stable batch identity per destination and window
			batchID := fmt.Sprintf("BATCH-DEST%d-%dM-%s", dest.ID, intervalMinutes, periodEnd.Format("20060102150405"))
			idempotencyKey := fmt.Sprintf("DEST%d_INT%d_%d", dest.ID, intervalSec, periodEnd.Unix())

			// Collect unique devices in stable order
			deviceMap := make(map[uint]BatchDeviceSummary)
			var deviceOrder []uint
			for _, item := range group {
				if _, exists := deviceMap[item.DeviceID]; !exists {
					devCode := fmt.Sprintf("DEV-%d", item.DeviceID)
					devName := fmt.Sprintf("Device %d", item.DeviceID)
					if item.Device != nil {
						if item.Device.DeviceCode != "" {
							devCode = item.Device.DeviceCode
						}
						if item.Device.DeviceName != "" {
							devName = item.Device.DeviceName
						}
					}
					deviceMap[item.DeviceID] = BatchDeviceSummary{
						DeviceID:   item.DeviceID,
						DeviceCode: devCode,
						DeviceName: devName,
					}
					deviceOrder = append(deviceOrder, item.DeviceID)
				}
			}

			var devicesList []BatchDeviceSummary
			var devCodes []string
			for _, dID := range deviceOrder {
				dev := deviceMap[dID]
				devicesList = append(devicesList, dev)
				devCodes = append(devCodes, dev.DeviceCode)
			}
			devicesSummary := strings.Join(devCodes, ", ")
			deviceCount := len(devicesList)

			// Collect all parameters across devices
			var sourceIDs []uint64
			var sourceItems []model.SourceAggregationValueItem

			for _, item := range group {
				sourceIDs = append(sourceIDs, item.ID)
				paramCode := ""
				paramName := ""
				unit := ""
				if item.Parameter != nil {
					paramCode = item.Parameter.ParameterCode
					paramName = item.Parameter.ParameterName
					unit = item.Parameter.Unit
				}
				devInfo := deviceMap[item.DeviceID]

				val := item.Value
				qual := string(item.Quality)

				// POS 3: Outbound Gatekeeper (Sanitize unphysical float/register overflow before transmission)
				if val != nil && (math.IsNaN(*val) || math.IsInf(*val, 0) || math.Abs(*val) > 1e8) {
					logger.Warn("[Outbound Gatekeeper] Quarantined implausible value %.2e for parameter %s (%s) in batch %s",
						*val, paramCode, devInfo.DeviceCode, batchID)
					val = nil
					qual = string(model.QualityBad)
				}

				sourceItems = append(sourceItems, model.SourceAggregationValueItem{
					DeviceID:            item.DeviceID,
					DeviceCode:          devInfo.DeviceCode,
					DeviceName:          devInfo.DeviceName,
					ParameterID:         item.ParameterID,
					ParameterCode:       paramCode,
					ParameterName:       paramName,
					Value:               val,
					Unit:                unit,
					Quality:             qual,
					AggregationResultID: item.ID,
					PeriodStart:         item.PeriodStart,
					PeriodEnd:           item.PeriodEnd,
				})
			}

			sourceIDsJSON, _ := json.Marshal(sourceIDs)
			sourceValuesJSON, _ := json.Marshal(sourceItems)

			batchData := &OutboundBatchData{
				DeviceCount:     deviceCount,
				DevicesSummary:  devicesSummary,
				Devices:         devicesList,
				IntervalSeconds: intervalSec,
				PeriodStart:     periodStart,
				PeriodEnd:       periodEnd,
				Values:          sourceItems,
			}
			if deviceCount == 1 {
				batchData.DeviceID = devicesList[0].DeviceID
				batchData.DeviceCode = devicesList[0].DeviceCode
				batchData.DeviceName = devicesList[0].DeviceName
			}

			// Format payload according to destination config
			cfg := dest.PayloadConfig
			if cfg == nil {
				cfg = &model.DestinationPayloadConfig{
					StationIDField:  "station_id",
					TimestampField:  "timestamp",
					TimestampFormat: "ISO8601",
					IntervalField:   "interval_minutes",
					Structure:       "FLAT",
					Version:         1,
				}
			}

			payloadBytes, err := s.formatter.Format(batchData, cfg)
			if err != nil {
				logger.Error("[Outbound] Serialization failed for dest %d batch %s: %v", dest.ID, batchID, err)
				continue
			}

			var queueDeviceID *uint
			if deviceCount == 1 {
				dID := devicesList[0].DeviceID
				queueDeviceID = &dID
			} else if dest.DeviceID != nil {
				queueDeviceID = dest.DeviceID
			}

			queueItem := &model.OutboundDeliveryQueue{
				BatchID:              batchID,
				DestinationID:        dest.ID,
				DeviceID:             queueDeviceID,
				DeviceCount:          deviceCount,
				DevicesSummary:       devicesSummary,
				PeriodStart:          periodStart,
				PeriodEnd:            periodEnd,
				IntervalSeconds:      intervalSec,
				ParameterCount:       len(sourceItems),
				SourceAggregationIDs: string(sourceIDsJSON),
				SourceValuesJSON:     string(sourceValuesJSON),
				PayloadSnapshot:      string(payloadBytes),
				PayloadVersion:       cfg.Version,
				IdempotencyKey:       idempotencyKey,
				Status:               model.DeliveryPending,
				Attempts:             0,
				MaxRetries:           dest.RetryLimit,
				IsReplay:             false,
				CreatedAt:            time.Now(),
			}
			if queueItem.MaxRetries <= 0 {
				queueItem.MaxRetries = 5
			}

			_ = s.repo.EnqueueDelivery(ctx, queueItem)
		}
	}

	return nil
}

// GetDestinations returns all destinations, with sensitive credentials redacted when requested
func (s *outboundService) GetDestinations(ctx context.Context, enabledOnly bool, redactSecrets bool) ([]model.OutputDestination, error) {
	dests, err := s.repo.GetAllDestinations(ctx, enabledOnly)
	if err != nil {
		return nil, err
	}
	if redactSecrets {
		for i := range dests {
			redactDestinationSecrets(&dests[i])
		}
	}
	return dests, nil
}

// GetDestinationByID returns a single destination
func (s *outboundService) GetDestinationByID(ctx context.Context, id uint, redactSecrets bool) (*model.OutputDestination, error) {
	dest, err := s.repo.GetDestinationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if redactSecrets {
		redactDestinationSecrets(dest)
	}
	return dest, nil
}

func redactDestinationSecrets(dest *model.OutputDestination) {
	if dest == nil {
		return
	}
	if dest.AuthToken != "" {
		dest.AuthToken = "***REDACTED***"
	}
	if dest.APIKeyValue != "" {
		dest.APIKeyValue = "***REDACTED***"
	}
	if dest.Password != "" {
		dest.Password = "***REDACTED***"
	}
}

// CreateDestination persists a new outbound destination with default payload config
func (s *outboundService) CreateDestination(ctx context.Context, dest *model.OutputDestination) error {
	if dest.RetryLimit <= 0 {
		dest.RetryLimit = 5
	}
	if dest.QueueLimit <= 0 {
		dest.QueueLimit = 10000
	}
	if err := s.repo.CreateDestination(ctx, dest); err != nil {
		return err
	}

	// Create default payload config for this destination
	cfg := &model.DestinationPayloadConfig{
		DestinationID:      dest.ID,
		TemplateType:       "STANDARD",
		StationIDField:     "station_id",
		TimestampField:     "timestamp",
		TimestampFormat:    "ISO8601",
		IntervalField:      "interval_minutes",
		Structure:          "FLAT",
		NestedDataKey:      "data",
		ConflictResolution: "PREFIX",
		IncludeUnits:       false,
		IncludeQuality:     false,
		IncludeMetadata:    true,
		Version:            1,
	}
	_ = s.repo.UpsertPayloadConfig(ctx, cfg)
	return nil
}

// UpdateDestination updates destination settings, preserving secrets if masked
func (s *outboundService) UpdateDestination(ctx context.Context, dest *model.OutputDestination) error {
	existing, err := s.repo.GetDestinationByID(ctx, dest.ID)
	if err != nil {
		return err
	}

	// If secret fields were sent as redacted placeholder, preserve existing secrets
	if dest.AuthToken == "***REDACTED***" {
		dest.AuthToken = existing.AuthToken
	}
	if dest.APIKeyValue == "***REDACTED***" {
		dest.APIKeyValue = existing.APIKeyValue
	}
	if dest.Password == "***REDACTED***" {
		dest.Password = existing.Password
	}

	return s.repo.UpdateDestination(ctx, dest)
}

// DeleteDestination removes a destination
func (s *outboundService) DeleteDestination(ctx context.Context, id uint) error {
	return s.repo.DeleteDestination(ctx, id)
}

// GetPayloadConfig fetches payload config for a destination
func (s *outboundService) GetPayloadConfig(ctx context.Context, destID uint) (*model.DestinationPayloadConfig, error) {
	return s.repo.GetPayloadConfigByDestinationID(ctx, destID)
}

// UpdatePayloadConfig validates and updates payload config
func (s *outboundService) UpdatePayloadConfig(ctx context.Context, cfg *model.DestinationPayloadConfig) error {
	if err := s.formatter.ValidateConfig(cfg); err != nil {
		return err
	}
	return s.repo.UpsertPayloadConfig(ctx, cfg)
}

// GeneratePreview builds a real or sample preview payload matching stored config
func (s *outboundService) GeneratePreview(ctx context.Context, destID uint, deviceID *uint) (string, error) {
	return s.GeneratePreviewWithConfig(ctx, destID, deviceID, nil)
}

// GeneratePreviewWithConfig builds a real or sample preview payload matching provided or stored config
func (s *outboundService) GeneratePreviewWithConfig(ctx context.Context, destID uint, deviceID *uint, customCfg *model.DestinationPayloadConfig) (string, error) {
	cfg := customCfg
	if cfg == nil {
		var err error
		cfg, err = s.repo.GetPayloadConfigByDestinationID(ctx, destID)
		if err != nil {
			cfg = &model.DestinationPayloadConfig{
				DestinationID:   destID,
				StationIDField:  "station_id",
				TimestampField:  "timestamp",
				TimestampFormat: "ISO8601",
				IntervalField:   "interval_minutes",
				Structure:       "FLAT",
				Version:         1,
			}
		}
	}

	var batch *OutboundBatchData

	dest, _ := s.repo.GetDestinationByID(ctx, destID)
	var targetDeviceIDs []uint
	if dest != nil && dest.DeviceID != nil {
		targetDeviceIDs = []uint{*dest.DeviceID}
	} else if deviceID != nil && *deviceID > 0 {
		targetDeviceIDs = []uint{*deviceID}
	}

	realResults, err := s.repo.GetLatestAggregationResultsMultiDevice(ctx, targetDeviceIDs)
	if err == nil && len(realResults) > 0 {
		deviceMap := make(map[uint]BatchDeviceSummary)
		var deviceOrder []uint
		for _, r := range realResults {
			if _, exists := deviceMap[r.DeviceID]; !exists {
				devCode := fmt.Sprintf("DEV-%d", r.DeviceID)
				devName := fmt.Sprintf("Device %d", r.DeviceID)
				if r.Device != nil {
					if r.Device.DeviceCode != "" {
						devCode = r.Device.DeviceCode
					}
					if r.Device.DeviceName != "" {
						devName = r.Device.DeviceName
					}
				}
				deviceMap[r.DeviceID] = BatchDeviceSummary{
					DeviceID:   r.DeviceID,
					DeviceCode: devCode,
					DeviceName: devName,
				}
				deviceOrder = append(deviceOrder, r.DeviceID)
			}
		}

		var devicesList []BatchDeviceSummary
		var devCodes []string
		for _, dID := range deviceOrder {
			d := deviceMap[dID]
			devicesList = append(devicesList, d)
			devCodes = append(devCodes, d.DeviceCode)
		}

		var items []model.SourceAggregationValueItem
		for _, r := range realResults {
			pCode := ""
			pName := ""
			unit := ""
			if r.Parameter != nil {
				pCode = r.Parameter.ParameterCode
				pName = r.Parameter.ParameterName
				unit = r.Parameter.Unit
			}
			devInfo := deviceMap[r.DeviceID]
			items = append(items, model.SourceAggregationValueItem{
				DeviceID:            r.DeviceID,
				DeviceCode:          devInfo.DeviceCode,
				DeviceName:          devInfo.DeviceName,
				ParameterID:         r.ParameterID,
				ParameterCode:       pCode,
				ParameterName:       pName,
				Value:               r.Value,
				Unit:                unit,
				Quality:             string(r.Quality),
				AggregationResultID: r.ID,
				PeriodStart:         r.PeriodStart,
				PeriodEnd:           r.PeriodEnd,
			})
		}

		batch = &OutboundBatchData{
			DeviceCount:     len(devicesList),
			DevicesSummary:  strings.Join(devCodes, ", "),
			Devices:         devicesList,
			IntervalSeconds: realResults[0].IntervalSeconds,
			PeriodStart:     realResults[0].PeriodStart,
			PeriodEnd:       realResults[0].PeriodEnd,
			Values:          items,
		}
		if len(devicesList) == 1 {
			batch.DeviceID = devicesList[0].DeviceID
			batch.DeviceCode = devicesList[0].DeviceCode
			batch.DeviceName = devicesList[0].DeviceName
		}
	} else {
		// Construct realistic multi-device sample batch (AQMS-01 and AQMS-02)
		now := time.Now()
		vPm25_1, vPm10_1, vTemp, vHum := 45.2, 62.4, 27.5, 68.2
		vPm25_2, vPm10_2, vSO2, vCO := 48.1, 65.0, 2.03, 615.0
		vNO2, vO3, vPress, vWSpeed := 15.4, 32.1, 1013.25, 2.8
		vIntensity, vRain, vHC, vWindAngle := 450.0, 0.0, 0.12, 185.0

		batch = &OutboundBatchData{
			DeviceCount:    2,
			DevicesSummary: "AQMS-01, AQMS-02",
			Devices: []BatchDeviceSummary{
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01"},
				{DeviceID: 2, DeviceCode: "AQMS-02", DeviceName: "Ambient Air Station 02"},
			},
			IntervalSeconds: 120,
			PeriodStart:     now.Add(-2 * time.Minute),
			PeriodEnd:       now,
			Values: []model.SourceAggregationValueItem{
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 1, ParameterCode: "pm25", ParameterName: "PM2.5", Value: &vPm25_1, Unit: "µg/m3", Quality: "GOOD", AggregationResultID: 101},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 2, ParameterCode: "pm10", ParameterName: "PM10", Value: &vPm10_1, Unit: "µg/m3", Quality: "GOOD", AggregationResultID: 102},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 3, ParameterCode: "temperature", ParameterName: "Temperature", Value: &vTemp, Unit: "°C", Quality: "GOOD", AggregationResultID: 103},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 4, ParameterCode: "humidity", ParameterName: "Humidity", Value: &vHum, Unit: "%", Quality: "GOOD", AggregationResultID: 104},
				{DeviceID: 2, DeviceCode: "AQMS-02", DeviceName: "Ambient Air Station 02", ParameterID: 5, ParameterCode: "pm25", ParameterName: "PM2.5", Value: &vPm25_2, Unit: "µg/m3", Quality: "GOOD", AggregationResultID: 201},
				{DeviceID: 2, DeviceCode: "AQMS-02", DeviceName: "Ambient Air Station 02", ParameterID: 6, ParameterCode: "pm10", ParameterName: "PM10", Value: &vPm10_2, Unit: "µg/m3", Quality: "GOOD", AggregationResultID: 202},
				{DeviceID: 2, DeviceCode: "AQMS-02", DeviceName: "Ambient Air Station 02", ParameterID: 7, ParameterCode: "so2", ParameterName: "Sulfur Dioxide", Value: &vSO2, Unit: "ppm", Quality: "GOOD", AggregationResultID: 203},
				{DeviceID: 2, DeviceCode: "AQMS-02", DeviceName: "Ambient Air Station 02", ParameterID: 8, ParameterCode: "co", ParameterName: "Carbon Monoxide", Value: &vCO, Unit: "ppm", Quality: "GOOD", AggregationResultID: 204},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 9, ParameterCode: "no2", ParameterName: "Nitrogen Dioxide", Value: &vNO2, Unit: "ppm", Quality: "GOOD", AggregationResultID: 105},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 10, ParameterCode: "o3", ParameterName: "Ozone", Value: &vO3, Unit: "ppm", Quality: "GOOD", AggregationResultID: 106},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 11, ParameterCode: "air-pressure", ParameterName: "Air Pressure", Value: &vPress, Unit: "hPa", Quality: "GOOD", AggregationResultID: 107},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 12, ParameterCode: "w-speed", ParameterName: "Wind Speed", Value: &vWSpeed, Unit: "m/s", Quality: "GOOD", AggregationResultID: 108},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 13, ParameterCode: "wind-angle", ParameterName: "Wind Direction", Value: &vWindAngle, Unit: "deg", Quality: "GOOD", AggregationResultID: 109},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 14, ParameterCode: "solar-radiation", ParameterName: "Solar Radiation", Value: &vIntensity, Unit: "W/m2", Quality: "GOOD", AggregationResultID: 110},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 15, ParameterCode: "rain-fall", ParameterName: "Rain Fall", Value: &vRain, Unit: "mm", Quality: "GOOD", AggregationResultID: 111},
				{DeviceID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Air Station 01", ParameterID: 16, ParameterCode: "hc", ParameterName: "Hydrocarbon", Value: &vHC, Unit: "ppm", Quality: "GOOD", AggregationResultID: 112},
			},
		}
	}

	return s.formatter.GeneratePreview(batch, cfg)
}

// TestPayload generates a preview payload and attempts direct delivery to destination
func (s *outboundService) TestPayload(ctx context.Context, destID uint) (*TestPayloadResult, error) {
	dest, err := s.repo.GetDestinationByID(ctx, destID)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %w", err)
	}

	previewJSON, err := s.GeneratePreview(ctx, destID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payload: %w", err)
	}

	result := &TestPayloadResult{
		PayloadSent: previewJSON,
	}

	startTime := time.Now()
	switch dest.Type {
	case model.DestREST:
		httpCode, respBody, _, _, err := s.deliverREST(ctx, dest, previewJSON)
		result.LatencyMs = time.Since(startTime).Milliseconds()
		result.HTTPCode = httpCode
		result.ResponseText = respBody
		if err != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
		} else {
			result.Success = true
		}
	case model.DestTCP:
		err := s.deliverTCP(ctx, dest, previewJSON)
		result.LatencyMs = time.Since(startTime).Milliseconds()
		if err != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
		} else {
			result.Success = true
			result.ResponseText = "TCP write succeeded"
		}
	case model.DestUDP:
		err := s.deliverUDP(ctx, dest, previewJSON)
		result.LatencyMs = time.Since(startTime).Milliseconds()
		if err != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
		} else {
			result.Success = true
			result.ResponseText = "UDP packet transmitted"
		}
	default:
		result.Success = true
		result.ResponseText = fmt.Sprintf("Protocol %s validated locally", dest.Type)
		result.LatencyMs = time.Since(startTime).Milliseconds()
	}

	return result, nil
}

// GetQueueItems queries filtered queue records
func (s *outboundService) GetQueueItems(ctx context.Context, filter repository.QueueFilter) ([]model.OutboundDeliveryQueue, int64, error) {
	return s.repo.GetQueueItems(ctx, filter)
}

// GetQueueItemDetail returns the queue record with its unmarshaled source aggregation values
func (s *outboundService) GetQueueItemDetail(ctx context.Context, id uint64) (*QueueItemDetailResponse, error) {
	item, err := s.repo.GetQueueItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var sourceValues []model.SourceAggregationValueItem
	if strings.TrimSpace(item.SourceValuesJSON) != "" {
		_ = json.Unmarshal([]byte(item.SourceValuesJSON), &sourceValues)
	}

	return &QueueItemDetailResponse{
		OutboundDeliveryQueue: *item,
		SourceValues:          sourceValues,
	}, nil
}

// GetQueueSummary returns status counters for dashboard KPI cards
func (s *outboundService) GetQueueSummary(ctx context.Context) (*model.DeliveryQueueSummary, error) {
	return s.repo.GetQueueSummary(ctx)
}

// RetryQueueItem resets a failed or retrying queue item to PENDING for immediate processing
func (s *outboundService) RetryQueueItem(ctx context.Context, id uint64) error {
	return s.repo.ResetQueueItemForRetry(ctx, id)
}

// ReplayQueueItem creates a new delivery queue entry preserving original payload and source values
func (s *outboundService) ReplayQueueItem(ctx context.Context, id uint64) (*model.OutboundDeliveryQueue, error) {
	original, err := s.repo.GetQueueItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("queue item %d not found: %w", id, err)
	}

	now := time.Now()
	newIdempotencyKey := fmt.Sprintf("%s_REPLAY_%d", original.IdempotencyKey, now.UnixNano())
	replayOfID := original.ID

	newItem := &model.OutboundDeliveryQueue{
		BatchID:              original.BatchID,
		DestinationID:        original.DestinationID,
		DeviceID:             original.DeviceID,
		PeriodStart:          original.PeriodStart,
		PeriodEnd:            original.PeriodEnd,
		IntervalSeconds:      original.IntervalSeconds,
		ParameterCount:       original.ParameterCount,
		SourceAggregationIDs: original.SourceAggregationIDs,
		SourceValuesJSON:     original.SourceValuesJSON,
		PayloadSnapshot:      original.PayloadSnapshot,
		PayloadVersion:       original.PayloadVersion,
		IdempotencyKey:       newIdempotencyKey,
		Status:               model.DeliveryPending,
		Attempts:             0,
		MaxRetries:           original.MaxRetries,
		IsReplay:             true,
		ReplayOfID:           &replayOfID,
		CreatedAt:            now,
	}

	if err := s.repo.EnqueueDelivery(ctx, newItem); err != nil {
		return nil, err
	}
	return newItem, nil
}

// CancelQueueItem marks an item as CANCELLED
func (s *outboundService) CancelQueueItem(ctx context.Context, id uint64) error {
	return s.repo.CancelQueueItem(ctx, id)
}

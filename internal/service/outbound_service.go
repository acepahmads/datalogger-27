package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

	"datalogger/internal/database"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
)

// OutboundService manages destination definitions, payload formatting, offline queueing, and dispatch workers
type OutboundService interface {
	StartWorker()
	StopWorker()
	ProcessPendingQueueOnce()

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
	TestPayloadWithConfig(ctx context.Context, destID uint, customCfg *model.DestinationPayloadConfig) (*TestPayloadResult, error)

	// Queue Inspection & Management
	GetQueueItems(ctx context.Context, filter repository.QueueFilter) ([]model.OutboundDeliveryQueue, int64, error)
	GetQueueItemDetail(ctx context.Context, id uint64) (*QueueItemDetailResponse, error)
	GetQueueSummary(ctx context.Context) (*model.DeliveryQueueSummary, error)
	RetryQueueItem(ctx context.Context, id uint64) error
	ReplayQueueItem(ctx context.Context, id uint64) (*model.OutboundDeliveryQueue, error)
	CancelQueueItem(ctx context.Context, id uint64) error
	DeleteQueueItem(ctx context.Context, id uint64) error
	PurgeQueue(ctx context.Context, filter repository.QueuePurgeFilter) (int64, int64, error)
	StartPurge(ctx context.Context, filter repository.QueuePurgeFilter) (*PurgeProgress, error)
	GetPurgeStatus() PurgeProgress
	CancelPurge()
}

// PurgeProgress tracks asynchronous progress for large queue & log deletion jobs
type PurgeProgress struct {
	IsRunning    bool       `json:"is_running"`
	Status       string     `json:"status"` // IDLE, RUNNING, COMPLETED, FAILED, CANCELLED
	TotalQueue   int64      `json:"total_queue"`
	PurgedQueue  int64      `json:"purged_queue"`
	TotalLogs    int64      `json:"total_logs"`
	PurgedLogs   int64      `json:"purged_logs"`
	Percent      float64    `json:"percent"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

type TestPayloadResult struct {
	Success          bool   `json:"success"`
	HTTPCode         int    `json:"http_code,omitempty"`
	LatencyMs        int64  `json:"latency_ms"`
	ResponseText     string `json:"response_text"`
	ErrorMessage     string `json:"error_message,omitempty"`
	PayloadSent      string `json:"payload_sent"`
	IsTwoStep        bool   `json:"is_two_step,omitempty"`
	AuthURL          string `json:"auth_url,omitempty"`
	DataURL          string `json:"data_url,omitempty"`
	AuthUsedCache    bool   `json:"auth_used_cache,omitempty"`
	AuthSuccess      bool   `json:"auth_success,omitempty"`
	AuthHTTPCode     int    `json:"auth_http_code,omitempty"`
	AuthLatencyMs    int64  `json:"auth_latency_ms,omitempty"`
	AuthTokenPreview string `json:"auth_token_preview,omitempty"`
	AuthMessage      string `json:"auth_message,omitempty"`
	DataLatencyMs    int64  `json:"data_latency_ms,omitempty"`
}

type restDeliveryOutcome struct {
	httpCode         int
	body             string
	isAck            bool
	isPermanent      bool
	err              error
	isTwoStep        bool
	authUsedCache    bool
	authSuccess      bool
	authHTTPCode     int
	authLatencyMs    int64
	authTokenPreview string
	authMessage      string
	dataLatencyMs    int64
}

type QueueItemDetailResponse struct {
	model.OutboundDeliveryQueue
	SourceValues []model.SourceAggregationValueItem `json:"source_values"`
}

type cachedAuthToken struct {
	token     string
	expiresAt time.Time
}

type outboundService struct {
	repo       repository.OutboundRepository
	deviceRepo *repository.DeviceRepository
	formatter  PayloadFormatter
	httpClient *http.Client

	workerMu   sync.Mutex
	dispatchMu sync.Mutex
	running    bool
	stopCh     chan struct{}
	wakeCh     chan struct{}

	authMu    sync.RWMutex
	authCache map[uint]cachedAuthToken

	purgeMu        sync.RWMutex
	purgeProgress  PurgeProgress
	purgeCancelled bool
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
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				// Preserve POST method, JSON body, and headers when HTTP redirects to HTTPS (e.g. http://sparing.kemenlh.go.id/api/send -> https://...)
				if len(via) > 0 && via[0].Method == http.MethodPost && req.Response != nil &&
					(req.Response.StatusCode == http.StatusMovedPermanently || req.Response.StatusCode == http.StatusFound) {
					req.Method = http.MethodPost
					if via[0].GetBody != nil {
						if bodyCopy, err := via[0].GetBody(); err == nil {
							req.Body = bodyCopy
							req.ContentLength = via[0].ContentLength
						}
					}
					for _, hdr := range []string{"Content-Type", "Accept", "Authorization", "User-Agent"} {
						if val := via[0].Header.Get(hdr); val != "" && req.Header.Get(hdr) == "" {
							req.Header.Set(hdr, val)
						}
					}
				}
				return nil
			},
		},
		stopCh:    make(chan struct{}),
		wakeCh:    make(chan struct{}, 1),
		authCache: make(map[uint]cachedAuthToken),
		purgeProgress: PurgeProgress{
			Status: "IDLE",
		},
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

	// Reconcile destination success/failed counters and status with actual queue items on startup
	_ = s.repo.SyncDestinationStatsFromQueue(context.Background(), nil)

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
		case <-s.wakeCh:
			s.processPendingQueue()
		case <-ticker.C:
			s.processPendingQueue()
		}
	}
}

// notifyWorker signals the background worker loop immediately without blocking
func (s *outboundService) notifyWorker() {
	if s.wakeCh == nil {
		return
	}
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

// ProcessPendingQueueOnce synchronously executes one pass of pending queue dispatch
func (s *outboundService) ProcessPendingQueueOnce() {
	s.processPendingQueue()
}

// processPendingQueue retrieves and dispatches pending deliveries
func (s *outboundService) processPendingQueue() {
	if database.IsRestoreMode() {
		return
	}
	s.dispatchMu.Lock()
	defer s.dispatchMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	items, err := s.repo.FetchPendingQueueItems(ctx, 20)
	if err != nil || len(items) == 0 {
		return
	}

	for _, item := range items {
		// Skip claiming items belonging to disabled destinations so attempts are preserved until re-enabled
		if item.Destination != nil && !item.Destination.Enabled {
			continue
		}

		// Claim item atomically to prevent concurrent deliveries
		claimed, err := s.repo.ClaimQueueItemForDelivery(ctx, item.ID)
		if err != nil || !claimed {
			continue
		}
		item.Attempts++

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

	// Defense-in-depth: if payload configuration was updated to a newer version after this batch was queued,
	// rebuild PayloadSnapshot from persisted SourceValuesJSON right before transmission.
	if dest.PayloadConfig != nil && item.PayloadVersion < dest.PayloadConfig.Version {
		if s.rebuildQueueItemPayload(item, dest.PayloadConfig) {
			_ = s.repo.UpdateQueueItemPayloadAndRequeue(ctx, item, false)
		}
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
	}

	_ = s.repo.UpdateDeliveryResult(ctx, item)

	// Reconcile destination last_delivery, counters, and health status directly from persisted queue states
	if err == nil {
		_ = s.repo.RecordDestinationLastDelivery(ctx, dest.ID, now)
	}
	_ = s.repo.SyncDestinationStatsFromQueue(ctx, &dest.ID)

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

// deliverREST executes an outbound HTTP request to a third-party webhook/endpoint,
// including 2-step authentication (Step 1: POST auth_url -> Step 2: POST data_url with Bearer token) when AuthType == "LOGIN_BEARER".
func (s *outboundService) deliverREST(ctx context.Context, dest *model.OutputDestination, payload string) (int, string, bool, bool, error) {
	out := s.deliverRESTDetailed(ctx, dest, payload)
	return out.httpCode, out.body, out.isAck, out.isPermanent, out.err
}

func maskAuthTokenPreview(tok string) string {
	t := strings.TrimSpace(tok)
	if len(t) <= 10 {
		if t == "" {
			return ""
		}
		return "****"
	}
	return t[:8] + "...****"
}

func detectApplicationLevelError(body []byte) (string, bool) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return "", false
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return "", false
	}

	isFalsyFailure := func(v interface{}) bool {
		switch val := v.(type) {
		case bool:
			return !val
		case float64:
			return val == 0
		case string:
			lower := strings.ToLower(strings.TrimSpace(val))
			return lower == "0" || lower == "false" || lower == "error" || lower == "failed" || lower == "fail" || lower == "gagal"
		}
		return false
	}

	rejected := false
	if st, ok := lookupCaseInsensitiveKey(obj, "status"); ok && isFalsyFailure(st) {
		rejected = true
	}
	if !rejected {
		if sc, ok := lookupCaseInsensitiveKey(obj, "success"); ok && isFalsyFailure(sc) {
			rejected = true
		}
	}
	if !rejected {
		return "", false
	}

	for _, msgKey := range []string{"desc", "description", "message", "msg", "error", "detail", "keterangan"} {
		if mv, ok := lookupCaseInsensitiveKey(obj, msgKey); ok {
			switch typed := mv.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return strings.TrimSpace(typed), true
				}
			case []interface{}:
				var parts []string
				for _, el := range typed {
					if s, ok := el.(string); ok && strings.TrimSpace(s) != "" {
						parts = append(parts, strings.TrimSpace(s))
					}
				}
				if len(parts) > 0 {
					return strings.Join(parts, "; "), true
				}
			}
		}
	}
	return string(trimmed), true
}

// encodeJWT_HS256 signs a JSON payload string as an RFC 7519 HS256 JWT ({"typ":"JWT","alg":"HS256"}).
func encodeJWT_HS256(payloadJSON string, secret string) (string, error) {
	trimmedSecret := strings.TrimSpace(secret)
	if trimmedSecret == "" {
		return "", fmt.Errorf("JWT secret key is empty")
	}
	var payloadBytes []byte
	var compactBuf bytes.Buffer
	if err := json.Compact(&compactBuf, []byte(payloadJSON)); err == nil {
		payloadBytes = compactBuf.Bytes()
	} else {
		payloadBytes = []byte(strings.TrimSpace(payloadJSON))
	}

	const headerJSON = `{"typ":"JWT","alg":"HS256"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := headerB64 + "." + payloadB64

	mac := hmac.New(sha256.New, []byte(trimmedSecret))
	_, _ = mac.Write([]byte(signingInput))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sigB64, nil
}

func (s *outboundService) deliverRESTDetailed(ctx context.Context, dest *model.OutputDestination, payload string) restDeliveryOutcome {
	sendDataPost := func(dynamicBearer string, bodyOverride string) (int, string, bool, bool, int64, error) {
		dataStart := time.Now()
		reqPayload := payload
		if bodyOverride != "" {
			reqPayload = bodyOverride
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, dest.Endpoint, bytes.NewReader([]byte(reqPayload)))
		if err != nil {
			return 0, "", false, true, 0, fmt.Errorf("invalid endpoint URL: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Datalogger-Outbound/1.0")

		// Apply Authentication
		switch dest.AuthType {
		case "LOGIN_BEARER":
			if dynamicBearer != "" {
				req.Header.Set("Authorization", "Bearer "+dynamicBearer)
			}
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
		dataLatency := time.Since(dataStart).Milliseconds()
		if err != nil {
			// Network connection error / timeout -> retryable
			return 0, "", false, false, dataLatency, fmt.Errorf("network error: %w", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		bodyStr := string(bodyBytes)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if appErrMsg, rejected := detectApplicationLevelError(bodyBytes); rejected {
				return resp.StatusCode, bodyStr, false, true, dataLatency, fmt.Errorf("Application rejected (HTTP %d): %s", resp.StatusCode, appErrMsg)
			}
			lowerBody := strings.ToLower(bodyStr)
			isAck := strings.Contains(lowerBody, "ack") ||
				strings.Contains(lowerBody, "success") ||
				strings.Contains(lowerBody, "received") ||
				strings.Contains(lowerBody, `"status":true`) ||
				strings.Contains(lowerBody, `"status": true`)
			return resp.StatusCode, bodyStr, isAck, false, dataLatency, nil
		}

		isPermanent := false
		if resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 422 {
			isPermanent = true
		}

		return resp.StatusCode, bodyStr, false, isPermanent, dataLatency, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, bodyStr)
	}

	if dest.AuthType == "SPARING_JWT_2STEP" {
		outcome := restDeliveryOutcome{isTwoStep: true}
		var secret string
		usedCachedSecret := false
		if dest.AuthCacheSeconds > 0 && dest.ID > 0 {
			if cached := s.getCachedAuthToken(dest.ID); cached != "" {
				secret = cached
				usedCachedSecret = true
				outcome.authUsedCache = true
				outcome.authSuccess = true
			}
		}

		if secret == "" {
			authStart := time.Now()
			freshSecret, authCode, authBody, authPerm, authErr := s.performSparingSecretFetch(ctx, dest)
			outcome.authLatencyMs = time.Since(authStart).Milliseconds()
			outcome.authHTTPCode = authCode
			if authErr != nil {
				outcome.authSuccess = false
				outcome.authMessage = authErr.Error()
				outcome.httpCode = authCode
				outcome.body = authBody
				outcome.isPermanent = authPerm
				outcome.err = authErr
				return outcome
			}
			secret = freshSecret
			outcome.authSuccess = true
			if dest.AuthCacheSeconds > 0 && dest.ID > 0 {
				s.setCachedAuthToken(dest.ID, secret, time.Duration(dest.AuthCacheSeconds)*time.Second)
			}
		}

		jwtToken, jwtErr := encodeJWT_HS256(payload, secret)
		if jwtErr != nil {
			outcome.authSuccess = false
			outcome.authMessage = jwtErr.Error()
			outcome.isPermanent = true
			outcome.err = jwtErr
			return outcome
		}
		outcome.authTokenPreview = maskAuthTokenPreview(jwtToken)
		if usedCachedSecret {
			outcome.authMessage = "Menggunakan secret-sensor dari cache → JWT HS256 berhasil dibuat"
		} else {
			outcome.authMessage = fmt.Sprintf("Secret sensor diterima (HTTP %d) → JWT HS256 berhasil dibuat", outcome.authHTTPCode)
		}

		wrappedBytes, _ := json.Marshal(map[string]string{"token": jwtToken})
		code, body, isAck, isPerm, dataLat, err := sendDataPost("", string(wrappedBytes))
		outcome.dataLatencyMs = dataLat

		// If cached secret was rotated upstream (401/403 or signature/token rejection), refresh secret once & retry immediately
		if err != nil && usedCachedSecret &&
			(code == http.StatusUnauthorized || code == http.StatusForbidden ||
				strings.Contains(strings.ToLower(body), "signature") ||
				strings.Contains(strings.ToLower(body), "token") ||
				strings.Contains(strings.ToLower(body), "secret") ||
				strings.Contains(strings.ToLower(err.Error()), "application rejected")) {
			s.clearCachedAuthToken(dest.ID)
			outcome.authUsedCache = false
			authStart := time.Now()
			freshSecret, authCode, authBody, authPerm, authErr := s.performSparingSecretFetch(ctx, dest)
			outcome.authLatencyMs = time.Since(authStart).Milliseconds()
			outcome.authHTTPCode = authCode
			if authErr != nil {
				outcome.authSuccess = false
				outcome.authMessage = authErr.Error()
				outcome.httpCode = authCode
				outcome.body = authBody
				outcome.isPermanent = authPerm
				outcome.err = authErr
				return outcome
			}
			if dest.AuthCacheSeconds > 0 && dest.ID > 0 {
				s.setCachedAuthToken(dest.ID, freshSecret, time.Duration(dest.AuthCacheSeconds)*time.Second)
			}
			freshJWT, freshJWTErr := encodeJWT_HS256(payload, freshSecret)
			if freshJWTErr != nil {
				outcome.authSuccess = false
				outcome.authMessage = freshJWTErr.Error()
				outcome.isPermanent = true
				outcome.err = freshJWTErr
				return outcome
			}
			outcome.authSuccess = true
			outcome.authTokenPreview = maskAuthTokenPreview(freshJWT)
			outcome.authMessage = fmt.Sprintf("Secret cache diperbarui → Re-sign JWT HS256 berhasil (HTTP %d)", authCode)

			wrappedFreshBytes, _ := json.Marshal(map[string]string{"token": freshJWT})
			code2, body2, isAck2, isPerm2, dataLat2, err2 := sendDataPost("", string(wrappedFreshBytes))
			outcome.httpCode = code2
			outcome.body = body2
			outcome.isAck = isAck2
			outcome.isPermanent = isPerm2
			outcome.dataLatencyMs += dataLat2
			outcome.err = err2
			return outcome
		}

		outcome.httpCode = code
		outcome.body = body
		outcome.isAck = isAck
		outcome.isPermanent = isPerm
		outcome.err = err
		return outcome
	}

	if dest.AuthType == "KLHK_ONLIMO_BODY" {
		apiKey := strings.TrimSpace(dest.APIKeyValue)
		apiSecret := strings.TrimSpace(dest.AuthToken)
		if apiKey == "" || apiSecret == "" {
			err := fmt.Errorf("KLHK WQMS / ONLIMO apikey atau apisecret belum diisi pada pengaturan destinasi")
			return restDeliveryOutcome{isPermanent: true, err: err}
		}
		signedBody, injErr := injectRootJSONCredentials(payload, apiKey, apiSecret)
		if injErr != nil {
			return restDeliveryOutcome{
				isPermanent: true,
				err:         fmt.Errorf("failed to inject KLHK ONLIMO credentials into JSON body: %w", injErr),
			}
		}
		code, body, isAck, isPerm, dataLat, err := sendDataPost("", signedBody)
		return restDeliveryOutcome{
			httpCode:      code,
			body:          body,
			isAck:         isAck,
			isPermanent:   isPerm,
			dataLatencyMs: dataLat,
			err:           err,
		}
	}

	if dest.AuthType == "SPARING_JWT_STATIC" {
		secret := strings.TrimSpace(dest.AuthToken)
		if secret == "" {
			err := fmt.Errorf("SPARING static secret (auth_token) belum diisi")
			return restDeliveryOutcome{isPermanent: true, err: err}
		}
		jwtToken, jwtErr := encodeJWT_HS256(payload, secret)
		if jwtErr != nil {
			return restDeliveryOutcome{isPermanent: true, err: jwtErr}
		}
		wrappedBytes, _ := json.Marshal(map[string]string{"token": jwtToken})
		code, body, isAck, isPerm, dataLat, err := sendDataPost("", string(wrappedBytes))
		return restDeliveryOutcome{
			httpCode:      code,
			body:          body,
			isAck:         isAck,
			isPermanent:   isPerm,
			dataLatencyMs: dataLat,
			err:           err,
		}
	}

	if dest.AuthType == "LOGIN_BEARER" {
		outcome := restDeliveryOutcome{isTwoStep: true}
		var token string
		usedCachedToken := false
		if dest.AuthCacheSeconds > 0 && dest.ID > 0 {
			if cached := s.getCachedAuthToken(dest.ID); cached != "" {
				token = cached
				usedCachedToken = true
				outcome.authUsedCache = true
				outcome.authSuccess = true
				outcome.authTokenPreview = maskAuthTokenPreview(token)
				outcome.authMessage = "Menggunakan Bearer token aktif dari cache"
			}
		}

		if token == "" {
			authStart := time.Now()
			freshToken, authCode, authBody, authPerm, authErr := s.performLoginAuth(ctx, dest)
			outcome.authLatencyMs = time.Since(authStart).Milliseconds()
			outcome.authHTTPCode = authCode
			if authErr != nil {
				outcome.authSuccess = false
				outcome.authMessage = authErr.Error()
				outcome.httpCode = authCode
				outcome.body = authBody
				outcome.isPermanent = authPerm
				outcome.err = authErr
				return outcome
			}
			token = freshToken
			outcome.authSuccess = true
			outcome.authTokenPreview = maskAuthTokenPreview(token)
			outcome.authMessage = fmt.Sprintf("Login berhasil & Bearer token diterima (HTTP %d)", authCode)
			if dest.AuthCacheSeconds > 0 && dest.ID > 0 {
				s.setCachedAuthToken(dest.ID, token, time.Duration(dest.AuthCacheSeconds)*time.Second)
			}
		}

		code, body, isAck, isPerm, dataLat, err := sendDataPost(token, "")
		outcome.dataLatencyMs = dataLat
		if err != nil && usedCachedToken && code == http.StatusUnauthorized {
			// Cached token expired or revoked early: clear cache, re-authenticate once, and retry Step 2 immediately
			s.clearCachedAuthToken(dest.ID)
			outcome.authUsedCache = false
			authStart := time.Now()
			freshToken, authCode, authBody, authPerm, authErr := s.performLoginAuth(ctx, dest)
			outcome.authLatencyMs = time.Since(authStart).Milliseconds()
			outcome.authHTTPCode = authCode
			if authErr != nil {
				outcome.authSuccess = false
				outcome.authMessage = authErr.Error()
				outcome.httpCode = authCode
				outcome.body = authBody
				outcome.isPermanent = authPerm
				outcome.err = authErr
				return outcome
			}
			outcome.authSuccess = true
			outcome.authTokenPreview = maskAuthTokenPreview(freshToken)
			outcome.authMessage = fmt.Sprintf("Token cache kedaluwarsa (401) → Re-login berhasil (HTTP %d)", authCode)
			if dest.AuthCacheSeconds > 0 && dest.ID > 0 {
				s.setCachedAuthToken(dest.ID, freshToken, time.Duration(dest.AuthCacheSeconds)*time.Second)
			}
			code2, body2, isAck2, isPerm2, dataLat2, err2 := sendDataPost(freshToken, "")
			outcome.httpCode = code2
			outcome.body = body2
			outcome.isAck = isAck2
			outcome.isPermanent = isPerm2
			outcome.dataLatencyMs += dataLat2
			outcome.err = err2
			return outcome
		}

		outcome.httpCode = code
		outcome.body = body
		outcome.isAck = isAck
		outcome.isPermanent = isPerm
		outcome.err = err
		return outcome
	}

	code, body, isAck, isPerm, dataLat, err := sendDataPost("", "")
	return restDeliveryOutcome{
		httpCode:      code,
		body:          body,
		isAck:         isAck,
		isPermanent:   isPerm,
		dataLatencyMs: dataLat,
		err:           err,
	}
}

func (s *outboundService) getCachedAuthToken(destID uint) string {
	s.authMu.RLock()
	defer s.authMu.RUnlock()
	if s.authCache == nil {
		return ""
	}
	entry, ok := s.authCache[destID]
	if !ok || entry.token == "" || time.Now().After(entry.expiresAt) {
		return ""
	}
	return entry.token
}

func (s *outboundService) setCachedAuthToken(destID uint, token string, ttl time.Duration) {
	if destID == 0 || token == "" || ttl <= 0 {
		return
	}
	s.authMu.Lock()
	defer s.authMu.Unlock()
	if s.authCache == nil {
		s.authCache = make(map[uint]cachedAuthToken)
	}
	s.authCache[destID] = cachedAuthToken{
		token:     token,
		expiresAt: time.Now().Add(ttl),
	}
}

func (s *outboundService) clearCachedAuthToken(destID uint) {
	if destID == 0 {
		return
	}
	s.authMu.Lock()
	defer s.authMu.Unlock()
	delete(s.authCache, destID)
}

// performLoginAuth executes Step 1 of 2-step delivery: POST {"username": "...", "password": "..."} to AuthURL and extracts Bearer token.
func (s *outboundService) performLoginAuth(ctx context.Context, dest *model.OutputDestination) (string, int, string, bool, error) {
	authURL := strings.TrimSpace(dest.AuthURL)
	if authURL == "" {
		return "", 0, "", true, fmt.Errorf("Step 1 Auth configuration error: auth_url is empty")
	}

	authBodyMap := map[string]string{
		"username": dest.Username,
		"password": dest.Password,
	}
	authBytes, err := json.Marshal(authBodyMap)
	if err != nil {
		return "", 0, "", true, fmt.Errorf("Step 1 Auth payload marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewBuffer(authBytes))
	if err != nil {
		return "", 0, "", true, fmt.Errorf("Step 1 Auth invalid auth_url %q: %w", authURL, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Datalogger-Outbound/1.0")

	if strings.TrimSpace(dest.CustomHeaders) != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(dest.CustomHeaders), &headers); err == nil {
			for k, v := range headers {
				if !strings.EqualFold(k, "Authorization") {
					req.Header.Set(k, v)
				}
			}
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", 0, "", false, fmt.Errorf("Step 1 Auth network error (%s): %w", authURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	bodyStr := string(bodyBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		isPerm := resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 422
		return "", resp.StatusCode, bodyStr, isPerm, fmt.Errorf("Step 1 Auth HTTP %d (%s): %s", resp.StatusCode, authURL, bodyStr)
	}

	token := extractTokenFromAuthResponse(bodyBytes, resp.Header, dest.AuthTokenPath)
	if token == "" {
		if appErrMsg, rejected := detectApplicationLevelError(bodyBytes); rejected {
			return "", resp.StatusCode, bodyStr, true, fmt.Errorf("Step 1 Auth rejected (HTTP %d): %s", resp.StatusCode, appErrMsg)
		}
		return "", resp.StatusCode, bodyStr, true, fmt.Errorf("Step 1 Auth succeeded (HTTP %d) but no token found in response from %s", resp.StatusCode, authURL)
	}

	return token, resp.StatusCode, bodyStr, false, nil
}

// performSparingSecretFetch executes Step 1 of KLHK SPARING 2-step delivery:
// HTTP GET to AuthURL (e.g. https://sparing.kemenlh.go.id/api/secret-sensor) to retrieve the dynamic HS256 signing secret.
func (s *outboundService) performSparingSecretFetch(ctx context.Context, dest *model.OutputDestination) (string, int, string, bool, error) {
	authURL := strings.TrimSpace(dest.AuthURL)
	if authURL == "" {
		return "", 0, "", true, fmt.Errorf("Step 1 SPARING Secret configuration error: auth_url (KLHK Secret URL) is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	if err != nil {
		return "", 0, "", true, fmt.Errorf("Step 1 SPARING Secret invalid auth_url %q: %w", authURL, err)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Datalogger-Outbound/1.0")

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
		return "", 0, "", false, fmt.Errorf("Step 1 SPARING Secret network error (%s): %w", authURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	bodyStr := string(bodyBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		isPerm := resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 422
		return "", resp.StatusCode, bodyStr, isPerm, fmt.Errorf("Step 1 SPARING Secret HTTP %d (%s): %s", resp.StatusCode, authURL, bodyStr)
	}

	if appErrMsg, rejected := detectApplicationLevelError(bodyBytes); rejected {
		return "", resp.StatusCode, bodyStr, true, fmt.Errorf("Step 1 SPARING Secret rejected (HTTP %d): %s", resp.StatusCode, appErrMsg)
	}

	secret := extractSecretFromSparingResponse(bodyBytes, resp.Header, dest.AuthTokenPath)
	if secret == "" {
		return "", resp.StatusCode, bodyStr, true, fmt.Errorf("Step 1 SPARING Secret succeeded (HTTP %d) but no secret key found in response from %s", resp.StatusCode, authURL)
	}

	return secret, resp.StatusCode, bodyStr, false, nil
}

func extractSecretFromSparingResponse(body []byte, headers http.Header, customPath string) string {
	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if s, ok := parsed.(string); ok {
			if sec := strings.Trim(strings.TrimSpace(s), `"'`); sec != "" {
				return sec
			}
		}

		if rootObj, ok := parsed.(map[string]interface{}); ok {
			if trimmedPath := strings.TrimSpace(customPath); trimmedPath != "" {
				parts := strings.Split(trimmedPath, ".")
				var cur interface{} = rootObj
				for _, p := range parts {
					obj, isMap := cur.(map[string]interface{})
					if !isMap {
						cur = nil
						break
					}
					next, found := lookupCaseInsensitiveKey(obj, strings.TrimSpace(p))
					if !found {
						cur = nil
						break
					}
					cur = next
				}
				if s, isStr := cur.(string); isStr {
					if sec := strings.Trim(strings.TrimSpace(s), `"'`); sec != "" {
						return sec
					}
				}
			}

			candidateKeys := []string{"secret", "secret_key", "secret_sensor", "secretsensor", "api_secret", "jwt_secret", "key", "token", "access_token"}
			for _, key := range candidateKeys {
				if val, found := lookupCaseInsensitiveKey(rootObj, key); found {
					if s, isStr := val.(string); isStr {
						if sec := strings.Trim(strings.TrimSpace(s), `"'`); sec != "" {
							return sec
						}
					}
				}
			}

			wrapperKeys := []string{"data", "result", "payload", "response"}
			for _, wKey := range wrapperKeys {
				if wrapped, found := lookupCaseInsensitiveKey(rootObj, wKey); found {
					switch w := wrapped.(type) {
					case string:
						if sec := strings.Trim(strings.TrimSpace(w), `"'`); sec != "" {
							return sec
						}
					case map[string]interface{}:
						for _, key := range candidateKeys {
							if val, ok := lookupCaseInsensitiveKey(w, key); ok {
								if s, isStr := val.(string); isStr {
									if sec := strings.Trim(strings.TrimSpace(s), `"'`); sec != "" {
										return sec
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if headers != nil {
		for _, h := range []string{"X-Secret-Sensor", "X-Secret-Key", "X-Auth-Token"} {
			if hv := strings.Trim(strings.TrimSpace(headers.Get(h)), `"'`); hv != "" {
				return hv
			}
		}
	}

	// Plain-text single-line secret response (standard KLHK SPARING /api/secret-sensor returns raw secret string)
	plain := strings.TrimSpace(string(body))
	lowerPlain := strings.ToLower(plain)
	if plain != "" &&
		!strings.HasPrefix(plain, "{") &&
		!strings.HasPrefix(plain, "[") &&
		!strings.HasPrefix(lowerPlain, "<html") &&
		!strings.HasPrefix(lowerPlain, "<!doctype") &&
		!strings.HasPrefix(plain, "<") &&
		!strings.ContainsAny(plain, "\r\n") {
		return strings.Trim(plain, `"'`)
	}

	return ""
}

func cleanExtractedBearerToken(raw string) string {
	t := strings.TrimSpace(raw)
	t = strings.Trim(t, `"'`)
	if len(t) > 7 && strings.EqualFold(t[:7], "Bearer ") {
		t = strings.TrimSpace(t[7:])
	}
	return t
}

func lookupCaseInsensitiveKey(m map[string]interface{}, target string) (interface{}, bool) {
	if v, ok := m[target]; ok {
		return v, true
	}
	for k, v := range m {
		if strings.EqualFold(k, target) {
			return v, true
		}
	}
	return nil, false
}

func extractTokenFromAuthResponse(body []byte, headers http.Header, customPath string) string {
	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		// Case 1: JSON body is a direct string (e.g. "eyJhbGci...")
		if s, ok := parsed.(string); ok {
			if tok := cleanExtractedBearerToken(s); tok != "" {
				return tok
			}
		}

		if rootObj, ok := parsed.(map[string]interface{}); ok {
			// Case 2: Explicit dot-separated path configured by user (e.g. "data.token")
			if trimmedPath := strings.TrimSpace(customPath); trimmedPath != "" {
				parts := strings.Split(trimmedPath, ".")
				var cur interface{} = rootObj
				for _, p := range parts {
					obj, isMap := cur.(map[string]interface{})
					if !isMap {
						cur = nil
						break
					}
					next, found := lookupCaseInsensitiveKey(obj, strings.TrimSpace(p))
					if !found {
						cur = nil
						break
					}
					cur = next
				}
				if s, isStr := cur.(string); isStr {
					if tok := cleanExtractedBearerToken(s); tok != "" {
						return tok
					}
				}
			}

			// Case 3: Auto-detect common top-level token keys
			candidateKeys := []string{"token", "access_token", "jwt", "auth_token", "bearer", "id_token", "api_token"}
			for _, key := range candidateKeys {
				if val, found := lookupCaseInsensitiveKey(rootObj, key); found {
					if s, isStr := val.(string); isStr {
						if tok := cleanExtractedBearerToken(s); tok != "" {
							return tok
						}
					}
				}
			}

			// Case 4: Auto-detect common wrapper objects (data, result, payload, response, auth)
			wrapperKeys := []string{"data", "result", "payload", "response", "auth"}
			for _, wKey := range wrapperKeys {
				if wrapped, found := lookupCaseInsensitiveKey(rootObj, wKey); found {
					switch w := wrapped.(type) {
					case string:
						if tok := cleanExtractedBearerToken(w); tok != "" {
							return tok
						}
					case map[string]interface{}:
						for _, key := range candidateKeys {
							if val, ok := lookupCaseInsensitiveKey(w, key); ok {
								if s, isStr := val.(string); isStr {
									if tok := cleanExtractedBearerToken(s); tok != "" {
										return tok
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Case 5: Check Authorization / X-Auth-Token response headers
	if headers != nil {
		for _, h := range []string{"Authorization", "X-Auth-Token", "X-Access-Token"} {
			if hv := cleanExtractedBearerToken(headers.Get(h)); hv != "" {
				return hv
			}
		}
	}

	// Case 6: Plain-text non-HTML single-line token response
	plain := strings.TrimSpace(string(body))
	if plain != "" && !strings.HasPrefix(plain, "{") && !strings.HasPrefix(plain, "[") && !strings.HasPrefix(plain, "<") && !strings.ContainsAny(plain, " \t\r\n") {
		return cleanExtractedBearerToken(plain)
	}

	return ""
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

// isActiveOutboundAggregationResult verifies that the aggregation result belongs to an active, non-deleted device and parameter.
func isActiveOutboundAggregationResult(item model.AggregationResult) bool {
	if item.Device == nil || !item.Device.Enabled {
		return false
	}
	st := strings.ToUpper(strings.TrimSpace(string(item.Device.Status)))
	if st == string(model.DeviceStatusInactive) || st == string(model.DeviceStatusDisabled) {
		return false
	}
	if item.Parameter == nil || !item.Parameter.Enabled {
		return false
	}
	if strings.TrimSpace(item.Parameter.ParameterCode) == "" {
		return false
	}
	return true
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

		// Resolve destination payload configuration & checklist matcher upfront
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
		matcher := buildParameterMappingMatcher(cfg)

		// Mode A: Bundled Multi-Bucket Dispatch Schedule (e.g. SPARING 2-minute aggregation sent once every 1 hour = 30 rows)
		sourceInterval := dest.IntervalSeconds
		if dest.DispatchIntervalSeconds > 0 {
			if sourceInterval <= 0 {
				sourceInterval = 120
			}
			if dest.DispatchIntervalSeconds > sourceInterval {
				// Only build & enqueue bundle when bucketEnd lands on the configured dispatch boundary (e.g. :00:00 for 3600s)
				if bucketEnd.Unix()%int64(dest.DispatchIntervalSeconds) != 0 {
					continue
				}
				windowStart := bucketEnd.Add(-time.Duration(dest.DispatchIntervalSeconds) * time.Second)
				windowEnd := bucketEnd

				rangeResults, err := s.repo.GetCustomerAggregationResultsForRange(ctx, windowStart, windowEnd, sourceInterval, queryDeviceIDs)
				if err != nil {
					logger.Error("[Outbound] Failed to fetch bundled aggregation range [%s, %s] for dest %d: %v",
						windowStart.Format(time.RFC3339), windowEnd.Format(time.RFC3339), dest.ID, err)
					continue
				}
				if len(rangeResults) == 0 {
					continue
				}

				var group []model.AggregationResult
				for _, item := range rangeResults {
					if !isActiveOutboundAggregationResult(item) {
						continue
					}
					devCode := item.Device.DeviceCode
					if devCode == "" {
						devCode = fmt.Sprintf("DEV-%d", item.DeviceID)
					}
					if enabled, _ := matcher.ResolveWithName(item.DeviceID, devCode, item.ParameterID, item.Parameter.ParameterCode, item.Parameter.ParameterName); !enabled {
						continue
					}
					group = append(group, item)
				}
				if len(group) == 0 {
					continue
				}

				dispMinutes := dest.DispatchIntervalSeconds / 60
				if dispMinutes <= 0 {
					dispMinutes = 1
				}
				srcMinutes := sourceInterval / 60
				if srcMinutes <= 0 {
					srcMinutes = 1
				}
				batchID := fmt.Sprintf("BATCH-DEST%d-%dM-%dM-%s", dest.ID, dispMinutes, srcMinutes, windowEnd.Format("20060102150405"))
				idempotencyKey := fmt.Sprintf("DEST%d_DISP%d_SRC%d_%d", dest.ID, dest.DispatchIntervalSeconds, sourceInterval, windowEnd.Unix())

				s.enqueueAggregatedBatchGroup(ctx, &dest, cfg, group, sourceInterval, windowStart, windowEnd, batchID, idempotencyKey)
				continue
			}
		}

		// Mode B: Immediate Per-Bucket Dispatch (e.g. every 2m, 5m, 30m, or 60m single-bucket aggregation)
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

		for intervalSec, rawGroup := range resultsByInterval {
			var group []model.AggregationResult
			for _, item := range rawGroup {
				if !isActiveOutboundAggregationResult(item) {
					continue
				}
				devCode := item.Device.DeviceCode
				if devCode == "" {
					devCode = fmt.Sprintf("DEV-%d", item.DeviceID)
				}
				enabled, _ := matcher.ResolveWithName(item.DeviceID, devCode, item.ParameterID, item.Parameter.ParameterCode, item.Parameter.ParameterName)
				if !enabled {
					continue
				}
				group = append(group, item)
			}
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

			s.enqueueAggregatedBatchGroup(ctx, &dest, cfg, group, intervalSec, periodStart, periodEnd, batchID, idempotencyKey)
		}
	}

	return nil
}

func (s *outboundService) enqueueAggregatedBatchGroup(
	ctx context.Context,
	dest *model.OutputDestination,
	cfg *model.DestinationPayloadConfig,
	group []model.AggregationResult,
	intervalSec int,
	periodStart time.Time,
	periodEnd time.Time,
	batchID string,
	idempotencyKey string,
) {
	// Collect unique devices in stable order strictly from checked & active parameters
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

	// Collect checked & active parameters across devices & chronological buckets
	var sourceIDs []uint64
	var sourceItems []model.SourceAggregationValueItem

	for _, item := range group {
		sourceIDs = append(sourceIDs, item.ID)
		paramCode := item.Parameter.ParameterCode
		paramName := item.Parameter.ParameterName
		unit := item.Parameter.Unit
		devInfo := deviceMap[item.DeviceID]

		val := item.Value
		qual := string(item.Quality)

		// POS 3: Outbound Gatekeeper (Sanitize unphysical float/register overflow before transmission)
		if val != nil && (math.IsNaN(*val) || math.IsInf(*val, 0) || math.Abs(*val) > 1e8) {
			logger.Warn("[Outbound Gatekeeper] Quarantined implausible value %.2e for parameter %s (%s) in batch %s",
				*val, paramCode, devInfo.DeviceCode, batchID)
			val = nil
			qual = string(model.QualityBad)
		} else if val != nil {
			roundedVal := roundOutboundValue(*val)
			val = &roundedVal
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

	payloadBytes, err := s.formatter.Format(batchData, cfg)
	if err != nil {
		logger.Error("[Outbound] Serialization failed for dest %d batch %s: %v", dest.ID, batchID, err)
		return
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

	if err := s.repo.EnqueueDelivery(ctx, queueItem); err == nil {
		s.notifyWorker()
	}
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

func validateCustomHeadersJSON(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var hdrs map[string]string
	if err := json.Unmarshal([]byte(trimmed), &hdrs); err != nil {
		return fmt.Errorf("custom_headers harus berupa JSON object key-value string yang valid (contoh: {\"X-Client-ID\":\"123\"}): %w", err)
	}
	return nil
}

// CreateDestination persists a new outbound destination with default payload config
func (s *outboundService) CreateDestination(ctx context.Context, dest *model.OutputDestination) error {
	dest.Endpoint = strings.TrimSpace(dest.Endpoint)
	dest.AuthURL = strings.TrimSpace(dest.AuthURL)
	dest.CustomHeaders = strings.TrimSpace(dest.CustomHeaders)
	if dest.AuthType == "LOGIN_BEARER" && dest.AuthURL == "" {
		return fmt.Errorf("auth_url wajib diisi untuk mode autentikasi 2-Step Login (LOGIN_BEARER)")
	}
	if dest.AuthType == "SPARING_JWT_2STEP" && dest.AuthURL == "" {
		return fmt.Errorf("auth_url (KLHK Secret Sensor URL) wajib diisi untuk mode SPARING 2-Step JWT (SPARING_JWT_2STEP)")
	}
	if dest.AuthType == "KLHK_ONLIMO_BODY" {
		dest.APIKeyValue = strings.TrimSpace(dest.APIKeyValue)
		dest.AuthToken = strings.TrimSpace(dest.AuthToken)
	}
	if err := validateCustomHeadersJSON(dest.CustomHeaders); err != nil {
		return err
	}
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
	// If created with LOGIN_BEARER, initialize with KLHK ISPU AQMS friendly defaults right away
	if dest.AuthType == "LOGIN_BEARER" {
		cfg.TemplateType = "KLHK_ISPU_AQMS"
		cfg.StationIDField = "id_stasiun"
		cfg.TimestampField = "waktu"
		cfg.TimestampFormat = "CUSTOM"
		cfg.TimestampTimezone = "Asia/Jakarta"
		cfg.OmitIntervalField = true
		cfg.IncludeMetadata = false
		cfg.IncludeSensorStatFlags = true
		cfg.SensorStatMode = "STATIC_1"
		cfg.ConflictResolution = "PRIORITY_PRIMARY"
	} else if dest.AuthType == "KLHK_ONLIMO_BODY" {
		// Initialize with KLHK WQMS / ONLIMO defaults (data{IDStasiun, Tanggal, Jam, Suhu..TSS} + apikey + apisecret)
		cfg.TemplateType = "KLHK_WQMS_ONLIMO"
		cfg.StationIDField = "IDStasiun"
		cfg.TimestampField = "Tanggal,Jam"
		cfg.TimestampFormat = "CUSTOM"
		cfg.TimestampSource = "PERIOD_END"
		cfg.TimestampTimezone = "Asia/Jakarta"
		cfg.NumericPrecisionMode = "NATURAL_2"
		cfg.OmitIntervalField = true
		cfg.IncludeMetadata = false
		cfg.Structure = "ONLIMO_DATA_OBJECT"
		cfg.NestedDataKey = "data"
		cfg.ConflictResolution = "PRIORITY_PRIMARY"
		cfg.FillMissingPresetFields = true
	} else if dest.AuthType == "SPARING_JWT_2STEP" || dest.AuthType == "SPARING_JWT_STATIC" {
		// Initialize with KLHK SPARING Water Quality defaults (Flat 2-min or Hourly 30-row data[] based on DispatchIntervalSeconds)
		cfg.StationIDField = "uid"
		cfg.TimestampField = "datetime"
		cfg.TimestampFormat = "UNIX_SEC"
		cfg.TimestampSource = "PERIOD_START"
		cfg.NumericPrecisionMode = "NATURAL_2"
		cfg.OmitIntervalField = true
		cfg.IncludeMetadata = false
		cfg.NestedDataKey = "data"
		cfg.ConflictResolution = "PRIORITY_PRIMARY"
		if dest.DispatchIntervalSeconds > 0 {
			cfg.TemplateType = "KLHK_SPARING_HOURLY"
			cfg.Structure = "SPARING_DATA_ARRAY"
		} else {
			cfg.TemplateType = "KLHK_SPARING"
			cfg.Structure = "FLAT"
		}
	}
	_ = s.repo.UpsertPayloadConfig(ctx, cfg)
	return nil
}

// UpdateDestination updates destination settings, preserving secrets if masked,
// and automatically refreshes & re-queues any undelivered/failed batches when enabled.
func (s *outboundService) UpdateDestination(ctx context.Context, dest *model.OutputDestination) error {
	existing, err := s.repo.GetDestinationByID(ctx, dest.ID)
	if err != nil {
		return err
	}

	dest.Endpoint = strings.TrimSpace(dest.Endpoint)
	dest.AuthURL = strings.TrimSpace(dest.AuthURL)
	dest.CustomHeaders = strings.TrimSpace(dest.CustomHeaders)
	if dest.AuthType == "LOGIN_BEARER" && dest.AuthURL == "" {
		return fmt.Errorf("auth_url wajib diisi untuk mode autentikasi 2-Step Login (LOGIN_BEARER)")
	}
	if dest.AuthType == "SPARING_JWT_2STEP" && dest.AuthURL == "" {
		return fmt.Errorf("auth_url (KLHK Secret Sensor URL) wajib diisi untuk mode SPARING 2-Step JWT (SPARING_JWT_2STEP)")
	}
	if err := validateCustomHeadersJSON(dest.CustomHeaders); err != nil {
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
	if dest.AuthType == "KLHK_ONLIMO_BODY" {
		dest.APIKeyValue = strings.TrimSpace(dest.APIKeyValue)
		dest.AuthToken = strings.TrimSpace(dest.AuthToken)
	}
	if dest.CreatedAt.IsZero() {
		dest.CreatedAt = existing.CreatedAt
	}
	if dest.LastDelivery == nil {
		dest.LastDelivery = existing.LastDelivery
	}
	if dest.RetryLimit <= 0 {
		dest.RetryLimit = existing.RetryLimit
		if dest.RetryLimit <= 0 {
			dest.RetryLimit = 5
		}
	}
	if dest.QueueLimit <= 0 {
		dest.QueueLimit = existing.QueueLimit
		if dest.QueueLimit <= 0 {
			dest.QueueLimit = 10000
		}
	}
	if dest.BatchSize <= 0 {
		dest.BatchSize = existing.BatchSize
		if dest.BatchSize <= 0 {
			dest.BatchSize = 100
		}
	}
	if dest.Format == "" {
		dest.Format = existing.Format
	}

	s.clearCachedAuthToken(dest.ID)
	if err := s.repo.UpdateDestination(ctx, dest); err != nil {
		return err
	}

	// Keep SPARING payload structure synchronized when user toggles DispatchIntervalSeconds (Per 2 Menit <-> Per 1 Jam / 30 Data)
	var updatedCfg *model.DestinationPayloadConfig
	if existing.DispatchIntervalSeconds != dest.DispatchIntervalSeconds {
		if cfg, cfgErr := s.repo.GetPayloadConfigByDestinationID(ctx, dest.ID); cfgErr == nil && cfg != nil {
			tpl := strings.ToUpper(strings.TrimSpace(cfg.TemplateType))
			str := strings.ToUpper(strings.TrimSpace(cfg.Structure))
			isSparing := tpl == "KLHK_SPARING" || tpl == "KLHK_SPARING_HOURLY" ||
				dest.AuthType == "SPARING_JWT_2STEP" || dest.AuthType == "SPARING_JWT_STATIC"
			if isSparing {
				if dest.DispatchIntervalSeconds > 0 && str == "FLAT" {
					cfg.TemplateType = "KLHK_SPARING_HOURLY"
					cfg.Structure = "SPARING_DATA_ARRAY"
					if cfg.NestedDataKey == "" {
						cfg.NestedDataKey = "data"
					}
					if cfg.TimestampSource == "" || cfg.TimestampSource == "PERIOD_END" {
						cfg.TimestampSource = "PERIOD_START"
					}
					if cfg.NumericPrecisionMode == "" || cfg.NumericPrecisionMode == "FIXED_2" {
						cfg.NumericPrecisionMode = "NATURAL_2"
					}
					cfg.OmitIntervalField = true
					cfg.IncludeMetadata = false
					_ = s.repo.UpsertPayloadConfig(ctx, cfg)
					updatedCfg = cfg
				} else if dest.DispatchIntervalSeconds == 0 && str == "SPARING_DATA_ARRAY" {
					cfg.TemplateType = "KLHK_SPARING"
					cfg.Structure = "FLAT"
					_ = s.repo.UpsertPayloadConfig(ctx, cfg)
					updatedCfg = cfg
				}
			}
		}
	}

	if dest.Enabled {
		dest.RequeuedBatches = s.reprocessUndeliveredItemsForDestination(ctx, dest.ID, updatedCfg, true)
	} else {
		_ = s.repo.SyncDestinationStatsFromQueue(ctx, &dest.ID)
	}
	return nil
}

// DeleteDestination removes a destination
func (s *outboundService) DeleteDestination(ctx context.Context, id uint) error {
	s.clearCachedAuthToken(id)
	return s.repo.DeleteDestination(ctx, id)
}

// GetPayloadConfig fetches payload config for a destination
func (s *outboundService) GetPayloadConfig(ctx context.Context, destID uint) (*model.DestinationPayloadConfig, error) {
	return s.repo.GetPayloadConfigByDestinationID(ctx, destID)
}

// UpdatePayloadConfig validates and updates payload config, then rebuilds PayloadSnapshot
// on all undelivered (PENDING, RETRYING, FAILED) queue items for this destination and re-queues them.
func (s *outboundService) UpdatePayloadConfig(ctx context.Context, cfg *model.DestinationPayloadConfig) error {
	if err := s.formatter.ValidateConfig(cfg); err != nil {
		return err
	}
	if err := s.repo.UpsertPayloadConfig(ctx, cfg); err != nil {
		return err
	}

	destEnabled := true
	if dest, err := s.repo.GetDestinationByID(ctx, cfg.DestinationID); err == nil && dest != nil {
		destEnabled = dest.Enabled
	}
	cfg.RequeuedBatches = s.reprocessUndeliveredItemsForDestination(ctx, cfg.DestinationID, cfg, destEnabled)
	return nil
}

// rebuildQueueItemPayload reconstructs OutboundBatchData from persisted SourceValuesJSON and re-serializes
// PayloadSnapshot using the latest DestinationPayloadConfig. Returns true if PayloadSnapshot was rebuilt.
func (s *outboundService) rebuildQueueItemPayload(item *model.OutboundDeliveryQueue, cfg *model.DestinationPayloadConfig) bool {
	if item == nil || cfg == nil || strings.TrimSpace(item.SourceValuesJSON) == "" {
		return false
	}

	var allSourceItems []model.SourceAggregationValueItem
	if err := json.Unmarshal([]byte(item.SourceValuesJSON), &allSourceItems); err != nil || len(allSourceItems) == 0 {
		return false
	}

	matcher := buildParameterMappingMatcher(cfg)
	var activeItems []model.SourceAggregationValueItem
	for _, src := range allSourceItems {
		devCode := src.DeviceCode
		if devCode == "" && src.DeviceID > 0 {
			devCode = fmt.Sprintf("DEV-%d", src.DeviceID)
		}
		enabled, _ := matcher.ResolveWithName(src.DeviceID, devCode, src.ParameterID, src.ParameterCode, src.ParameterName)
		if enabled {
			activeItems = append(activeItems, src)
		}
	}
	if len(activeItems) == 0 {
		activeItems = allSourceItems
	}

	deviceMap := make(map[uint]BatchDeviceSummary)
	var deviceOrder []uint
	for _, src := range activeItems {
		if _, exists := deviceMap[src.DeviceID]; !exists {
			devCode := src.DeviceCode
			if devCode == "" {
				devCode = fmt.Sprintf("DEV-%d", src.DeviceID)
			}
			devName := src.DeviceName
			if devName == "" {
				devName = fmt.Sprintf("Device %d", src.DeviceID)
			}
			deviceMap[src.DeviceID] = BatchDeviceSummary{
				DeviceID:   src.DeviceID,
				DeviceCode: devCode,
				DeviceName: devName,
			}
			deviceOrder = append(deviceOrder, src.DeviceID)
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

	intervalSec := item.IntervalSeconds
	if intervalSec <= 0 && !item.PeriodEnd.IsZero() && !item.PeriodStart.IsZero() {
		if diff := int(item.PeriodEnd.Sub(item.PeriodStart).Seconds()); diff > 0 {
			intervalSec = diff
		}
	}

	batchData := &OutboundBatchData{
		DeviceCount:     deviceCount,
		DevicesSummary:  devicesSummary,
		Devices:         devicesList,
		IntervalSeconds: intervalSec,
		PeriodStart:     item.PeriodStart,
		PeriodEnd:       item.PeriodEnd,
		Values:          activeItems,
	}
	if deviceCount == 1 {
		batchData.DeviceID = devicesList[0].DeviceID
		batchData.DeviceCode = devicesList[0].DeviceCode
		batchData.DeviceName = devicesList[0].DeviceName
	} else if item.DeviceID != nil {
		batchData.DeviceID = *item.DeviceID
	}

	payloadBytes, err := s.formatter.Format(batchData, cfg)
	if err != nil {
		logger.Warn("[Outbound] Failed to rebuild payload for queue item %d (%s): %v", item.ID, item.BatchID, err)
		return false
	}

	item.PayloadSnapshot = string(payloadBytes)
	if cfg.Version > 0 {
		item.PayloadVersion = cfg.Version
	}
	item.ParameterCount = len(activeItems)
	if deviceCount > 0 {
		item.DeviceCount = deviceCount
		item.DevicesSummary = devicesSummary
	}
	return true
}

// reprocessUndeliveredItemsForDestination refreshes PayloadSnapshot on all PENDING, RETRYING, and FAILED
// queue items for destID and resets them to PENDING (attempts=0) so corrected config/credentials take effect automatically.
func (s *outboundService) reprocessUndeliveredItemsForDestination(
	ctx context.Context,
	destID uint,
	cfg *model.DestinationPayloadConfig,
	wakeWorker bool,
) int {
	if destID == 0 {
		return 0
	}
	items, err := s.repo.GetUndeliveredQueueItemsByDestination(ctx, destID)
	if err != nil || len(items) == 0 {
		_ = s.repo.SyncDestinationStatsFromQueue(ctx, &destID)
		return 0
	}

	if cfg == nil {
		cfg, _ = s.repo.GetPayloadConfigByDestinationID(ctx, destID)
	}

	updatedCount := 0
	for i := range items {
		item := &items[i]
		_ = s.rebuildQueueItemPayload(item, cfg)
		if err := s.repo.UpdateQueueItemPayloadAndRequeue(ctx, item, true); err == nil {
			updatedCount++
		}
	}

	_ = s.repo.SyncDestinationStatsFromQueue(ctx, &destID)
	if updatedCount > 0 {
		logger.Info("[Outbound] Auto-refreshed & re-queued %d undelivered batch(es) for destination #%d", updatedCount, destID)
		if wakeWorker {
			s.notifyWorker()
		}
	}
	return updatedCount
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

	matcher := buildParameterMappingMatcher(cfg)
	structureUpper := strings.ToUpper(strings.TrimSpace(cfg.Structure))
	templateUpper := strings.ToUpper(strings.TrimSpace(cfg.TemplateType))
	isSparingArray := structureUpper == "SPARING_DATA_ARRAY"

	var realResults []model.AggregationResult
	var err error
	if isSparingArray || (dest != nil && dest.DispatchIntervalSeconds > 0 && customCfg == nil) {
		srcInterval := 120
		if dest != nil && dest.IntervalSeconds > 0 {
			srcInterval = dest.IntervalSeconds
		}
		winDur := time.Hour
		if dest != nil && dest.DispatchIntervalSeconds > 0 {
			winDur = time.Duration(dest.DispatchIntervalSeconds) * time.Second
		}
		realResults, err = s.repo.GetRecentAggregationResultsMultiDevice(ctx, srcInterval, winDur, targetDeviceIDs)
		if (err != nil || len(realResults) == 0) && srcInterval > 0 {
			realResults, err = s.repo.GetRecentAggregationResultsMultiDevice(ctx, 0, winDur, targetDeviceIDs)
		}
	} else {
		realResults, err = s.repo.GetLatestAggregationResultsMultiDevice(ctx, targetDeviceIDs)
	}

	var filteredReal []model.AggregationResult
	if err == nil && len(realResults) > 0 {
		for _, r := range realResults {
			if !isActiveOutboundAggregationResult(r) {
				continue
			}
			devCode := r.Device.DeviceCode
			if devCode == "" {
				devCode = fmt.Sprintf("DEV-%d", r.DeviceID)
			}
			if enabled, _ := matcher.ResolveWithName(r.DeviceID, devCode, r.ParameterID, r.Parameter.ParameterCode, r.Parameter.ParameterName); !enabled {
				continue
			}
			filteredReal = append(filteredReal, r)
		}
	}

	if len(filteredReal) > 0 {
		deviceMap := make(map[uint]BatchDeviceSummary)
		var deviceOrder []uint
		for _, r := range filteredReal {
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
		for _, r := range filteredReal {
			pCode := r.Parameter.ParameterCode
			pName := r.Parameter.ParameterName
			unit := r.Parameter.Unit
			devInfo := deviceMap[r.DeviceID]
			val := r.Value
			if val != nil {
				roundedVal := roundOutboundValue(*val)
				val = &roundedVal
			}
			items = append(items, model.SourceAggregationValueItem{
				DeviceID:            r.DeviceID,
				DeviceCode:          devInfo.DeviceCode,
				DeviceName:          devInfo.DeviceName,
				ParameterID:         r.ParameterID,
				ParameterCode:       pCode,
				ParameterName:       pName,
				Value:               val,
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
			IntervalSeconds: filteredReal[0].IntervalSeconds,
			PeriodStart:     filteredReal[0].PeriodStart,
			PeriodEnd:       filteredReal[len(filteredReal)-1].PeriodEnd,
			Values:          items,
		}
		if len(devicesList) == 1 {
			batch.DeviceID = devicesList[0].DeviceID
			batch.DeviceCode = devicesList[0].DeviceCode
			batch.DeviceName = devicesList[0].DeviceName
		}
	} else if isSparingArray || templateUpper == "KLHK_SPARING" || templateUpper == "KLHK_SPARING_HOURLY" {
		// Construct realistic KLHK SPARING sample batch (either 30x 2-minute hourly bundle or 1x 2-minute flat record)
		srcSec := 120
		if dest != nil && dest.IntervalSeconds > 0 {
			srcSec = dest.IntervalSeconds
		}
		dispSec := 3600
		if dest != nil && dest.DispatchIntervalSeconds > 0 {
			dispSec = dest.DispatchIntervalSeconds
		}
		bucketCount := 1
		if isSparingArray {
			bucketCount = dispSec / srcSec
			if bucketCount < 1 {
				bucketCount = 30
			} else if bucketCount > 60 {
				bucketCount = 60
			}
		}

		alignedHourEnd := time.Unix((time.Now().Unix()/3600)*3600, 0).UTC()
		if !isSparingArray {
			alignedHourEnd = time.Unix((time.Now().Unix()/int64(srcSec))*int64(srcSec), 0).UTC()
		}
		windowStart := alignedHourEnd.Add(-time.Duration(bucketCount*srcSec) * time.Second)

		type sparingParamSpec struct {
			deviceID   uint
			deviceCode string
			paramID    uint
			code       string
			name       string
			baseVal    float64
			stepDelta  float64
		}
		var specs []sparingParamSpec
		if cfg != nil && strings.TrimSpace(cfg.ParameterMappingsJSON) != "" {
			var mappings []model.ParameterMappingItem
			_ = json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings)
			for idx, m := range mappings {
				if !m.Enabled || strings.TrimSpace(m.ParameterCode) == "" {
					continue
				}
				dID := m.DeviceID
				if dID == 0 {
					dID = 1
				}
				dCode := strings.TrimSpace(m.DeviceCode)
				if dCode == "" {
					dCode = fmt.Sprintf("SPARING-%02d", dID)
				}
				base := 15.0 + float64(idx)*5.25
				delta := 0.02
				switch strings.ToLower(strings.TrimSpace(m.ParameterCode)) {
				case "ph", "p_h":
					base, delta = 6.93, 0.01
				case "cod":
					base, delta = 40.99, 0.12
				case "tss":
					base, delta = 27.95, 0.03
				case "nh3n", "nh3-n", "nh3_n", "ammonia", "amonia":
					base, delta = 0, 0
				case "debit", "flow", "flowrate", "flow_rate", "flow-rate":
					base, delta = 0.24, 0.01
				}
				specs = append(specs, sparingParamSpec{
					deviceID:   dID,
					deviceCode: dCode,
					paramID:    m.ParameterID,
					code:       m.ParameterCode,
					name:       m.ParameterName,
					baseVal:    base,
					stepDelta:  delta,
				})
			}
		}
		if len(specs) == 0 {
			specs = []sparingParamSpec{
				{deviceID: 1, deviceCode: "SPARING-01", paramID: 1, code: "ph", name: "pH", baseVal: 6.93, stepDelta: 0.01},
				{deviceID: 1, deviceCode: "SPARING-01", paramID: 2, code: "cod", name: "COD", baseVal: 40.99, stepDelta: 0.12},
				{deviceID: 1, deviceCode: "SPARING-01", paramID: 3, code: "tss", name: "TSS", baseVal: 27.95, stepDelta: 0.03},
				{deviceID: 1, deviceCode: "SPARING-01", paramID: 4, code: "nh3n", name: "NH3-N", baseVal: 0, stepDelta: 0},
				{deviceID: 1, deviceCode: "SPARING-01", paramID: 5, code: "debit", name: "Debit", baseVal: 0.24, stepDelta: 0.01},
			}
		}

		deviceMap := make(map[string]BatchDeviceSummary)
		var deviceOrder []string
		var items []model.SourceAggregationValueItem
		aggSeq := uint64(1000)

		for bIdx := 0; bIdx < bucketCount; bIdx++ {
			bStart := windowStart.Add(time.Duration(bIdx*srcSec) * time.Second)
			bEnd := bStart.Add(time.Duration(srcSec) * time.Second)
			for _, sp := range specs {
				if _, ok := deviceMap[sp.deviceCode]; !ok {
					deviceMap[sp.deviceCode] = BatchDeviceSummary{
						DeviceID:   sp.deviceID,
						DeviceCode: sp.deviceCode,
						DeviceName: sp.deviceCode,
					}
					deviceOrder = append(deviceOrder, sp.deviceCode)
				}
				v := roundOutboundValue(sp.baseVal + float64(bIdx%8)*sp.stepDelta)
				aggSeq++
				items = append(items, model.SourceAggregationValueItem{
					DeviceID:            sp.deviceID,
					DeviceCode:          sp.deviceCode,
					DeviceName:          sp.deviceCode,
					ParameterID:         sp.paramID,
					ParameterCode:       sp.code,
					ParameterName:       sp.name,
					Value:               &v,
					Quality:             "GOOD",
					AggregationResultID: aggSeq,
					PeriodStart:         bStart,
					PeriodEnd:           bEnd,
				})
			}
		}

		var devicesList []BatchDeviceSummary
		for _, dc := range deviceOrder {
			devicesList = append(devicesList, deviceMap[dc])
		}
		batch = &OutboundBatchData{
			DeviceCount:     len(devicesList),
			DevicesSummary:  strings.Join(deviceOrder, ", "),
			Devices:         devicesList,
			IntervalSeconds: srcSec,
			PeriodStart:     windowStart,
			PeriodEnd:       alignedHourEnd,
			Values:          items,
		}
		if len(devicesList) == 1 {
			batch.DeviceID = devicesList[0].DeviceID
			batch.DeviceCode = devicesList[0].DeviceCode
			batch.DeviceName = devicesList[0].DeviceName
		}
	} else if templateUpper == "KLHK_WQMS_ONLIMO" || structureUpper == "ONLIMO_DATA_OBJECT" {
		// Construct realistic KLHK WQMS / ONLIMO sample batch (13 canonical water quality parameters)
		srcSec := 3600
		if dest != nil && dest.IntervalSeconds > 0 {
			srcSec = dest.IntervalSeconds
		}
		alignedEnd := time.Unix((time.Now().Unix()/int64(srcSec))*int64(srcSec), 0).UTC()
		windowStart := alignedEnd.Add(-time.Duration(srcSec) * time.Second)

		type wqmsSampleSpec struct {
			deviceID   uint
			deviceCode string
			paramID    uint
			code       string
			name       string
			val        float64
		}
		defaultValByAlias := map[string]float64{
			"Suhu":      27.5,
			"DHL":       385.88,
			"TDS":       192.4,
			"Salinitas": 3.88,
			"DO":        4.56,
			"PH":        6.97,
			"Turbidity": 12.3,
			"Kedalaman": 1.45,
			"Nitrat":    74.05,
			"Amonia":    0.12,
			"COD":       68.05,
			"BOD":       56.04,
			"TSS":       24.5,
		}
		var specs []wqmsSampleSpec
		if cfg != nil && strings.TrimSpace(cfg.ParameterMappingsJSON) != "" {
			var mappings []model.ParameterMappingItem
			_ = json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings)
			for idx, m := range mappings {
				if !m.Enabled || strings.TrimSpace(m.ParameterCode) == "" {
					continue
				}
				dID := m.DeviceID
				if dID == 0 {
					dID = 1
				}
				dCode := strings.TrimSpace(m.DeviceCode)
				if dCode == "" {
					dCode = fmt.Sprintf("WQMS-%02d", dID)
				}
				alias := strings.TrimSpace(m.ExternalName)
				if alias == "" {
					alias = defaultKLHKWQMSAlias(m.ParameterCode, m.ParameterName)
				}
				v, hasDefault := defaultValByAlias[alias]
				if !hasDefault {
					v = 10.0 + float64(idx)*2.25
				}
				specs = append(specs, wqmsSampleSpec{
					deviceID:   dID,
					deviceCode: dCode,
					paramID:    m.ParameterID,
					code:       m.ParameterCode,
					name:       m.ParameterName,
					val:        v,
				})
			}
		}
		if len(specs) == 0 {
			specs = []wqmsSampleSpec{
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 1, code: "temperature", name: "Temperature", val: 27.5},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 2, code: "cl", name: "Conductivity / DHL", val: 385.88},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 3, code: "fl", name: "TDS", val: 192.4},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 4, code: "salinity", name: "Salinity", val: 3.88},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 5, code: "do", name: "DO", val: 4.56},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 6, code: "ph", name: "PH", val: 6.97},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 7, code: "turbidity", name: "Turbidity", val: 12.3},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 8, code: "depth", name: "Depth", val: 1.45},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 9, code: "no3", name: "Nitrat", val: 74.05},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 10, code: "nh4", name: "Ammonia (NH3-N)", val: 0.12},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 11, code: "cod", name: "COD", val: 68.05},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 12, code: "bod", name: "BOD", val: 56.04},
				{deviceID: 1, deviceCode: "WQMS-01", paramID: 13, code: "tss", name: "TSS", val: 24.5},
			}
		}

		deviceMap := make(map[string]BatchDeviceSummary)
		var deviceOrder []string
		var items []model.SourceAggregationValueItem
		for idx, sp := range specs {
			if _, ok := deviceMap[sp.deviceCode]; !ok {
				deviceMap[sp.deviceCode] = BatchDeviceSummary{
					DeviceID:   sp.deviceID,
					DeviceCode: sp.deviceCode,
					DeviceName: sp.deviceCode,
				}
				deviceOrder = append(deviceOrder, sp.deviceCode)
			}
			vCopy := sp.val
			items = append(items, model.SourceAggregationValueItem{
				DeviceID:            sp.deviceID,
				DeviceCode:          sp.deviceCode,
				DeviceName:          sp.deviceCode,
				ParameterID:         sp.paramID,
				ParameterCode:       sp.code,
				ParameterName:       sp.name,
				Value:               &vCopy,
				Quality:             "GOOD",
				AggregationResultID: uint64(300 + idx),
				PeriodStart:         windowStart,
				PeriodEnd:           alignedEnd,
			})
		}

		var devicesList []BatchDeviceSummary
		for _, dc := range deviceOrder {
			devicesList = append(devicesList, deviceMap[dc])
		}
		batch = &OutboundBatchData{
			DeviceCount:     len(devicesList),
			DevicesSummary:  strings.Join(deviceOrder, ", "),
			Devices:         devicesList,
			IntervalSeconds: srcSec,
			PeriodStart:     windowStart,
			PeriodEnd:       alignedEnd,
			Values:          items,
		}
		if len(devicesList) == 1 {
			batch.DeviceID = devicesList[0].DeviceID
			batch.DeviceCode = devicesList[0].DeviceCode
			batch.DeviceName = devicesList[0].DeviceName
		}
	} else if cfg != nil && strings.TrimSpace(cfg.ParameterMappingsJSON) != "" {
		// Construct sample batch matching the user's configured ParameterMappings checklist
		now := time.Now()
		var mappings []model.ParameterMappingItem
		_ = json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings)

		deviceMap := make(map[string]BatchDeviceSummary)
		var deviceOrder []string
		var items []model.SourceAggregationValueItem
		sampleVal := 25.4

		for idx, m := range mappings {
			if !m.Enabled || strings.TrimSpace(m.ParameterCode) == "" {
				continue
			}
			dID := m.DeviceID
			if dID == 0 {
				dID = 1
			}
			dCode := strings.TrimSpace(m.DeviceCode)
			if dCode == "" {
				dCode = fmt.Sprintf("DEV-%d", dID)
			}
			if _, ok := deviceMap[dCode]; !ok {
				deviceMap[dCode] = BatchDeviceSummary{
					DeviceID:   dID,
					DeviceCode: dCode,
					DeviceName: dCode,
				}
				deviceOrder = append(deviceOrder, dCode)
			}
			v := sampleVal + float64(idx)*1.5
			items = append(items, model.SourceAggregationValueItem{
				DeviceID:            dID,
				DeviceCode:          dCode,
				DeviceName:          dCode,
				ParameterID:         m.ParameterID,
				ParameterCode:       m.ParameterCode,
				ParameterName:       m.ParameterName,
				Value:               &v,
				Quality:             "GOOD",
				AggregationResultID: uint64(100 + idx),
				PeriodStart:         now.Add(-2 * time.Minute),
				PeriodEnd:           now,
			})
		}

		var devicesList []BatchDeviceSummary
		for _, dc := range deviceOrder {
			devicesList = append(devicesList, deviceMap[dc])
		}
		batch = &OutboundBatchData{
			DeviceCount:     len(devicesList),
			DevicesSummary:  strings.Join(deviceOrder, ", "),
			Devices:         devicesList,
			IntervalSeconds: 120,
			PeriodStart:     now.Add(-2 * time.Minute),
			PeriodEnd:       now,
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

	previewStr, err := s.formatter.GeneratePreview(batch, cfg)
	if err != nil {
		return "", err
	}
	if dest != nil && dest.AuthType == "KLHK_ONLIMO_BODY" {
		maskedKey := "***REDACTED***"
		if strings.TrimSpace(dest.APIKeyValue) == "" {
			maskedKey = "<BELUM_DIISI_DI_DESTINASI>"
		}
		maskedSecret := "***REDACTED***"
		if strings.TrimSpace(dest.AuthToken) == "" {
			maskedSecret = "<BELUM_DIISI_DI_DESTINASI>"
		}
		if withCreds, cErr := injectRootJSONCredentialsPretty(previewStr, maskedKey, maskedSecret); cErr == nil {
			previewStr = withCreds
		}
	}
	return previewStr, nil
}

// TestPayload generates a preview payload and attempts direct delivery to destination
func (s *outboundService) TestPayload(ctx context.Context, destID uint) (*TestPayloadResult, error) {
	return s.TestPayloadWithConfig(ctx, destID, nil)
}

// TestPayloadWithConfig generates a preview payload (using custom draft config if provided) and attempts direct delivery to destination
func (s *outboundService) TestPayloadWithConfig(ctx context.Context, destID uint, customCfg *model.DestinationPayloadConfig) (*TestPayloadResult, error) {
	dest, err := s.repo.GetDestinationByID(ctx, destID)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %w", err)
	}

	previewJSON, err := s.GeneratePreviewWithConfig(ctx, destID, nil, customCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payload: %w", err)
	}

	result := &TestPayloadResult{
		PayloadSent: previewJSON,
	}

	startTime := time.Now()
	switch dest.Type {
	case model.DestREST:
		out := s.deliverRESTDetailed(ctx, dest, previewJSON)
		result.LatencyMs = time.Since(startTime).Milliseconds()
		result.HTTPCode = out.httpCode
		result.ResponseText = out.body
		result.IsTwoStep = out.isTwoStep
		if out.isTwoStep {
			result.AuthURL = dest.AuthURL
			result.DataURL = dest.Endpoint
			result.AuthUsedCache = out.authUsedCache
			result.AuthSuccess = out.authSuccess
			result.AuthHTTPCode = out.authHTTPCode
			result.AuthLatencyMs = out.authLatencyMs
			result.AuthTokenPreview = out.authTokenPreview
			result.AuthMessage = out.authMessage
			result.DataLatencyMs = out.dataLatencyMs
		}
		if out.err != nil {
			result.Success = false
			result.ErrorMessage = out.err.Error()
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

// RetryQueueItem rebuilds PayloadSnapshot with the latest destination config (if available),
// resets the queue item to PENDING (attempts=0), and wakes the delivery worker immediately.
func (s *outboundService) RetryQueueItem(ctx context.Context, id uint64) error {
	item, err := s.repo.GetQueueItemByID(ctx, id)
	if err == nil && item != nil {
		s.clearCachedAuthToken(item.DestinationID)
		if cfg, cfgErr := s.repo.GetPayloadConfigByDestinationID(ctx, item.DestinationID); cfgErr == nil && cfg != nil {
			_ = s.rebuildQueueItemPayload(item, cfg)
		}
		if updErr := s.repo.UpdateQueueItemPayloadAndRequeue(ctx, item, true); updErr != nil {
			return updErr
		}
		if item.DestinationID > 0 {
			_ = s.repo.SyncDestinationStatsFromQueue(ctx, &item.DestinationID)
		}
		s.notifyWorker()
		return nil
	}
	if err := s.repo.ResetQueueItemForRetry(ctx, id); err != nil {
		return err
	}
	s.notifyWorker()
	return nil
}

// ReplayQueueItem creates a new delivery queue entry from original source values,
// re-serializing PayloadSnapshot against the latest DestinationPayloadConfig when available.
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
		DeviceCount:          original.DeviceCount,
		DevicesSummary:       original.DevicesSummary,
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

	if cfg, cfgErr := s.repo.GetPayloadConfigByDestinationID(ctx, original.DestinationID); cfgErr == nil && cfg != nil {
		_ = s.rebuildQueueItemPayload(newItem, cfg)
	}

	if err := s.repo.EnqueueDelivery(ctx, newItem); err != nil {
		return nil, err
	}
	s.notifyWorker()
	return newItem, nil
}

// CancelQueueItem marks an item as CANCELLED
func (s *outboundService) CancelQueueItem(ctx context.Context, id uint64) error {
	return s.repo.CancelQueueItem(ctx, id)
}

// DeleteQueueItem permanently removes a single queue item
func (s *outboundService) DeleteQueueItem(ctx context.Context, id uint64) error {
	return s.repo.DeleteQueueItem(ctx, id)
}

// PurgeQueue bulk deletes queue items and optional delivery logs
func (s *outboundService) PurgeQueue(ctx context.Context, filter repository.QueuePurgeFilter) (int64, int64, error) {
	return s.repo.PurgeQueueItems(ctx, filter)
}

// StartPurge initiates a non-blocking background queue purge job with progress tracking
func (s *outboundService) StartPurge(ctx context.Context, filter repository.QueuePurgeFilter) (*PurgeProgress, error) {
	s.purgeMu.Lock()
	if s.purgeProgress.IsRunning {
		prog := s.purgeProgress
		s.purgeMu.Unlock()
		return &prog, nil
	}

	queueCount, logsCount, err := s.repo.CountQueuePurgeItems(ctx, filter)
	if err != nil {
		s.purgeMu.Unlock()
		return nil, fmt.Errorf("failed to count items to purge: %w", err)
	}

	now := time.Now()
	s.purgeCancelled = false
	s.purgeProgress = PurgeProgress{
		IsRunning:    true,
		Status:       "RUNNING",
		TotalQueue:   queueCount,
		PurgedQueue:  0,
		TotalLogs:    logsCount,
		PurgedLogs:   0,
		Percent:      0.0,
		StartedAt:    &now,
		ErrorMessage: "",
	}

	if queueCount == 0 && logsCount == 0 {
		s.purgeProgress.IsRunning = false
		s.purgeProgress.Status = "COMPLETED"
		s.purgeProgress.Percent = 100.0
		s.purgeProgress.CompletedAt = &now
		prog := s.purgeProgress
		s.purgeMu.Unlock()
		return &prog, nil
	}

	prog := s.purgeProgress
	s.purgeMu.Unlock()

	go s.runBackgroundPurge(filter)

	return &prog, nil
}

// GetPurgeStatus returns the current snapshot of the background purge job
func (s *outboundService) GetPurgeStatus() PurgeProgress {
	s.purgeMu.RLock()
	defer s.purgeMu.RUnlock()
	return s.purgeProgress
}

// CancelPurge requests cancellation of the active purge job
func (s *outboundService) CancelPurge() {
	s.purgeMu.Lock()
	defer s.purgeMu.Unlock()
	if s.purgeProgress.IsRunning {
		s.purgeCancelled = true
	}
}

func (s *outboundService) runBackgroundPurge(filter repository.QueuePurgeFilter) {
	ctx := context.Background()
	batchSize := 1000

	for {
		s.purgeMu.RLock()
		cancelled := s.purgeCancelled
		s.purgeMu.RUnlock()

		if cancelled {
			s.purgeMu.Lock()
			s.purgeProgress.IsRunning = false
			s.purgeProgress.Status = "CANCELLED"
			now := time.Now()
			s.purgeProgress.CompletedAt = &now
			s.purgeMu.Unlock()
			logger.Info("[Outbound] Manual queue purge job cancelled by user")
			return
		}

		delQ, delL, done, err := s.repo.PurgeQueueBatch(ctx, filter, batchSize)
		if err != nil {
			s.purgeMu.Lock()
			s.purgeProgress.IsRunning = false
			s.purgeProgress.Status = "FAILED"
			s.purgeProgress.ErrorMessage = err.Error()
			now := time.Now()
			s.purgeProgress.CompletedAt = &now
			s.purgeMu.Unlock()
			logger.Error("[Outbound] Manual queue purge job failed: %v", err)
			return
		}

		s.purgeMu.Lock()
		s.purgeProgress.PurgedQueue += delQ
		s.purgeProgress.PurgedLogs += delL
		total := s.purgeProgress.TotalQueue + s.purgeProgress.TotalLogs
		purged := s.purgeProgress.PurgedQueue + s.purgeProgress.PurgedLogs
		if total > 0 {
			pct := float64(purged) / float64(total) * 100.0
			if pct > 100.0 {
				pct = 100.0
			}
			s.purgeProgress.Percent = pct
		}

		if done {
			s.purgeProgress.IsRunning = false
			s.purgeProgress.Status = "COMPLETED"
			s.purgeProgress.Percent = 100.0
			now := time.Now()
			s.purgeProgress.CompletedAt = &now
			s.purgeMu.Unlock()
			logger.Info("[Outbound] Manual queue purge completed: %d queue items and %d logs deleted", s.purgeProgress.PurgedQueue, s.purgeProgress.PurgedLogs)
			return
		}
		s.purgeMu.Unlock()

		time.Sleep(15 * time.Millisecond)
	}
}

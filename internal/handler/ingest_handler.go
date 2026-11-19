package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	ingestService  *service.IngestService
	inboundLogRepo *repository.InboundLogRepository
}

func NewIngestHandler(ingestService *service.IngestService) *IngestHandler {
	return &IngestHandler{ingestService: ingestService}
}

func (h *IngestHandler) SetInboundLogRepository(repo *repository.InboundLogRepository) {
	h.inboundLogRepo = repo
}

func redactIngestKeyPrefix(rawKey string) string {
	k := strings.TrimSpace(rawKey)
	if k == "" {
		return ""
	}
	if len(k) <= 12 {
		return k[:minInt(4, len(k))] + "…"
	}
	return k[:12] + "…"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func resolveIngestErrorHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var httpErr *service.IngestHTTPError
	if errors.As(err, &httpErr) && httpErr != nil && httpErr.StatusCode >= 400 && httpErr.StatusCode <= 599 {
		return httpErr.StatusCode
	}
	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "missing") || strings.Contains(lower, "invalid or disabled"):
		return http.StatusUnauthorized
	case strings.Contains(lower, "suspended") || strings.Contains(lower, "locked to station"):
		return http.StatusForbidden
	case strings.Contains(lower, "not registered"):
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func (h *IngestHandler) recordInboundLog(c *gin.Context, source, rawKey string, bodyBytes []byte, res *service.EdgeIngestResponse, err error, httpStatus int, startedAt time.Time) {
	if h == nil || h.inboundLogRepo == nil {
		return
	}
	entry := &model.InboundIngestLog{
		ReceivedAt:   time.Now().UTC(),
		Endpoint:     c.Request.URL.Path,
		Source:       strings.ToUpper(strings.TrimSpace(source)),
		HTTPStatus:   httpStatus,
		ClientIP:     c.ClientIP(),
		KeyPrefix:    redactIngestKeyPrefix(rawKey),
		PayloadBytes: len(bodyBytes),
		LatencyMs:    time.Since(startedAt).Milliseconds(),
		RawPayload:   string(bodyBytes),
	}
	if entry.Source == "" {
		entry.Source = "EDGE_PUSH"
	}
	if res != nil && err == nil {
		entry.Status = "ACCEPTED"
		entry.TenantID = res.TenantID
		entry.SiteID = res.SiteID
		entry.SiteCode = res.SiteCode
		entry.DeviceCode = res.DeviceCode
		entry.LoggerType = strings.ToUpper(strings.TrimSpace(res.LoggerType))
		entry.DetectedFormat = strings.ToUpper(strings.TrimSpace(res.DetectedFormat))
		if res.Source != "" {
			entry.Source = strings.ToUpper(strings.TrimSpace(res.Source))
		}
		entry.RecordsProcessed = res.RecordsProcessed
		entry.ParametersIngested = res.ParametersIngested
		entry.ParametersSkipped = res.ParametersSkipped
	} else {
		entry.Status = "REJECTED"
		if err != nil {
			entry.ErrorMessage = err.Error()
		}
		// Best-effort metadata peek from JSON body so rejected pushes remain searchable by site/device/domain
		if len(bodyBytes) > 0 && len(bodyBytes) <= repository.MaxInboundRawPayloadBytes {
			var peek map[string]interface{}
			if json.Unmarshal(bodyBytes, &peek) == nil {
				if v, ok := peek["site_code"].(string); ok && entry.SiteCode == "" {
					entry.SiteCode = strings.TrimSpace(v)
				} else if v, ok := peek["uid"].(string); ok && entry.SiteCode == "" {
					entry.SiteCode = strings.TrimSpace(v)
				}
				if v, ok := peek["device_code"].(string); ok && entry.DeviceCode == "" {
					entry.DeviceCode = strings.TrimSpace(v)
				}
				if v, ok := peek["logger_type"].(string); ok && entry.LoggerType == "" {
					entry.LoggerType = strings.ToUpper(strings.TrimSpace(v))
				}
			}
		}
	}
	_ = h.inboundLogRepo.Create(context.Background(), entry)
}

// extractRawIngestKey extracts Bearer / Legacy Token first (Authorization: Bearer ..., Token ..., or Token token="..."),
// with backward compatibility for X-Ingest-Key, X-API-Key, and X-MQTT-Token headers.
func extractRawIngestKey(c *gin.Context) string {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader != "" {
		lower := strings.ToLower(authHeader)
		switch {
		case strings.HasPrefix(lower, "bearer "):
			if candidate := strings.Trim(strings.TrimSpace(authHeader[7:]), `"'`); candidate != "" {
				return candidate
			}
		case strings.HasPrefix(lower, "token "):
			rest := strings.TrimSpace(authHeader[6:])
			if strings.HasPrefix(strings.ToLower(rest), "token=") {
				rest = strings.TrimSpace(rest[6:])
			}
			if idx := strings.Index(rest, ","); idx > 0 {
				rest = strings.TrimSpace(rest[:idx])
			}
			if candidate := strings.Trim(rest, `"'`); candidate != "" {
				return candidate
			}
		case !strings.Contains(authHeader, " "):
			if candidate := strings.Trim(authHeader, `"'`); candidate != "" {
				return candidate
			}
		}
	}
	for _, hdr := range []string{"X-Ingest-Key", "X-API-Key", "X-MQTT-Token"} {
		if k := strings.Trim(strings.TrimSpace(c.GetHeader(hdr)), `"'`); k != "" {
			return k
		}
	}
	return ""
}

// ingestAPIResponse includes both standard APIResponse fields ("success": true/false)
// and legacy CBI Edge Datalogger status ("status": 1 on success, "status": 0 on error)
// so both datalogger-27 and legacy Go sendDataPortal (responseDataPortal.Status == 1) acknowledge cleanly.
type ingestAPIResponse struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeIngestSuccess(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, ingestAPIResponse{
		Status:  1,
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func writeIngestFail(c *gin.Context, statusCode int, msg string, details interface{}) {
	c.JSON(statusCode, ingestAPIResponse{
		Status:  0,
		Success: false,
		Message: msg,
		Error:   msg,
		Data:    details,
	})
}

func writeIngestError(c *gin.Context, err error) {
	var httpErr *service.IngestHTTPError
	if errors.As(err, &httpErr) && httpErr != nil {
		code := httpErr.StatusCode
		if code < 400 || code > 599 {
			code = http.StatusBadRequest
		}
		writeIngestFail(c, code, httpErr.Message, httpErr.Details)
		return
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "missing") || strings.Contains(lower, "invalid or disabled"):
		writeIngestFail(c, http.StatusUnauthorized, err.Error(), nil)
	case strings.Contains(lower, "suspended") || strings.Contains(lower, "locked to station"):
		writeIngestFail(c, http.StatusForbidden, err.Error(), nil)
	case strings.Contains(lower, "not registered"):
		writeIngestFail(c, http.StatusNotFound, err.Error(), nil)
	default:
		writeIngestFail(c, http.StatusBadRequest, err.Error(), nil)
	}
}

// IngestTelemetry handles POST /api/v1/ingest/telemetry and POST /api/ingest/telemetry
// using plain JSON + Authorization: Bearer dlg_live_... token for AQMS, SPARING, and ONLIMO dataloggers.
func (h *IngestHandler) IngestTelemetry(c *gin.Context) {
	startedAt := time.Now()
	rawKey := extractRawIngestKey(c)
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if rawKey == "" {
		msg := "Missing Bearer token (provide 'Authorization: Bearer dlg_live_...' header)"
		h.recordInboundLog(c, "EDGE_PUSH", "", bodyBytes, nil, errors.New(msg), http.StatusUnauthorized, startedAt)
		writeIngestFail(c, http.StatusUnauthorized, msg, nil)
		return
	}
	if err != nil || len(bodyBytes) == 0 {
		msg := "Empty or unreadable telemetry JSON body"
		h.recordInboundLog(c, "EDGE_PUSH", rawKey, bodyBytes, nil, errors.New(msg), http.StatusBadRequest, startedAt)
		writeIngestFail(c, http.StatusBadRequest, msg, nil)
		return
	}

	res, err := h.ingestService.IngestPayload(rawKey, bodyBytes, c.ClientIP())
	if err != nil {
		h.recordInboundLog(c, "EDGE_PUSH", rawKey, bodyBytes, nil, err, resolveIngestErrorHTTPStatus(err), startedAt)
		writeIngestError(c, err)
		return
	}

	h.recordInboundLog(c, "EDGE_PUSH", rawKey, bodyBytes, res, nil, http.StatusOK, startedAt)
	writeIngestSuccess(c, "Telemetry ingested successfully", res)
}

// IngestMQTT handles POST /api/v1/ingest/mqtt and POST /api/ingest/mqtt for MQTT broker webhooks & bridges
// when the administrator enables MQTT ingestion (mqtt_ingest_enabled = true).
func (h *IngestHandler) IngestMQTT(c *gin.Context) {
	startedAt := time.Now()
	if !config.Get().MQTTIngestEnabled {
		writeIngestFail(c, http.StatusForbidden, "MQTT telemetry ingestion module is currently disabled by administrator", nil)
		return
	}

	rawKey := extractRawIngestKey(c)
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if rawKey == "" {
		msg := "Missing Bearer token for MQTT ingestion (provide 'Authorization: Bearer dlg_live_...' or 'X-MQTT-Token')"
		h.recordInboundLog(c, "MQTT", "", bodyBytes, nil, errors.New(msg), http.StatusUnauthorized, startedAt)
		writeIngestFail(c, http.StatusUnauthorized, msg, nil)
		return
	}
	if err != nil || len(bodyBytes) == 0 {
		msg := "Empty or unreadable MQTT telemetry JSON payload"
		h.recordInboundLog(c, "MQTT", rawKey, bodyBytes, nil, errors.New(msg), http.StatusBadRequest, startedAt)
		writeIngestFail(c, http.StatusBadRequest, msg, nil)
		return
	}

	res, err := h.ingestService.IngestPayloadWithSource(rawKey, bodyBytes, c.ClientIP(), "MQTT")
	if err != nil {
		h.recordInboundLog(c, "MQTT", rawKey, bodyBytes, nil, err, resolveIngestErrorHTTPStatus(err), startedAt)
		writeIngestError(c, err)
		return
	}

	h.recordInboundLog(c, "MQTT", rawKey, bodyBytes, res, nil, http.StatusOK, startedAt)
	writeIngestSuccess(c, "MQTT telemetry ingested successfully", res)
}

// GetPayloadTemplate handles GET /api/ingest-keys/payload-template to generate standardized
// AQMS, SPARING, and ONLIMO JSON payloads + Bearer Token cURL commands from manually registered parameters.
func (h *IngestHandler) GetPayloadTemplate(c *gin.Context) {
	scope := middleware.GetScope(c)
	if !scope.CanViewAllTenants() && !strings.EqualFold(scope.Role, "ADMIN") {
		if cfg := config.Get(); cfg == nil || !cfg.IsCloudMode() || !strings.EqualFold(scope.Role, "ENGINEER") {
			response.Forbidden(c, "Forbidden: insufficient role privileges for payload template")
			return
		}
	}
	if !scope.CanViewAllTenants() && scope.TenantID > 0 {
		if qTid, err := strconv.ParseUint(c.Query("tenant_id"), 10, 32); err == nil && qTid > 0 && uint(qTid) != scope.TenantID {
			response.Forbidden(c, "Forbidden: cross-tenant payload template access denied")
			return
		}
	} else if scope.CanViewAllTenants() {
		if qTid, err := strconv.ParseUint(c.Query("tenant_id"), 10, 32); err == nil && qTid > 0 {
			scope.TenantID = uint(qTid)
		}
	}

	var siteID, deviceID uint
	if sID, err := strconv.ParseUint(c.Query("site_id"), 10, 32); err == nil {
		siteID = uint(sID)
	}
	if dID, err := strconv.ParseUint(c.Query("device_id"), 10, 32); err == nil {
		deviceID = uint(dID)
	}
	loggerType := strings.TrimSpace(c.Query("logger_type"))

	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost:8080"
	}
	baseURL := scheme + "://" + host

	tpl, err := h.ingestService.BuildPayloadTemplate(scope, siteID, deviceID, loggerType, baseURL)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, tpl)
}

package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"datalogger/internal/config"
	"datalogger/internal/middleware"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	ingestService *service.IngestService
}

func NewIngestHandler(ingestService *service.IngestService) *IngestHandler {
	return &IngestHandler{ingestService: ingestService}
}

// extractRawIngestKey extracts Bearer Token first (Authorization: Bearer dlg_live_...),
// with backward compatibility for X-Ingest-Key, X-API-Key, and X-MQTT-Token headers.
func extractRawIngestKey(c *gin.Context) string {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		if candidate := strings.TrimSpace(authHeader[7:]); candidate != "" {
			return candidate
		}
	}
	for _, hdr := range []string{"X-Ingest-Key", "X-API-Key", "X-MQTT-Token"} {
		if k := strings.TrimSpace(c.GetHeader(hdr)); k != "" {
			return k
		}
	}
	return ""
}

func writeIngestError(c *gin.Context, err error) {
	var httpErr *service.IngestHTTPError
	if errors.As(err, &httpErr) && httpErr != nil {
		code := httpErr.StatusCode
		if code < 400 || code > 599 {
			code = http.StatusBadRequest
		}
		c.JSON(code, response.APIResponse{
			Success: false,
			Message: httpErr.Message,
			Error:   httpErr.Message,
			Data:    httpErr.Details,
		})
		return
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "missing") || strings.Contains(lower, "invalid or disabled"):
		response.Unauthorized(c, err.Error())
	case strings.Contains(lower, "suspended") || strings.Contains(lower, "locked to station"):
		response.Forbidden(c, err.Error())
	case strings.Contains(lower, "not registered"):
		response.NotFound(c, err.Error())
	default:
		response.BadRequest(c, err.Error())
	}
}

// IngestTelemetry handles POST /api/v1/ingest/telemetry and POST /api/ingest/telemetry
// using plain JSON + Authorization: Bearer dlg_live_... token for AQMS, SPARING, and ONLIMO dataloggers.
func (h *IngestHandler) IngestTelemetry(c *gin.Context) {
	rawKey := extractRawIngestKey(c)
	if rawKey == "" {
		response.Unauthorized(c, "Missing Bearer token (provide 'Authorization: Bearer dlg_live_...' header)")
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if err != nil || len(bodyBytes) == 0 {
		response.BadRequest(c, "Empty or unreadable telemetry JSON body")
		return
	}

	res, err := h.ingestService.IngestPayload(rawKey, bodyBytes, c.ClientIP())
	if err != nil {
		writeIngestError(c, err)
		return
	}

	response.Message(c, "Telemetry ingested successfully", res)
}

// IngestMQTT handles POST /api/v1/ingest/mqtt and POST /api/ingest/mqtt for MQTT broker webhooks & bridges
// when the administrator enables MQTT ingestion (mqtt_ingest_enabled = true).
func (h *IngestHandler) IngestMQTT(c *gin.Context) {
	if !config.Get().MQTTIngestEnabled {
		response.Forbidden(c, "MQTT telemetry ingestion module is currently disabled by administrator")
		return
	}

	rawKey := extractRawIngestKey(c)
	if rawKey == "" {
		response.Unauthorized(c, "Missing Bearer token for MQTT ingestion (provide 'Authorization: Bearer dlg_live_...' or 'X-MQTT-Token')")
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if err != nil || len(bodyBytes) == 0 {
		response.BadRequest(c, "Empty or unreadable MQTT telemetry JSON payload")
		return
	}

	res, err := h.ingestService.IngestPayloadWithSource(rawKey, bodyBytes, c.ClientIP(), "MQTT")
	if err != nil {
		writeIngestError(c, err)
		return
	}

	response.Message(c, "MQTT telemetry ingested successfully", res)
}

// GetPayloadTemplate handles GET /api/ingest-keys/payload-template to generate standardized
// AQMS, SPARING, and ONLIMO JSON payloads + Bearer Token cURL commands from manually registered parameters.
func (h *IngestHandler) GetPayloadTemplate(c *gin.Context) {
	scope := middleware.GetScope(c)
	if !scope.IsSuperAdmin && !strings.EqualFold(scope.Role, "ADMIN") {
		if cfg := config.Get(); cfg == nil || !cfg.IsCloudMode() || !strings.EqualFold(scope.Role, "ENGINEER") {
			response.Forbidden(c, "Forbidden: insufficient role privileges for payload template")
			return
		}
	}
	if !scope.IsSuperAdmin && scope.TenantID > 0 {
		if qTid, err := strconv.ParseUint(c.Query("tenant_id"), 10, 32); err == nil && qTid > 0 && uint(qTid) != scope.TenantID {
			response.Forbidden(c, "Forbidden: cross-tenant payload template access denied")
			return
		}
	} else if scope.IsSuperAdmin {
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

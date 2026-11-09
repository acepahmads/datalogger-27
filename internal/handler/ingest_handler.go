package handler

import (
	"io"
	"strings"

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

func extractRawIngestKey(c *gin.Context) string {
	if k := strings.TrimSpace(c.GetHeader("X-Ingest-Key")); k != "" {
		return k
	}
	if k := strings.TrimSpace(c.GetHeader("X-API-Key")); k != "" {
		return k
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		candidate := strings.TrimSpace(authHeader[7:])
		if strings.HasPrefix(candidate, "dlg_") {
			return candidate
		}
	}
	return ""
}

// IngestTelemetry handles POST /api/v1/ingest/telemetry and POST /api/ingest/telemetry for Edge-to-Central push.
func (h *IngestHandler) IngestTelemetry(c *gin.Context) {
	rawKey := extractRawIngestKey(c)
	if rawKey == "" {
		response.Unauthorized(c, "Missing Edge Ingest API key (provide X-Ingest-Key header)")
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if err != nil || len(bodyBytes) == 0 {
		response.BadRequest(c, "Empty or unreadable telemetry JSON body")
		return
	}

	res, err := h.ingestService.IngestPayload(rawKey, bodyBytes, c.ClientIP())
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "missing") || strings.Contains(lower, "invalid or disabled") {
			response.Unauthorized(c, err.Error())
			return
		}
		if strings.Contains(lower, "suspended") || strings.Contains(lower, "kuota") || strings.Contains(lower, "not registered") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, res)
}

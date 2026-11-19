package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/handler"
	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestInboundLogHandler_CloudSuperadminOnlyAndMultiTenantFiltering(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Get()
	origCloud := cfg.CloudMode
	origDomain := cfg.MonitoringDomain
	cfg.CloudMode = true
	cfg.MonitoringDomain = "CLOUD"
	defer func() {
		cfg.CloudMode = origCloud
		cfg.MonitoringDomain = origDomain
	}()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.Tenant{}, &model.InboundIngestLog{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	db.Create(&model.Tenant{ID: 1, Name: "DLH Sukabumi", Code: "dlh-sukabumi", Status: "ACTIVE"})
	db.Create(&model.Tenant{ID: 2, Name: "UJIKON SPARING", Code: "ujikon-sparing", Status: "ACTIVE"})

	repo := repository.NewInboundLogRepository(db)
	h := handler.NewInboundLogHandler(repo)

	now := time.Now().UTC()
	_ = repo.Create(context.Background(), &model.InboundIngestLog{
		ReceivedAt:         now.Add(-2 * time.Minute),
		TenantID:           1,
		SiteID:             10,
		SiteCode:           "aqms-sukabumi",
		DeviceCode:         "cbi-aqms-sukabumi",
		Endpoint:           "/api/v1/ingest/telemetry",
		Source:             "EDGE_PUSH",
		LoggerType:         "AQMS",
		DetectedFormat:     "STANDARD_V1",
		Status:             "ACCEPTED",
		HTTPStatus:         200,
		ClientIP:           "10.0.0.11",
		KeyPrefix:          "dlg_live_abc…",
		RecordsProcessed:   1,
		ParametersIngested: 14,
		PayloadBytes:       320,
		LatencyMs:          9,
		RawPayload:         `{"logger_type":"AQMS","site_code":"aqms-sukabumi","values":{"pm25":34.5}}`,
	})
	_ = repo.Create(context.Background(), &model.InboundIngestLog{
		ReceivedAt:         now.Add(-1 * time.Minute),
		TenantID:           2,
		SiteID:             20,
		SiteCode:           "sparing-ujikon-2026",
		DeviceCode:         "sparing-ujikon-2026",
		Endpoint:           "/api/v1/ingest/telemetry",
		Source:             "EDGE_PUSH",
		LoggerType:         "SPARING",
		DetectedFormat:     "STANDARD_V1",
		Status:             "ACCEPTED",
		HTTPStatus:         200,
		ClientIP:           "10.0.0.22",
		KeyPrefix:          "dlg_live_xyz…",
		RecordsProcessed:   30,
		ParametersIngested: 150,
		PayloadBytes:       4820,
		LatencyMs:          18,
		RawPayload:         `{"logger_type":"SPARING","site_code":"sparing-ujikon-2026","records":[{"timestamp":"2026-11-19T15:00:00+07:00","values":{"ph":7.1}}]}`,
	})
	_ = repo.Create(context.Background(), &model.InboundIngestLog{
		ReceivedAt:   now,
		TenantID:     0,
		SiteCode:     "unknown-site",
		Endpoint:     "/api/v1/ingest/telemetry",
		Source:       "EDGE_PUSH",
		Status:       "REJECTED",
		HTTPStatus:   401,
		ClientIP:     "10.0.0.99",
		ErrorMessage: "Invalid Bearer token",
		PayloadBytes: 42,
		RawPayload:   `{"site_code":"unknown-site"}`,
	})

	buildRouter := func(role string) *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", role)
			c.Next()
		})
		r.GET("/api/superadmin/inbound-logs", h.List)
		r.GET("/api/superadmin/inbound-logs/:id", h.GetByID)
		r.DELETE("/api/superadmin/inbound-logs", h.Purge)
		return r
	}

	// 1. Strict RBAC: non-SUPERADMIN roles MUST receive 403 Forbidden
	for _, forbiddenRole := range []string{"ADMIN", "ENGINEER", "OPERATOR", "CUSTOMER_SERVICE"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/superadmin/inbound-logs", nil)
		buildRouter(forbiddenRole).ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for role %s, got %d: %s", forbiddenRole, w.Code, w.Body.String())
		}
	}

	// 2. SUPERADMIN in Cloud Mode sees ALL tenants + summary + raw_payload
	saRouter := buildRouter("SUPERADMIN")
	wAll := httptest.NewRecorder()
	saRouter.ServeHTTP(wAll, httptest.NewRequest(http.MethodGet, "/api/superadmin/inbound-logs", nil))
	if wAll.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SUPERADMIN in Cloud Mode, got %d: %s", wAll.Code, wAll.Body.String())
	}
	var listResp struct {
		Success bool `json:"success"`
		Data    struct {
			Total   int64                        `json:"total"`
			Items   []model.InboundIngestLog     `json:"items"`
			Summary repository.InboundLogSummary `json:"summary"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wAll.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if listResp.Data.Total != 3 || listResp.Data.Summary.AcceptedCount != 2 || listResp.Data.Summary.RejectedCount != 1 || listResp.Data.Summary.ActiveTenants != 2 {
		t.Fatalf("unexpected summary/total: total=%d summary=%+v", listResp.Data.Total, listResp.Data.Summary)
	}

	// 3. Filtering by tenant_id=2 + logger_type=SPARING returns strictly the SPARING entry with tenant_name resolved
	wFilter := httptest.NewRecorder()
	saRouter.ServeHTTP(wFilter, httptest.NewRequest(http.MethodGet, "/api/superadmin/inbound-logs?tenant_id=2&logger_type=SPARING&status=ACCEPTED", nil))
	if wFilter.Code != http.StatusOK {
		t.Fatalf("filtered query failed (%d): %s", wFilter.Code, wFilter.Body.String())
	}
	var filterResp struct {
		Data struct {
			Total int64                    `json:"total"`
			Items []model.InboundIngestLog `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wFilter.Body.Bytes(), &filterResp)
	if filterResp.Data.Total != 1 || len(filterResp.Data.Items) != 1 {
		t.Fatalf("expected 1 filtered SPARING log, got %d", filterResp.Data.Total)
	}
	if filterResp.Data.Items[0].TenantName != "UJIKON SPARING" || filterResp.Data.Items[0].RecordsProcessed != 30 {
		t.Fatalf("unexpected filtered item metadata: %+v", filterResp.Data.Items[0])
	}

	// 4. Edge Mode MUST reject access even for SUPERADMIN
	cfg.CloudMode = false
	cfg.MonitoringDomain = "AQMS"
	wEdge := httptest.NewRecorder()
	saRouter.ServeHTTP(wEdge, httptest.NewRequest(http.MethodGet, "/api/superadmin/inbound-logs", nil))
	if wEdge.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden in Edge Mode, got %d", wEdge.Code)
	}

	// 5. Restore Cloud Mode and verify DELETE purge all works for SUPERADMIN
	cfg.CloudMode = true
	cfg.MonitoringDomain = "CLOUD"
	wPurge := httptest.NewRecorder()
	saRouter.ServeHTTP(wPurge, httptest.NewRequest(http.MethodDelete, "/api/superadmin/inbound-logs?all=1", nil))
	if wPurge.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for purge, got %d: %s", wPurge.Code, wPurge.Body.String())
	}
}

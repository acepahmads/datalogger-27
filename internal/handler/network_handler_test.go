package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/handler"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupNetworkTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, service.OutboundService) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:memnet_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = db.AutoMigrate(
		&model.SystemLog{},
		&model.Device{},
		&model.Parameter{},
		&model.OutputDestination{},
		&model.DestinationPayloadConfig{},
		&model.OutboundDeliveryQueue{},
		&model.DeliveryLog{},
	)

	cfg := &config.Config{Environment: "test"}
	systemRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)

	deviceRepo := repository.NewDeviceRepository(db)
	outboundRepo := repository.NewOutboundRepository(db)
	outboundService := service.NewOutboundService(outboundRepo, deviceRepo, nil)

	systemHandler := handler.NewSystemHandler(systemService)
	systemHandler.SetOutboundService(outboundService)

	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/system/network", systemHandler.GetNetworkStatus)
	}

	return r, db, outboundService
}

func TestGetNetworkStatus_StructureAndSanitization(t *testing.T) {
	r, db, outboundService := setupNetworkTestRouter(t)
	ctx := context.Background()

	// Seed 2 destinations:
	// 1. Enabled REST with basic auth + token in URL, status ERROR, 0 success, 10 fails -> should map to FAILED
	dest1 := &model.OutputDestination{
		Name:              "KLHK Server",
		Type:              model.DestREST,
		Endpoint:          "https://admin:supersecret@ispu.menlhk.go.id/api/v1?token=token12345&site_id=JKT01",
		Enabled:           true,
		Status:            "ERROR",
		SuccessDeliveries: 0,
		FailedDeliveries:  10,
		AuthType:          "BEARER_TOKEN",
		AuthToken:         "secret_token_never_leak",
	}
	if err := outboundService.CreateDestination(ctx, dest1); err != nil {
		t.Fatalf("Failed to create dest1: %v", err)
	}

	// Seed a queue item with error for dest1
	now := time.Now()
	_ = db.Create(&model.OutboundDeliveryQueue{
		DestinationID:   dest1.ID,
		BatchID:         "BATCH-TEST-1",
		IdempotencyKey:  "IDEMP-TEST-1",
		PayloadSnapshot: "{}",
		Status:          model.DeliveryFailed,
		Attempts:        5,
		MaxRetries:      5,
		LastAttemptAt:   &now,
		LastError:       "Post request failed with secret_key=TOPSECRET999",
	})

	// 2. Disabled MQTT destination
	dest2 := &model.OutputDestination{
		Name:              "Internal SCADA Broker",
		Type:              model.DestMQTT,
		Endpoint:          "tcp://192.168.1.100:1883",
		Enabled:           false,
		Status:            "IDLE",
		SuccessDeliveries: 50,
		FailedDeliveries:  0,
	}
	if err := outboundService.CreateDestination(ctx, dest2); err != nil {
		t.Fatalf("Failed to create dest2: %v", err)
	}

	// Make request to GET /api/system/network
	req := httptest.NewRequest("GET", "/api/system/network", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			LocalNetwork struct {
				Status           string `json:"status"`
				PrimaryIP        string `json:"primary_ip"`
				Gateway          string `json:"gateway"`
				GatewayReachable bool   `json:"gateway_reachable"`
			} `json:"local_network"`
			Internet struct {
				Status      string  `json:"status"`
				Connected   bool    `json:"connected"`
				LatencyMs   float64 `json:"latency_ms"`
				ProbeTarget string  `json:"probe_target"`
			} `json:"internet"`
			Outbound struct {
				TotalLinks    int    `json:"total_links"`
				HealthyCount  int    `json:"healthy_count"`
				FailedCount   int    `json:"failed_count"`
				DisabledCount int    `json:"disabled_count"`
				OverallStatus string `json:"overall_status"`
				Destinations  []struct {
					ID          uint    `json:"id"`
					Name        string  `json:"name"`
					Type        string  `json:"type"`
					Endpoint    string  `json:"endpoint"`
					Status      string  `json:"status"`
					Enabled     bool    `json:"enabled"`
					LastError   string  `json:"last_error"`
					SuccessRate float64 `json:"success_rate"`
					QueueCount  int64   `json:"queue_count"`
				} `json:"destinations"`
			} `json:"outbound"`
		} `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	// Verify local network
	if resp.Data.LocalNetwork.Status != "UP" && resp.Data.LocalNetwork.Status != "DOWN" {
		t.Errorf("Unexpected local network status: %s", resp.Data.LocalNetwork.Status)
	}

	// Verify outbound destinations
	outbound := resp.Data.Outbound
	if outbound.TotalLinks != 2 {
		t.Errorf("Expected 2 total outbound links, got %d", outbound.TotalLinks)
	}
	if outbound.FailedCount != 1 {
		t.Errorf("Expected 1 failed link, got %d", outbound.FailedCount)
	}
	if outbound.DisabledCount != 1 {
		t.Errorf("Expected 1 disabled link, got %d", outbound.DisabledCount)
	}
	if outbound.OverallStatus != "FAILED" {
		t.Errorf("Expected overall outbound status FAILED, got %s", outbound.OverallStatus)
	}

	// Verify dest1 sanitization & status
	var d1Found, d2Found bool
	for _, d := range outbound.Destinations {
		if d.ID == dest1.ID {
			d1Found = true
			if d.Status != "FAILED" {
				t.Errorf("Expected dest1 status FAILED, got %s", d.Status)
			}
			// Must NOT contain raw password or token in endpoint
			if strings.Contains(d.Endpoint, "supersecret") {
				t.Errorf("Credential leaked in endpoint: %s", d.Endpoint)
			}
			if strings.Contains(d.Endpoint, "token12345") {
				t.Errorf("Token leaked in endpoint query: %s", d.Endpoint)
			}
			// Must NOT leak secret_token_never_leak or secret_key
			if strings.Contains(w.Body.String(), "secret_token_never_leak") {
				t.Errorf("AuthToken leaked in response!")
			}
			if strings.Contains(d.LastError, "TOPSECRET999") {
				t.Errorf("Secret leaked in last_error: %s", d.LastError)
			}
			if d.QueueCount < 1 {
				t.Errorf("Expected queue count >= 1 for dest1, got %d", d.QueueCount)
			}
		}
		if d.ID == dest2.ID {
			d2Found = true
			if d.Status != "DISABLED" {
				t.Errorf("Expected dest2 status DISABLED, got %s", d.Status)
			}
		}
	}

	if !d1Found || !d2Found {
		t.Errorf("Expected both destinations in response, found d1=%v, d2=%v", d1Found, d2Found)
	}
}

func TestGetNetworkStatus_CacheFastResponse(t *testing.T) {
	r, _, _ := setupNetworkTestRouter(t)

	// First call populates cache
	req1 := httptest.NewRequest("GET", "/api/system/network", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("First request failed: %d", w1.Code)
	}

	// Second immediate call must be served directly from in-memory cache in < 15ms
	start := time.Now()
	req2 := httptest.NewRequest("GET", "/api/system/network", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	elapsed := time.Since(start)

	if w2.Code != http.StatusOK {
		t.Fatalf("Second request failed: %d", w2.Code)
	}
	if elapsed > 30*time.Millisecond {
		t.Errorf("Cached response took too long: %v (expected < 30ms)", elapsed)
	}
}


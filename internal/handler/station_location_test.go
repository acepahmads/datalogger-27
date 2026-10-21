package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupStationLocationTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *service.SystemService) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:memloc_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = db.AutoMigrate(
		&model.StationLocation{},
		&model.AuditTrail{},
		&model.SystemLog{},
		&model.User{},
		&model.Role{},
	)

	cfg := &config.Config{Environment: "test"}
	systemRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)
	systemHandler := handler.NewSystemHandler(systemService)

	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/system/location", systemHandler.GetStationLocation)
		api.GET("/system/location/ip-estimate", systemHandler.GetIPGeolocationEstimate)

		// Simulated authenticated endpoint with context
		api.PUT("/system/location/direct", func(c *gin.Context) {
			c.Set("userID", uint(42))
			c.Set("username", "admin_operator")
			systemHandler.SaveStationLocation(c)
		})

		// Protected endpoint using RequirePermission
		api.PUT("/system/location/protected", middleware.RequirePermission("config.manage"), systemHandler.SaveStationLocation)
	}

	return r, db, systemService
}

func TestStationLocation_DefaultUnconfigured(t *testing.T) {
	router, _, _ := setupStationLocationTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/system/location", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data model.StationLocation `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Data.IsConfigured {
		t.Errorf("Expected IsConfigured=false by default, got true")
	}
	if resp.Data.LocationSource != model.LocationSourceUnconfigured {
		t.Errorf("Expected LocationSource=%s, got %s", model.LocationSourceUnconfigured, resp.Data.LocationSource)
	}
	if resp.Data.Latitude != nil || resp.Data.Longitude != nil {
		t.Errorf("Expected nil coordinates by default, got lat=%v, lng=%v", resp.Data.Latitude, resp.Data.Longitude)
	}
}

func TestStationLocation_ManualUpdateAndAuditTrail(t *testing.T) {
	router, db, _ := setupStationLocationTestRouter(t)

	lat := -6.2088
	lng := 106.8456
	stationName := "Central AQMS Jakarta"
	address := "Jl. Medan Merdeka Selatan, Gambir"
	source := "MANUAL"

	payload := service.UpdateStationLocationRequest{
		StationName:    &stationName,
		Latitude:       &lat,
		Longitude:      &lng,
		LocationSource: source,
		Address:        &address,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/api/system/location/direct", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data model.StationLocation `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Data.IsConfigured {
		t.Errorf("Expected IsConfigured=true after update")
	}
	if resp.Data.Latitude == nil || *resp.Data.Latitude != lat {
		t.Errorf("Expected latitude %v, got %v", lat, resp.Data.Latitude)
	}
	if resp.Data.Longitude == nil || *resp.Data.Longitude != lng {
		t.Errorf("Expected longitude %v, got %v", lng, resp.Data.Longitude)
	}
	if resp.Data.LocationSource != "MANUAL" {
		t.Errorf("Expected LocationSource=MANUAL, got %s", resp.Data.LocationSource)
	}
	if resp.Data.StationName != stationName {
		t.Errorf("Expected StationName=%s, got %s", stationName, resp.Data.StationName)
	}
	if resp.Data.UpdatedBy != "admin_operator" {
		t.Errorf("Expected UpdatedBy=admin_operator, got %s", resp.Data.UpdatedBy)
	}

	// Verify Audit Trail persistence
	var audit model.AuditTrail
	err := db.Where("action = ?", "STATION_LOCATION_UPDATE").First(&audit).Error
	if err != nil {
		t.Fatalf("Audit trail record not found: %v", err)
	}
	if audit.Username != "admin_operator" {
		t.Errorf("Expected audit username 'admin_operator', got '%s'", audit.Username)
	}
	if audit.Resource != "STATION_LOCATION" {
		t.Errorf("Expected audit resource 'STATION_LOCATION', got '%s'", audit.Resource)
	}
}

func TestStationLocation_CoordinateValidation(t *testing.T) {
	router, _, _ := setupStationLocationTestRouter(t)

	testCases := []struct {
		name      string
		lat       *float64
		lng       *float64
		source    string
		acc       *float64
		expectErr bool
	}{
		{
			name:      "Lat too high (>90)",
			lat:       ptrFloat(90.1),
			lng:       ptrFloat(100.0),
			source:    "MANUAL",
			expectErr: true,
		},
		{
			name:      "Lat too low (<-90)",
			lat:       ptrFloat(-90.5),
			lng:       ptrFloat(100.0),
			source:    "MANUAL",
			expectErr: true,
		},
		{
			name:      "Lng too high (>180)",
			lat:       ptrFloat(0.0),
			lng:       ptrFloat(180.1),
			source:    "MANUAL",
			expectErr: true,
		},
		{
			name:      "Lng too low (<-180)",
			lat:       ptrFloat(0.0),
			lng:       ptrFloat(-180.5),
			source:    "MANUAL",
			expectErr: true,
		},
		{
			name:      "Negative accuracy",
			lat:       ptrFloat(10.0),
			lng:       ptrFloat(20.0),
			source:    "GPS",
			acc:       ptrFloat(-1.5),
			expectErr: true,
		},
		{
			name:      "Invalid source",
			lat:       ptrFloat(10.0),
			lng:       ptrFloat(20.0),
			source:    "SATELLITE_ALIEN",
			expectErr: true,
		},
		{
			name:      "Valid GPS coordinates and accuracy",
			lat:       ptrFloat(-7.250445),
			lng:       ptrFloat(112.768845),
			source:    "GPS",
			acc:       ptrFloat(12.5),
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := service.UpdateStationLocationRequest{
				Latitude:       tc.lat,
				Longitude:      tc.lng,
				LocationSource: tc.source,
				AccuracyMeters: tc.acc,
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPut, "/api/system/location/direct", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tc.expectErr && w.Code != http.StatusBadRequest {
				t.Errorf("Expected 400 Bad Request for %s, got %d: %s", tc.name, w.Code, w.Body.String())
			}
			if !tc.expectErr && w.Code != http.StatusOK {
				t.Errorf("Expected 200 OK for %s, got %d: %s", tc.name, w.Code, w.Body.String())
			}
		})
	}
}

func TestStationLocation_ResetToUnconfigured(t *testing.T) {
	router, _, _ := setupStationLocationTestRouter(t)

	// Set valid location first
	lat := -6.1754
	lng := 106.8272
	savePayload := service.UpdateStationLocationRequest{
		Latitude:       &lat,
		Longitude:      &lng,
		LocationSource: "MANUAL",
	}
	body, _ := json.Marshal(savePayload)
	req1 := httptest.NewRequest(http.MethodPut, "/api/system/location/direct", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("Setup location failed: %s", w1.Body.String())
	}

	// Now trigger reset
	resetPayload := service.UpdateStationLocationRequest{
		Reset: true,
	}
	resetBody, _ := json.Marshal(resetPayload)
	req2 := httptest.NewRequest(http.MethodPut, "/api/system/location/direct", bytes.NewBuffer(resetBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Reset request failed: %s", w2.Body.String())
	}

	var resp struct {
		Data model.StationLocation `json:"data"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &resp)

	if resp.Data.IsConfigured {
		t.Errorf("Expected IsConfigured=false after reset, got true")
	}
	if resp.Data.Latitude != nil || resp.Data.Longitude != nil {
		t.Errorf("Expected coordinates to be nil after reset, got lat=%v, lng=%v", resp.Data.Latitude, resp.Data.Longitude)
	}
	if resp.Data.LocationSource != model.LocationSourceUnconfigured {
		t.Errorf("Expected LocationSource=UNCONFIGURED after reset, got %s", resp.Data.LocationSource)
	}
}

func TestStationLocation_IPGeolocationEstimateContract(t *testing.T) {
	router, _, _ := setupStationLocationTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/system/location/ip-estimate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["available"] != false {
		t.Errorf("Expected available=false without configured provider, got %v", resp["available"])
	}
	if resp["coordinates"] != nil {
		t.Errorf("Expected coordinates=nil, no fake coordinates allowed, got %v", resp["coordinates"])
	}
	if resp["status"] != "UNAVAILABLE" {
		t.Errorf("Expected status=UNAVAILABLE, got %v", resp["status"])
	}
}

func TestStationLocation_UnauthorizedCannotModify(t *testing.T) {
	router, _, _ := setupStationLocationTestRouter(t)

	lat := -6.2
	lng := 106.8
	payload := service.UpdateStationLocationRequest{
		Latitude:  &lat,
		Longitude: &lng,
	}
	body, _ := json.Marshal(payload)

	// Call protected endpoint without authentication context -> RequirePermission should return 401 Unauthorized
	req := httptest.NewRequest(http.MethodPut, "/api/system/location/protected", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Errorf("Expected 401 Unauthorized or 403 Forbidden for unauthenticated location update, got %d: %s", w.Code, w.Body.String())
	}
}

func ptrFloat(f float64) *float64 {
	return &f
}

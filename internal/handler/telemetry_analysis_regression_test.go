package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func floatPtr(f float64) *float64 {
	return &f
}

// TestUnifiedTelemetryAnalysisMultiParameterRegression verifies the data integrity requirements
// specified in BUGFIX — UNIFIED TELEMETRY ANALYSIS DATA INTEGRITY & MULTI-PARAMETER CHARTS.
func TestUnifiedTelemetryAnalysisMultiParameterRegression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_telem_reg_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = database.RunMigrations(db)

	cfg := config.DefaultConfig()
	sysRepo := repository.NewSystemRepository(db)
	authService := service.NewAuthService(sysRepo, cfg)
	telemetryRepo := repository.NewTelemetryRepository(db)
	telemetryService := service.NewTelemetryService(telemetryRepo, nil, service.DefaultTelemetryConfig())
	telemetryHandler := handler.NewTelemetryHandler(telemetryService)

	// 1. Seed Roles & Auth
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminRole := model.Role{Name: "Administrator", Description: "Full access"}
	db.Create(&adminRole)
	adminUser := model.User{
		Username: "admin_reg",
		Password: string(hash),
		RoleID:   adminRole.ID,
		IsActive: true,
	}
	db.Create(&adminUser)
	token, _, _ := authService.Login("admin_reg", "password123", "127.0.0.1", "")

	// 2. Seed 2 Devices
	dev1 := model.Device{
		DeviceCode:       "DEV-BOILER-01",
		DeviceName:       "Main Steam Boiler",
		DeviceType:       "MODBUS_TCP",
		Status:           model.DeviceStatusActive,
		ConnectionStatus: model.DeviceConnOnline,
	}
	dev2 := model.Device{
		DeviceCode:       "DEV-CHILLER-02",
		DeviceName:       "Process Water Chiller",
		DeviceType:       "MODBUS_TCP",
		Status:           model.DeviceStatusActive,
		ConnectionStatus: model.DeviceConnOnline,
	}
	db.Create(&dev1)
	db.Create(&dev2)

	// 3. Seed at least three parameters with different ranges and units:
	// - Param 1: Temperature (°C, range ~20-60)
	// - Param 2: Pressure (hPa, range ~900-1050)
	// - Param 3: Flow Rate (m³/h, range ~0.5-15.0)
	paramTemp := model.Parameter{
		DeviceID:      dev1.ID,
		ParameterCode: "TEMP_01",
		ParameterName: "Boiler Temperature",
		DataType:      model.DataTypeFloat32,
		Unit:          "°C",
		MinValue:      floatPtr(0.0),
		MaxValue:      floatPtr(150.0),
		Precision:     2,
		Enabled:       true,
	}
	paramPress := model.Parameter{
		DeviceID:      dev1.ID,
		ParameterCode: "PRESS_01",
		ParameterName: "Chamber Pressure",
		DataType:      model.DataTypeFloat32,
		Unit:          "hPa",
		MinValue:      floatPtr(800.0),
		MaxValue:      floatPtr(1200.0),
		Precision:     2,
		Enabled:       true,
	}
	paramFlow := model.Parameter{
		DeviceID:      dev2.ID,
		ParameterCode: "FLOW_02",
		ParameterName: "Chiller Coolant Flow",
		DataType:      model.DataTypeFloat32,
		Unit:          "m³/h",
		MinValue:      floatPtr(0.0),
		MaxValue:      floatPtr(20.0),
		Precision:     2,
		Enabled:       true,
	}
	db.Create(&paramTemp)
	db.Create(&paramPress)
	db.Create(&paramFlow)

	// 4. Seed records with duplicate timestamps, bad-quality records, and staggered timestamps
	tBase := time.Now().Add(-2 * time.Hour).Truncate(time.Second)

	records := []model.RawData{
		// T0: Duplicate timestamp: Both Temp (47.31 °C) and Press (933.09 hPa) recorded at exact same timestamp
		{
			DeviceID:       dev1.ID,
			ParameterID:    paramTemp.ID,
			Value:          47.31,
			ProcessedValue: 47.31,
			RawValue:       4731,
			RawHex:         "0x127B",
			Quality:        "GOOD",
			QualityReason:  "NONE",
			Timestamp:      tBase,
			ReceivedAt:     tBase,
		},
		{
			DeviceID:       dev1.ID,
			ParameterID:    paramPress.ID,
			Value:          933.09,
			ProcessedValue: 933.09,
			RawValue:       93309,
			RawHex:         "0x6C5D",
			Quality:        "GOOD",
			QualityReason:  "NONE",
			Timestamp:      tBase, // Duplicate timestamp!
			ReceivedAt:     tBase,
		},
		// T1: Flow rate on Device 2 (4.85 m³/h)
		{
			DeviceID:       dev2.ID,
			ParameterID:    paramFlow.ID,
			Value:          4.85,
			ProcessedValue: 4.85,
			RawValue:       485,
			RawHex:         "0x01E5",
			Quality:        "GOOD",
			QualityReason:  "NONE",
			Timestamp:      tBase.Add(1 * time.Minute),
			ReceivedAt:     tBase.Add(1 * time.Minute),
		},
		// T2: Second Temp point (35.84 °C)
		{
			DeviceID:       dev1.ID,
			ParameterID:    paramTemp.ID,
			Value:          35.84,
			ProcessedValue: 35.84,
			RawValue:       3584,
			RawHex:         "0x0E00",
			Quality:        "GOOD",
			QualityReason:  "NONE",
			Timestamp:      tBase.Add(2 * time.Minute),
			ReceivedAt:     tBase.Add(2 * time.Minute),
		},
		// T3: BAD quality record on Pressure (1150.00 hPa with OUT_OF_BOUNDS)
		{
			DeviceID:       dev1.ID,
			ParameterID:    paramPress.ID,
			Value:          1150.00,
			ProcessedValue: 1150.00,
			RawValue:       115000,
			RawHex:         "0x8CA0",
			Quality:        "BAD",
			QualityReason:  "OUT_OF_BOUNDS",
			Timestamp:      tBase.Add(3 * time.Minute),
			ReceivedAt:     tBase.Add(3 * time.Minute),
		},
		// T4: Second Flow point (6.20 m³/h)
		{
			DeviceID:       dev2.ID,
			ParameterID:    paramFlow.ID,
			Value:          6.20,
			ProcessedValue: 6.20,
			RawValue:       620,
			RawHex:         "0x026C",
			Quality:        "GOOD",
			QualityReason:  "NONE",
			Timestamp:      tBase.Add(4 * time.Minute),
			ReceivedAt:     tBase.Add(4 * time.Minute),
		},
	}
	for i := range records {
		db.Create(&records[i])
	}

	// 5. Setup Router
	r := gin.New()
	api := r.Group("/api")
	api.Use(middleware.JWTAuth(authService))
	{
		api.GET("/telemetry/history", telemetryHandler.GetAllHistorical)
		api.GET("/telemetry/raw", telemetryHandler.GetAllRawTelemetry)
		api.GET("/devices/:id/telemetry/history", telemetryHandler.GetHistorical)
		api.GET("/devices/:id/telemetry/raw", telemetryHandler.GetRawTelemetry)
	}

	// --- VERIFICATION 1: All Devices + All Parameters query returns preloaded relationships ---
	t.Run("All Devices + All Parameters Query returns preloaded Device and Parameter metadata", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/telemetry/history?page=1&page_size=50", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data struct {
				Total int             `json:"total"`
				Items []model.RawData `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if resp.Data.Total != 6 {
			t.Fatalf("Expected 6 total records, got %d", resp.Data.Total)
		}

		// Verify every record contains stable preloaded Device and Parameter
		for _, item := range resp.Data.Items {
			if item.Device == nil {
				t.Errorf("Record ID %d missing preloaded Device relation", item.ID)
			} else if item.Device.DeviceCode == "" {
				t.Errorf("Record ID %d preloaded Device missing DeviceCode", item.ID)
			}
			if item.Parameter == nil {
				t.Errorf("Record ID %d missing preloaded Parameter relation", item.ID)
			} else if item.Parameter.Unit == "" {
				t.Errorf("Record ID %d preloaded Parameter missing Unit", item.ID)
			}
		}
	})

	// --- VERIFICATION 2: Raw Telemetry preserves unscaled registers and hex stream ---
	t.Run("Raw Telemetry endpoint preserves raw registers, hex stream, and protocol types", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/telemetry/raw?page=1&page_size=50", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data struct {
				Items []model.RawData `json:"items"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		hasHex := false
		for _, item := range resp.Data.Items {
			if item.RawHex != "" {
				hasHex = true
			}
		}
		if !hasHex {
			t.Errorf("Expected RawHex to be preserved in raw telemetry mode")
		}
	})

	// --- VERIFICATION 3: Single Parameter Filtering ---
	t.Run("Single Parameter Filtering isolates exactly that parameter's records", func(t *testing.T) {
		url := fmt.Sprintf("/api/telemetry/history?parameter_id=%d", paramTemp.ID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data struct {
				Total int             `json:"total"`
				Items []model.RawData `json:"items"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Data.Total != 2 {
			t.Fatalf("Expected 2 Temperature records, got %d", resp.Data.Total)
		}
		for _, item := range resp.Data.Items {
			if item.ParameterID != paramTemp.ID {
				t.Errorf("Expected only parameter %d, got %d", paramTemp.ID, item.ParameterID)
			}
		}
	})

	// --- VERIFICATION 4: Multi-series separation & per-series statistics simulation ---
	t.Run("Series Grouping algorithm cleanly isolates incompatible parameters and units", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/telemetry/history?page=1&page_size=50", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Data struct {
				Items []model.RawData `json:"items"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		// Simulate the series grouping applied by the frontend
		seriesMap := make(map[string][]model.RawData)
		for _, r := range resp.Data.Items {
			key := fmt.Sprintf("%d_%d", r.DeviceID, r.ParameterID)
			seriesMap[key] = append(seriesMap[key], r)
		}

		// Expected 3 distinct series:
		// 1. DEV-BOILER-01 : TEMP_01 (°C)
		// 2. DEV-BOILER-01 : PRESS_01 (hPa)
		// 3. DEV-CHILLER-02 : FLOW_02 (m³/h)
		if len(seriesMap) != 3 {
			t.Fatalf("Expected exactly 3 distinct series, got %d", len(seriesMap))
		}

		// Calculate per-series statistics independently
		for key, items := range seriesMap {
			var minVal, maxVal, sumVal float64
			minVal = items[0].ProcessedValue
			maxVal = items[0].ProcessedValue
			goodCount := 0

			for _, item := range items {
				v := item.ProcessedValue
				if v < minVal {
					minVal = v
				}
				if v > maxVal {
					maxVal = v
				}
				sumVal += v
				if item.Quality == "GOOD" {
					goodCount++
				}
			}
			avgVal := sumVal / float64(len(items))

			paramUnit := items[0].Parameter.Unit

			// Verify that no combined calculation mixed °C and hPa!
			if paramUnit == "°C" {
				if minVal < 30 || maxVal > 60 {
					t.Errorf("Temperature series %s scale contaminated: min=%.2f, max=%.2f", key, minVal, maxVal)
				}
			} else if paramUnit == "hPa" {
				if minVal < 900 || maxVal > 1200 {
					t.Errorf("Pressure series %s scale contaminated: min=%.2f, max=%.2f", key, minVal, maxVal)
				}
				if goodCount != 1 {
					t.Errorf("Expected 1 GOOD record and 1 BAD record for pressure, got %d GOOD", goodCount)
				}
			} else if paramUnit == "m³/h" {
				if minVal < 0 || maxVal > 10 {
					t.Errorf("Flow series %s scale contaminated: min=%.2f, max=%.2f", key, minVal, maxVal)
				}
			}

			t.Logf("Series %s (%s): count=%d, min=%.2f, max=%.2f, avg=%.2f, good=%d",
				key, paramUnit, len(items), minVal, maxVal, avgVal, goodCount)
		}
	})
}

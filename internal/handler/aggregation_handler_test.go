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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type AggregationTestEnv struct {
	Router        *gin.Engine
	DB            *gorm.DB
	AggService    *service.AggregationService
	AdminToken    string
	OperatorToken string
}

func setupAggregationTestRouter(t *testing.T) *AggregationTestEnv {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:mem_agg_handler_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = db.AutoMigrate(
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.RawData{},
		&model.AggregationDefinition{},
		&model.AggregationResult{},
		&model.AuditTrail{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
	)

	// Auth and Permissions setup
	permManage := model.Permission{Name: "device.manage", Description: "Manage Devices & Definitions"}
	permView := model.Permission{Name: "device.view", Description: "View Devices & Aggregations"}
	db.Create(&permManage)
	db.Create(&permView)

	adminRole := model.Role{Name: "Administrator", Description: "Full access"}
	opRole := model.Role{Name: "Operator", Description: "Read access"}
	db.Create(&adminRole)
	db.Create(&opRole)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminUser := model.User{
		Username: "admin_agg",
		Email:    "admin_agg@test.local",
		Password: string(hash),
		RoleID:   adminRole.ID,
		IsActive: true,
	}
	opUser := model.User{
		Username: "op_agg",
		Email:    "op_agg@test.local",
		Password: string(hash),
		RoleID:   opRole.ID,
		IsActive: true,
	}
	if err := db.Create(&adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	if err := db.Create(&opUser).Error; err != nil {
		t.Fatalf("Failed to create op user: %v", err)
	}

	cfg := &config.Config{JWTSecret: "test-jwt-secret-key-12345", JWTExpirationHours: 24}
	sysRepo := repository.NewSystemRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	aggRepo := repository.NewAggregationRepository(db)

	authService := service.NewAuthService(sysRepo, cfg)
	aggService := service.NewAggregationService(aggRepo, devRepo, sysRepo, nil)
	aggHandler := handler.NewAggregationHandler(aggService)

	adminToken, _, err1 := authService.Login("admin_agg", "password123", "127.0.0.1", "")
	if err1 != nil {
		t.Fatalf("Admin login failed: %v", err1)
	}
	opToken, _, err2 := authService.Login("op_agg", "password123", "127.0.0.1", "")
	if err2 != nil {
		t.Fatalf("Operator login failed: %v", err2)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		devicesGroup := api.Group("/devices")
		devicesGroup.Use(middleware.JWTAuth(authService))
		{
			devicesGroup.GET("/:id/telemetry/downsampled", middleware.RequirePermission("device.view"), aggHandler.GetDownsampledHistory)
		}

		aggregationsGroup := api.Group("/aggregations")
		aggregationsGroup.Use(middleware.JWTAuth(authService))
		{
			aggregationsGroup.GET("/definitions", middleware.RequirePermission("device.view"), aggHandler.ListDefinitions)
			aggregationsGroup.GET("/definitions/:id", middleware.RequirePermission("device.view"), aggHandler.GetDefinition)
			aggregationsGroup.POST("/definitions", middleware.RequirePermission("device.manage"), aggHandler.CreateDefinition)
			aggregationsGroup.PUT("/definitions/:id", middleware.RequirePermission("device.manage"), aggHandler.UpdateDefinition)
			aggregationsGroup.DELETE("/definitions/:id", middleware.RequirePermission("device.manage"), aggHandler.DeleteDefinition)
			aggregationsGroup.POST("/definitions/:id/run", middleware.RequirePermission("device.manage"), aggHandler.TriggerBucket)
			aggregationsGroup.POST("/run", middleware.RequirePermission("device.manage"), aggHandler.RunBuckets)
			aggregationsGroup.GET("/results", middleware.RequirePermission("device.view"), aggHandler.GetResults)
			aggregationsGroup.GET("/results/:id/samples", middleware.RequirePermission("device.view"), aggHandler.GetResultSamples)
			aggregationsGroup.GET("/samples", middleware.RequirePermission("device.view"), aggHandler.GetResultSamples)
		}

		customerGroup := api.Group("/customer")
		customerGroup.Use(middleware.JWTAuth(authService))
		{
			customerGroup.GET("/aggregated-data/:identifier", middleware.RequirePermission("device.view"), aggHandler.GetCustomerAggregatedData)
		}
	}

	return &AggregationTestEnv{
		Router:        r,
		DB:            db,
		AggService:    aggService,
		AdminToken:    adminToken,
		OperatorToken: opToken,
	}
}

func TestAggregationAPIEndpoints(t *testing.T) {
	env := setupAggregationTestRouter(t)

	// Seed Device & Parameter
	dev := model.Device{ID: 1, DeviceCode: "AQMS-01", DeviceName: "Air Quality 01"}
	param := model.Parameter{ID: 10, DeviceID: 1, ParameterCode: "temperature", ParameterName: "Temperature", Unit: "°C"}
	env.DB.Create(&dev)
	env.DB.Create(&param)

	// 1. Create Definition (POST /api/aggregations/definitions) - Admin authorized
	createReq := service.CreateAggregationDefinitionRequest{
		Name:            "Temp 5m Customer Avg",
		Code:            "TEMP_5M_CUST_AVG",
		DeviceID:        1,
		ParameterID:     10,
		SourceType:      model.SourceCustomerProcessed,
		Function:        model.FunctionAvg,
		IntervalSeconds: 300,
		Timezone:        "UTC",
	}
	bodyBytes, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/aggregations/definitions", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for definition creation, got %d: %s", w.Code, w.Body.String())
	}

	var createdResp struct {
		Data model.AggregationDefinition `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createdResp)
	defID := createdResp.Data.ID

	// 2. Operator RBAC check (POST without device.manage permission -> 403 Forbidden)
	reqForbidden, _ := http.NewRequest("POST", "/api/aggregations/definitions", bytes.NewReader(bodyBytes))
	reqForbidden.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	reqForbidden.Header.Set("Content-Type", "application/json")
	wForbidden := httptest.NewRecorder()
	env.Router.ServeHTTP(wForbidden, reqForbidden)

	if wForbidden.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for operator creating definition, got %d", wForbidden.Code)
	}

	// 3. List Definitions (GET /api/aggregations/definitions) - Operator authorized (device.view)
	reqList, _ := http.NewRequest("GET", "/api/aggregations/definitions?device_id=1", nil)
	reqList.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wList := httptest.NewRecorder()
	env.Router.ServeHTTP(wList, reqList)

	if wList.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list definitions, got %d", wList.Code)
	}

	// Seed raw telemetry samples for bucket
	bucketStart := time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)
	env.DB.Create(&model.RawData{
		DeviceID:       1,
		ParameterID:    10,
		RawValue:       258.0,
		ProcessedValue: 25.8,
		ReceivedAt:     bucketStart.Add(2 * time.Minute),
		Quality:        model.QualityGood,
	})

	// 4. Trigger Bucket Calculation (POST /api/aggregations/definitions/:id/run)
	triggerBody, _ := json.Marshal(map[string]interface{}{"timestamp": bucketStart})
	reqRun, _ := http.NewRequest("POST", fmt.Sprintf("/api/aggregations/definitions/%d/run", defID), bytes.NewReader(triggerBody))
	reqRun.Header.Set("Authorization", "Bearer "+env.AdminToken)
	reqRun.Header.Set("Content-Type", "application/json")
	wRun := httptest.NewRecorder()
	env.Router.ServeHTTP(wRun, reqRun)

	if wRun.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for trigger bucket, got %d: %s", wRun.Code, wRun.Body.String())
	}

	// 5. Inspect Results (GET /api/aggregations/results)
	reqResults, _ := http.NewRequest("GET", "/api/aggregations/results?device_id=1&source_type=CUSTOMER_PROCESSED", nil)
	reqResults.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wResults := httptest.NewRecorder()
	env.Router.ServeHTTP(wResults, reqResults)

	if wResults.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for get results, got %d", wResults.Code)
	}

	// 6. Customer API (GET /api/customer/aggregated-data/:identifier)
	identifier := "20261008140500"
	reqCust, _ := http.NewRequest("GET", "/api/customer/aggregated-data/"+identifier, nil)
	reqCust.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wCust := httptest.NewRecorder()
	env.Router.ServeHTTP(wCust, reqCust)

	if wCust.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for customer aggregated data, got %d: %s", wCust.Code, wCust.Body.String())
	}

	var custResp struct {
		Data service.CustomerAggregatedDataResponse `json:"data"`
	}
	_ = json.Unmarshal(wCust.Body.Bytes(), &custResp)
	if len(custResp.Data.Data) == 0 {
		t.Fatalf("Customer API returned 0 items")
	}
	if *custResp.Data.Data[0].Value != 25.8 {
		t.Errorf("Customer API value mismatch: expected 25.8, got %v", *custResp.Data.Data[0].Value)
	}

	// 7. Downsampling API (GET /api/devices/:id/telemetry/downsampled)
	reqDown, _ := http.NewRequest("GET", "/api/devices/1/telemetry/downsampled?parameter_id=10&resolution=5m", nil)
	reqDown.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wDown := httptest.NewRecorder()
	env.Router.ServeHTTP(wDown, reqDown)

	if wDown.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for downsampled telemetry, got %d: %s", wDown.Code, wDown.Body.String())
	}

	// 8. Run Buckets API (POST /api/aggregations/run)
	runBody, _ := json.Marshal(map[string]interface{}{"device_id": 1})
	reqRunAll, _ := http.NewRequest("POST", "/api/aggregations/run", bytes.NewReader(runBody))
	reqRunAll.Header.Set("Authorization", "Bearer "+env.AdminToken)
	reqRunAll.Header.Set("Content-Type", "application/json")
	wRunAll := httptest.NewRecorder()
	env.Router.ServeHTTP(wRunAll, reqRunAll)

	if wRunAll.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for run all buckets, got %d: %s", wRunAll.Code, wRunAll.Body.String())
	}

	// 9. Inspect Bucket Samples API (GET /api/aggregations/samples)
	reqSamples, _ := http.NewRequest("GET", "/api/aggregations/samples?device_id=1&parameter_id=10", nil)
	reqSamples.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wSamples := httptest.NewRecorder()
	env.Router.ServeHTTP(wSamples, reqSamples)

	if wSamples.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for get bucket samples, got %d: %s", wSamples.Code, wSamples.Body.String())
	}
}

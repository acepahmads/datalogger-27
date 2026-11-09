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
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type multiTenantTestEnv struct {
	Router          *gin.Engine
	DB              *gorm.DB
	AuthService     *service.AuthService
	TenantService   *service.TenantService
	TelemetrySvc    *service.TelemetryService
	SuperAdminToken string
	Tenant1AdmToken string
}

func setupMultiTenantTestEnv(t *testing.T) *multiTenantTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_mt_iso_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	// Simulate legacy single-tenant records BEFORE running migrations/seed
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}
	// Run migrations once more to trigger backfill over seeded single-tenant users & devices
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("second RunMigrations failed: %v", err)
	}

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	calibRepo := repository.NewCalibrationRepository(db)
	outboundRepo := repository.NewOutboundRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	authService := service.NewAuthService(systemRepo, cfg)
	tenantService := service.NewTenantService(tenantRepo, systemRepo)
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)
	deviceService.SetTenantService(tenantService)

	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)
	systemService.SetTenantService(tenantService)

	hub := websocket.NewHub()
	tCfg := service.DefaultTelemetryConfig()
	tCfg.QueueEnabled = false
	tCfg.FlushInterval = 20 * time.Millisecond
	telemetryService := service.NewTelemetryService(telemetryRepo, hub, tCfg)
	telemetryService.SetDeviceLookup(deviceRepo.GetByID)
	telemetryService.SetParamLookup(deviceRepo.FindParameterByID)
	t.Cleanup(func() {
		telemetryService.Stop()
	})

	categoryService := service.NewCategoryService(categoryRepo, systemRepo)
	calibrationService := service.NewCalibrationService(calibRepo, deviceRepo, systemRepo, telemetryService)
	calibrationService.SetTenantService(tenantService)

	ingestService := service.NewIngestService(tenantService, deviceRepo, categoryRepo, telemetryService)
	outboundService := service.NewOutboundService(outboundRepo, deviceRepo, nil)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(authService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	ingestHandler := handler.NewIngestHandler(ingestService)
	deviceHandler := handler.NewDeviceHandler(deviceService, nil)
	systemHandler := handler.NewSystemHandler(systemService)
	systemHandler.SetTenantService(tenantService)
	systemHandler.SetOutboundService(outboundService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	calibHandler := handler.NewCalibrationHandler(calibrationService)
	outboundHandler := handler.NewOutboundHandler(outboundService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	api := r.Group("/api")
	api.Use(middleware.OptionalTenantContext(authService))
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/ingest/telemetry", ingestHandler.IngestTelemetry)
		api.POST("/v1/ingest/telemetry", ingestHandler.IngestTelemetry)

		api.GET("/system/branding", systemHandler.GetBranding)
		api.PUT("/system/branding", middleware.JWTAuth(authService), middleware.RequirePermission("config.manage"), systemHandler.SaveBranding)
		api.GET("/system/location", systemHandler.GetStationLocation)
		api.PUT("/system/location", middleware.JWTAuth(authService), middleware.RequirePermission("config.manage"), systemHandler.SaveStationLocation)
		api.GET("/system/config", middleware.JWTAuth(authService), middleware.RequirePermission("config.view"), systemHandler.GetConfig)
		api.PUT("/system/config", middleware.JWTAuth(authService), middleware.RequirePermission("config.manage"), systemHandler.SaveConfig)
		api.GET("/alarms", middleware.JWTAuth(authService), systemHandler.GetAlarms)
		api.POST("/alarms", middleware.JWTAuth(authService), systemHandler.CreateAlarm)
		api.GET("/alarms/:id", middleware.JWTAuth(authService), systemHandler.GetAlarmByID)

		tenantsGrp := api.Group("/tenants")
		tenantsGrp.Use(middleware.JWTAuth(authService))
		{
			tenantsGrp.GET("/context", tenantHandler.GetContext)
			tenantsGrp.GET("/current", tenantHandler.GetCurrentTenant)
			tenantsGrp.GET("", tenantHandler.ListTenants)
			tenantsGrp.POST("", middleware.RequireRole("SUPERADMIN"), tenantHandler.CreateTenant)
			tenantsGrp.PUT("/:id", tenantHandler.UpdateTenant)
			tenantsGrp.DELETE("/:id", middleware.RequireRole("SUPERADMIN"), tenantHandler.DeleteTenant)
		}

		sitesGrp := api.Group("/sites")
		sitesGrp.Use(middleware.JWTAuth(authService))
		{
			sitesGrp.GET("", middleware.RequirePermission("site.view"), tenantHandler.ListSites)
			sitesGrp.POST("", middleware.RequirePermission("site.manage"), tenantHandler.CreateSite)
			sitesGrp.PUT("/:id", middleware.RequirePermission("site.manage"), tenantHandler.UpdateSite)
			sitesGrp.DELETE("/:id", middleware.RequirePermission("site.manage"), tenantHandler.DeleteSite)
		}

		keysGrp := api.Group("/ingest-keys")
		keysGrp.Use(middleware.JWTAuth(authService))
		{
			keysGrp.GET("", middleware.RequirePermission("ingest_key.manage"), tenantHandler.ListAPIKeys)
			keysGrp.POST("", middleware.RequirePermission("ingest_key.manage"), tenantHandler.CreateAPIKey)
			keysGrp.PUT("/:id", middleware.RequirePermission("ingest_key.manage"), tenantHandler.ToggleAPIKey)
			keysGrp.PUT("/:id/enable", middleware.RequirePermission("ingest_key.manage"), tenantHandler.ToggleAPIKey)
			keysGrp.DELETE("/:id", middleware.RequirePermission("ingest_key.manage"), tenantHandler.DeleteAPIKey)
		}

		usersGrp := api.Group("/users")
		usersGrp.Use(middleware.JWTAuth(authService))
		{
			usersGrp.GET("", middleware.RequirePermission("user.view"), userHandler.ListUsers)
			usersGrp.GET("/:id", middleware.RequirePermission("user.view"), userHandler.GetUser)
			usersGrp.POST("", middleware.RequirePermission("user.manage"), userHandler.CreateUser)
			usersGrp.PUT("/:id", middleware.RequirePermission("user.manage"), userHandler.UpdateUser)
		}

		devGrp := api.Group("/devices")
		devGrp.Use(middleware.JWTAuth(authService))
		{
			devGrp.GET("", middleware.RequirePermission("device.view"), deviceHandler.ListDevices)
			devGrp.GET("/:id", middleware.RequirePermission("device.view"), deviceHandler.GetDeviceByID)
			devGrp.POST("", middleware.RequirePermission("device.create"), deviceHandler.CreateDevice)
		}

		catGrp := api.Group("/categories")
		catGrp.Use(middleware.JWTAuth(authService))
		{
			catGrp.GET("", categoryHandler.ListCategories)
			catGrp.POST("", middleware.RequirePermission("category.manage"), categoryHandler.CreateCategory)
			catGrp.DELETE("/:id", middleware.RequirePermission("category.manage"), categoryHandler.DeleteCategory)
		}

		calGrp := api.Group("/calibrations")
		calGrp.Use(middleware.JWTAuth(authService))
		{
			calGrp.GET("/parameters", calibHandler.ListParameters)
		}

		outGrp := api.Group("/destinations")
		outGrp.Use(middleware.JWTAuth(authService))
		{
			outGrp.GET("", outboundHandler.GetDestinations)
			outGrp.POST("", outboundHandler.CreateDestination)
			outGrp.GET("/:id", outboundHandler.GetDestinationByID)
		}
	}

	superTok, _, err := authService.Login("admin", "Password123!", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("superadmin login failed: %v", err)
	}
	t1AdmTok, _, err := authService.Login("admin_user", "Password123!", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("admin_user login failed: %v", err)
	}

	return &multiTenantTestEnv{
		Router:          r,
		DB:              db,
		AuthService:     authService,
		TenantService:   tenantService,
		TelemetrySvc:    telemetryService,
		SuperAdminToken: superTok,
		Tenant1AdmToken: t1AdmTok,
	}
}

func doJSONReq(t *testing.T, r *gin.Engine, method, path, token string, headers map[string]string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, path, &buf)
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Test 1: Auto-backfill migration creates Tenant #1, Site #1, global SUPERADMIN (tenant_id=NULL), and scoped non-SUPERADMIN (tenant_id=1)
func TestMultiTenant_AutoBackfillMigration(t *testing.T) {
	env := setupMultiTenantTestEnv(t)

	var t1 model.Tenant
	if err := env.DB.First(&t1, 1).Error; err != nil {
		t.Fatalf("expected Tenant #1 to exist: %v", err)
	}
	if t1.Code != "default" {
		t.Fatalf("expected Tenant #1 code 'default', got %q", t1.Code)
	}

	var s1 model.Site
	if err := env.DB.First(&s1, 1).Error; err != nil {
		t.Fatalf("expected Site #1 to exist: %v", err)
	}
	if s1.TenantID != 1 {
		t.Fatalf("expected Site #1 tenant_id=1, got %d", s1.TenantID)
	}

	var superUser model.User
	if err := env.DB.Where("username = ?", "admin").First(&superUser).Error; err != nil {
		t.Fatalf("expected superadmin user: %v", err)
	}
	if superUser.TenantID != nil {
		t.Fatalf("expected SUPERADMIN tenant_id == nil (global), got %v", *superUser.TenantID)
	}

	var t1Admin model.User
	if err := env.DB.Where("username = ?", "admin_user").First(&t1Admin).Error; err != nil {
		t.Fatalf("expected admin_user: %v", err)
	}
	if t1Admin.TenantID == nil || *t1Admin.TenantID != 1 {
		t.Fatalf("expected admin_user tenant_id == 1, got %v", t1Admin.TenantID)
	}
}

// Test 2 & 3: Strict Cross-Tenant Isolation, Duplicate DeviceCode Allowed Across Tenants, Anti-Spoofing, and SUPERADMIN Switching
func TestMultiTenant_StrictIsolationAndSuperAdminSwitching(t *testing.T) {
	env := setupMultiTenantTestEnv(t)

	// SUPERADMIN creates Tenant #2 ("PT Sawit Lestari")
	wCreateT2 := doJSONReq(t, env.Router, "POST", "/api/tenants", env.SuperAdminToken, nil, map[string]interface{}{
		"name":          "PT Sawit Lestari",
		"code":          "sawit-lestari",
		"customer_name": "PT Sawit Lestari Tbk",
	})
	if wCreateT2.Code != http.StatusOK && wCreateT2.Code != http.StatusCreated {
		t.Fatalf("create Tenant B failed (%d): %s", wCreateT2.Code, wCreateT2.Body.String())
	}
	var createT2Resp struct {
		Data model.Tenant `json:"data"`
	}
	_ = json.Unmarshal(wCreateT2.Body.Bytes(), &createT2Resp)
	tenant2ID := createT2Resp.Data.ID
	if tenant2ID < 2 {
		t.Fatalf("expected tenant2ID >= 2, got %d", tenant2ID)
	}

	// Find ADMIN role ID
	var adminRole model.Role
	env.DB.Where("UPPER(name) = ?", "ADMIN").First(&adminRole)

	// SUPERADMIN creates an ADMIN user inside Tenant #2
	wCreateT2Admin := doJSONReq(t, env.Router, "POST", "/api/users", env.SuperAdminToken, nil, map[string]interface{}{
		"username":  "admin_sawit",
		"full_name": "Admin Sawit",
		"password":  "Password123!",
		"role_id":   adminRole.ID,
		"is_active": true,
		"tenant_id": tenant2ID,
	})
	if wCreateT2Admin.Code != http.StatusOK {
		t.Fatalf("create Tenant B admin failed (%d): %s", wCreateT2Admin.Code, wCreateT2Admin.Body.String())
	}

	t2AdmTok, _, err := env.AuthService.Login("admin_sawit", "Password123!", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("login admin_sawit failed: %v", err)
	}

	// Both Tenant #1 and Tenant #2 register a device with the EXACT SAME DeviceCode "AQMS-SHARED-01"!
	wDevT1 := doJSONReq(t, env.Router, "POST", "/api/devices", env.Tenant1AdmToken, nil, map[string]interface{}{
		"device_code": "AQMS-SHARED-01",
		"device_name": "Tenant 1 AQMS Analyzer",
		"device_type": "MODBUS_TCP",
		"status":      "ACTIVE",
		"enabled":     false,
		"connection": map[string]interface{}{
			"protocol":        "MODBUS_TCP",
			"connection_type": "ETHERNET",
			"host":            "10.1.0.10",
			"port":            502,
			"enabled":         false,
		},
	})
	if wDevT1.Code != http.StatusCreated && wDevT1.Code != http.StatusOK {
		t.Fatalf("Tenant 1 create AQMS-SHARED-01 failed (%d): %s", wDevT1.Code, wDevT1.Body.String())
	}

	wDevT2 := doJSONReq(t, env.Router, "POST", "/api/devices", t2AdmTok, nil, map[string]interface{}{
		"device_code": "AQMS-SHARED-01",
		"device_name": "Tenant 2 AQMS Analyzer",
		"device_type": "MODBUS_TCP",
		"status":      "ACTIVE",
		"enabled":     false,
		"connection": map[string]interface{}{
			"protocol":        "MODBUS_TCP",
			"connection_type": "ETHERNET",
			"host":            "10.2.0.10",
			"port":            502,
			"enabled":         false,
		},
	})
	if wDevT2.Code != http.StatusCreated && wDevT2.Code != http.StatusOK {
		t.Fatalf("Tenant 2 create duplicate code AQMS-SHARED-01 failed (%d): %s", wDevT2.Code, wDevT2.Body.String())
	}

	var devT2Resp struct {
		Data model.Device `json:"data"`
	}
	_ = json.Unmarshal(wDevT2.Body.Bytes(), &devT2Resp)
	devT2ID := devT2Resp.Data.ID

	// Anti-spoofing check: Tenant 1 Admin sends spoofed header X-Tenant-ID: <tenant2ID>
	wSpoofList := doJSONReq(t, env.Router, "GET", "/api/devices", env.Tenant1AdmToken, map[string]string{
		"X-Tenant-ID": fmt.Sprintf("%d", tenant2ID),
	}, nil)
	if wSpoofList.Code != http.StatusOK {
		t.Fatalf("list devices failed: %d", wSpoofList.Code)
	}
	if bytes.Contains(wSpoofList.Body.Bytes(), []byte("Tenant 2 AQMS Analyzer")) {
		t.Fatalf("SECURITY VIOLATION: Tenant 1 user saw Tenant 2 device via spoofed X-Tenant-ID header!")
	}

	// Direct cross-tenant ID access check: Tenant 1 Admin tries GET /api/devices/:devT2ID -> must be 403 or 404
	wCrossGet := doJSONReq(t, env.Router, "GET", fmt.Sprintf("/api/devices/%d", devT2ID), env.Tenant1AdmToken, nil, nil)
	if wCrossGet.Code != http.StatusForbidden && wCrossGet.Code != http.StatusNotFound {
		t.Fatalf("SECURITY VIOLATION: expected 403/404 for cross-tenant device GET, got %d: %s", wCrossGet.Code, wCrossGet.Body.String())
	}

	// Cross-tenant Outbound Destination check
	wOutT2 := doJSONReq(t, env.Router, "POST", "/api/destinations", t2AdmTok, nil, map[string]interface{}{
		"name":      "Tenant 2 KLHK Webhook",
		"type":      "REST",
		"endpoint":  "https://example.org/klhk",
		"auth_type": "NONE",
		"enabled":   false,
	})
	if wOutT2.Code != http.StatusOK && wOutT2.Code != http.StatusCreated {
		t.Fatalf("Tenant 2 create outbound failed (%d): %s", wOutT2.Code, wOutT2.Body.String())
	}
	wOutListT1 := doJSONReq(t, env.Router, "GET", "/api/destinations", env.Tenant1AdmToken, nil, nil)
	if bytes.Contains(wOutListT1.Body.Bytes(), []byte("Tenant 2 KLHK Webhook")) {
		t.Fatalf("SECURITY VIOLATION: Tenant 1 saw Tenant 2 outbound destination!")
	}

	// SUPERADMIN Global (no X-Tenant-ID) sees BOTH Tenant 1 & Tenant 2 devices!
	wSuperGlobal := doJSONReq(t, env.Router, "GET", "/api/devices?all=true", env.SuperAdminToken, nil, nil)
	if !bytes.Contains(wSuperGlobal.Body.Bytes(), []byte("Tenant 1 AQMS Analyzer")) ||
		!bytes.Contains(wSuperGlobal.Body.Bytes(), []byte("Tenant 2 AQMS Analyzer")) {
		t.Fatalf("expected global SUPERADMIN to see both Tenant 1 and Tenant 2 devices, body=%s", wSuperGlobal.Body.String())
	}

	// SUPERADMIN Switched to Tenant 2 (X-Tenant-ID: tenant2ID) sees ONLY Tenant 2 devices & branding!
	wSuperScoped := doJSONReq(t, env.Router, "GET", "/api/devices?all=true", env.SuperAdminToken, map[string]string{
		"X-Tenant-ID": fmt.Sprintf("%d", tenant2ID),
	}, nil)
	if bytes.Contains(wSuperScoped.Body.Bytes(), []byte("Tenant 1 AQMS Analyzer")) ||
		!bytes.Contains(wSuperScoped.Body.Bytes(), []byte("Tenant 2 AQMS Analyzer")) {
		t.Fatalf("expected Tenant-2-scoped SUPERADMIN to see ONLY Tenant 2 devices, body=%s", wSuperScoped.Body.String())
	}

	wBrandingT2 := doJSONReq(t, env.Router, "GET", "/api/system/branding", env.SuperAdminToken, map[string]string{
		"X-Tenant-ID": fmt.Sprintf("%d", tenant2ID),
	}, nil)
	if !bytes.Contains(wBrandingT2.Body.Bytes(), []byte("PT Sawit Lestari Tbk")) {
		t.Fatalf("expected Tenant 2 customer branding overlay, got %s", wBrandingT2.Body.String())
	}
}

// Test 4: Self-Service Scoped ADMIN User & Site Management + Station-Restricted OPERATOR
func TestMultiTenant_ScopedAdminSelfServiceAndSiteRestriction(t *testing.T) {
	env := setupMultiTenantTestEnv(t)

	// 1. Tenant 1 ADMIN cannot create a SUPERADMIN user
	var superRole model.Role
	env.DB.Where("UPPER(name) IN ?", []string{"SUPERADMIN", "ADMINISTRATOR"}).First(&superRole)

	wEscalation := doJSONReq(t, env.Router, "POST", "/api/users", env.Tenant1AdmToken, nil, map[string]interface{}{
		"username":  "rogue_super",
		"password":  "Password123!",
		"role_id":   superRole.ID,
		"is_active": true,
	})
	if wEscalation.Code != http.StatusForbidden && wEscalation.Code != http.StatusBadRequest {
		t.Fatalf("SECURITY VIOLATION: Tenant ADMIN must be blocked from creating SUPERADMIN, got %d", wEscalation.Code)
	}

	// 1b. Tenant 1 ADMIN can update their own tenant branding via PUT /api/tenants/current
	wSelfProfile := doJSONReq(t, env.Router, "PUT", "/api/tenants/current", env.Tenant1AdmToken, nil, map[string]interface{}{
		"name":          "PT Default Industri",
		"customer_name": "PT Default Industri Mandiri",
	})
	if wSelfProfile.Code != http.StatusOK {
		t.Fatalf("Tenant 1 ADMIN update self profile failed (%d): %s", wSelfProfile.Code, wSelfProfile.Body.String())
	}

	// 2. Tenant 1 ADMIN creates Site #2 ("SPARING Outlet #2")
	wSite2 := doJSONReq(t, env.Router, "POST", "/api/sites", env.Tenant1AdmToken, nil, map[string]interface{}{
		"name":              "SPARING Outlet #2",
		"code":              "sparing-outlet-2",
		"monitoring_domain": "SPARING",
		"enabled":           true,
	})
	if wSite2.Code != http.StatusOK && wSite2.Code != http.StatusCreated {
		t.Fatalf("Tenant ADMIN create site failed (%d): %s", wSite2.Code, wSite2.Body.String())
	}
	var site2Resp struct {
		Data model.Site `json:"data"`
	}
	_ = json.Unmarshal(wSite2.Body.Bytes(), &site2Resp)
	site2ID := site2Resp.Data.ID

	// Create a device specifically on Site #2
	wDevSite2 := doJSONReq(t, env.Router, "POST", "/api/devices", env.Tenant1AdmToken, nil, map[string]interface{}{
		"site_id":     site2ID,
		"device_code": "SPARING-S2-01",
		"device_name": "Sensor COD Outlet 2",
		"device_type": "MODBUS_TCP",
		"status":      "ACTIVE",
		"enabled":     false,
		"connection": map[string]interface{}{
			"protocol":        "MODBUS_TCP",
			"connection_type": "ETHERNET",
			"host":            "10.1.2.10",
			"port":            502,
			"enabled":         false,
		},
	})
	if wDevSite2.Code != http.StatusCreated && wDevSite2.Code != http.StatusOK {
		t.Fatalf("create device on Site 2 failed (%d): %s", wDevSite2.Code, wDevSite2.Body.String())
	}

	// 3. Tenant 1 ADMIN creates an OPERATOR restricted strictly to Site #2
	var opRole model.Role
	env.DB.Where("UPPER(name) = ?", "OPERATOR").First(&opRole)

	wOpRestricted := doJSONReq(t, env.Router, "POST", "/api/users", env.Tenant1AdmToken, nil, map[string]interface{}{
		"username":         "op_site2_only",
		"full_name":        "Operator Outlet 2",
		"password":         "Password123!",
		"role_id":          opRole.ID,
		"is_active":        true,
		"allowed_site_ids": fmt.Sprintf("%d", site2ID),
	})
	if wOpRestricted.Code != http.StatusOK {
		t.Fatalf("create site-restricted operator failed (%d): %s", wOpRestricted.Code, wOpRestricted.Body.String())
	}

	opSite2Tok, _, err := env.AuthService.Login("op_site2_only", "Password123!", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("login op_site2_only failed: %v", err)
	}

	// Restricted operator lists devices -> sees ONLY Site #2 device ("SPARING-S2-01"), NOT Site #1 seeded devices ("PM-01")!
	wOpDevs := doJSONReq(t, env.Router, "GET", "/api/devices?all=true", opSite2Tok, nil, nil)
	if wOpDevs.Code != http.StatusOK {
		t.Fatalf("op list devices failed: %d", wOpDevs.Code)
	}
	if !bytes.Contains(wOpDevs.Body.Bytes(), []byte("SPARING-S2-01")) {
		t.Fatalf("expected restricted operator to see SPARING-S2-01, got %s", wOpDevs.Body.String())
	}
	if bytes.Contains(wOpDevs.Body.Bytes(), []byte("PM-01")) {
		t.Fatalf("SECURITY VIOLATION: Site-2-restricted operator saw Site #1 device PM-01!")
	}
}

// Test 5: Hybrid Edge Push Gateway (POST /api/v1/ingest/telemetry) with Auto-Provisioning & Key Validation
func TestMultiTenant_HybridEdgePushAutoProvision(t *testing.T) {
	env := setupMultiTenantTestEnv(t)

	// Create Tenant #2
	wT2 := doJSONReq(t, env.Router, "POST", "/api/tenants", env.SuperAdminToken, nil, map[string]interface{}{
		"name": "PT Tambang Mandiri",
		"code": "tambang-mandiri",
	})
	var t2Resp struct {
		Data model.Tenant `json:"data"`
	}
	_ = json.Unmarshal(wT2.Body.Bytes(), &t2Resp)
	tenant2ID := t2Resp.Data.ID

	// Mint an Edge Ingest API key for Tenant #2
	wKey := doJSONReq(t, env.Router, "POST", "/api/ingest-keys", env.SuperAdminToken, nil, map[string]interface{}{
		"tenant_id":      tenant2ID,
		"name":           "Edge Box Pit #4",
		"auto_provision": true,
	})
	if wKey.Code != http.StatusOK && wKey.Code != http.StatusCreated {
		t.Fatalf("create ingest key failed (%d): %s", wKey.Code, wKey.Body.String())
	}
	var keyResp struct {
		Data struct {
			PlainKey string             `json:"plain_key"`
			Key      model.TenantAPIKey `json:"key"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wKey.Body.Bytes(), &keyResp)
	plainKey := keyResp.Data.PlainKey
	if plainKey == "" {
		t.Fatalf("expected non-empty plain_key dlg_live_...")
	}

	// 1. Reject invalid key with 401
	wBadKey := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", "", map[string]string{
		"X-Ingest-Key": "dlg_live_invalid_bogus_key",
	}, map[string]interface{}{
		"device_code": "EDGE-01",
		"readings":    []map[string]interface{}{{"code": "PH", "value": 7.12}},
	})
	if wBadKey.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid ingest key, got %d", wBadKey.Code)
	}

	// 2. Push valid Edge telemetry -> auto-provisions EDGE_PUSH device & parameters inside Tenant #2 ONLY!
	wPush := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", "", map[string]string{
		"X-Ingest-Key": plainKey,
	}, map[string]interface{}{
		"site_code":   "pit-4-wqms",
		"site_name":   "WQMS Settling Pond Pit 4",
		"domain":      "WQMS",
		"device_code": "EDGE-WQMS-04",
		"device_name": "Multiparameter Sonde Pit 4",
		"readings": []map[string]interface{}{
			{"code": "PH", "name": "Acidity pH", "value": 7.35, "unit": "pH"},
			{"code": "TSS", "name": "Total Suspended Solids", "value": 42.8, "unit": "mg/L"},
		},
	})
	if wPush.Code != http.StatusOK {
		t.Fatalf("valid edge push failed (%d): %s", wPush.Code, wPush.Body.String())
	}

	// Verify auto-provisioned device belongs strictly to Tenant #2 and is invisible to Tenant #1!
	wT1Check := doJSONReq(t, env.Router, "GET", "/api/devices?all=true", env.Tenant1AdmToken, nil, nil)
	if bytes.Contains(wT1Check.Body.Bytes(), []byte("EDGE-WQMS-04")) {
		t.Fatalf("SECURITY VIOLATION: Tenant 1 saw Tenant 2 auto-provisioned EDGE-WQMS-04 device!")
	}

	wT2Check := doJSONReq(t, env.Router, "GET", "/api/devices?all=true", env.SuperAdminToken, map[string]string{
		"X-Tenant-ID": fmt.Sprintf("%d", tenant2ID),
	}, nil)
	if !bytes.Contains(wT2Check.Body.Bytes(), []byte("EDGE-WQMS-04")) {
		t.Fatalf("expected EDGE-WQMS-04 in Tenant 2 devices, got %s", wT2Check.Body.String())
	}
}

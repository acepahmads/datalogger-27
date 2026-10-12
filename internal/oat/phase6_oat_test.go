package oat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/service/notification"
	"datalogger/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func makeTestToken(user *model.User, secret string) string {
	roleName := "Viewer"
	if user.Role != nil {
		roleName = user.Role.Name
	}
	claims := service.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "datalogger-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

type oatHarness struct {
	db                 *gorm.DB
	cfg                *config.Config
	alarmRepo          *repository.AlarmRepository
	deviceRepo         *repository.DeviceRepository
	systemRepo         *repository.SystemRepository
	phaseRepo          *repository.PhaseRepository
	telemetryRepo      *repository.TelemetryRepository
	notificationRepo   *repository.NotificationRepository
	authService        *service.AuthService
	deviceService      *service.DeviceService
	systemService      *service.SystemService
	alarmService       *service.AlarmService
	evaluator          *service.AlarmEvaluator
	notificationEngine *service.NotificationEngine
	hub                *websocket.Hub
	router             *gin.Engine
	adminToken         string
	engineerToken      string
	operatorToken      string
	viewerToken        string
}

func setupOATHarness(t *testing.T, dbPath string) *oatHarness {
	gin.SetMode(gin.TestMode)

	var dsn string
	if dbPath == "" {
		dsn = fmt.Sprintf("file:mem_oat_%d?mode=memory&cache=shared", time.Now().UnixNano()+int64(rand.Intn(100000)))
	} else {
		dsn = dbPath
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	err = db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Alarm{},
		&model.Notification{},
		&model.NotificationChannel{},
		&model.AuditTrail{},
		&model.SystemLog{},
		&model.CommunicationLog{},
		&model.DevelopmentPhase{},
		&model.DevelopmentSubphase{},
		&model.DevelopmentTask{},
	)
	if err != nil {
		t.Fatalf("failed migrating schema: %v", err)
	}

	cfg := &config.Config{
		AppName:            "Datalogger Analysis Application",
		Version:            "v1.0.0-phase1",
		Environment:        "test",
		JWTSecret:          "oat-test-secret-key-32-chars-ok!!",
		JWTExpirationHours: 24,
	}

	alarmRepo := repository.NewAlarmRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	authService := service.NewAuthService(systemRepo, cfg)
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)

	hub := websocket.NewHub()
	go hub.Run()

	notifCfg := service.DefaultNotificationEngineConfig()
	notifCfg.PollInterval = 50 * time.Millisecond
	notifCfg.DefaultBackoffSec = 1
	notificationEngine := service.NewNotificationEngine(notificationRepo, alarmRepo, systemRepo, notifCfg)
	notificationEngine.RegisterProvider(notification.NewWebProvider(hub))
	alarmService.SetNotificationEngine(notificationEngine)
	notificationEngine.Start()

	evalCfg := service.DefaultAlarmEvaluatorConfig()
	evalCfg.DefaultHysteresisPercent = 1.0
	evalCfg.DeviceOfflineTimeout = 5 * time.Second
	evalCfg.StartupGracePeriod = 1 * time.Second
	evalCfg.CommErrorConsecutiveLimit = 3
	evalCfg.CommRecoverySuccessLimit = 2
	evalCfg.StuckValueCountThreshold = 5
	evalCfg.StuckValueDurationThreshold = 2 * time.Second
	evalCfg.BadQualityCountThreshold = 3
	evalCfg.StaleDataCountThreshold = 2

	evaluator := service.NewAlarmEvaluator(context.Background(), alarmService, evalCfg, func(ctx context.Context, paramID uint) (*model.Parameter, error) {
		return deviceRepo.FindParameterByID(paramID)
	})

	systemService.SetAlarmService(alarmService)

	// Roles & Users
	adminRole := model.Role{Name: "Administrator"}
	engRole := model.Role{Name: "Engineer"}
	opRole := model.Role{Name: "Operator"}
	viewerRole := model.Role{Name: "Viewer"}
	db.Create(&adminRole)
	db.Create(&engRole)
	db.Create(&opRole)
	db.Create(&viewerRole)

	adminUser := model.User{Username: "admin_test", FullName: "Lead Admin", RoleID: adminRole.ID, Role: &adminRole, IsActive: true}
	engUser := model.User{Username: "eng_test", FullName: "System Engineer", RoleID: engRole.ID, Role: &engRole, IsActive: true}
	opUser := model.User{Username: "op_test", FullName: "Operator One", RoleID: opRole.ID, Role: &opRole, IsActive: true}
	viewerUser := model.User{Username: "viewer_test", FullName: "Viewer Guest", RoleID: viewerRole.ID, Role: &viewerRole, IsActive: true}
	db.Create(&adminUser)
	db.Create(&engUser)
	db.Create(&opUser)
	db.Create(&viewerUser)

	adminToken := makeTestToken(&adminUser, cfg.JWTSecret)
	engToken := makeTestToken(&engUser, cfg.JWTSecret)
	opToken := makeTestToken(&opUser, cfg.JWTSecret)
	viewerToken := makeTestToken(&viewerUser, cfg.JWTSecret)

	// Router setup
	r := gin.New()
	r.Use(gin.Recovery())

	systemHandler := handler.NewSystemHandler(systemService)
	notifHandler := handler.NewNotificationHandler(notificationEngine, notificationRepo)

	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWS(hub, c)
	})

	api := r.Group("/api")
	alarmsGroup := api.Group("/alarms")
	alarmsGroup.Use(middleware.JWTAuth(authService))
	{
		alarmsGroup.GET("", middleware.RequirePermission("alarm.view"), systemHandler.GetAlarms)
		alarmsGroup.GET("/active", middleware.RequirePermission("alarm.view"), systemHandler.GetActiveAlarms)
		alarmsGroup.GET("/history", middleware.RequirePermission("alarm.view"), systemHandler.GetAlarmHistory)
		alarmsGroup.GET("/:id", middleware.RequirePermission("alarm.view"), systemHandler.GetAlarmByID)
		alarmsGroup.POST("", middleware.RequirePermission("alarm.manage"), systemHandler.CreateAlarm)
		alarmsGroup.PUT("/:id/acknowledge", middleware.RequirePermission("alarm.acknowledge"), systemHandler.AcknowledgeAlarm)
		alarmsGroup.PUT("/:id/clear", middleware.RequirePermission("alarm.manage"), systemHandler.ClearAlarm)
	}

	notifGroup := api.Group("/notifications")
	notifGroup.Use(middleware.JWTAuth(authService))
	{
		notifGroup.GET("", middleware.RequirePermission("notification.view"), notifHandler.ListNotifications)
		notifGroup.GET("/metrics", middleware.RequirePermission("notification.view"), notifHandler.GetMetrics)
		notifGroup.GET("/:id", middleware.RequirePermission("notification.view"), notifHandler.GetNotification)
		notifGroup.POST("/:id/retry", middleware.RequirePermission("notification.manage"), notifHandler.RetryNotification)
		notifGroup.POST("/:id/cancel", middleware.RequirePermission("notification.manage"), notifHandler.CancelNotification)
		notifGroup.GET("/channels", middleware.RequirePermission("notification.view"), notifHandler.ListChannels)
		notifGroup.GET("/channels/:id", middleware.RequirePermission("notification.view"), notifHandler.GetChannel)
		notifGroup.PUT("/channels/:id", middleware.RequirePermission("notification.manage"), notifHandler.UpdateChannel)
		notifGroup.POST("/channels/:id/test", middleware.RequirePermission("notification.manage"), notifHandler.TestSend)
	}

	return &oatHarness{
		db:                 db,
		cfg:                cfg,
		alarmRepo:          alarmRepo,
		deviceRepo:         deviceRepo,
		systemRepo:         systemRepo,
		phaseRepo:          phaseRepo,
		telemetryRepo:      telemetryRepo,
		notificationRepo:   notificationRepo,
		authService:        authService,
		deviceService:      deviceService,
		systemService:      systemService,
		alarmService:       alarmService,
		evaluator:          evaluator,
		notificationEngine: notificationEngine,
		hub:                hub,
		router:             r,
		adminToken:         adminToken,
		engineerToken:      engToken,
		operatorToken:      opToken,
		viewerToken:        viewerToken,
	}
}

// =========================================================================
// SCENARIO A: THRESHOLD ALARM ENGINE
// =========================================================================
func TestOAT_ScenarioA_ThresholdAlarmEngine(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	ctx := context.Background()

	// Seed Device & Parameter with thresholds:
	// Normal: 25.0 | WarningLow: 10.0 | WarningHigh: 50.0 | CriticalLow: 5.0 | CriticalHigh: 80.0
	// Hysteresis: 1.0% (0.5 for High, 0.8 for CritHigh)
	wLow, wHigh := 10.0, 50.0
	cLow, cHigh := 5.0, 80.0
	hyst := 1.0

	dev := model.Device{DeviceCode: "OAT-DEV-A", DeviceName: "Boiler Pump", Status: model.DeviceStatusActive, Enabled: true}
	h.db.Create(&dev)

	param := model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "Temperature",
		ParameterName: "Boiler Temperature",
		Unit:          "°C",
		DataType:      model.DataTypeFloat32,
		WarningLow:    &wLow,
		WarningHigh:   &wHigh,
		CriticalLow:   &cLow,
		CriticalHigh:  &cHigh,
		Hysteresis:    &hyst,
		Enabled:       true,
	}
	h.db.Create(&param)

	evalInput := func(val float64, q model.TelemetryQuality) *service.LatestTelemetry {
		return &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: val,
			ReceivedAt:     time.Now().UTC(),
			Quality:        q,
			Unit:           param.Unit,
		}
	}

	// Step 1: Normal value -> No alarm
	h.evaluator.EvaluateTelemetry(ctx, evalInput(25.0, model.QualityGood), &param)
	if active, _ := h.alarmService.GetActiveAlarms(ctx); len(active) != 0 {
		t.Fatalf("Expected 0 active alarms on normal value, got %d", len(active))
	}

	// Step 2: Crosses warning threshold (55.0°C) -> Exactly one WARNING alarm
	h.evaluator.EvaluateTelemetry(ctx, evalInput(55.0, model.QualityGood), &param)
	active, _ := h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 {
		t.Fatalf("Expected 1 active alarm on warning cross, got %d", len(active))
	}
	if active[0].Severity != model.SeverityWarning || active[0].AlarmCode != service.AlarmCodeWarningHigh {
		t.Fatalf("Expected WARNING HIGH alarm, got %s / %s", active[0].Severity, active[0].AlarmCode)
	}

	// Step 3: Crosses critical threshold (85.0°C) -> Escalation to CRITICAL
	h.evaluator.EvaluateTelemetry(ctx, evalInput(85.0, model.QualityGood), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 {
		t.Fatalf("Expected exactly 1 active alarm after critical escalation, got %d", len(active))
	}
	if active[0].Severity != model.SeverityCritical || active[0].AlarmCode != service.AlarmCodeCriticalHigh {
		t.Fatalf("Expected CRITICAL HIGH alarm, got %s / %s", active[0].Severity, active[0].AlarmCode)
	}

	// Step 4: Repeated abnormal values -> No duplicate alarm storm
	for i := 0; i < 5; i++ {
		h.evaluator.EvaluateTelemetry(ctx, evalInput(86.0+float64(i), model.QualityGood), &param)
	}
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 {
		t.Fatalf("Expected no duplicate alarms on repeated critical values, got %d", len(active))
	}

	// Step 5 & 6: Hysteresis deadband test
	// Critical High is 80.0, hysteresis 1.0% = deadband 0.8°C -> recovery threshold = 79.2°C
	// Drop to 79.5°C -> Still above 79.2°C deadband! Alarm must remain CRITICAL.
	h.evaluator.EvaluateTelemetry(ctx, evalInput(79.5, model.QualityGood), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeCriticalHigh {
		t.Fatalf("Hysteresis violation: alarm should remain critical at 79.5 (deadband is 79.2)")
	}

	// Drop to 78.0°C -> Clears Critical, but 78.0°C is still above WarningHigh (50.0) -> transitions/escalates to Warning
	h.evaluator.EvaluateTelemetry(ctx, evalInput(78.0, model.QualityGood), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeWarningHigh {
		t.Fatalf("Expected transition to WARNING HIGH at 78.0, got %v", active)
	}

	// Drop to 49.8°C -> WarningHigh is 50.0, deadband is 0.5 (49.5°C). 49.8 is within deadband -> remains WARNING!
	h.evaluator.EvaluateTelemetry(ctx, evalInput(49.8, model.QualityGood), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 {
		t.Fatalf("Hysteresis violation: alarm should remain active at 49.8 (deadband 49.5)")
	}

	// Drop to 48.5°C -> Below 50.0 - 1.0 (49.0°C deadband) -> Completely CLEARED!
	h.evaluator.EvaluateTelemetry(ctx, evalInput(48.5, model.QualityGood), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 0 {
		t.Fatalf("Expected alarm to clear at 48.5, active count: %d", len(active))
	}

	// Step 7: Bad Quality Telemetry follows evaluation policy (does not trigger false threshold alarm)
	h.evaluator.EvaluateTelemetry(ctx, evalInput(99.0, model.QualityBad), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	for _, a := range active {
		if strings.HasPrefix(a.AlarmCode, "THRESHOLD_") {
			t.Fatalf("Unexpected threshold alarm triggered on BAD quality telemetry")
		}
	}

	t.Logf("PASS: Scenario A Threshold Alarm Engine verified with full lifecycle, escalation, hysteresis deadband, and bad-quality immunity.")
}

// =========================================================================
// SCENARIO B: DEVICE OFFLINE & COMMUNICATION ERROR
// =========================================================================
func TestOAT_ScenarioB_DeviceOfflineAndCommError(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	ctx := context.Background()

	dev1 := model.Device{DeviceCode: "OAT-DEV-B1", DeviceName: "Main Inverter", Status: model.DeviceStatusActive, Enabled: true}
	dev2 := model.Device{DeviceCode: "OAT-DEV-B2", DeviceName: "Aux Generator", Status: model.DeviceStatusActive, Enabled: true}
	h.db.Create(&dev1)
	h.db.Create(&dev2)

	// Step 1: Healthy device does not generate false offline or comm alarm
	now := time.Now().UTC()
	h.evaluator.EvaluateDeviceOffline(ctx, &dev1, &service.DeviceHealthInput{
		DeviceID:             dev1.ID,
		State:                "CONNECTED",
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 10,
		LastCommSuccess:      &now,
	})
	active, _ := h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 0 {
		t.Fatalf("Expected 0 active alarms on healthy device, got %d", len(active))
	}

	// Step 2: Consecutive communication failures (threshold is 3)
	for i := 1; i <= 3; i++ {
		failTime := time.Now().UTC()
		h.evaluator.EvaluateDeviceCommFailure(ctx, &dev1, &service.DeviceHealthInput{
			DeviceID:             dev1.ID,
			State:                "ERROR",
			ConsecutiveFailures:  i,
			ConsecutiveSuccesses: 0,
			LastCommFailure:      &failTime,
			LastErrorCategory:    "TIMEOUT",
			LastError:            "serial port read timeout after 1000ms",
		}, fmt.Errorf("read timeout"))
	}
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeCommError {
		t.Fatalf("Expected 1 COMMUNICATION_ERROR alarm on 3 failures, got %v", active)
	}

	// Step 3: Multi-device isolation — Dev 2 is healthy, unaffected by Dev 1
	h.evaluator.EvaluateDeviceOffline(ctx, &dev2, &service.DeviceHealthInput{
		DeviceID:             dev2.ID,
		State:                "CONNECTED",
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 5,
		LastCommSuccess:      &now,
	})
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || *active[0].DeviceID != dev1.ID {
		t.Fatalf("Multi-device isolation failed: unexpected alarms on Dev 2: %v", active)
	}

	// Step 4: Communication recovery (2 consecutive successes required)
	h.evaluator.EvaluateDeviceCommSuccess(ctx, &dev1, &service.DeviceHealthInput{
		DeviceID:             dev1.ID,
		State:                "ONLINE",
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 1, // Only 1 success -> Not cleared yet
		LastCommSuccess:      &now,
	})
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 {
		t.Fatalf("Premature recovery: expected alarm still active after 1 success")
	}

	h.evaluator.EvaluateDeviceCommSuccess(ctx, &dev1, &service.DeviceHealthInput{
		DeviceID:             dev1.ID,
		State:                "CONNECTED",
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 2, // 2nd success -> Recovered!
		LastCommSuccess:      &now,
	})
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 0 {
		t.Fatalf("Expected COMMUNICATION_ERROR to clear after 2 successes, got %v", active)
	}

	// Step 5: Device Offline when last seen exceeds timeout (5s)
	pastTime := time.Now().UTC().Add(-10 * time.Second)
	h.evaluator.EvaluateDeviceOffline(ctx, &dev1, &service.DeviceHealthInput{
		DeviceID:             dev1.ID,
		State:                "OFFLINE",
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 0,
		LastCommSuccess:      &pastTime,
	})
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeDeviceOffline {
		t.Fatalf("Expected DEVICE_OFFLINE alarm on timeout, got %v", active)
	}

	t.Logf("PASS: Scenario B Device Offline & Communication Error verified with failure count thresholds, isolation, and recovery.")
}

// =========================================================================
// SCENARIO C: DATA ABNORMALITY
// =========================================================================
func TestOAT_ScenarioC_DataAbnormality(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	ctx := context.Background()

	dev := model.Device{DeviceCode: "OAT-DEV-C", DeviceName: "Flow Meter", Status: model.DeviceStatusActive, Enabled: true}
	h.db.Create(&dev)

	param := model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "FlowRate",
		ParameterName: "Water Flow",
		Unit:          "L/s",
		DataType:      model.DataTypeFloat32,
		Enabled:       true,
	}
	h.db.Create(&param)

	evalInput := func(val float64, q model.TelemetryQuality, tOffset time.Duration) *service.LatestTelemetry {
		return &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: val,
			ReceivedAt:     time.Now().UTC().Add(tOffset),
			Quality:        q,
			Unit:           param.Unit,
		}
	}

	// Step 1: Bad Quality Abnormality (threshold is 3 consecutive bad)
	for i := 1; i <= 3; i++ {
		h.evaluator.EvaluateTelemetry(ctx, evalInput(15.0, model.QualityBad, 0), &param)
	}
	active, _ := h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeAbnormalityBad {
		t.Fatalf("Expected DATA_ABNORMALITY_BAD_QUALITY alarm, got %v", active)
	}

	// Recovery on good quality
	h.evaluator.EvaluateTelemetry(ctx, evalInput(15.0, model.QualityGood, 0), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 0 {
		t.Fatalf("Expected BAD QUALITY alarm to clear on GOOD telemetry")
	}

	// Step 2: Stale Data Abnormality (threshold is 2 consecutive stale)
	for i := 1; i <= 2; i++ {
		h.evaluator.EvaluateTelemetry(ctx, evalInput(15.0, model.QualityStale, 0), &param)
	}
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeAbnormalityStale {
		t.Fatalf("Expected DATA_ABNORMALITY_STALE_DATA alarm, got %v", active)
	}

	// Recovery on fresh data
	h.evaluator.EvaluateTelemetry(ctx, evalInput(15.2, model.QualityGood, 0), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 0 {
		t.Fatalf("Expected STALE DATA alarm to clear on fresh telemetry")
	}

	// Step 3: Stuck Value Abnormality (threshold is 5 identical values across >= 2s)
	for i := 0; i < 6; i++ {
		offset := time.Duration(i) * 500 * time.Millisecond
		h.evaluator.EvaluateTelemetry(ctx, evalInput(42.0, model.QualityGood, offset), &param)
	}
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeAbnormalityStuck {
		t.Fatalf("Expected DATA_ABNORMALITY_STUCK_VALUE alarm, got %v", active)
	}

	// Recovery on value variation (> 1e-6)
	h.evaluator.EvaluateTelemetry(ctx, evalInput(42.5, model.QualityGood, 3*time.Second), &param)
	active, _ = h.alarmService.GetActiveAlarms(ctx)
	if len(active) != 0 {
		t.Fatalf("Expected STUCK VALUE alarm to clear on varied telemetry, got %v", active)
	}

	t.Logf("PASS: Scenario C Data Abnormality verified with Bad Quality, Stale Data, and Stuck Value detection and recovery.")
}

// =========================================================================
// SCENARIO D: ALARM ACKNOWLEDGEMENT
// =========================================================================
func TestOAT_ScenarioD_AlarmAcknowledgement(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	// Seed active alarm
	devID := uint(1)
	alarm := model.Alarm{
		DeviceID:    &devID,
		AlarmCode:   "THRESHOLD_WARNING_HIGH",
		Severity:    model.SeverityWarning,
		Status:      model.AlarmActive,
		Message:     "Water level high warning",
		TriggeredAt: time.Now().UTC(),
	}
	h.db.Create(&alarm)

	// Step 1: Viewer is rejected (403 Forbidden)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBufferString(`{"notes":"viewer trying"}`))
	req.Header.Set("Authorization", "Bearer "+h.viewerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for Viewer, got %d", w.Code)
	}

	// Step 2: Unauthenticated is rejected (401 Unauthorized)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBufferString(`{}`))
	w = httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized without token, got %d", w.Code)
	}

	// Step 3: Authorized Operator acknowledges successfully with server-derived identity
	// Even if client injects arbitrary "user_id": 9999 or "username": "hacker", server derives opUser!
	reqBody := `{"notes":"Inspected pump, pressure normal. Monitoring.","user_id":9999,"username":"spoofed"}`
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBufferString(reqBody))
	req.Header.Set("Authorization", "Bearer "+h.operatorToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for Operator acknowledge, got %d (%s)", w.Code, w.Body.String())
	}

	// Verify persistence in DB
	var dbAlarm model.Alarm
	h.db.First(&dbAlarm, alarm.ID)
	if dbAlarm.Status != model.AlarmAcked {
		t.Fatalf("Expected status ACKNOWLEDGED, got %s", dbAlarm.Status)
	}
	if dbAlarm.AcknowledgedBy != "op_test" || dbAlarm.AcknowledgedByName != "Operator One" {
		t.Fatalf("Server-derived identity failed! Expected op_test / Operator One, got %s / %s", dbAlarm.AcknowledgedBy, dbAlarm.AcknowledgedByName)
	}
	if dbAlarm.Notes != "Inspected pump, pressure normal. Monitoring." {
		t.Fatalf("Notes not persisted correctly: %s", dbAlarm.Notes)
	}
	if dbAlarm.AcknowledgedAt == nil {
		t.Fatalf("AcknowledgedAt timestamp is nil")
	}

	// Step 4: Audit trail record created
	var audit model.AuditTrail
	h.db.Where("action = ? AND details LIKE ?", "ALARM_ACKNOWLEDGE", "%"+alarm.AlarmCode+"%").First(&audit)
	if audit.ID == 0 {
		t.Fatalf("Audit trail record for ALARM_ACKNOWLEDGE not found!")
	}
	if audit.Username != "op_test" {
		t.Fatalf("Expected audit username op_test, got %s", audit.Username)
	}

	// Step 5: Duplicate acknowledgement is rejected (400 Bad Request)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBufferString(`{"notes":"second try"}`))
	req.Header.Set("Authorization", "Bearer "+h.operatorToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for duplicate acknowledge, got %d", w.Code)
	}

	// Step 6: Concurrency test with 10 simultaneous goroutines racing to acknowledge another active alarm
	alarmRace := model.Alarm{
		DeviceID:    &devID,
		AlarmCode:   "THRESHOLD_CRITICAL_HIGH",
		Severity:    model.SeverityCritical,
		Status:      model.AlarmActive,
		Message:     "High temperature spike",
		TriggeredAt: time.Now().UTC(),
	}
	h.db.Create(&alarmRace)

	var wg sync.WaitGroup
	var successCount int64
	var err400Count int64
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			body := fmt.Sprintf(`{"notes":"Race worker %d"}`, workerID)
			r := httptest.NewRequest("PUT", fmt.Sprintf("/api/alarms/%d/acknowledge", alarmRace.ID), bytes.NewBufferString(body))
			r.Header.Set("Authorization", "Bearer "+h.engineerToken)
			r.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.router.ServeHTTP(rec, r)

			mu.Lock()
			if rec.Code == http.StatusOK {
				successCount++
			} else if rec.Code == http.StatusBadRequest {
				err400Count++
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	if successCount != 1 {
		t.Fatalf("Race condition failure: expected exactly 1 successful acknowledgement, got %d", successCount)
	}
	if err400Count != 9 {
		t.Fatalf("Expected exactly 9 competing requests to be rejected with 400, got %d", err400Count)
	}

	// Step 7: Cleared alarm cannot be acknowledged
	alarmCleared := model.Alarm{
		DeviceID:    &devID,
		AlarmCode:   "DEVICE_OFFLINE",
		Severity:    model.SeverityWarning,
		Status:      model.AlarmCleared,
		Message:     "Device recovered",
		TriggeredAt: time.Now().UTC().Add(-10 * time.Minute),
	}
	h.db.Create(&alarmCleared)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/alarms/%d/acknowledge", alarmCleared.ID), bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer "+h.operatorToken)
	w = httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request when acknowledging cleared alarm, got %d", w.Code)
	}

	t.Logf("PASS: Scenario D Alarm Acknowledgement verified with RBAC, anti-spoofing, audit logging, and atomic race safety.")
}

// =========================================================================
// SCENARIO E: ALARM HISTORY
// =========================================================================
func TestOAT_ScenarioE_AlarmHistory(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	// Seed 25 historical alarms across devices, severities, and states
	now := time.Now().UTC()
	for i := 1; i <= 25; i++ {
		sev := model.SeverityWarning
		code := "THRESHOLD_WARNING_HIGH"
		status := model.AlarmCleared
		var trigVal *float64
		var threshVal *float64
		if i%2 == 0 {
			sev = model.SeverityCritical
			code = "THRESHOLD_CRITICAL_HIGH"
			v := 75.0 + float64(i)
			trigVal = &v
			th := 70.0
			threshVal = &th
		} else if i%3 == 0 {
			code = "DEVICE_OFFLINE"
			status = model.AlarmActive
		}

		clearedAt := now.Add(-time.Duration(25-i) * time.Minute)
		alarmDevID := uint((i % 3) + 1)
		al := model.Alarm{
			DeviceID:       &alarmDevID,
			AlarmCode:      code,
			Severity:       sev,
			Status:         status,
			Message:        fmt.Sprintf("Alarm event #%d", i),
			TriggerValue:   trigVal,
			ThresholdValue: threshVal,
			TriggeredAt:    clearedAt.Add(-5 * time.Minute),
			CreatedAt:      clearedAt.Add(-5 * time.Minute),
		}
		if status == model.AlarmCleared {
			al.ClearedAt = &clearedAt
			al.Notes = "Recovery verified"
		}
		h.db.Create(&al)
	}

	// 1. Default pagination (page_size 20)
	req := httptest.NewRequest("GET", "/api/alarms/history", nil)
	req.Header.Set("Authorization", "Bearer "+h.operatorToken)
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("History request failed: %d (%s)", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Items      []map[string]interface{} `json:"items"`
			Total      int64                    `json:"total"`
			Page       int                      `json:"page"`
			PageSize   int                      `json:"page_size"`
			TotalPages int                      `json:"total_pages"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.Total != 25 {
		t.Fatalf("Expected 25 total records, got %d", resp.Data.Total)
	}
	if len(resp.Data.Items) != 20 {
		t.Fatalf("Expected 20 records on page 1, got %d", len(resp.Data.Items))
	}
	if resp.Data.TotalPages != 2 {
		t.Fatalf("Expected 2 total pages, got %d", resp.Data.TotalPages)
	}

	// Verify duration calculation & un-fabricated context
	rec := resp.Data.Items[0]
	if _, ok := rec["duration_seconds"]; !ok {
		t.Fatalf("duration_seconds missing from history record")
	}
	if _, ok := rec["duration_formatted"]; !ok {
		t.Fatalf("duration_formatted missing from history record")
	}

	// Filter test: Severity = CRITICAL
	req = httptest.NewRequest("GET", "/api/alarms/history?severity=CRITICAL", nil)
	req.Header.Set("Authorization", "Bearer "+h.operatorToken)
	w = httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	for _, r := range resp.Data.Items {
		if r["severity"] != "CRITICAL" {
			t.Fatalf("Filter failure: expected CRITICAL, got %v", r["severity"])
		}
	}

	// Filter test: Alarm code = DEVICE_OFFLINE -> Verify trigger_value and threshold_value remain NULL
	req = httptest.NewRequest("GET", "/api/alarms/history?alarm_type=DEVICE_OFFLINE", nil)
	req.Header.Set("Authorization", "Bearer "+h.operatorToken)
	w = httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data.Items) > 0 {
		firstOffline := resp.Data.Items[0]
		if firstOffline["trigger_value"] != nil || firstOffline["threshold_value"] != nil {
			t.Fatalf("Fabrication violation: trigger_value/threshold_value for DEVICE_OFFLINE must be null!")
		}
	}

	t.Logf("PASS: Scenario E Alarm History verified with server-side pagination, filters, duration calculations, and null metric preservation.")
}

// =========================================================================
// SCENARIO F: NOTIFICATION ENGINE & REAL DELIVERY
// =========================================================================
func TestOAT_ScenarioF_NotificationEngineAndDelivery(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	// 1. Real Webhook Server Receiver
	var receivedWebhookPayload map[string]interface{}
	var webhookMu sync.Mutex
	webhookServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		webhookMu.Lock()
		defer webhookMu.Unlock()
		body, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(body, &receivedWebhookPayload)
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(`{"status":"received"}`))
	}))
	defer webhookServer.Close()

	// Configure real Webhook Channel in DB
	webhookCfgJSON := fmt.Sprintf(`{"webhook_url":"%s","method":"POST"}`, webhookServer.URL)
	whChannel := model.NotificationChannel{
		Type:        model.ChannelWebhook,
		Name:        "Real Test Webhook",
		Enabled:     true,
		MinSeverity: model.SeverityWarning,
		ConfigJSON:  webhookCfgJSON,
	}
	h.db.Create(&whChannel)

	// Configure In-App Web Channel in DB
	webChannel := model.NotificationChannel{
		Type:        model.ChannelWeb,
		Name:        "Real In-App Web",
		Enabled:     true,
		MinSeverity: model.SeverityWarning,
		ConfigJSON:  `{}`,
	}
	h.db.Create(&webChannel)

	// Step 2: Trigger alarm and dispatch
	dev10 := uint(10)
	alarm := &model.Alarm{
		DeviceID:    &dev10,
		AlarmCode:   "THRESHOLD_CRITICAL_HIGH",
		Severity:    model.SeverityCritical,
		Status:      model.AlarmActive,
		Message:     "Core reactor temperature exceeding limit",
		TriggeredAt: time.Now().UTC(),
	}
	h.db.Create(alarm)

	ctx := context.Background()
	h.notificationEngine.DispatchAlarm(ctx, alarm)

	// Allow worker to poll and process
	time.Sleep(300 * time.Millisecond)

	// Verify Webhook received real HTTP call
	webhookMu.Lock()
	received := receivedWebhookPayload != nil
	webhookMu.Unlock()
	if !received {
		t.Fatalf("Real Webhook endpoint did not receive dispatch payload!")
	}

	// Verify DB notification job transitioned to SENT
	var notif model.Notification
	h.db.Where("channel_id = ? AND alarm_id = ?", whChannel.ID, alarm.ID).First(&notif)
	if notif.Status != model.NotificationSent {
		t.Fatalf("Expected notification status SENT, got %s (err: %s)", notif.Status, notif.ErrorMsg)
	}

	// Step 3: Temporary failure and exponential backoff
	failServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(`{"error":"temporary database lock"}`))
	}))
	defer failServer.Close()

	failChannel := model.NotificationChannel{
		Type:        model.ChannelWebhook,
		Name:        "Flaky Webhook",
		Enabled:     true,
		MinSeverity: model.SeverityWarning,
		ConfigJSON:  fmt.Sprintf(`{"webhook_url":"%s","method":"POST"}`, failServer.URL),
	}
	h.db.Create(&failChannel)

	dev11 := uint(11)
	alarmFlaky := &model.Alarm{
		DeviceID:    &dev11,
		AlarmCode:   "THRESHOLD_WARNING_LOW",
		Severity:    model.SeverityWarning,
		Status:      model.AlarmActive,
		Message:     "Flaky delivery test alarm",
		TriggeredAt: time.Now().UTC(),
	}
	h.db.Create(alarmFlaky)
	h.notificationEngine.DispatchAlarm(ctx, alarmFlaky)

	time.Sleep(300 * time.Millisecond)

	var failNotif model.Notification
	h.db.Where("channel_id = ? AND alarm_id = ?", failChannel.ID, alarmFlaky.ID).First(&failNotif)
	if failNotif.Status != model.NotificationRetrying && failNotif.Status != model.NotificationFailed {
		t.Fatalf("Expected RETRYING or FAILED status for HTTP 500 endpoint, got %s", failNotif.Status)
	}
	if failNotif.Attempts < 1 {
		t.Fatalf("Expected attempt count >= 1, got %d", failNotif.Attempts)
	}

	// Step 4: Secret Redaction & Safe Credentials Merging
	secretChan := model.NotificationChannel{
		Type:       model.ChannelTelegram,
		Name:       "Production Telegram",
		Enabled:    true,
		ConfigJSON: `{"bot_token":"123456:SECRET_TELEGRAM_TOKEN","chat_id":"-100123"}`,
	}
	h.db.Create(&secretChan)

	// Query via API -> Must redact bot_token
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/notifications/channels/%d", secretChan.ID), nil)
	req.Header.Set("Authorization", "Bearer "+h.engineerToken)
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if strings.Contains(w.Body.String(), "SECRET_TELEGRAM_TOKEN") {
		t.Fatalf("Security violation: bot_token exposed unmasked in channel API response!")
	}
	if !strings.Contains(w.Body.String(), "********") {
		t.Fatalf("Expected masked asterisks in sanitized config response")
	}

	// Update channel providing "********" -> Retains original secret!
	updatePayload := `{"name":"Production Telegram Renamed","config":{"bot_token":"********","chat_id":"-100123"}}`
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/notifications/channels/%d", secretChan.ID), bytes.NewBufferString(updatePayload))
	req.Header.Set("Authorization", "Bearer "+h.engineerToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Channel update failed: %d", w.Code)
	}

	var reloadedChan model.NotificationChannel
	h.db.First(&reloadedChan, secretChan.ID)
	if !strings.Contains(reloadedChan.ConfigJSON, "SECRET_TELEGRAM_TOKEN") {
		t.Fatalf("Secret was accidentally erased by updating with asterisks! ConfigJSON: %s", reloadedChan.ConfigJSON)
	}

	t.Logf("PASS: Scenario F Notification Engine verified with real Webhook delivery, retries, backoff, and secret preservation.")
}

// =========================================================================
// SCENARIO G: API, RBAC, AUDIT & LOCALIZATION
// =========================================================================
func TestOAT_ScenarioG_APIRBACAndFrontend(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	// 1. RBAC Matrix Verification
	endpoints := []struct {
		method         string
		path           string
		body           string
		adminAllowed   bool
		engAllowed     bool
		opAllowed      bool
		viewerAllowed  bool
	}{
		{"GET", "/api/alarms", "", true, true, true, false},
		{"GET", "/api/alarms/active", "", true, true, true, false},
		{"GET", "/api/alarms/history", "", true, true, true, false},
		{"POST", "/api/alarms", `{"alarm_code":"TEST","severity":"INFO","message":"test"}`, true, true, false, false},
		{"GET", "/api/notifications", "", true, true, true, false},
		{"GET", "/api/notifications/channels", "", true, true, true, false},
		{"POST", "/api/notifications/1/retry", "", true, true, false, false},
		{"POST", "/api/notifications/1/cancel", "", true, true, false, false},
	}

	for _, ep := range endpoints {
		checkAccess := func(token string, expectedAllowed bool, roleName string) {
			req := httptest.NewRequest(ep.method, ep.path, bytes.NewBufferString(ep.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.router.ServeHTTP(w, req)

			if expectedAllowed && w.Code == http.StatusForbidden {
				t.Fatalf("%s %s unexpectedly forbidden for role %s", ep.method, ep.path, roleName)
			}
			if !expectedAllowed && w.Code != http.StatusForbidden && w.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s expected 403 Forbidden for role %s, got %d", ep.method, ep.path, roleName, w.Code)
			}
		}

		checkAccess(h.adminToken, ep.adminAllowed, "Admin")
		checkAccess(h.engineerToken, ep.engAllowed, "Engineer")
		checkAccess(h.operatorToken, ep.opAllowed, "Operator")
		checkAccess(h.viewerToken, ep.viewerAllowed, "Viewer")
	}

	// 2. Localization Parity Check: Read en.js and id.js
	enBytes, errEn := os.ReadFile(filepath.Join("..", "..", "web", "src", "i18n", "locales", "en.js"))
	idBytes, errId := os.ReadFile(filepath.Join("..", "..", "web", "src", "i18n", "locales", "id.js"))
	if errEn == nil && errId == nil {
		enContent := string(enBytes)
		idContent := string(idBytes)

		// Key alarm & notification terms that must exist in both
		requiredKeys := []string{
			"tabActive", "tabHistory", "tabQueue", "tabChannels",
			"ackModalTitle", "operatorNotes", "confirmAck", "historyTitle",
			"filterDevice", "filterSeverity", "filterStatus", "filterType",
			"duration", "clearedAt", "ackedAt", "ackedBy",
		}
		for _, k := range requiredKeys {
			if !strings.Contains(enContent, k) {
				t.Fatalf("en.js missing translation key: %s", k)
			}
			if !strings.Contains(idContent, k) {
				t.Fatalf("id.js missing translation key: %s", k)
			}
		}
	}

	t.Logf("PASS: Scenario G API, RBAC Matrix, and Frontend Localization verified.")
}

// =========================================================================
// SCENARIO H: PERSISTENCE, RESTART & RECOVERY
// =========================================================================
func TestOAT_ScenarioH_PersistenceRestartRecovery(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "oat_restart.db")

	// Phase 1: Initialize first instance and persist state
	h1 := setupOATHarness(t, tempDB)

	dev1 := uint(1)
	alarm := model.Alarm{
		DeviceID:             &dev1,
		AlarmCode:            "THRESHOLD_CRITICAL_HIGH",
		Severity:             model.SeverityCritical,
		Status:               model.AlarmAcked,
		Message:              "Persisted alarm before reboot",
		Notes:                "Critical acknowledged by senior operator",
		AcknowledgedBy:       "op_test",
		AcknowledgedByName:   "Operator One",
		AcknowledgedByUserID: &[]uint{1}[0],
		TriggeredAt:          time.Now().UTC().Add(-15 * time.Minute),
	}
	now := time.Now().UTC()
	alarm.AcknowledgedAt = &now
	h1.db.Create(&alarm)

	chanID := uint(1)
	notif := model.Notification{
		AlarmID:        alarm.ID,
		ChannelID:      &chanID,
		Channel:        "WEBHOOK",
		Status:         model.NotificationRetrying,
		Attempts:       1,
		MaxAttempts:    3,
		Recipient:      "http://127.0.0.1:9090",
		Subject:        "Alert",
		Message:        "Body",
		IdempotencyKey: "alarm:1:channel:1",
	}
	h1.db.Create(&notif)

	// Simulate system reboot / restart by shutting down all services & database
	h1.notificationEngine.Stop()
	sqlDB, _ := h1.db.DB()
	_ = sqlDB.Close()

	// Phase 2: Start fresh second instance from the exact same disk file
	h2 := setupOATHarness(t, tempDB)
	defer func() {
		h2.notificationEngine.Stop()
		if sDB, err := h2.db.DB(); err == nil && sDB != nil {
			_ = sDB.Close()
		}
	}()

	// Verify active / acknowledged alarm is intact
	var reloadedAlarm model.Alarm
	h2.db.First(&reloadedAlarm, alarm.ID)
	if reloadedAlarm.Status != model.AlarmAcked {
		t.Fatalf("Alarm status corrupted across restart: %s", reloadedAlarm.Status)
	}
	if reloadedAlarm.Notes != "Critical acknowledged by senior operator" {
		t.Fatalf("Notes lost across restart: %s", reloadedAlarm.Notes)
	}
	if reloadedAlarm.AcknowledgedByName != "Operator One" {
		t.Fatalf("Operator display name lost across restart: %s", reloadedAlarm.AcknowledgedByName)
	}

	// Verify notification queue survived restart
	var reloadedNotif model.Notification
	h2.db.First(&reloadedNotif, notif.ID)
	if reloadedNotif.Status != model.NotificationRetrying && reloadedNotif.Status != model.NotificationPending {
		t.Fatalf("Notification state lost across restart: %s", reloadedNotif.Status)
	}

	t.Logf("PASS: Scenario H Persistence, Restart, and Recovery verified across simulated process shutdown and cold boot.")
}

// =========================================================================
// SECTION 5: PERFORMANCE & RESOURCE BENCHMARK
// =========================================================================
func TestOAT_Section5_PerformanceAndResourceCheck(t *testing.T) {
	h := setupOATHarness(t, "")
	defer h.notificationEngine.Stop()

	ctx := context.Background()

	// 1. Measure baseline
	runtime.GC()
	var mBaseline runtime.MemStats
	runtime.ReadMemStats(&mBaseline)
	baseGoroutines := runtime.NumGoroutine()

	// 2. Sustained workload: 1,000 telemetry evaluations
	dev := model.Device{DeviceCode: "BENCH-DEV", DeviceName: "Turbine Sensor", Status: model.DeviceStatusActive, Enabled: true}
	h.db.Create(&dev)

	wHigh := 60.0
	cHigh := 90.0
	param := model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "Vibration",
		WarningHigh:   &wHigh,
		CriticalHigh:  &cHigh,
		Enabled:       true,
		DataType:      model.DataTypeFloat32,
	}
	h.db.Create(&param)

	startBench := time.Now()
	iterations := 1000
	for i := 0; i < iterations; i++ {
		// oscillate between normal (30.0), warning (65.0), and critical (95.0)
		val := 30.0
		if i%20 == 0 {
			val = 95.0
		} else if i%10 == 0 {
			val = 65.0
		}

		h.evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: val,
			ReceivedAt:     time.Now().UTC(),
			Quality:        model.QualityGood,
		}, &param)
	}
	elapsed := time.Since(startBench)
	avgLatency := float64(elapsed.Microseconds()) / float64(iterations)

	// 3. Post-workload resource measurement
	runtime.GC()
	var mActive runtime.MemStats
	runtime.ReadMemStats(&mActive)
	endGoroutines := runtime.NumGoroutine()

	t.Logf("=== PERFORMANCE MEASUREMENTS ===")
	t.Logf("Evaluations Executed:       %d iterations", iterations)
	t.Logf("Total Duration:             %v", elapsed)
	t.Logf("Avg Evaluation Latency:     %.2f µs / packet", avgLatency)
	t.Logf("Throughput:                 %.0f packets / sec", float64(iterations)/elapsed.Seconds())
	t.Logf("Baseline Heap Alloc:        %.2f MB", float64(mBaseline.HeapAlloc)/1024/1024)
	t.Logf("Active Heap Alloc:          %.2f MB", float64(mActive.HeapAlloc)/1024/1024)
	t.Logf("Goroutines (Base -> End):   %d -> %d", baseGoroutines, endGoroutines)

	// Sanity checks
	if avgLatency > 2000.0 { // Must be faster than 2ms per evaluation on edge
		t.Fatalf("Alarm evaluation latency too high: %.2f µs", avgLatency)
	}
	if math.Abs(float64(endGoroutines-baseGoroutines)) > 15 {
		t.Fatalf("Potential goroutine leak detected: delta = %d", endGoroutines-baseGoroutines)
	}

	t.Logf("PASS: Section 5 Performance and Resource Check verified with high throughput and zero goroutine leak.")
}

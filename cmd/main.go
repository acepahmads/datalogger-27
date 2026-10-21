package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"datalogger/internal/communication"
	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/logger"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/scheduler"
	"datalogger/internal/service"
	"datalogger/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// 1. Load Configuration
	configPath := "config.json"
	if envCfg := os.Getenv("DATALOGGER_CONFIG"); envCfg != "" {
		configPath = envCfg
	} else if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Fallback: look next to executable
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			altPath := filepath.Join(exeDir, "config.json")
			if _, statErr := os.Stat(altPath); statErr == nil {
				configPath = altPath
			}
		}
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Warning: failed to load config file (%s): %v. Using defaults.\n", configPath, err)
		cfg = config.DefaultConfig()
	}

	// 2. Initialize Structured Logger
	log, err := logger.Init(cfg.LogDir, cfg.LogLevel)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
	}
	defer logger.Close()

	logger.Info("Starting %s (%s)...", cfg.AppName, cfg.Version)
	logger.Info("Environment: %s | Target Host: %s:%s", cfg.Environment, cfg.Host, cfg.Port)

	// 3. Initialize MariaDB Database (with retry loop for edge cold-boot resilience)
	logger.Info("Connecting to MariaDB database (%s:%s/%s)...", cfg.DBHost, cfg.DBPort, cfg.DBName)
	var db *gorm.DB
	for attempt := 1; attempt <= 10; attempt++ {
		db, err = database.Init(cfg)
		if err == nil {
			break
		}
		logger.Warn("Database connection attempt %d/10 failed: %v. Retrying in 2s...", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Error("Database initialization permanently failed: %v", err)
		os.Exit(1)
	}

	// 4. Seed Database (10 phases, tasks, admin user, sample devices)
	logger.Info("Verifying schemas and seeding initial database records...")
	if err := database.Seed(db); err != nil {
		logger.Error("Database seed error: %v", err)
	}

	// 5. Initialize Repositories
	deviceRepo := repository.NewDeviceRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	// 6. Initialize Services
	authService := service.NewAuthService(systemRepo, cfg)
	middleware.SetPermissionChecker(authService.HasPermission)
	websocket.SetTokenValidator(func(token string) (string, string, error) {
		claims, err := authService.ValidateToken(token)
		if err != nil || claims == nil {
			return "", "", err
		}
		return claims.Role, claims.Username, nil
	})
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)

	// 7. Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()
	logger.Info("WebSocket hub initialized for real-time edge telemetry")

	// 7b. Initialize Telemetry Ingestion Pipeline with Persistent Queue (Phase 4.2)
	telemetryCfg := service.DefaultTelemetryConfig()
	telemetryCfg.QueueEnabled = cfg.QueueEnabled
	telemetryCfg.QueueDir = cfg.QueueDir
	telemetryCfg.QueueMaxSizeBytes = cfg.QueueMaxSizeBytes
	telemetryCfg.BatchSize = cfg.QueueBatchSize
	telemetryCfg.QueueSyncMode = cfg.QueueSyncMode
	telemetryCfg.QueueDiskWarnPercent = cfg.QueueDiskWarnPercent

	telemetryService := service.NewTelemetryService(telemetryRepo, hub, telemetryCfg)
	defer telemetryService.Stop()
	systemService.SetTelemetryService(telemetryService)

	// 7c. Initialize Aggregation & Rollup Engine (Phase 3.3)
	aggRepo := repository.NewAggregationRepository(db)
	aggService := service.NewAggregationService(aggRepo, deviceRepo, systemRepo, hub)
	aggService.StartWorker()
	defer aggService.StopWorker()

	// 7d. Initialize Backup & Restore Subsystem (Phase 4.3)
	backupRepo := repository.NewBackupRepository(db)
	dbBackupAdapter := service.NewDatabaseBackupAdapter(db, cfg)
	backupService := service.NewBackupService(db, cfg, backupRepo, dbBackupAdapter, telemetryService.GetWALAdapter(), telemetryService, systemRepo)
	systemService.SetBackupService(backupService)

	backupScheduler := service.NewBackupScheduler(backupService, cfg)
	backupScheduler.Start()
	defer backupScheduler.Stop()

	// 7e. Initialize Retention & Storage Management Subsystem (Phase 4.4)
	retentionRepo := repository.NewRetentionRepository(db)
	retentionService := service.NewRetentionService(db, cfg, retentionRepo, backupRepo, systemRepo, telemetryService, backupService)
	retentionScheduler := service.NewRetentionScheduler(retentionService, cfg)
	retentionScheduler.Start()
	defer retentionScheduler.Stop()

	// 7f. Initialize Alarm Storage Subsystem (Phase 5)
	alarmRepo := repository.NewAlarmRepository(db)
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	alarmService.SetHub(hub)
	systemService.SetAlarmService(alarmService)

	// 7h. Initialize Notification Engine Subsystem (Phase 6 Wave 3)
	notificationRepo := repository.NewNotificationRepository(db)
	notificationEngine := service.NewNotificationEngine(notificationRepo, alarmRepo, systemRepo, service.DefaultNotificationEngineConfig())
	alarmService.SetNotificationEngine(notificationEngine)
	notificationEngine.Start()
	defer notificationEngine.Stop()

	// 7g. Initialize Core Alarm Engine & Threshold Evaluator (Phase 6 Wave 1)
	alarmEvaluator := service.NewAlarmEvaluator(
		context.Background(),
		alarmService,
		service.DefaultAlarmEvaluatorConfig(),
		func(ctx context.Context, paramID uint) (*model.Parameter, error) {
			return deviceRepo.FindParameterByID(paramID)
		},
	)
	alarmEvaluator.SetDeviceStatusUpdater(deviceRepo.SetConnectionStatus)
	telemetryService.SetAlarmEvaluator(alarmEvaluator)
	logger.Info("AlarmEvaluator initialized with active DB alarm synchronization and attached to TelemetryService")

	// 7k. Initialize Outbound Delivery & Integration Subsystem (Phase 8)
	outboundRepo := repository.NewOutboundRepository(db)
	outboundService := service.NewOutboundService(outboundRepo, deviceRepo, nil)
	outboundService.StartWorker()
	defer outboundService.StopWorker()
	outboundHandler := handler.NewOutboundHandler(outboundService)

	// 8. Initialize Scheduler
	sched := scheduler.NewScheduler(systemRepo)
	sched.Start()
	defer sched.Stop()

	// 8b. Initialize Communication Engine (Phase 2.2)
	connManager := communication.NewConnectionManager(deviceService, communication.DefaultAdapterFactory)
	pollingEngine := communication.NewPollingEngine(connManager, deviceService, hub)
	pollingEngine.SetTelemetryService(telemetryService)
	pollingEngine.SetAlarmEvaluator(alarmEvaluator)
	pollingEngine.Start()
	defer pollingEngine.Stop()

	// 9. Setup Gin HTTP Engine
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())

	// 10. Handlers
	authHandler := handler.NewAuthHandler(authService)
	devHandler := handler.NewDevHandler(phaseService, hub)
	systemHandler := handler.NewSystemHandler(systemService)
	systemHandler.SetOutboundService(outboundService)
	deviceHandler := handler.NewDeviceHandler(deviceService, pollingEngine)
	commHandler := handler.NewCommunicationHandler(connManager, pollingEngine, deviceService)
	telemetryHandler := handler.NewTelemetryHandler(telemetryService)
	aggHandler := handler.NewAggregationHandler(aggService)
	backupHandler := handler.NewBackupHandler(backupService, backupScheduler)
	retentionHandler := handler.NewRetentionHandler(retentionService, retentionScheduler)
	notificationHandler := handler.NewNotificationHandler(notificationEngine, notificationRepo)
	reportService := service.NewReportService(db, systemRepo)
	reportHandler := handler.NewReportHandler(reportService)
	userHandler := handler.NewUserHandler(authService)
	roleHandler := handler.NewRoleHandler(authService)
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo, systemRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	// Phase 7.5 — ISPU Module (Permen LHK No. 14 Tahun 2020)
	ispuRepo := repository.NewISPURepository(db)
	ispuService := service.NewISPUService(db, ispuRepo, systemRepo)
	ispuService.SetHub(hub)
	ispuService.StartWorker(context.Background())
	ispuHandler := handler.NewISPUHandler(ispuService)

	// Startup reconciliation for active AQMS stations (catch up any pending aggregation periods cleanly)
	go func() {
		time.Sleep(2 * time.Second) // wait for DB & devices to stabilize
		ctx := context.Background()
		_, _ = ispuService.ReconcilePendingCalculations(ctx)
	}()

	// Automatically trigger ISPU calculation whenever an eligible parameter aggregation completes
	aggService.SetOnAggregationCompleted(func(ctx context.Context, event service.AggregationCompletedEvent) {
		ispuService.HandleAggregationCompleted(ctx, event)
	})
	aggService.SetOnBucketCompleted(func(ctx context.Context, deviceID uint, bucketEnd time.Time) {
		ispuService.EnqueueCalculation(deviceID, bucketEnd, "AGGREGATION_COMPLETED")
		_ = outboundService.HandleBucketCompleted(ctx, deviceID, bucketEnd)
	})

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWS(hub, c)
	})

	// Public & Protected API routes
	api := r.Group("/api")
	{
		// Auth
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/logout", middleware.JWTAuth(authService), authHandler.Logout)
		api.POST("/auth/switch-account", middleware.JWTAuth(authService), middleware.RequireRole("SUPERADMIN"), authHandler.SwitchAccount)
		api.GET("/auth/me", middleware.JWTAuth(authService), authHandler.Me)
		api.GET("/users/me", middleware.JWTAuth(authService), authHandler.Me)

		// Development Phases & Tasks (Protected & Authorized — SUPERADMIN only)
		devGroup := api.Group("")
		devGroup.Use(middleware.JWTAuth(authService))
		{
			devGroup.GET("/phases", middleware.RequirePermission("development.view"), devHandler.GetPhases)
			devGroup.GET("/phases/:id", middleware.RequirePermission("development.view"), devHandler.GetPhaseByID)
			devGroup.GET("/progress", middleware.RequirePermission("development.view"), devHandler.GetProgress)
			devGroup.GET("/tasks", middleware.RequirePermission("development.view"), devHandler.GetTasks)
			devGroup.POST("/tasks", middleware.RequirePermission("development.manage"), devHandler.CreateTask)
			devGroup.GET("/tasks/:id", middleware.RequirePermission("development.view"), devHandler.GetTaskByID)
			devGroup.PUT("/tasks/:id", middleware.RequirePermission("development.manage"), devHandler.UpdateTask)
			devGroup.DELETE("/tasks/:id", middleware.RequirePermission("development.manage"), devHandler.DeleteTask)
			devGroup.GET("/activity", middleware.RequirePermission("development.view"), devHandler.GetActivity)
		}

		// System Diagnostics & Health
		api.GET("/system/status", systemHandler.GetStatus)
		api.GET("/system/health", systemHandler.GetHealth)
		api.GET("/system/network", systemHandler.GetNetworkStatus)
		api.GET("/system/serial-ports", systemHandler.GetSerialPorts)
		api.GET("/system/serial-ports/details", systemHandler.GetDetailedSerialPorts)
		api.GET("/devices/serial-ports", systemHandler.GetSerialPorts)
		api.GET("/devices/serial-ports/details", systemHandler.GetDetailedSerialPorts)
		api.GET("/system/branding", systemHandler.GetBranding)
		api.POST("/system/branding", systemHandler.SaveBranding)
		api.PUT("/system/branding", systemHandler.SaveBranding)
		api.GET("/system/location", systemHandler.GetStationLocation)
		api.GET("/system/location/ip-estimate", systemHandler.GetIPGeolocationEstimate)
		api.GET("/system/location/search", systemHandler.SearchLocation)
		api.PUT("/system/location", middleware.JWTAuth(authService), middleware.RequirePermission("config.manage"), systemHandler.SaveStationLocation)
		api.POST("/system/location", middleware.JWTAuth(authService), middleware.RequirePermission("config.manage"), systemHandler.SaveStationLocation)

		// Phase 2.1 — Device Management & Parameters (Authenticated & Authorized)
		devicesGroup := api.Group("/devices")
		devicesGroup.Use(middleware.JWTAuth(authService))
		{
			// Read endpoints (device.view)
			devicesGroup.GET("", middleware.RequirePermission("device.view"), deviceHandler.ListDevices)
			devicesGroup.GET("/:id", middleware.RequirePermission("device.view"), deviceHandler.GetDeviceByID)
			devicesGroup.GET("/:id/status", middleware.RequirePermission("device.view"), systemHandler.GetDeviceStatus)
			devicesGroup.GET("/:id/activity", middleware.RequirePermission("device.view"), deviceHandler.GetDeviceActivity)

			// Create endpoint (device.create)
			devicesGroup.POST("", middleware.RequirePermission("device.create"), deviceHandler.CreateDevice)

			// Update endpoints (device.update)
			devicesGroup.PUT("/:id", middleware.RequirePermission("device.update"), deviceHandler.UpdateDevice)
			devicesGroup.PUT("/:id/enable", middleware.RequirePermission("device.update"), deviceHandler.ToggleDeviceEnabled)
			devicesGroup.PUT("/:id/status", middleware.RequirePermission("device.update"), deviceHandler.UpdateDeviceStatus)
			devicesGroup.PUT("/:id/connection", middleware.RequirePermission("device.update"), deviceHandler.UpdateConnection)

			// Delete endpoint (device.delete)
			devicesGroup.DELETE("/:id", middleware.RequirePermission("device.delete"), deviceHandler.DeleteDevice)

			// Parameters (inside device)
			devicesGroup.GET("/:id/parameters", middleware.RequirePermission("device.view"), deviceHandler.ListParameters)
			devicesGroup.GET("/:id/parameters/:paramId", middleware.RequirePermission("device.view"), deviceHandler.GetParameter)
			devicesGroup.POST("/:id/parameters", middleware.RequirePermission("device.manage"), deviceHandler.CreateParameter)
			devicesGroup.PUT("/:id/parameters/:paramId", middleware.RequirePermission("device.manage"), deviceHandler.UpdateParameter)
			devicesGroup.DELETE("/:id/parameters/:paramId", middleware.RequirePermission("device.manage"), deviceHandler.DeleteParameter)
			devicesGroup.PUT("/:id/parameters/:paramId/enable", middleware.RequirePermission("device.manage"), deviceHandler.ToggleParameterEnabled)
			devicesGroup.POST("/parameters/validate-formula", middleware.RequirePermission("device.view"), deviceHandler.ValidateFormula)
			devicesGroup.POST("/:id/parameters/validate-formula", middleware.RequirePermission("device.view"), deviceHandler.ValidateFormula)
			devicesGroup.GET("/:id/parameters/:paramId/quality", middleware.RequirePermission("device.view"), deviceHandler.GetParameterQuality)
			devicesGroup.PUT("/:id/parameters/:paramId/quality", middleware.RequirePermission("device.manage"), deviceHandler.UpdateParameterQuality)

			// Phase 2.2 — Communication Controls & Diagnostics
			devicesGroup.POST("/:id/communication/connect", middleware.RequirePermission("device.communication.manage"), commHandler.Connect)
			devicesGroup.POST("/:id/communication/disconnect", middleware.RequirePermission("device.communication.manage"), commHandler.Disconnect)
			devicesGroup.POST("/:id/communication/reconnect", middleware.RequirePermission("device.communication.manage"), commHandler.Reconnect)
			devicesGroup.POST("/:id/communication/test", middleware.RequirePermission("device.communication.test"), commHandler.TestConnection)
			devicesGroup.GET("/:id/communication/status", middleware.RequirePermission("device.communication.view"), commHandler.GetStatus)
			devicesGroup.POST("/:id/parameters/:paramId/test-read", middleware.RequirePermission("device.communication.test"), commHandler.TestReadParameter)

			// Phase 3.1 & 3.2 — Telemetry Data Pipeline (Authenticated & Authorized)
			devicesGroup.GET("/:id/telemetry/latest", middleware.RequirePermission("device.view"), telemetryHandler.GetLatestForDevice)
			devicesGroup.GET("/:id/parameters/:paramId/telemetry/latest", middleware.RequirePermission("device.view"), telemetryHandler.GetLatestForParameter)
			devicesGroup.GET("/:id/telemetry/history", middleware.RequirePermission("device.view"), telemetryHandler.GetHistorical)
			devicesGroup.GET("/:id/telemetry/raw", middleware.RequirePermission("telemetry.raw"), telemetryHandler.GetRawTelemetry)
			devicesGroup.GET("/:id/telemetry/quality-summary", middleware.RequirePermission("device.view"), telemetryHandler.GetDeviceQualitySummary)

			// Phase 3.3 — Historical Downsampling
			devicesGroup.GET("/:id/telemetry/downsampled", middleware.RequirePermission("device.view"), aggHandler.GetDownsampledHistory)
		}

		// Dynamic Parameter Categories & Layout (Authenticated & Authorized)
		categoriesGroup := api.Group("/categories")
		categoriesGroup.Use(middleware.JWTAuth(authService))
		{
			categoriesGroup.GET("", middleware.RequirePermission("category.view"), categoryHandler.ListCategories)
			categoriesGroup.GET("/:id", middleware.RequirePermission("category.view"), categoryHandler.GetCategory)
			categoriesGroup.POST("", middleware.RequirePermission("category.manage"), categoryHandler.CreateCategory)
			categoriesGroup.PUT("/reorder", middleware.RequirePermission("category.manage"), categoryHandler.ReorderCategories)
			categoriesGroup.PUT("/:id", middleware.RequirePermission("category.manage"), categoryHandler.UpdateCategory)
			categoriesGroup.PUT("/:id/toggle", middleware.RequirePermission("category.manage"), categoryHandler.ToggleCategory)
			categoriesGroup.DELETE("/:id", middleware.RequirePermission("category.manage"), categoryHandler.DeleteCategory)
			categoriesGroup.POST("/:id/parameters", middleware.RequirePermission("category.assign"), categoryHandler.AssignParameters)
		}

		// Phase 3.3 — Aggregations & Rollups (Internal & Management)
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

		// Phase 3.3 — Customer Aggregated Data API (Customer-Facing & Identifier Query)
		customerGroup := api.Group("/customer")
		customerGroup.Use(middleware.JWTAuth(authService))
		{
			customerGroup.GET("/aggregated-data/:identifier", middleware.RequirePermission("device.view"), aggHandler.GetCustomerAggregatedData)
		}

		// Telemetry Pipeline & Unified Analysis (Phase 3.1, 3.2, UX Refactor)
		telemetryGroup := api.Group("/telemetry")
		telemetryGroup.Use(middleware.JWTAuth(authService))
		{
			telemetryGroup.GET("/history", middleware.RequirePermission("device.view"), telemetryHandler.GetAllHistorical)
			telemetryGroup.GET("/raw", middleware.RequirePermission("telemetry.raw"), telemetryHandler.GetAllRawTelemetry)
			telemetryGroup.GET("/downsampled", middleware.RequirePermission("device.view"), aggHandler.GetDownsampledHistory)
			telemetryGroup.GET("/metrics", middleware.RequirePermission("device.view"), telemetryHandler.GetMetrics)
			telemetryGroup.GET("/quality-summary", middleware.RequirePermission("device.view"), telemetryHandler.GetQualitySummary)
		}

		// Operational Telemetry Data
		api.GET("/data", systemHandler.GetData)
		api.GET("/data/latest", systemHandler.GetLatestData)
		api.GET("/data/trend", systemHandler.GetDataTrend)

		// Phase 5 & 6 — Alarms Storage, Lifecycle & History (Authenticated & Authorized)
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

		// Logs
		api.GET("/logs", systemHandler.GetLogs)
		api.GET("/audit-trails", systemHandler.GetAuditTrails)
		api.GET("/communication-logs", systemHandler.GetCommunicationLogs)

		// Database Management & Monitoring
		api.GET("/system/database/status", systemHandler.GetDatabaseStatus)
		api.POST("/system/database/test", systemHandler.TestDatabaseConnection)
		api.POST("/system/database/config", systemHandler.SaveDatabaseConfig)

		// Phase 4.3 — Backup & Restore Management (Authenticated & Authorized)
		backupGroup := api.Group("/backups")
		backupGroup.Use(middleware.JWTAuth(authService))
		{
			backupGroup.GET("", middleware.RequirePermission("backup.view"), backupHandler.ListBackups)
			backupGroup.GET("/status", middleware.RequirePermission("backup.view"), backupHandler.GetStatus)
			backupGroup.GET("/schedule", middleware.RequirePermission("backup.view"), backupHandler.GetSchedule)
			backupGroup.PUT("/schedule", middleware.RequirePermission("backup.manage"), backupHandler.UpdateSchedule)
			backupGroup.GET("/jobs/:jobId", middleware.RequirePermission("backup.view"), backupHandler.GetJob)
			backupGroup.GET("/:id", middleware.RequirePermission("backup.view"), backupHandler.GetBackup)
			backupGroup.GET("/:id/restore-preview", middleware.RequirePermission("backup.view"), backupHandler.GetRestorePreview)
			backupGroup.GET("/:id/download", middleware.RequirePermission("backup.view"), backupHandler.DownloadBackup)
			backupGroup.POST("", middleware.RequirePermission("backup.create"), backupHandler.CreateBackup)
			backupGroup.POST("/:id/validate", middleware.RequirePermission("backup.validate"), backupHandler.ValidateBackup)
			backupGroup.POST("/:id/restore", middleware.RequirePermission("backup.restore"), backupHandler.RestoreBackup)
			backupGroup.DELETE("/:id", middleware.RequirePermission("backup.manage"), backupHandler.DeleteBackup)
		}

		// Phase 4.4 — Retention & Storage Management (Authenticated & Authorized)
		retentionGroup := api.Group("/retention")
		retentionGroup.Use(middleware.JWTAuth(authService))
		{
			retentionGroup.GET("/policies", middleware.RequirePermission("retention.view"), retentionHandler.ListPolicies)
			retentionGroup.GET("/policies/:id", middleware.RequirePermission("retention.view"), retentionHandler.GetPolicy)
			retentionGroup.PUT("/policies/:id", middleware.RequirePermission("retention.manage"), retentionHandler.UpdatePolicy)
			retentionGroup.PUT("/policies/:id/toggle", middleware.RequirePermission("retention.manage"), retentionHandler.TogglePolicy)
			retentionGroup.POST("/policies/:id/dry-run", middleware.RequirePermission("retention.view"), retentionHandler.DryRun)
			retentionGroup.POST("/policies/:id/execute", middleware.RequirePermission("retention.execute"), retentionHandler.ExecutePolicy)
			retentionGroup.POST("/housekeeping/trigger", middleware.RequirePermission("retention.execute"), retentionHandler.TriggerHousekeeping)
			retentionGroup.GET("/storage", middleware.RequirePermission("retention.view"), retentionHandler.GetStorageOverview)
			retentionGroup.GET("/history", middleware.RequirePermission("retention.view"), retentionHandler.GetHistory)
		}

		// Phase 6 Wave 3 — Notification Engine & Channels (Authenticated & Authorized)
		notifGroup := api.Group("/notifications")
		notifGroup.Use(middleware.JWTAuth(authService))
		{
			notifGroup.GET("", middleware.RequirePermission("notification.view"), notificationHandler.ListNotifications)
			notifGroup.GET("/metrics", middleware.RequirePermission("notification.view"), notificationHandler.GetMetrics)
			notifGroup.GET("/:id", middleware.RequirePermission("notification.view"), notificationHandler.GetNotification)
			notifGroup.POST("/:id/retry", middleware.RequirePermission("notification.manage"), notificationHandler.RetryNotification)
			notifGroup.POST("/:id/cancel", middleware.RequirePermission("notification.manage"), notificationHandler.CancelNotification)
			notifGroup.GET("/channels", middleware.RequirePermission("notification.view"), notificationHandler.ListChannels)
			notifGroup.GET("/channels/:id", middleware.RequirePermission("notification.view"), notificationHandler.GetChannel)
			notifGroup.PUT("/channels/:id", middleware.RequirePermission("notification.manage"), notificationHandler.UpdateChannel)
			notifGroup.POST("/channels/:id/test", middleware.RequirePermission("notification.manage"), notificationHandler.TestSend)
		}

		// Phase 7 Package C1 — Operational Reporting (Task #79)
		reportsGroup := api.Group("/reports")
		reportsGroup.Use(middleware.JWTAuth(authService))
		{
			reportsGroup.GET("/preview", middleware.RequirePermission("report.view"), reportHandler.PreviewReport)
			reportsGroup.POST("/preview", middleware.RequirePermission("report.view"), reportHandler.PreviewReport)
			reportsGroup.GET("/generate", middleware.RequirePermission("report.generate"), reportHandler.GenerateReport)
			reportsGroup.POST("/generate", middleware.RequirePermission("report.generate"), reportHandler.GenerateReport)
		}

		// Phase 7 Package C2 — User Management (Task #80)
		usersGroup := api.Group("/users")
		usersGroup.Use(middleware.JWTAuth(authService))
		{
			usersGroup.GET("", middleware.RequirePermission("user.view"), userHandler.ListUsers)
			usersGroup.GET("/:id", middleware.RequirePermission("user.view"), userHandler.GetUser)
			usersGroup.POST("", middleware.RequirePermission("user.manage"), userHandler.CreateUser)
			usersGroup.PUT("/:id", middleware.RequirePermission("user.manage"), userHandler.UpdateUser)
			usersGroup.PUT("/:id/enable", middleware.RequirePermission("user.manage"), userHandler.ToggleUserActive)
			usersGroup.PUT("/:id/password", middleware.RequirePermission("user.manage"), userHandler.ResetPassword)
			usersGroup.DELETE("/:id", middleware.RequirePermission("user.manage"), userHandler.DeleteUser)
		}

		// Phase 7 Package C3 — Role & Permission Management (Tasks #81, #82)
		rolesGroup := api.Group("/roles")
		rolesGroup.Use(middleware.JWTAuth(authService))
		{
			rolesGroup.GET("", middleware.RequirePermission("role.view"), roleHandler.ListRoles)
			rolesGroup.GET("/:id", middleware.RequirePermission("role.view"), roleHandler.GetRole)
			rolesGroup.POST("", middleware.RequirePermission("role.manage"), roleHandler.CreateRole)
			rolesGroup.PUT("/:id", middleware.RequirePermission("role.manage"), roleHandler.UpdateRole)
			rolesGroup.PUT("/:id/permissions", middleware.RequirePermission("role.manage"), roleHandler.UpdateRolePermissions)
			rolesGroup.DELETE("/:id", middleware.RequirePermission("role.manage"), roleHandler.DeleteRole)
		}

		permissionsGroup := api.Group("/permissions")
		permissionsGroup.Use(middleware.JWTAuth(authService))
		{
			permissionsGroup.GET("", middleware.RequirePermission("role.view"), roleHandler.ListPermissions)
		}

		// Phase 7 Package C4 — System Configuration (Task #83)
		configGroup := api.Group("/system/config")
		configGroup.Use(middleware.JWTAuth(authService))
		{
			configGroup.GET("", middleware.RequirePermission("config.view"), systemHandler.GetConfig)
			configGroup.GET("/domain-compatibility", middleware.RequirePermission("config.view"), systemHandler.CheckDomainCompatibility)
			configGroup.PUT("", middleware.RequirePermission("config.manage"), systemHandler.SaveConfig)
			configGroup.POST("", middleware.RequirePermission("config.manage"), systemHandler.SaveConfig)
			configGroup.GET("/location", middleware.RequirePermission("config.view"), systemHandler.GetStationLocation)
			configGroup.PUT("/location", middleware.RequirePermission("config.manage"), systemHandler.SaveStationLocation)
		}

		// Phase 7.5 — Environmental Compliance & ISPU Module (Permen LHK No. 14 Tahun 2020)
		ispuGroup := api.Group("/ispu")
		ispuGroup.Use(middleware.JWTAuth(authService))
		{
			ispuGroup.GET("/latest", middleware.RequirePermission("ispu.view"), ispuHandler.GetLatest)
			ispuGroup.GET("/trend", middleware.RequirePermission("ispu.view"), ispuHandler.GetTrend)
			ispuGroup.GET("/history", middleware.RequirePermission("ispu.view"), ispuHandler.GetHistory)
			ispuGroup.GET("/export/csv", middleware.RequirePermission("ispu.export"), ispuHandler.ExportCSV)
			ispuGroup.GET("/export/excel", middleware.RequirePermission("ispu.export"), ispuHandler.ExportExcel)
			ispuGroup.GET("/mappings", middleware.RequirePermission("ispu.view"), ispuHandler.GetMappings)
			ispuGroup.POST("/mappings", middleware.RequirePermission("ispu.manage_mapping"), ispuHandler.SaveMapping)
			ispuGroup.DELETE("/mappings/:id", middleware.RequirePermission("ispu.manage_mapping"), ispuHandler.DeleteMapping)
			ispuGroup.POST("/recalculate", middleware.RequirePermission("ispu.recalculate"), ispuHandler.Recalculate)
			ispuGroup.POST("/backfill", middleware.RequirePermission("ispu.recalculate"), ispuHandler.Backfill)
			ispuGroup.GET("/audit-tracking", middleware.RequirePermission("ispu.view"), ispuHandler.GetAuditTracking)
			ispuGroup.GET("/:id", middleware.RequirePermission("ispu.view"), ispuHandler.GetDetail)
		}

		// Phase 8 — Output & Integration (Authenticated & Authorized)
		destinationsGroup := api.Group("/destinations")
		destinationsGroup.Use(middleware.JWTAuth(authService))
		{
			destinationsGroup.GET("", middleware.RequirePermission("integration.view"), outboundHandler.GetDestinations)
			destinationsGroup.GET("/:id", middleware.RequirePermission("integration.view"), outboundHandler.GetDestinationByID)
			destinationsGroup.POST("", middleware.RequirePermission("integration.manage"), outboundHandler.CreateDestination)
			destinationsGroup.PUT("/:id", middleware.RequirePermission("integration.manage"), outboundHandler.UpdateDestination)
			destinationsGroup.DELETE("/:id", middleware.RequirePermission("integration.manage"), outboundHandler.DeleteDestination)
			destinationsGroup.GET("/:id/payload-config", middleware.RequirePermission("integration.view"), outboundHandler.GetPayloadConfig)
			destinationsGroup.PUT("/:id/payload-config", middleware.RequirePermission("integration.manage"), outboundHandler.UpdatePayloadConfig)
			destinationsGroup.GET("/:id/preview", middleware.RequirePermission("integration.view"), outboundHandler.GeneratePreview)
			destinationsGroup.POST("/:id/preview", middleware.RequirePermission("integration.view"), outboundHandler.GeneratePreview)
			destinationsGroup.POST("/:id/test", middleware.RequirePermission("integration.test"), outboundHandler.TestPayload)
		}

		queueGroup := api.Group("/integration/queue")
		queueGroup.Use(middleware.JWTAuth(authService))
		{
			queueGroup.GET("", middleware.RequirePermission("integration.view"), outboundHandler.GetQueueItems)
			queueGroup.GET("/summary", middleware.RequirePermission("integration.view"), outboundHandler.GetQueueSummary)
			queueGroup.GET("/:id", middleware.RequirePermission("integration.view"), outboundHandler.GetQueueItemDetail)
			queueGroup.POST("/:id/retry", middleware.RequirePermission("integration.queue.manage"), outboundHandler.RetryQueueItem)
			queueGroup.POST("/:id/replay", middleware.RequirePermission("integration.queue.manage"), outboundHandler.ReplayQueueItem)
			queueGroup.POST("/:id/cancel", middleware.RequirePermission("integration.queue.manage"), outboundHandler.CancelQueueItem)
		}

		// Public deployment domain inspection
		api.GET("/system/domain", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"monitoring_domain": config.Get().MonitoringDomain,
					"app_name":          config.Get().AppName,
					"version":           config.Get().Version,
				},
			})
		})
	}

	// 11. Serve Web UI Static Assets (if web/dist exists)
	distPath := filepath.Join("web", "dist")
	if envWeb := os.Getenv("DATALOGGER_WEB_DIR"); envWeb != "" {
		distPath = envWeb
	} else if _, err := os.Stat(distPath); os.IsNotExist(err) {
		// Fallback: look next to executable
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			altPath := filepath.Join(exeDir, "web", "dist")
			if _, statErr := os.Stat(altPath); statErr == nil {
				distPath = altPath
			}
		}
	}
	if _, err := os.Stat(distPath); err == nil {
		r.Static("/assets", filepath.Join(distPath, "assets"))
		r.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))

		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// Don't intercept API routes
			if len(path) >= 4 && path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
				return
			}
			c.File(filepath.Join(distPath, "index.html"))
		})
		logger.Info("Serving embedded Web UI from: %s", distPath)
	} else {
		r.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":      "RUNNING",
				"application": cfg.AppName,
				"version":     cfg.Version,
				"message":     "Datalogger Analysis Application backend is running. Frontend UI can be accessed at port 3000 (dev) or built into web/dist.",
			})
		})
	}

	// 12. Start Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		displayHost := cfg.Host
		if displayHost == "0.0.0.0" || displayHost == "" {
			displayHost = "localhost"
		}
		logger.Info("Web & REST API server listening at http://%s:%s", displayHost, cfg.Port)
		logger.Info("Local Browser Access: http://localhost:%s or http://127.0.0.1:%s", cfg.Port, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down %s...", cfg.AppName)

	// 1. Stop all active polling workers and cancel pending reconnects
	pollingEngine.Stop()

	// 2. Stop aggregation worker safely
	aggService.StopWorker()
	ispuService.StopWorker()
	outboundService.StopWorker()

	// 3. Stop background scheduler, backup scheduler, and housekeeping scheduler
	sched.Stop()
	backupScheduler.Stop()
	retentionScheduler.Stop()

	// 4. Flush pending telemetry buffer to MariaDB (Phase 3.1 Zero Data Loss)
	telemetryService.Stop()

	// 5. Gracefully shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced shutdown: %v", err)
	}

	// 6. Close all communication connections and adapters
	connManager.CloseAll()

	// 7. Close database connection pool last, after all workers stop
	if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}

	_ = log
	logger.Info("Server stopped cleanly with all connections closed. Goodbye.")
}

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
	"datalogger/internal/repository"
	"datalogger/internal/scheduler"
	"datalogger/internal/service"
	"datalogger/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Printf("Warning: failed to load config.json: %v. Using defaults.\n", err)
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
	if err := database.Seed(db); err != nil {
		logger.Error("Database seed error: %v", err)
	}

	// 5. Initialize Repositories
	deviceRepo := repository.NewDeviceRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)

	// 6. Initialize Services
	authService := service.NewAuthService(systemRepo, cfg)
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)

	// 7. Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()
	logger.Info("WebSocket hub initialized for real-time edge telemetry")

	// 8. Initialize Scheduler
	sched := scheduler.NewScheduler(systemRepo)
	sched.Start()
	defer sched.Stop()

	// 8b. Initialize Communication Engine (Phase 2.2)
	connManager := communication.NewConnectionManager(deviceService, communication.DefaultAdapterFactory)
	pollingEngine := communication.NewPollingEngine(connManager, deviceService, hub)
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
	deviceHandler := handler.NewDeviceHandler(deviceService)
	commHandler := handler.NewCommunicationHandler(connManager, pollingEngine, deviceService)

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWS(hub, c)
	})

	// Public & Protected API routes
	api := r.Group("/api")
	{
		// Auth
		api.POST("/auth/login", authHandler.Login)
		api.GET("/auth/me", middleware.JWTAuth(authService), authHandler.Me)
		api.GET("/users/me", middleware.JWTAuth(authService), authHandler.Me)

		// Development Phases & Tasks
		api.GET("/phases", devHandler.GetPhases)
		api.GET("/phases/:id", devHandler.GetPhaseByID)
		api.GET("/progress", devHandler.GetProgress)
		api.GET("/tasks", devHandler.GetTasks)
		api.POST("/tasks", devHandler.CreateTask)
		api.GET("/tasks/:id", devHandler.GetTaskByID)
		api.PUT("/tasks/:id", devHandler.UpdateTask)
		api.DELETE("/tasks/:id", devHandler.DeleteTask)
		api.GET("/activity", devHandler.GetActivity)

		// System Diagnostics & Health
		api.GET("/system/status", systemHandler.GetStatus)
		api.GET("/system/health", systemHandler.GetHealth)
		api.GET("/system/resources", systemHandler.GetResources)

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

			// Phase 2.2 — Communication Controls & Diagnostics
			devicesGroup.POST("/:id/communication/connect", middleware.RequirePermission("device.communication.manage"), commHandler.Connect)
			devicesGroup.POST("/:id/communication/disconnect", middleware.RequirePermission("device.communication.manage"), commHandler.Disconnect)
			devicesGroup.POST("/:id/communication/reconnect", middleware.RequirePermission("device.communication.manage"), commHandler.Reconnect)
			devicesGroup.POST("/:id/communication/test", middleware.RequirePermission("device.communication.test"), commHandler.TestConnection)
			devicesGroup.GET("/:id/communication/status", middleware.RequirePermission("device.communication.view"), commHandler.GetStatus)
			devicesGroup.POST("/:id/parameters/:paramId/test-read", middleware.RequirePermission("device.communication.test"), commHandler.TestReadParameter)
		}

		// Operational Telemetry Data
		api.GET("/data", systemHandler.GetData)
		api.GET("/data/latest", systemHandler.GetLatestData)
		api.GET("/data/trend", systemHandler.GetDataTrend)

		// Alarms
		api.GET("/alarms", systemHandler.GetAlarms)
		api.GET("/alarms/active", systemHandler.GetActiveAlarms)
		api.PUT("/alarms/:id/acknowledge", systemHandler.AcknowledgeAlarm)

		// Logs
		api.GET("/logs", systemHandler.GetLogs)
		api.GET("/audit-trails", systemHandler.GetAuditTrails)
		api.GET("/communication-logs", systemHandler.GetCommunicationLogs)

		// Database Management & Monitoring
		api.GET("/system/database/status", systemHandler.GetDatabaseStatus)
		api.POST("/system/database/test", systemHandler.TestDatabaseConnection)
		api.POST("/system/database/config", systemHandler.SaveDatabaseConfig)
	}

	// 11. Serve Web UI Static Assets (if web/dist exists)
	distPath := filepath.Join("web", "dist")
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

	// Stop all active polling workers and close connections gracefully
	pollingEngine.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced shutdown: %v", err)
	}

	_ = log
	logger.Info("Server stopped cleanly. Goodbye.")
}

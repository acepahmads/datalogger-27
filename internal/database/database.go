package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/logger"
	"datalogger/internal/model"

	_ "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	DB           *gorm.DB
	dbMu         sync.RWMutex
	lastPingTime time.Time
	lastLatency  float64
	dbVersion    string
	dbSizeMB     float64
)

type DatabaseMetrics struct {
	Type              string    `json:"type"`
	Host              string    `json:"host"`
	Port              string    `json:"port"`
	Database          string    `json:"database"`
	Username          string    `json:"username"`
	Status            string    `json:"status"` // "Connected", "Healthy", "Disconnected", "Error"
	LatencyMs         float64   `json:"latency_ms"`
	Version           string    `json:"version"`
	ActiveConnections int       `json:"active_connections"`
	MaxConnections    int       `json:"max_connections"`
	DatabaseSizeMB    float64   `json:"database_size_mb"`
	LastQueryTime     time.Time `json:"last_query_time"`
	LastError         string    `json:"last_error,omitempty"`
}

// Init initializes the MariaDB database connection, ensures schema exists, and runs auto-migrations.
func Init(cfg *config.Config) (*gorm.DB, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	// Default to MariaDB
	dbUser := cfg.DBUser
	if dbUser == "" {
		dbUser = "root"
	}
	dbHost := cfg.DBHost
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort := cfg.DBPort
	if dbPort == "" {
		dbPort = "3306"
	}
	dbName := cfg.DBName
	if dbName == "" {
		dbName = "datalogger"
	}

	// 1. Ensure database exists in MariaDB
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, cfg.DBPassword, dbHost, dbPort)

	rawDB, err := sql.Open("mysql", rootDSN)
	if err == nil {
		defer rawDB.Close()
		_, _ = rawDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbName))
	}

	// 2. Open connection to the target database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s",
		dbUser, cfg.DBPassword, dbHost, dbPort, dbName)

	start := time.Now()
	gormDB, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		PrepareStmt: true,
	})
	if err != nil {
		logger.Error("Failed to connect to MariaDB at %s:%s/%s: %v", dbHost, dbPort, dbName, err)
		return nil, fmt.Errorf("mariadb connection failed: %w", err)
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	lastLatency = latency
	lastPingTime = time.Now()

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 3. Configure Connection Pooling for MariaDB Edge
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	// 4. Query version
	var ver string
	if err := sqlDB.QueryRow("SELECT VERSION();").Scan(&ver); err == nil {
		dbVersion = ver
	} else {
		dbVersion = "MariaDB 10.4.x"
	}

	logger.Info("Connected to MariaDB (%s) at: %s:%s/%s in %.2fms", dbVersion, dbHost, dbPort, dbName, latency)

	// 5. Auto-migrate models
	err = gormDB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.DevelopmentPhase{},
		&model.DevelopmentSubphase{},
		&model.DevelopmentTask{},
		&model.DevelopmentTaskLog{},
		&model.DevelopmentDependency{},
		&model.DevelopmentEvidence{},
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Sensor{},
		&model.RawData{},
		&model.ProcessedData{},
		&model.AggregatedData{},
		&model.Alarm{},
		&model.NotificationChannel{},
		&model.Notification{},
		&model.AuditTrail{},
		&model.SystemLog{},
		&model.CommunicationLog{},
		&model.OutputDestination{},
		&model.DeliveryLog{},
		&model.SystemHealth{},
	)
	if err != nil {
		logger.Error("Database schema migration error: %v", err)
		return nil, fmt.Errorf("failed to auto-migrate MariaDB schema: %w", err)
	}

	logger.Info("MariaDB schema migration completed successfully")

	DB = gormDB
	return gormDB, nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return DB
}

// TestConnection tests a MariaDB connection with given credentials without altering current global DB
func TestConnection(host, port, user, password, dbName string) (bool, string, float64, error) {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "3306"
	}
	if user == "" {
		user = "root"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=3s",
		user, password, host, port, dbName)

	start := time.Now()
	testDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, "", 0, err
	}
	defer testDB.Close()

	if err := testDB.Ping(); err != nil {
		return false, "", 0, err
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0

	var ver string
	_ = testDB.QueryRow("SELECT VERSION();").Scan(&ver)
	if ver == "" {
		ver = "MariaDB"
	}

	return true, ver, latency, nil
}

// GetMetrics returns active MariaDB runtime health and statistics
func GetMetrics() DatabaseMetrics {
	dbMu.RLock()
	defer dbMu.RUnlock()

	cfg := config.Get()
	metrics := DatabaseMetrics{
		Type:              "MariaDB",
		Host:              cfg.DBHost,
		Port:              cfg.DBPort,
		Database:          cfg.DBName,
		Username:          cfg.DBUser,
		Status:            "Disconnected",
		LatencyMs:         lastLatency,
		Version:           dbVersion,
		MaxConnections:    25,
		LastQueryTime:     lastPingTime,
	}

	if DB == nil {
		return metrics
	}

	sqlDB, err := DB.DB()
	if err != nil {
		metrics.LastError = err.Error()
		return metrics
	}

	// Measure current ping latency
	start := time.Now()
	if err := sqlDB.Ping(); err == nil {
		metrics.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
		metrics.Status = "Connected"
		metrics.LastQueryTime = time.Now()
	} else {
		metrics.Status = "Error"
		metrics.LastError = err.Error()
	}

	// Connection stats
	stats := sqlDB.Stats()
	metrics.ActiveConnections = stats.InUse
	metrics.MaxConnections = stats.MaxOpenConnections

	// Database size in MB
	var sizeMB float64
	query := fmt.Sprintf("SELECT IFNULL(ROUND(SUM(data_length + index_length) / 1024 / 1024, 2), 0) FROM information_schema.tables WHERE table_schema = '%s';", cfg.DBName)
	if err := sqlDB.QueryRow(query).Scan(&sizeMB); err == nil {
		metrics.DatabaseSizeMB = sizeMB
	}

	return metrics
}

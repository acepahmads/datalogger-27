package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config represents the application configuration.
type Config struct {
	AppName            string `json:"app_name"`
	Version            string `json:"version"`
	Environment        string `json:"environment"`
	Port               string `json:"port"`
	Host               string `json:"host"`
	DBType             string `json:"db_type"` // "mariadb", "mysql", "sqlite"
	DBPath             string `json:"db_path"`
	DBHost             string `json:"db_host"`
	DBPort             string `json:"db_port"`
	DBUser             string `json:"db_user"`
	DBPassword         string `json:"db_password"`
	DBName             string `json:"db_name"`
	JWTSecret          string `json:"jwt_secret"`
	JWTExpirationHours int    `json:"jwt_expiration_hours"`
	LogLevel           string `json:"log_level"`
	LogDir             string `json:"log_dir"`
	DataDir            string `json:"data_dir"`
	EnableCloudSync    bool   `json:"enable_cloud_sync"`
	CloudEndpoint      string `json:"cloud_endpoint"`

	// Phase 4.2: Persistent Spool / Write-Ahead Log Queue Configuration
	QueueEnabled         bool    `json:"queue_enabled"`
	QueueDir             string  `json:"queue_dir"`
	QueueMaxSizeBytes    int64   `json:"queue_max_size_bytes"`
	QueueBatchSize       int     `json:"queue_batch_size"`
	QueueSyncMode        string  `json:"queue_sync_mode"` // "batch", "always", "none"
	QueueFlushIntervalMs int     `json:"queue_flush_interval_ms"`
	QueueDiskWarnPercent float64 `json:"queue_disk_warn_percent"`

	// Phase 4.3: Backup & Restore Configuration
	BackupDir                   string `json:"backup_dir"`
	BackupScheduleEnabled       bool   `json:"backup_schedule_enabled"`
	BackupScheduleIntervalHours int    `json:"backup_schedule_interval_hours"`
	BackupScheduleTime          string `json:"backup_schedule_time"`
	BackupCompression           string `json:"backup_compression"` // "gzip", "none"
	BackupMinFreeSpaceMB        int64  `json:"backup_min_free_space_mb"`
	BackupMaxKeepCount          int    `json:"backup_max_keep_count"`
}

var (
	globalConfig *Config
	cfgMu        sync.RWMutex
)

// DefaultConfig returns the default configuration with MariaDB as production database.
func DefaultConfig() *Config {
	return &Config{
		AppName:                     "Datalogger Analysis Application",
		Version:                     "1.0.0-phase1",
		Environment:                 "development",
		Port:                        "8080",
		Host:                        "0.0.0.0",
		DBType:                      "mariadb",
		DBHost:                      "127.0.0.1",
		DBPort:                      "3306",
		DBUser:                      "root",
		DBPassword:                  "",
		DBName:                      "datalogger",
		DBPath:                      filepath.Join("data", "datalogger.db"),
		JWTSecret:                   "datalogger-local-secret-key-prod-2026",
		JWTExpirationHours:          24,
		LogLevel:                    "info",
		LogDir:                      "logs",
		DataDir:                     "data",
		EnableCloudSync:             false,
		CloudEndpoint:               "",
		QueueEnabled:                true,
		QueueDir:                    filepath.Join("data", "queue"),
		QueueMaxSizeBytes:           100 * 1024 * 1024, // 100 MB
		QueueBatchSize:              50,
		QueueSyncMode:               "batch",
		QueueFlushIntervalMs:        250,
		QueueDiskWarnPercent:        80.0,
		BackupDir:                   filepath.Join("data", "backups"),
		BackupScheduleEnabled:       true,
		BackupScheduleIntervalHours: 24,
		BackupScheduleTime:          "02:00",
		BackupCompression:           "gzip",
		BackupMinFreeSpaceMB:        500,
		BackupMaxKeepCount:          10,
	}
}

// Load loads configuration from environment variables, optional config file, or defaults.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// If config file exists, load it
	if configPath != "" {
		if file, err := os.Open(configPath); err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(cfg); err != nil {
				return nil, err
			}
		}
	}

	// Environment variable overrides (support both standard and DB_ prefixed)
	if envDriver := os.Getenv("DB_DRIVER"); envDriver != "" {
		cfg.DBType = envDriver
	} else if dbType := os.Getenv("DATALOGGER_DB_TYPE"); dbType != "" {
		cfg.DBType = dbType
	}

	if envHost := os.Getenv("DB_HOST"); envHost != "" {
		cfg.DBHost = envHost
	} else if dbHost := os.Getenv("DATALOGGER_DB_HOST"); dbHost != "" {
		cfg.DBHost = dbHost
	}

	if envPort := os.Getenv("DB_PORT"); envPort != "" {
		cfg.DBPort = envPort
	} else if dbPort := os.Getenv("DATALOGGER_DB_PORT"); dbPort != "" {
		cfg.DBPort = dbPort
	}

	if envUser := os.Getenv("DB_USER"); envUser != "" {
		cfg.DBUser = envUser
	} else if dbUser := os.Getenv("DATALOGGER_DB_USER"); dbUser != "" {
		cfg.DBUser = dbUser
	}

	if envPass := os.Getenv("DB_PASSWORD"); envPass != "" {
		cfg.DBPassword = envPass
	} else if dbPass := os.Getenv("DATALOGGER_DB_PASSWORD"); dbPass != "" {
		cfg.DBPassword = dbPass
	}

	if envName := os.Getenv("DB_NAME"); envName != "" {
		cfg.DBName = envName
	} else if dbName := os.Getenv("DATALOGGER_DB_NAME"); dbName != "" {
		cfg.DBName = dbName
	}

	if port := os.Getenv("DATALOGGER_PORT"); port != "" {
		cfg.Port = port
	}
	if host := os.Getenv("DATALOGGER_HOST"); host != "" {
		cfg.Host = host
	}
	if dbPath := os.Getenv("DATALOGGER_DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}
	if jwtSecret := os.Getenv("DATALOGGER_JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}
	if logDir := os.Getenv("DATALOGGER_LOG_DIR"); logDir != "" {
		cfg.LogDir = logDir
	}
	if dataDir := os.Getenv("DATALOGGER_DATA_DIR"); dataDir != "" {
		cfg.DataDir = dataDir
	}
	if env := os.Getenv("DATALOGGER_ENV"); env != "" {
		cfg.Environment = env
	}
	if qDir := os.Getenv("DATALOGGER_QUEUE_DIR"); qDir != "" {
		cfg.QueueDir = qDir
	}
	if qSync := os.Getenv("DATALOGGER_QUEUE_SYNC_MODE"); qSync != "" {
		cfg.QueueSyncMode = qSync
	}
	if bDir := os.Getenv("DATALOGGER_BACKUP_DIR"); bDir != "" {
		cfg.BackupDir = bDir
	}
	if bSched := os.Getenv("DATALOGGER_BACKUP_SCHEDULE_ENABLED"); bSched != "" {
		cfg.BackupScheduleEnabled = bSched == "true" || bSched == "1"
	}
	if bComp := os.Getenv("DATALOGGER_BACKUP_COMPRESSION"); bComp != "" {
		cfg.BackupCompression = bComp
	}
	if bTime := os.Getenv("DATALOGGER_BACKUP_TIME"); bTime != "" {
		cfg.BackupScheduleTime = bTime
	}

	// Ensure data, log, queue, and backup directories exist
	_ = os.MkdirAll(cfg.DataDir, 0755)
	_ = os.MkdirAll(cfg.LogDir, 0755)
	_ = os.MkdirAll(cfg.BackupDir, 0755)
	_ = os.MkdirAll(filepath.Dir(cfg.DBPath), 0755)
	if cfg.QueueEnabled {
		_ = os.MkdirAll(cfg.QueueDir, 0755)
	}

	cfgMu.Lock()
	globalConfig = cfg
	cfgMu.Unlock()

	return cfg, nil
}

// Save persists the current configuration to config.json
func Save(configPath string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	cfgMu.Lock()
	globalConfig = cfg
	cfgMu.Unlock()

	return nil
}

// Get returns the loaded configuration or default
func Get() *Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if globalConfig == nil {
		globalConfig = DefaultConfig()
	}
	return globalConfig
}

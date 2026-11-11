package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	// Primary Monitoring Domain: "AQMS", "SPARING", "ONLIMO", "WQMS", "CUSTOM", "CLOUD"
	MonitoringDomain string `json:"monitoring_domain"`
	CloudMode        bool   `json:"cloud_mode"`

	// Database Seeding Governance (Default ON in Edge mode, Default OFF in Cloud mode)
	SeedUsers           bool   `json:"seed_users"`
	SeedSampleData      bool   `json:"seed_sample_data"`
	BootstrapSuperadmin *bool  `json:"bootstrap_superadmin,omitempty"`
	SuperadminUsername  string `json:"superadmin_username,omitempty"`
	SuperadminPassword  string `json:"superadmin_password,omitempty"`

	// Configurable Telemetry Acquisition Modules (Multi-Tenant Receiver & Hybrid Edge)
	SensorPollingEnabled bool   `json:"sensor_polling_enabled"`
	MQTTIngestEnabled    bool   `json:"mqtt_ingest_enabled"`
	MQTTBrokerURL        string `json:"mqtt_broker_url"`
	MQTTTopicPrefix      string `json:"mqtt_topic_prefix"`

	// Raw Telemetry Query Governance: Maximum allowed time span in hours (default 168 = 7 days)
	RawTelemetryMaxRangeHours int `json:"raw_telemetry_max_range_hours"`

	// Communication / Serial Bus Defaults (Phase 7 #83)
	SerialPort      string `json:"serial_port"`
	SerialBaudRate  int    `json:"serial_baud_rate"`
	SerialDataBits  int    `json:"serial_data_bits"`
	SerialStopBits  int    `json:"serial_stop_bits"`
	SerialParity    string `json:"serial_parity"`
	SerialTimeoutMs int    `json:"serial_timeout_ms"`
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
		MonitoringDomain:            "AQMS",
		SeedUsers:                   true,
		SeedSampleData:              true,
		SuperadminUsername:          "admin",
		SensorPollingEnabled:        true,
		MQTTIngestEnabled:           true,
		MQTTBrokerURL:               "",
		MQTTTopicPrefix:             "datalogger/ingest/+",
		RawTelemetryMaxRangeHours:   168,
		SerialPort:                  "COM1",
		SerialBaudRate:              9600,
		SerialDataBits:              8,
		SerialStopBits:              1,
		SerialParity:                "N",
		SerialTimeoutMs:             1000,
	}
}

func parseBoolEnvValue(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	return v == "true" || v == "1" || v == "yes" || v == "on"
}

// Load loads configuration from environment variables, optional config file, or defaults.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// If config file exists, load it
	hasExplicitMQTTInFile := false
	hasExplicitSeedUsersInFile := false
	hasExplicitSeedSampleInFile := false
	fileExplicitCloudReady := false
	if configPath != "" {
		if rawBytes, err := os.ReadFile(configPath); err == nil {
			if err := json.Unmarshal(rawBytes, cfg); err != nil {
				return nil, err
			}
			var rawMap map[string]json.RawMessage
			if json.Unmarshal(rawBytes, &rawMap) == nil {
				_, hasExplicitMQTTInFile = rawMap["mqtt_ingest_enabled"]
				_, hasExplicitSeedUsersInFile = rawMap["seed_users"]
				_, hasExplicitSeedSampleInFile = rawMap["seed_sample_data"]

				var rawDomain string
				if rawDomBytes, ok := rawMap["monitoring_domain"]; ok {
					_ = json.Unmarshal(rawDomBytes, &rawDomain)
				}
				rawCloudVal, hasRawCloudKey := false, false
				if rawCloudBytes, ok := rawMap["cloud_mode"]; ok {
					hasRawCloudKey = true
					_ = json.Unmarshal(rawCloudBytes, &rawCloudVal)
				}
				// Consider seed flags in file explicitly intended for Cloud when either:
				// - cloud_mode is true in file, OR
				// - monitoring_domain is "CLOUD" in file without a leftover "cloud_mode": false from the Edge template.
				fileExplicitCloudReady = rawCloudVal || (strings.EqualFold(strings.TrimSpace(rawDomain), "CLOUD") && !hasRawCloudKey)
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
	if mDomain := os.Getenv("DATALOGGER_MONITORING_DOMAIN"); mDomain != "" {
		cfg.MonitoringDomain = mDomain
	}
	cfg.MonitoringDomain = NormalizeMonitoringDomain(cfg.MonitoringDomain)

	if sPoll := strings.TrimSpace(os.Getenv("SENSOR_POLLING_ENABLED")); sPoll != "" {
		cfg.SensorPollingEnabled = !strings.EqualFold(sPoll, "false") && sPoll != "0"
	} else if sPoll2 := strings.TrimSpace(os.Getenv("DATALOGGER_SENSOR_POLLING_ENABLED")); sPoll2 != "" {
		cfg.SensorPollingEnabled = !strings.EqualFold(sPoll2, "false") && sPoll2 != "0"
	}

	if mqttEn := strings.TrimSpace(os.Getenv("MQTT_INGEST_ENABLED")); mqttEn != "" {
		cfg.MQTTIngestEnabled = !strings.EqualFold(mqttEn, "false") && mqttEn != "0"
	} else if mqttEn2 := strings.TrimSpace(os.Getenv("DATALOGGER_MQTT_INGEST_ENABLED")); mqttEn2 != "" {
		cfg.MQTTIngestEnabled = !strings.EqualFold(mqttEn2, "false") && mqttEn2 != "0"
	}

	if mqttBroker := strings.TrimSpace(os.Getenv("MQTT_BROKER_URL")); mqttBroker != "" {
		cfg.MQTTBrokerURL = mqttBroker
	} else if mqttBroker2 := strings.TrimSpace(os.Getenv("DATALOGGER_MQTT_BROKER_URL")); mqttBroker2 != "" {
		cfg.MQTTBrokerURL = mqttBroker2
	}

	if mqttTopic := strings.TrimSpace(os.Getenv("MQTT_TOPIC_PREFIX")); mqttTopic != "" {
		cfg.MQTTTopicPrefix = mqttTopic
	} else if mqttTopic2 := strings.TrimSpace(os.Getenv("DATALOGGER_MQTT_TOPIC_PREFIX")); mqttTopic2 != "" {
		cfg.MQTTTopicPrefix = mqttTopic2
	}

	if maxRangeStr := os.Getenv("DATALOGGER_RAW_MAX_RANGE_HOURS"); maxRangeStr != "" {
		if v, err := strconv.Atoi(strings.TrimSpace(maxRangeStr)); err == nil && v > 0 {
			cfg.RawTelemetryMaxRangeHours = v
		}
	}
	if cfg.RawTelemetryMaxRangeHours <= 0 {
		cfg.RawTelemetryMaxRangeHours = 168
	}
	hasEnvSeedUsers := false
	if sUsers := strings.TrimSpace(os.Getenv("SEED_USERS")); sUsers != "" {
		cfg.SeedUsers = parseBoolEnvValue(sUsers)
		hasEnvSeedUsers = true
	} else if sUsers2 := strings.TrimSpace(os.Getenv("DATALOGGER_SEED_USERS")); sUsers2 != "" {
		cfg.SeedUsers = parseBoolEnvValue(sUsers2)
		hasEnvSeedUsers = true
	}

	hasEnvSeedSample := false
	if sSample := strings.TrimSpace(os.Getenv("SEED_SAMPLE_DATA")); sSample != "" {
		cfg.SeedSampleData = parseBoolEnvValue(sSample)
		hasEnvSeedSample = true
	} else if sSample2 := strings.TrimSpace(os.Getenv("DATALOGGER_SEED_SAMPLE_DATA")); sSample2 != "" {
		cfg.SeedSampleData = parseBoolEnvValue(sSample2)
		hasEnvSeedSample = true
	}

	if bSuper := strings.TrimSpace(os.Getenv("BOOTSTRAP_SUPERADMIN")); bSuper != "" {
		v := parseBoolEnvValue(bSuper)
		cfg.BootstrapSuperadmin = &v
	} else if bSuper2 := strings.TrimSpace(os.Getenv("DATALOGGER_BOOTSTRAP_SUPERADMIN")); bSuper2 != "" {
		v := parseBoolEnvValue(bSuper2)
		cfg.BootstrapSuperadmin = &v
	}

	if suUser := strings.TrimSpace(os.Getenv("SUPERADMIN_USER")); suUser != "" {
		cfg.SuperadminUsername = suUser
	} else if suUser2 := strings.TrimSpace(os.Getenv("DATALOGGER_SUPERADMIN_USER")); suUser2 != "" {
		cfg.SuperadminUsername = suUser2
	}

	if suPass := os.Getenv("SUPERADMIN_PASSWORD"); suPass != "" {
		cfg.SuperadminPassword = suPass
	} else if suPass2 := os.Getenv("DATALOGGER_SUPERADMIN_PASSWORD"); suPass2 != "" {
		cfg.SuperadminPassword = suPass2
	}

	if envCloud := strings.TrimSpace(os.Getenv("CLOUD_MODE")); envCloud != "" {
		cfg.CloudMode = strings.EqualFold(envCloud, "true") || envCloud == "1" || strings.EqualFold(envCloud, "yes")
	}
	if strings.EqualFold(strings.TrimSpace(cfg.MonitoringDomain), "CLOUD") {
		cfg.CloudMode = true
	}
	if cfg.CloudMode {
		cfg.SensorPollingEnabled = false
		if !hasExplicitMQTTInFile && os.Getenv("MQTT_INGEST_ENABLED") == "" && os.Getenv("DATALOGGER_MQTT_INGEST_ENABLED") == "" {
			cfg.MQTTIngestEnabled = false
		}
		if !hasEnvSeedUsers && (!hasExplicitSeedUsersInFile || !fileExplicitCloudReady) {
			cfg.SeedUsers = false
		}
		if !hasEnvSeedSample && (!hasExplicitSeedSampleInFile || !fileExplicitCloudReady) {
			cfg.SeedSampleData = false
		}
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

// NormalizeMonitoringDomain ensures the domain is uppercase and one of the valid values: AQMS, SPARING, ONLIMO, WQMS, CUSTOM, or CLOUD.
func NormalizeMonitoringDomain(domain string) string {
	d := strings.ToUpper(strings.TrimSpace(domain))
	if d == "AQMS" || d == "SPARING" || d == "ONLIMO" || d == "WQMS" || d == "CUSTOM" || d == "CLOUD" {
		return d
	}
	return "AQMS"
}

// IsCloudMode returns true when the platform is configured in Cloud Receiver mode (physical sensor polling disabled).
func (c *Config) IsCloudMode() bool {
	if c == nil {
		return false
	}
	return c.CloudMode || strings.EqualFold(strings.TrimSpace(c.MonitoringDomain), "CLOUD")
}

// EffectiveRawTelemetryMaxRangeHours returns the configured max query window in hours (defaulting to 168h / 7d).
func (c *Config) EffectiveRawTelemetryMaxRangeHours() int {
	if c == nil || c.RawTelemetryMaxRangeHours <= 0 {
		return 168
	}
	if c.RawTelemetryMaxRangeHours > 8760 {
		return 8760
	}
	return c.RawTelemetryMaxRangeHours
}

// ShouldBootstrapSuperadmin returns true unless bootstrap_superadmin is explicitly disabled (false).
func (c *Config) ShouldBootstrapSuperadmin() bool {
	if c == nil || c.BootstrapSuperadmin == nil {
		return true
	}
	return *c.BootstrapSuperadmin
}

// EffectiveSuperadminUsername returns the configured initial superadmin username (defaulting to "admin").
func (c *Config) EffectiveSuperadminUsername() string {
	if c == nil {
		return "admin"
	}
	if u := strings.TrimSpace(c.SuperadminUsername); u != "" {
		return u
	}
	return "admin"
}

// EffectiveSuperadminPassword returns the configured initial superadmin password (defaulting to "Password123!").
func (c *Config) EffectiveSuperadminPassword() string {
	if c == nil || c.SuperadminPassword == "" {
		return "Password123!"
	}
	return c.SuperadminPassword
}

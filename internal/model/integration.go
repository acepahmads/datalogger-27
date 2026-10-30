package model

import (
	"time"

	"gorm.io/gorm"
)

type DestinationType string

const (
	DestREST      DestinationType = "REST_API"
	DestMQTT      DestinationType = "MQTT"
	DestTCP       DestinationType = "TCP"
	DestUDP       DestinationType = "UDP"
	DestWebSocket DestinationType = "WEBSOCKET"
	DestCloud     DestinationType = "CLOUD"
)

type OutboundDeliveryStatus string

const (
	DeliveryPending      OutboundDeliveryStatus = "PENDING"
	DeliverySending      OutboundDeliveryStatus = "SENDING"
	DeliveryDelivered    OutboundDeliveryStatus = "DELIVERED"
	DeliveryRetrying     OutboundDeliveryStatus = "RETRYING"
	DeliveryFailed       OutboundDeliveryStatus = "FAILED"
	DeliveryAcknowledged OutboundDeliveryStatus = "ACKNOWLEDGED"
	DeliveryCancelled    OutboundDeliveryStatus = "CANCELLED"
)

// OutputDestination represents an upstream receiving platform or socket target
type OutputDestination struct {
	ID                uint                     `gorm:"primaryKey" json:"id"`
	Name              string                   `gorm:"size:128;not null" json:"name"`
	Type              DestinationType          `gorm:"size:32;not null" json:"type"`
	Endpoint          string                   `gorm:"size:255;not null" json:"endpoint"` // URL or host:port
	Enabled           bool                     `gorm:"default:false" json:"enabled"`     // Local-first: default false unless explicitly activated
	Format            string                   `gorm:"size:32;default:'JSON'" json:"format"` // JSON, PROTOBUF, CSV
	IntervalSeconds   int                      `gorm:"default:0" json:"interval_seconds"` // 0 = all configured intervals, or 120, 300, 1800, 3600
	DeviceID          *uint                    `gorm:"index" json:"device_id,omitempty"` // nil = all stations
	BatchSize         int                      `gorm:"default:100" json:"batch_size"`
	RetryLimit        int                      `gorm:"default:5" json:"retry_limit"`
	QueueLimit        int                      `gorm:"default:10000" json:"queue_limit"`
	Status            string                   `gorm:"size:32;default:'IDLE'" json:"status"` // IDLE, ACTIVE, ERROR, DISABLED
	
	// Authentication & Security (redacted on API response)
	AuthType          string                   `gorm:"size:32;default:'NONE'" json:"auth_type"` // NONE, BEARER_TOKEN, API_KEY, BASIC_AUTH, LOGIN_BEARER
	AuthURL           string                   `gorm:"size:255" json:"auth_url,omitempty"`      // Step-1 login endpoint for LOGIN_BEARER (2-step auth)
	AuthTokenPath     string                   `gorm:"size:128" json:"auth_token_path,omitempty"` // Optional JSON dot-path (e.g. "token", "data.token")
	AuthCacheSeconds  int                      `gorm:"default:0" json:"auth_cache_seconds"`       // 0 = always auth before each data send (strict 2x); >0 = cache TTL seconds
	AuthToken         string                   `gorm:"size:255" json:"auth_token,omitempty"`
	APIKeyHeader      string                   `gorm:"size:64;default:'X-API-Key'" json:"api_key_header,omitempty"`
	APIKeyValue       string                   `gorm:"size:255" json:"api_key_value,omitempty"`
	Username          string                   `gorm:"size:128" json:"username,omitempty"`
	Password          string                   `gorm:"size:255" json:"password,omitempty"`
	CustomHeaders     string                   `gorm:"type:text" json:"custom_headers,omitempty"` // JSON string map[string]string
	
	// Protocol-specific parameters
	MQTTTopic         string                   `gorm:"size:255" json:"mqtt_topic,omitempty"`
	MQTTQoS           int                      `gorm:"default:1" json:"mqtt_qos,omitempty"` // 0, 1, 2
	
	LastDelivery      *time.Time               `json:"last_delivery"`
	SuccessDeliveries int64                    `gorm:"default:0" json:"success_deliveries"`
	FailedDeliveries  int64                    `gorm:"default:0" json:"failed_deliveries"`
	
	PayloadConfig     *DestinationPayloadConfig `gorm:"foreignKey:DestinationID" json:"payload_config,omitempty"`
	Device            *Device                  `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	RequeuedBatches   int                      `gorm:"-" json:"requeued_batches,omitempty"`

	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	DeletedAt         gorm.DeletedAt           `gorm:"index" json:"-"`
}

// DestinationPayloadConfig defines destination-specific serialization formatting and field aliasing
type DestinationPayloadConfig struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	DestinationID         uint      `gorm:"uniqueIndex;not null" json:"destination_id"`
	TemplateType          string    `gorm:"size:32;default:'STANDARD'" json:"template_type"` // STANDARD, KLHK_SPARING, NESTED_SCADA, MULTI_DEVICE, CUSTOM
	StationIDField        string    `gorm:"size:64;default:'station_id'" json:"station_id_field"`
	StationIDValue        string    `gorm:"size:64" json:"station_id_value,omitempty"` // optional override
	ModuleIDValue         string    `gorm:"size:64" json:"module_id_value,omitempty"`  // optional override for tick schemas
	TimestampField         string    `gorm:"size:64;default:'timestamp'" json:"timestamp_field"`
	TimestampFormat        string    `gorm:"size:32;default:'ISO8601'" json:"timestamp_format"` // ISO8601, UNIX_MS, UNIX_SEC, CUSTOM
	TimestampTimezone      string    `gorm:"size:64;default:''" json:"timestamp_timezone,omitempty"` // e.g. Asia/Jakarta, Asia/Makassar, Asia/Jayapura, UTC
	IntervalField          string    `gorm:"size:64;default:'interval_minutes'" json:"interval_field"`
	OmitIntervalField      bool      `gorm:"default:false" json:"omit_interval_field"`
	Structure              string    `gorm:"size:32;default:'FLAT'" json:"structure"` // FLAT, NESTED, NESTED_BY_DEVICE, ARRAY_BY_DEVICE, MULTI_TICK_ARRAY
	NestedDataKey          string    `gorm:"size:64;default:'data'" json:"nested_data_key"`
	IncludeDeviceIdentity  bool      `gorm:"default:false" json:"include_device_identity"` // true: include device code/name in payload
	DeviceIdentityMode     string    `gorm:"size:32;default:'NONE'" json:"device_identity_mode"` // NONE, PREFIX, NESTED, TAG
	ConflictResolution     string    `gorm:"size:32;default:'FAIL_ON_CONFLICT'" json:"conflict_resolution"` // FAIL_ON_CONFLICT, CUSTOM_MAPPING, ARRAY, PRIORITY_PRIMARY, AVERAGE
	IncludeUnits           bool      `gorm:"default:false" json:"include_units"`
	IncludeQuality         bool      `gorm:"default:false" json:"include_quality"`
	IncludeMetadata        bool      `gorm:"default:true" json:"include_metadata"`
	IncludeSensorStatFlags bool      `gorm:"default:false" json:"include_sensor_stat_flags"`          // Inject stat_pm10..stat_hc flags (KLHK ISPU AQMS)
	SensorStatMode         string    `gorm:"size:32;default:'STATIC_1'" json:"sensor_stat_mode"`     // STATIC_1 or AUTO_QUALITY
	StaticMetadataJSON    string    `gorm:"type:text" json:"static_metadata_json,omitempty"`     // JSON string map[string]interface{}
	ParameterMappingsJSON string    `gorm:"type:text" json:"parameter_mappings_json,omitempty"` // JSON array of ParameterMappingItem
	RequiredFieldsJSON    string    `gorm:"type:text" json:"required_fields_json,omitempty"`    // JSON array of string
	Version               int       `gorm:"default:1" json:"version"`
	RequeuedBatches       int       `gorm:"-" json:"requeued_batches,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

// ParameterMappingItem configures external aliasing and inclusion per parameter
type ParameterMappingItem struct {
	DeviceID      uint   `json:"device_id,omitempty"`      // 0 = wildcard across all devices, or specific device ID
	DeviceCode    string `json:"device_code,omitempty"`    // e.g. "AQMS-01"
	ParameterID   uint   `json:"parameter_id"`
	ParameterCode string `json:"parameter_code"`
	ParameterName string `json:"parameter_name"`
	ExternalName  string `json:"external_name"`
	Enabled       bool   `json:"enabled"`
}

// OutboundDeliveryQueue stores persistent pending, sending, retrying, and delivered telemetry batches
type OutboundDeliveryQueue struct {
	ID                   uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchID              string                 `gorm:"size:64;not null;index" json:"batch_id"` // deterministic: BATCH-DEST<dest_id>-<interval>-<period_end>
	DestinationID        uint                   `gorm:"not null;index" json:"destination_id"`
	Destination          *OutputDestination     `gorm:"foreignKey:DestinationID" json:"destination,omitempty"`
	DeviceID             *uint                  `gorm:"index" json:"device_id,omitempty"` // nullable: nil for unified multi-device batch
	Device               *Device                `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	DeviceCount          int                    `gorm:"default:1" json:"device_count"`
	DevicesSummary       string                 `gorm:"size:255;default:''" json:"devices_summary"` // e.g. "AQMS-01, AQMS-02"
	PeriodStart          time.Time              `gorm:"not null;index" json:"period_start"`
	PeriodEnd            time.Time              `gorm:"not null;index" json:"period_end"`
	IntervalSeconds      int                    `gorm:"not null;index" json:"interval_seconds"`
	ParameterCount       int                    `gorm:"default:0" json:"parameter_count"`
	SourceAggregationIDs string                 `gorm:"type:text" json:"source_aggregation_ids"` // JSON array of uint64
	SourceValuesJSON     string                 `gorm:"type:text" json:"source_values_json"`     // JSON array of SourceAggregationValueItem
	PayloadSnapshot      string                 `gorm:"type:text;not null" json:"payload_snapshot"`
	PayloadVersion       int                    `gorm:"default:1" json:"payload_version"`
	IdempotencyKey       string                 `gorm:"size:160;not null;uniqueIndex" json:"idempotency_key"` // DEST<id>_INT<interval>_<period_end_unix>
	Status               OutboundDeliveryStatus `gorm:"size:32;not null;default:'PENDING';index" json:"status"`
	Attempts             int                    `gorm:"default:0" json:"attempts"`
	MaxRetries           int                    `gorm:"default:5" json:"max_retries"`
	NextRetryAt          *time.Time             `gorm:"index" json:"next_retry_at,omitempty"`
	LastAttemptAt        *time.Time             `json:"last_attempt_at,omitempty"`
	DeliveredAt          *time.Time             `json:"delivered_at,omitempty"`
	LastHTTPCode         int                    `gorm:"default:0" json:"last_http_code"`
	LastResponse         string                 `gorm:"type:text" json:"last_response,omitempty"`
	LastError            string                 `gorm:"type:text" json:"last_error,omitempty"`
	IsReplay             bool                   `gorm:"default:false" json:"is_replay"`
	ReplayOfID           *uint64                `gorm:"index" json:"replay_of_id,omitempty"`
	CreatedAt            time.Time              `gorm:"index" json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// SourceAggregationValueItem represents actual unmodified aggregation values before payload serialization
type SourceAggregationValueItem struct {
	DeviceID             uint      `json:"device_id"`
	DeviceCode           string    `json:"device_code"`
	DeviceName           string    `json:"device_name"`
	ParameterID          uint      `json:"parameter_id"`
	ParameterCode        string    `json:"parameter_code"`
	ParameterName        string    `json:"parameter_name"`
	Value                *float64  `json:"value"`
	Unit                 string    `json:"unit"`
	Quality              string    `json:"quality"`
	AggregationResultID  uint64    `json:"aggregation_result_id"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// DeliveryLog records historical attempt entries
type DeliveryLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DestinationID uint      `gorm:"index;not null" json:"destination_id"`
	QueueItemID   *uint64   `gorm:"index" json:"queue_item_id,omitempty"`
	BatchCount    int       `json:"batch_count"`
	Status        string    `gorm:"size:32;not null" json:"status"` // SUCCESS, FAILED, RETRYING, ACKNOWLEDGED
	HTTPCode      int       `json:"http_code"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
	LatencyMs     int       `json:"latency_ms"`
	DeliveredAt   time.Time `gorm:"index;not null" json:"delivered_at"`
}

// DeliveryQueueSummary provides real-time counts across queue states
type DeliveryQueueSummary struct {
	Pending      int64 `json:"pending"`
	Sending      int64 `json:"sending"`
	Delivered    int64 `json:"delivered"`
	Retrying     int64 `json:"retrying"`
	Failed       int64 `json:"failed"`
	Acknowledged int64 `json:"acknowledged"`
	Total        int64 `json:"total"`
}

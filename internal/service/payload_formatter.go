package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"datalogger/internal/model"
)

// PayloadFormatter defines the interface for serializing aggregation batches into destination-specific JSON payloads
type PayloadFormatter interface {
	Format(batch *OutboundBatchData, cfg *model.DestinationPayloadConfig) ([]byte, error)
	ValidateConfig(cfg *model.DestinationPayloadConfig) error
	GeneratePreview(batch *OutboundBatchData, cfg *model.DestinationPayloadConfig) (string, error)
}

// OutboundBatchData encapsulates the aggregation data for a single destination dispatch
type OutboundBatchData struct {
	DeviceID        uint
	DeviceCode      string
	DeviceName      string
	IntervalSeconds int
	PeriodStart     time.Time
	PeriodEnd       time.Time
	Values          []model.SourceAggregationValueItem
}

// ConfigurablePayloadFormatter implements safe, deterministic JSON payload formatting
type ConfigurablePayloadFormatter struct{}

// NewConfigurablePayloadFormatter creates a new formatter instance
func NewConfigurablePayloadFormatter() *ConfigurablePayloadFormatter {
	return &ConfigurablePayloadFormatter{}
}

// ValidateConfig verifies that the payload configuration contains valid syntax, safe structure, and proper rules
func (f *ConfigurablePayloadFormatter) ValidateConfig(cfg *model.DestinationPayloadConfig) error {
	if cfg == nil {
		return fmt.Errorf("payload configuration cannot be nil")
	}

	if cfg.StationIDField == "" {
		cfg.StationIDField = "station_id"
	}
	if cfg.TimestampField == "" {
		cfg.TimestampField = "timestamp"
	}
	if cfg.IntervalField == "" {
		cfg.IntervalField = "interval_minutes"
	}
	if cfg.Structure == "" {
		cfg.Structure = "FLAT"
	}
	if cfg.Structure == "NESTED" && cfg.NestedDataKey == "" {
		cfg.NestedDataKey = "data"
	}

	// Validate StaticMetadataJSON if present
	if strings.TrimSpace(cfg.StaticMetadataJSON) != "" {
		var staticMap map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.StaticMetadataJSON), &staticMap); err != nil {
			return fmt.Errorf("invalid static_metadata_json: %w", err)
		}
	}

	// Validate ParameterMappingsJSON if present
	if strings.TrimSpace(cfg.ParameterMappingsJSON) != "" {
		var mappings []model.ParameterMappingItem
		if err := json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings); err != nil {
			return fmt.Errorf("invalid parameter_mappings_json: %w", err)
		}
	}

	// Validate RequiredFieldsJSON if present
	if strings.TrimSpace(cfg.RequiredFieldsJSON) != "" {
		var reqFields []string
		if err := json.Unmarshal([]byte(cfg.RequiredFieldsJSON), &reqFields); err != nil {
			return fmt.Errorf("invalid required_fields_json: %w", err)
		}
	}

	return nil
}

// Format transforms aggregation items into destination-specific JSON payload bytes
func (f *ConfigurablePayloadFormatter) Format(batch *OutboundBatchData, cfg *model.DestinationPayloadConfig) ([]byte, error) {
	if batch == nil {
		return nil, fmt.Errorf("outbound batch data cannot be nil")
	}
	if cfg == nil {
		cfg = &model.DestinationPayloadConfig{
			StationIDField:  "station_id",
			TimestampField:  "timestamp",
			TimestampFormat: "ISO8601",
			IntervalField:   "interval_minutes",
			Structure:       "FLAT",
		}
	}

	// 1. Build parameter mapping index and filter
	paramAliasMap := make(map[string]string)     // parameter_code -> external_name
	paramEnabledMap := make(map[string]bool)     // parameter_code -> enabled
	hasExplicitMappings := false

	if strings.TrimSpace(cfg.ParameterMappingsJSON) != "" {
		var mappings []model.ParameterMappingItem
		if err := json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings); err == nil && len(mappings) > 0 {
			hasExplicitMappings = true
			for _, m := range mappings {
				codeKey := strings.ToUpper(strings.TrimSpace(m.ParameterCode))
				paramEnabledMap[codeKey] = m.Enabled
				if m.ExternalName != "" {
					paramAliasMap[codeKey] = strings.TrimSpace(m.ExternalName)
				} else {
					paramAliasMap[codeKey] = m.ParameterCode
				}
			}
		}
	}

	// 2. Format timestamp
	tsStr, err := formatTimestamp(batch.PeriodEnd, cfg.TimestampFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to format timestamp: %w", err)
	}

	// 3. Resolve station ID
	stationID := batch.DeviceCode
	if strings.TrimSpace(cfg.StationIDValue) != "" {
		stationID = strings.TrimSpace(cfg.StationIDValue)
	}
	if stationID == "" {
		stationID = fmt.Sprintf("STATION_%d", batch.DeviceID)
	}

	// 4. Construct payload map
	rootMap := make(map[string]interface{})

	// Base identification
	stationKey := cfg.StationIDField
	if stationKey == "" {
		stationKey = "station_id"
	}
	rootMap[stationKey] = stationID

	timestampKey := cfg.TimestampField
	if timestampKey == "" {
		timestampKey = "timestamp"
	}
	rootMap[timestampKey] = tsStr

	// Interval in minutes (2, 5, 30, 60)
	intervalKey := cfg.IntervalField
	if intervalKey == "" {
		intervalKey = "interval_minutes"
	}
	intervalMinutes := batch.IntervalSeconds / 60
	if intervalMinutes <= 0 {
		intervalMinutes = 1
	}
	rootMap[intervalKey] = intervalMinutes

	// Optional metadata
	if cfg.IncludeMetadata {
		rootMap["period_start"] = batch.PeriodStart.Format(time.RFC3339)
		rootMap["period_end"] = batch.PeriodEnd.Format(time.RFC3339)
		rootMap["interval_seconds"] = batch.IntervalSeconds
		rootMap["device_name"] = batch.DeviceName
	}

	// Static metadata
	if strings.TrimSpace(cfg.StaticMetadataJSON) != "" {
		var staticMap map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.StaticMetadataJSON), &staticMap); err == nil {
			for k, v := range staticMap {
				rootMap[k] = v
			}
		}
	}

	// 5. Build telemetry parameters map
	measurements := make(map[string]interface{})

	for _, item := range batch.Values {
		codeKey := strings.ToUpper(strings.TrimSpace(item.ParameterCode))

		// Check filter if explicit mappings exist
		if hasExplicitMappings {
			enabled, exists := paramEnabledMap[codeKey]
			if exists && !enabled {
				continue // parameter filtered out for this destination
			}
		}

		targetFieldName := item.ParameterCode
		if alias, exists := paramAliasMap[codeKey]; exists && alias != "" {
			targetFieldName = alias
		}

		// Handle complex value (with units / quality) vs scalar value
		if cfg.IncludeUnits || cfg.IncludeQuality {
			detail := make(map[string]interface{})
			if item.Value != nil {
				detail["value"] = *item.Value
			} else {
				detail["value"] = nil // Do not fabricate or zero
			}
			if cfg.IncludeUnits && item.Unit != "" {
				detail["unit"] = item.Unit
			}
			if cfg.IncludeQuality && item.Quality != "" {
				detail["quality"] = item.Quality
			}
			measurements[targetFieldName] = detail
		} else {
			if item.Value != nil {
				measurements[targetFieldName] = *item.Value
			} else {
				measurements[targetFieldName] = nil // Missing data stays null, not fabricated 0
			}
		}
	}

	// 6. Merge measurements into root (FLAT vs NESTED)
	if strings.ToUpper(cfg.Structure) == "NESTED" {
		nestedKey := cfg.NestedDataKey
		if nestedKey == "" {
			nestedKey = "data"
		}
		rootMap[nestedKey] = measurements
	} else {
		// FLAT structure: measurements placed directly in root
		for k, v := range measurements {
			rootMap[k] = v
		}
	}

	// 7. Validate required fields
	if strings.TrimSpace(cfg.RequiredFieldsJSON) != "" {
		var reqFields []string
		if err := json.Unmarshal([]byte(cfg.RequiredFieldsJSON), &reqFields); err == nil {
			for _, rf := range reqFields {
				rf = strings.TrimSpace(rf)
				if rf == "" {
					continue
				}
				if _, ok := rootMap[rf]; !ok {
					// Check inside nested data if structure is nested
					if strings.ToUpper(cfg.Structure) == "NESTED" {
						if _, inNested := measurements[rf]; inNested {
							continue
						}
					}
					return nil, fmt.Errorf("required field %q is missing from generated payload", rf)
				}
			}
		}
	}

	// 8. Serialize to JSON
	serialized, err := json.Marshal(rootMap)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload to JSON: %w", err)
	}

	return serialized, nil
}

// GeneratePreview formats the batch data with indentation for live UI inspection
func (f *ConfigurablePayloadFormatter) GeneratePreview(batch *OutboundBatchData, cfg *model.DestinationPayloadConfig) (string, error) {
	rawBytes, err := f.Format(batch, cfg)
	if err != nil {
		return "", err
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(rawBytes, &parsed); err != nil {
		return string(rawBytes), nil
	}

	prettyBytes, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return string(rawBytes), nil
	}

	return string(prettyBytes), nil
}

func formatTimestamp(t time.Time, format string) (interface{}, error) {
	switch strings.ToUpper(strings.TrimSpace(format)) {
	case "UNIX_MS":
		return t.UnixNano() / int64(time.Millisecond), nil
	case "UNIX_SEC":
		return t.Unix(), nil
	case "CUSTOM":
		return t.Format("2006-01-02 15:04:05"), nil
	case "ISO8601", "":
		fallthrough
	default:
		return t.Format(time.RFC3339), nil
	}
}

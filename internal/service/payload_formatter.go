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
	DestinationID   uint
	DestinationName string
	IntervalSeconds int
	PeriodStart     time.Time
	PeriodEnd       time.Time
	DeviceCount     int
	DevicesSummary  string
	Devices         []BatchDeviceSummary
	Values          []model.SourceAggregationValueItem

	// Single device backward compatibility
	DeviceID   uint
	DeviceCode string
	DeviceName string
}

type BatchDeviceSummary struct {
	DeviceID   uint   `json:"device_id"`
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
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
	if cfg.ConflictResolution == "" {
		cfg.ConflictResolution = "FAIL_ON_CONFLICT"
	}
	if cfg.DeviceIdentityMode == "" {
		cfg.DeviceIdentityMode = "NONE"
	}
	if (cfg.Structure == "NESTED" || cfg.Structure == "NESTED_BY_DEVICE" || cfg.Structure == "ARRAY_BY_DEVICE") && cfg.NestedDataKey == "" {
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
			StationIDField:     "station_id",
			TimestampField:     "timestamp",
			TimestampFormat:    "ISO8601",
			IntervalField:      "interval_minutes",
			Structure:          "FLAT",
			ConflictResolution: "FAIL_ON_CONFLICT",
			DeviceIdentityMode: "NONE",
		}
	}

	// 1. Build parameter mapping index and filter (supporting device-specific and generic rules)
	paramAliasMap := make(map[string]string)     // key -> external_name
	paramEnabledMap := make(map[string]bool)     // key -> enabled
	hasExplicitMappings := false

	if strings.TrimSpace(cfg.ParameterMappingsJSON) != "" {
		var mappings []model.ParameterMappingItem
		if err := json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings); err == nil && len(mappings) > 0 {
			hasExplicitMappings = true
			for _, m := range mappings {
				code := strings.ToUpper(strings.TrimSpace(m.ParameterCode))
				extName := strings.TrimSpace(m.ExternalName)
				if extName == "" {
					extName = m.ParameterCode
				}

				// Device-specific key e.g. "DEV:5:PM25" or "AQMS-01:PM25"
				if m.DeviceID > 0 {
					devKey := fmt.Sprintf("DEV:%d:%s", m.DeviceID, code)
					paramEnabledMap[devKey] = m.Enabled
					paramAliasMap[devKey] = extName
				}
				if m.DeviceCode != "" {
					devCodeKey := fmt.Sprintf("%s:%s", strings.ToUpper(strings.TrimSpace(m.DeviceCode)), code)
					paramEnabledMap[devCodeKey] = m.Enabled
					paramAliasMap[devCodeKey] = extName
				}

				// Generic fallback key
				paramEnabledMap[code] = m.Enabled
				paramAliasMap[code] = extName
			}
		}
	}

	// 2. Format timestamp
	tsStr, err := formatTimestamp(batch.PeriodEnd, cfg.TimestampFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to format timestamp: %w", err)
	}

	// 3. Resolve station ID
	stationID := strings.TrimSpace(cfg.StationIDValue)
	if stationID == "" {
		if batch.DeviceCount > 1 && batch.DevicesSummary != "" {
			stationID = batch.DevicesSummary
		} else if batch.DeviceCode != "" {
			stationID = batch.DeviceCode
		} else if len(batch.Devices) > 0 {
			stationID = batch.Devices[0].DeviceCode
		} else if batch.DeviceID > 0 {
			stationID = fmt.Sprintf("STATION_%d", batch.DeviceID)
		} else {
			stationID = "STATION_EDGE"
		}
	}

	// 4. Construct root payload map
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

	// Optional metadata headers
	if cfg.IncludeMetadata {
		rootMap["period_start"] = batch.PeriodStart.Format(time.RFC3339)
		rootMap["period_end"] = batch.PeriodEnd.Format(time.RFC3339)
		rootMap["interval_seconds"] = batch.IntervalSeconds
		rootMap["device_count"] = batch.DeviceCount
		if batch.DevicesSummary != "" {
			rootMap["devices"] = batch.DevicesSummary
		}
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

	// 5. Structure measurements based on Structure & DeviceIdentityMode
	structure := strings.ToUpper(strings.TrimSpace(cfg.Structure))
	deviceMode := strings.ToUpper(strings.TrimSpace(cfg.DeviceIdentityMode))
	conflictResolution := strings.ToUpper(strings.TrimSpace(cfg.ConflictResolution))
	if conflictResolution == "" {
		conflictResolution = "FAIL_ON_CONFLICT"
	}

	// Helper to extract measurement value (scalar vs unit/quality dictionary)
	extractVal := func(item model.SourceAggregationValueItem) interface{} {
		if cfg.IncludeUnits || cfg.IncludeQuality {
			detail := make(map[string]interface{})
			if item.Value != nil {
				detail["value"] = *item.Value
			} else {
				detail["value"] = nil
			}
			if cfg.IncludeUnits && item.Unit != "" {
				detail["unit"] = item.Unit
			}
			if cfg.IncludeQuality && item.Quality != "" {
				detail["quality"] = item.Quality
			}
			if cfg.IncludeDeviceIdentity && item.DeviceCode != "" {
				detail["device"] = item.DeviceCode
			}
			return detail
		}
		if item.Value != nil {
			return *item.Value
		}
		return nil
	}

	switch {
	case structure == "NESTED_BY_DEVICE" || deviceMode == "NESTED":
		// Nested by Device: data[DeviceCode][FieldName] = Value
		deviceGroups := make(map[string]map[string]interface{})

		for _, item := range batch.Values {
			code := strings.ToUpper(strings.TrimSpace(item.ParameterCode))
			devKey := fmt.Sprintf("DEV:%d:%s", item.DeviceID, code)
			devCodeKey := fmt.Sprintf("%s:%s", strings.ToUpper(strings.TrimSpace(item.DeviceCode)), code)

			// Check filtering
			if hasExplicitMappings {
				enabled := true
				if en, exists := paramEnabledMap[devKey]; exists {
					enabled = en
				} else if en, exists := paramEnabledMap[devCodeKey]; exists {
					enabled = en
				} else if en, exists := paramEnabledMap[code]; exists {
					enabled = en
				}
				if !enabled {
					continue
				}
			}

			targetFieldName := item.ParameterCode
			if alias, exists := paramAliasMap[devKey]; exists && alias != "" {
				targetFieldName = alias
			} else if alias, exists := paramAliasMap[devCodeKey]; exists && alias != "" {
				targetFieldName = alias
			} else if alias, exists := paramAliasMap[code]; exists && alias != "" {
				targetFieldName = alias
			}

			devName := item.DeviceCode
			if devName == "" {
				devName = fmt.Sprintf("DEV_%d", item.DeviceID)
			}

			if _, ok := deviceGroups[devName]; !ok {
				deviceGroups[devName] = make(map[string]interface{})
			}
			deviceGroups[devName][targetFieldName] = extractVal(item)
		}

		nestedKey := cfg.NestedDataKey
		if nestedKey == "" {
			nestedKey = "data"
		}
		rootMap[nestedKey] = deviceGroups

	case structure == "ARRAY_BY_DEVICE":
		// Array of devices: devices: [ { "device": "AQMS-01", ... }, ... ]
		deviceMap := make(map[string]map[string]interface{})
		var deviceOrder []string

		for _, item := range batch.Values {
			code := strings.ToUpper(strings.TrimSpace(item.ParameterCode))
			devKey := fmt.Sprintf("DEV:%d:%s", item.DeviceID, code)
			devCodeKey := fmt.Sprintf("%s:%s", strings.ToUpper(strings.TrimSpace(item.DeviceCode)), code)

			if hasExplicitMappings {
				enabled := true
				if en, exists := paramEnabledMap[devKey]; exists {
					enabled = en
				} else if en, exists := paramEnabledMap[devCodeKey]; exists {
					enabled = en
				} else if en, exists := paramEnabledMap[code]; exists {
					enabled = en
				}
				if !enabled {
					continue
				}
			}

			targetFieldName := item.ParameterCode
			if alias, exists := paramAliasMap[devKey]; exists && alias != "" {
				targetFieldName = alias
			} else if alias, exists := paramAliasMap[devCodeKey]; exists && alias != "" {
				targetFieldName = alias
			} else if alias, exists := paramAliasMap[code]; exists && alias != "" {
				targetFieldName = alias
			}

			devName := item.DeviceCode
			if devName == "" {
				devName = fmt.Sprintf("DEV_%d", item.DeviceID)
			}

			if _, ok := deviceMap[devName]; !ok {
				deviceMap[devName] = map[string]interface{}{
					"device_code": devName,
					"device_name": item.DeviceName,
				}
				deviceOrder = append(deviceOrder, devName)
			}
			deviceMap[devName][targetFieldName] = extractVal(item)
		}

		var deviceList []map[string]interface{}
		for _, devName := range deviceOrder {
			deviceList = append(deviceList, deviceMap[devName])
		}

		nestedKey := cfg.NestedDataKey
		if nestedKey == "" || nestedKey == "data" {
			nestedKey = "devices"
		}
		rootMap[nestedKey] = deviceList

	default:
		// FLAT or NESTED flat measurements map
		measurements := make(map[string]interface{})
		keySourceDevice := make(map[string]string)     // targetKey -> DeviceCode
		keyValuesList := make(map[string][]float64)     // targetKey -> []float64 (for AVERAGE)

		for _, item := range batch.Values {
			code := strings.ToUpper(strings.TrimSpace(item.ParameterCode))
			devKey := fmt.Sprintf("DEV:%d:%s", item.DeviceID, code)
			devCodeKey := fmt.Sprintf("%s:%s", strings.ToUpper(strings.TrimSpace(item.DeviceCode)), code)

			if hasExplicitMappings {
				enabled := true
				if en, exists := paramEnabledMap[devKey]; exists {
					enabled = en
				} else if en, exists := paramEnabledMap[devCodeKey]; exists {
					enabled = en
				} else if en, exists := paramEnabledMap[code]; exists {
					enabled = en
				}
				if !enabled {
					continue
				}
			}

			targetFieldName := item.ParameterCode
			if alias, exists := paramAliasMap[devKey]; exists && alias != "" {
				targetFieldName = alias
			} else if alias, exists := paramAliasMap[devCodeKey]; exists && alias != "" {
				targetFieldName = alias
			} else if alias, exists := paramAliasMap[code]; exists && alias != "" {
				targetFieldName = alias
			}

			// Apply PREFIX mode if configured or includeDeviceIdentity or conflictResolution == PREFIX
			if deviceMode == "PREFIX" || conflictResolution == "PREFIX" || (cfg.IncludeDeviceIdentity && deviceMode != "TAG") {
				if item.DeviceCode != "" {
					targetFieldName = fmt.Sprintf("%s_%s", strings.ToLower(item.DeviceCode), strings.ToLower(targetFieldName))
				}
			}

			val := extractVal(item)

			// Conflict Handling: check if key already populated by another device
			prevDev, conflict := keySourceDevice[targetFieldName]
			if conflict && prevDev != item.DeviceCode {
				switch conflictResolution {
				case "PREFIX":
					prefixedKey := fmt.Sprintf("%s_%s", strings.ToLower(item.DeviceCode), strings.ToLower(targetFieldName))
					measurements[prefixedKey] = val
					keySourceDevice[prefixedKey] = item.DeviceCode
				case "ARRAY":
					// Store as value array
					switch existing := measurements[targetFieldName].(type) {
					case []interface{}:
						measurements[targetFieldName] = append(existing, val)
					default:
						measurements[targetFieldName] = []interface{}{existing, val}
					}
				case "PRIORITY_PRIMARY":
					// Keep first device, ignore second
					continue
				case "AVERAGE":
					// Average numeric values
					if item.Value != nil {
						keyValuesList[targetFieldName] = append(keyValuesList[targetFieldName], *item.Value)
						sum := 0.0
						for _, v := range keyValuesList[targetFieldName] {
							sum += v
						}
						avg := sum / float64(len(keyValuesList[targetFieldName]))
						measurements[targetFieldName] = avg
					}
				case "CUSTOM_MAPPING":
					// If custom mapping was chosen but still collides:
					return nil, fmt.Errorf("parameter key collision on '%s' between device %s and %s; please specify distinct external aliases per device", targetFieldName, prevDev, item.DeviceCode)
				default: // FAIL_ON_CONFLICT
					return nil, fmt.Errorf("parameter key collision on '%s' between device %s and %s; configure distinct external aliases or enable device grouping", targetFieldName, prevDev, item.DeviceCode)
				}
			} else {
				measurements[targetFieldName] = val
				keySourceDevice[targetFieldName] = item.DeviceCode
				if item.Value != nil {
					keyValuesList[targetFieldName] = []float64{*item.Value}
				}
			}
		}

		if structure == "NESTED" {
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
					nestedKey := cfg.NestedDataKey
					if nestedKey == "" {
						nestedKey = "data"
					}
					if nestedMap, ok := rootMap[nestedKey].(map[string]interface{}); ok {
						if _, inNested := nestedMap[rf]; inNested {
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

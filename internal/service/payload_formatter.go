package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"
)

// fixed2Float serializes numeric measurement values with 2 digits after the decimal point (e.g. 23.33, 88.50).
type fixed2Float float64

func (f fixed2Float) MarshalJSON() ([]byte, error) {
	rounded := roundOutboundValue(float64(f))
	return []byte(strconv.FormatFloat(rounded, 'f', 2, 64)), nil
}

func roundOutboundValue(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	rounded := math.Round(v*100) / 100
	if rounded == 0 {
		return 0
	}
	return rounded
}

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
	if !cfg.OmitIntervalField && cfg.IntervalField == "" {
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

type parameterMappingMatcher struct {
	templateType              string
	hasExplicitMappings       bool
	hasDeviceSpecificMappings bool
	devParamIDEnabled         map[string]bool
	devParamIDAlias           map[string]string
	devEnabledMap             map[string]bool
	devAliasMap               map[string]string
	wildcardEnabledMap        map[string]bool
	wildcardAliasMap          map[string]string
}

func defaultKLHKISPUAlias(code string) string {
	trimmed := strings.TrimSpace(code)
	switch strings.ToLower(trimmed) {
	case "pm10":
		return "pm10"
	case "pm2.5", "pm25", "pm2_5":
		return "pm25"
	case "so2":
		return "so2"
	case "co":
		return "co"
	case "o3":
		return "o3"
	case "no2":
		return "no2"
	case "hc", "nmhc":
		return "hc"
	case "w-speed", "wspeed", "ws", "wind_speed", "wind-speed":
		return "ws"
	case "wind-angle", "wind_angle", "wd", "wind_dir", "wind-dir":
		return "wd"
	case "humidity", "hum", "rh":
		return "humidity"
	case "temperature", "temp":
		return "temperature"
	case "air-pressure", "air_pressure", "pressure", "press":
		return "pressure"
	case "solar-radiation", "solar_radiation", "intensity", "sr":
		return "sr"
	case "rain-fall", "rain_fall", "rainfall", "tprecipitation", "rain_intensity", "rain-intensity":
		return "rain_intensity"
	default:
		return trimmed
	}
}

func buildParameterMappingMatcher(cfg *model.DestinationPayloadConfig) *parameterMappingMatcher {
	matcher := &parameterMappingMatcher{
		devParamIDEnabled:  make(map[string]bool),
		devParamIDAlias:    make(map[string]string),
		devEnabledMap:      make(map[string]bool),
		devAliasMap:        make(map[string]string),
		wildcardEnabledMap: make(map[string]bool),
		wildcardAliasMap:   make(map[string]string),
	}
	if cfg == nil {
		return matcher
	}
	matcher.templateType = strings.ToUpper(strings.TrimSpace(cfg.TemplateType))
	if strings.TrimSpace(cfg.ParameterMappingsJSON) == "" {
		return matcher
	}
	var mappings []model.ParameterMappingItem
	if err := json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings); err != nil || len(mappings) == 0 {
		return matcher
	}
	matcher.hasExplicitMappings = true
	for _, m := range mappings {
		code := strings.ToUpper(strings.TrimSpace(m.ParameterCode))
		if code == "" {
			continue
		}
		extName := strings.TrimSpace(m.ExternalName)
		if extName == "" {
			extName = strings.TrimSpace(m.ParameterCode)
		}
		devCodeUpper := strings.ToUpper(strings.TrimSpace(m.DeviceCode))
		isDeviceScoped := m.DeviceID > 0 || devCodeUpper != ""

		if isDeviceScoped {
			matcher.hasDeviceSpecificMappings = true
			if m.DeviceID > 0 && m.ParameterID > 0 {
				pidKey := fmt.Sprintf("DEV:%d:PID:%d", m.DeviceID, m.ParameterID)
				matcher.devParamIDEnabled[pidKey] = m.Enabled
				matcher.devParamIDAlias[pidKey] = extName
			}
			if m.DeviceID > 0 {
				devKey := fmt.Sprintf("DEV:%d:%s", m.DeviceID, code)
				matcher.devEnabledMap[devKey] = m.Enabled
				matcher.devAliasMap[devKey] = extName
			}
			if devCodeUpper != "" {
				devCodeKey := fmt.Sprintf("%s:%s", devCodeUpper, code)
				matcher.devEnabledMap[devCodeKey] = m.Enabled
				matcher.devAliasMap[devCodeKey] = extName
			}
		} else {
			matcher.wildcardEnabledMap[code] = m.Enabled
			matcher.wildcardAliasMap[code] = extName
		}
	}
	return matcher
}

func (m *parameterMappingMatcher) Resolve(deviceID uint, deviceCode string, paramID uint, paramCode string) (bool, string) {
	trimmedCode := strings.TrimSpace(paramCode)
	if trimmedCode == "" {
		return false, ""
	}
	if m == nil || !m.hasExplicitMappings {
		if m != nil && m.templateType == "KLHK_ISPU_AQMS" {
			return true, defaultKLHKISPUAlias(trimmedCode)
		}
		return true, trimmedCode
	}
	codeUpper := strings.ToUpper(trimmedCode)
	devCodeUpper := strings.ToUpper(strings.TrimSpace(deviceCode))

	if deviceID > 0 && paramID > 0 {
		pidKey := fmt.Sprintf("DEV:%d:PID:%d", deviceID, paramID)
		if en, ok := m.devParamIDEnabled[pidKey]; ok {
			alias := m.devParamIDAlias[pidKey]
			if alias == "" {
				alias = trimmedCode
			}
			return en, alias
		}
	}
	if deviceID > 0 {
		devKey := fmt.Sprintf("DEV:%d:%s", deviceID, codeUpper)
		if en, ok := m.devEnabledMap[devKey]; ok {
			alias := m.devAliasMap[devKey]
			if alias == "" {
				alias = trimmedCode
			}
			return en, alias
		}
	}
	if devCodeUpper != "" {
		devCodeKey := fmt.Sprintf("%s:%s", devCodeUpper, codeUpper)
		if en, ok := m.devEnabledMap[devCodeKey]; ok {
			alias := m.devAliasMap[devCodeKey]
			if alias == "" {
				alias = trimmedCode
			}
			return en, alias
		}
	}
	if en, ok := m.wildcardEnabledMap[codeUpper]; ok {
		alias := m.wildcardAliasMap[codeUpper]
		if alias == "" {
			alias = trimmedCode
		}
		return en, alias
	}

	if m.hasDeviceSpecificMappings && (deviceID > 0 || devCodeUpper != "") {
		return false, trimmedCode
	}
	return true, trimmedCode
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

	// 1. Build parameter mapping index and filter (strictly isolating device-scoped checklists from wildcards)
	matcher := buildParameterMappingMatcher(cfg)

	// 2. Format timestamp
	tsStr, err := formatTimestamp(batch.PeriodEnd, cfg.TimestampFormat, cfg.TimestampTimezone)
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

	// Interval in minutes (2, 5, 30, 60) — omitted when OmitIntervalField is true (e.g. strict KLHK ISPU AQMS schema)
	if !cfg.OmitIntervalField {
		intervalKey := strings.TrimSpace(cfg.IntervalField)
		if intervalKey == "" {
			intervalKey = "interval_minutes"
		}
		intervalMinutes := batch.IntervalSeconds / 60
		if intervalMinutes <= 0 {
			intervalMinutes = 1
		}
		rootMap[intervalKey] = intervalMinutes
	}

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
				detail["value"] = fixed2Float(*item.Value)
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
			return fixed2Float(*item.Value)
		}
		return nil
	}

	switch {
	case structure == "MULTI_TICK_ARRAY" || structure == "TICK_ARRAY":
		var tickList []map[string]interface{}

		// Resolve module_id
		var resolvedModuleID interface{} = 1
		if strings.TrimSpace(cfg.ModuleIDValue) != "" {
			if num, err := strconv.ParseInt(strings.TrimSpace(cfg.ModuleIDValue), 10, 64); err == nil {
				resolvedModuleID = num
			} else {
				resolvedModuleID = strings.TrimSpace(cfg.ModuleIDValue)
			}
		}

		// Resolve station_id
		var resolvedStationID interface{} = stationID
		if num, err := strconv.ParseInt(stationID, 10, 64); err == nil {
			resolvedStationID = num
		}

		// Check static metadata overrides
		if strings.TrimSpace(cfg.StaticMetadataJSON) != "" {
			var staticMap map[string]interface{}
			if err := json.Unmarshal([]byte(cfg.StaticMetadataJSON), &staticMap); err == nil {
				if v, ok := staticMap["module_id"]; ok && v != nil {
					resolvedModuleID = v
				}
				if v, ok := staticMap["station_id"]; ok && v != nil {
					resolvedStationID = v
				}
			}
		}

		createdAtTs := tsStr
		if !batch.PeriodStart.IsZero() {
			if cs, err := formatTimestamp(batch.PeriodStart, cfg.TimestampFormat, cfg.TimestampTimezone); err == nil {
				createdAtTs = cs
			}
		}
		updatedAtTs := tsStr

		for _, item := range batch.Values {
			enabled, sensorName := matcher.Resolve(item.DeviceID, item.DeviceCode, item.ParameterID, item.ParameterCode)
			if !enabled {
				continue
			}

			modID := resolvedModuleID
			if cfg.ModuleIDValue == "" && item.DeviceID > 0 {
				modID = item.DeviceID
			}

			var val interface{}
			if item.Value != nil {
				val = fixed2Float(*item.Value)
			} else {
				val = fixed2Float(0)
			}

			tickMap := map[string]interface{}{
				"module_id":   modID,
				"station_id":  resolvedStationID,
				"sensor_name": sensorName,
				"value":       val,
				"created_at":  createdAtTs,
				"updated_at":  updatedAtTs,
			}

			tickList = append(tickList, map[string]interface{}{
				"tick": tickMap,
			})
		}

		if tickList == nil {
			tickList = []map[string]interface{}{}
		}

		rootMap = map[string]interface{}{
			"multi": tickList,
		}

	case structure == "NESTED_BY_DEVICE" || deviceMode == "NESTED":
		// Nested by Device: data[DeviceCode][FieldName] = Value
		deviceGroups := make(map[string]map[string]interface{})

		for _, item := range batch.Values {
			enabled, targetFieldName := matcher.Resolve(item.DeviceID, item.DeviceCode, item.ParameterID, item.ParameterCode)
			if !enabled {
				continue
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
			enabled, targetFieldName := matcher.Resolve(item.DeviceID, item.DeviceCode, item.ParameterID, item.ParameterCode)
			if !enabled {
				continue
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
		keySourceDevice := make(map[string]string) // targetKey -> DeviceCode
		keyValuesList := make(map[string][]float64) // targetKey -> []float64 (for AVERAGE)
		validPollutantStat := make(map[string]bool)

		for _, item := range batch.Values {
			enabled, targetFieldName := matcher.Resolve(item.DeviceID, item.DeviceCode, item.ParameterID, item.ParameterCode)
			if !enabled {
				continue
			}

			if item.Value != nil && strings.ToUpper(strings.TrimSpace(item.Quality)) != "BAD" {
				validPollutantStat[normalizeStatPollutantKey(targetFieldName)] = true
				validPollutantStat[normalizeStatPollutantKey(item.ParameterCode)] = true
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
						measurements[targetFieldName] = fixed2Float(avg)
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

		// Inject KLHK ISPU AQMS sensor status flags (stat_pm10 .. stat_hc) when enabled
		if cfg.IncludeSensorStatFlags {
			isAutoQuality := strings.EqualFold(strings.TrimSpace(cfg.SensorStatMode), "AUTO_QUALITY")
			for _, code := range []string{"pm10", "pm25", "so2", "co", "o3", "no2", "hc"} {
				statKey := "stat_" + code
				if _, alreadyInStatic := rootMap[statKey]; alreadyInStatic && structure != "NESTED" {
					continue
				}
				if _, alreadyInMeas := measurements[statKey]; alreadyInMeas {
					continue
				}
				if isAutoQuality {
					if validPollutantStat[code] {
						measurements[statKey] = 1
					} else {
						measurements[statKey] = 0
					}
				} else {
					measurements[statKey] = 1
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

	var buf bytes.Buffer
	if err := json.Indent(&buf, rawBytes, "", "  "); err != nil {
		return string(rawBytes), nil
	}

	return buf.String(), nil
}

func normalizeStatPollutantKey(s string) string {
	k := strings.ToLower(strings.TrimSpace(s))
	k = strings.ReplaceAll(k, ".", "")
	k = strings.ReplaceAll(k, "_", "")
	k = strings.ReplaceAll(k, "-", "")
	return k
}

func resolveTimezoneLocation(tz string) *time.Location {
	trimmed := strings.TrimSpace(tz)
	if trimmed == "" {
		return nil
	}
	upper := strings.ToUpper(trimmed)
	switch upper {
	case "UTC":
		return time.UTC
	case "ASIA/JAKARTA", "WIB", "UTC+7", "+07:00":
		if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
			return loc
		}
		return time.FixedZone("WIB", 7*3600)
	case "ASIA/MAKASSAR", "WITA", "UTC+8", "+08:00":
		if loc, err := time.LoadLocation("Asia/Makassar"); err == nil {
			return loc
		}
		return time.FixedZone("WITA", 8*3600)
	case "ASIA/JAYAPURA", "WIT", "UTC+9", "+09:00":
		if loc, err := time.LoadLocation("Asia/Jayapura"); err == nil {
			return loc
		}
		return time.FixedZone("WIT", 9*3600)
	default:
		if loc, err := time.LoadLocation(trimmed); err == nil {
			return loc
		}
		return nil
	}
}

func formatTimestamp(t time.Time, format string, tz string) (interface{}, error) {
	if loc := resolveTimezoneLocation(tz); loc != nil {
		t = t.In(loc)
	}
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

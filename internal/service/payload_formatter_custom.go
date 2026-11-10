package service

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"
)

// resolveCustomSchemaMode determines whether the flexible user-defined Custom Schema engine is active.
// Legacy configs with TemplateType="CUSTOM" and empty CustomSchemaMode (Structure != "CUSTOM_SCHEMA")
// retain their original FLAT behavior for 100% backward compatibility.
func resolveCustomSchemaMode(cfg *model.DestinationPayloadConfig) (string, bool) {
	if cfg == nil {
		return "", false
	}
	struc := strings.ToUpper(strings.TrimSpace(cfg.Structure))
	tmpl := strings.ToUpper(strings.TrimSpace(cfg.TemplateType))
	mode := strings.ToUpper(strings.TrimSpace(cfg.CustomSchemaMode))

	if struc != "CUSTOM_SCHEMA" && !(tmpl == "CUSTOM" && mode != "") {
		return "", false
	}

	switch mode {
	case "JSON_TEMPLATE":
		return "JSON_TEMPLATE", true
	case "VISUAL_FIELDS":
		return "VISUAL_FIELDS", true
	default:
		if strings.TrimSpace(cfg.CustomTemplateJSON) != "" && strings.TrimSpace(cfg.CustomFieldsJSON) == "" {
			return "JSON_TEMPLATE", true
		}
		return "VISUAL_FIELDS", true
	}
}

// extractCustomReferencedParamTokens collects parameter codes/aliases explicitly referenced in
// CustomFieldsJSON (Source="PARAMETER") or CustomTemplateJSON ({{param:...}}, {{unit:...}}, {{quality:...}})
// so OutboundService includes them in batch.Values even if unchecked in {{values}}.
func extractCustomReferencedParamTokens(cfg *model.DestinationPayloadConfig) map[string]bool {
	refs := make(map[string]bool)
	mode, active := resolveCustomSchemaMode(cfg)
	if !active {
		return refs
	}

	if mode == "VISUAL_FIELDS" && strings.TrimSpace(cfg.CustomFieldsJSON) != "" {
		var fields []model.CustomPayloadFieldItem
		if err := json.Unmarshal([]byte(cfg.CustomFieldsJSON), &fields); err == nil {
			for _, f := range fields {
				if !f.Enabled {
					continue
				}
				if strings.EqualFold(strings.TrimSpace(f.Source), "PARAMETER") {
					tok := parseParamRefKey(f.Value)
					if tok != "" {
						refs[strings.ToUpper(tok)] = true
					}
				}
			}
		}
	}

	if mode == "JSON_TEMPLATE" && strings.TrimSpace(cfg.CustomTemplateJSON) != "" {
		raw := cfg.CustomTemplateJSON
		idx := 0
		for idx < len(raw) {
			start := strings.Index(raw[idx:], "{{")
			if start < 0 {
				break
			}
			start += idx
			end := strings.Index(raw[start+2:], "}}")
			if end < 0 {
				break
			}
			inner := strings.TrimSpace(raw[start+2 : start+2+end])
			idx = start + 2 + end + 2

			lower := strings.ToLower(inner)
			if strings.HasPrefix(lower, "str:") {
				inner = strings.TrimSpace(inner[4:])
				lower = strings.ToLower(inner)
			}
			for _, prefix := range []string{"param:", "unit:", "quality:"} {
				if strings.HasPrefix(lower, prefix) {
					tok := parseParamRefKey(inner[len(prefix):])
					if tok != "" {
						refs[strings.ToUpper(tok)] = true
					}
				}
			}
		}
	}

	return refs
}

func parseParamRefKey(expr string) string {
	trimmed := strings.TrimSpace(expr)
	if pipeIdx := strings.Index(trimmed, "|"); pipeIdx >= 0 {
		trimmed = strings.TrimSpace(trimmed[:pipeIdx])
	}
	if colonIdx := strings.LastIndex(trimmed, ":"); colonIdx >= 0 {
		// Support both "DEV01:ph" and plain "ph" — register the parameter code part as well
		after := strings.TrimSpace(trimmed[colonIdx+1:])
		if after != "" {
			return after
		}
	}
	return trimmed
}

func validateCustomSchemaConfig(cfg *model.DestinationPayloadConfig, mode string) error {
	cfg.Structure = "CUSTOM_SCHEMA"
	cfg.CustomSchemaMode = mode

	switch mode {
	case "VISUAL_FIELDS":
		if strings.TrimSpace(cfg.CustomFieldsJSON) == "" {
			return fmt.Errorf("custom_fields_json is required when custom_schema_mode is VISUAL_FIELDS")
		}
		var fields []model.CustomPayloadFieldItem
		if err := json.Unmarshal([]byte(cfg.CustomFieldsJSON), &fields); err != nil {
			return fmt.Errorf("invalid custom_fields_json: %w", err)
		}
		enabledCount := 0
		for i, f := range fields {
			if !f.Enabled {
				continue
			}
			enabledCount++
			src := strings.ToUpper(strings.TrimSpace(f.Source))
			key := strings.TrimSpace(f.Key)
			if src != "ALL_PARAMETERS" && key == "" {
				return fmt.Errorf("custom field row #%d requires a JSON field name", i+1)
			}
			switch src {
			case "STATIC_NUMBER":
				if _, err := strconv.ParseFloat(strings.TrimSpace(f.Value), 64); err != nil {
					return fmt.Errorf("custom field %q has invalid STATIC_NUMBER value %q", key, f.Value)
				}
			case "STATIC_JSON":
				rawVal := strings.TrimSpace(f.Value)
				if rawVal == "" || !json.Valid([]byte(rawVal)) {
					return fmt.Errorf("custom field %q has invalid STATIC_JSON literal", key)
				}
			case "PARAMETER":
				if strings.TrimSpace(f.Value) == "" {
					return fmt.Errorf("custom field %q with source PARAMETER requires a parameter code", key)
				}
			}
		}
		if enabledCount == 0 {
			return fmt.Errorf("custom schema requires at least one enabled field")
		}

	case "JSON_TEMPLATE":
		rawTpl := strings.TrimSpace(cfg.CustomTemplateJSON)
		if rawTpl == "" {
			return fmt.Errorf("custom_template_json is required when custom_schema_mode is JSON_TEMPLATE")
		}
		normalized := normalizeUnquotedTemplatePlaceholders(rawTpl)
		parsed, err := decodeOrderedJSONDocument(normalized)
		if err != nil {
			return fmt.Errorf("invalid JSON template syntax: %w", err)
		}
		switch parsed.(type) {
		case orderedJSONObject, []interface{}:
			// Valid root JSON object or array
		default:
			return fmt.Errorf("JSON template root must be a JSON object {...} or array [...]")
		}
	}

	return nil
}

// normalizeUnquotedTemplatePlaceholders wraps unquoted {{...}} tokens outside JSON string literals
// in double quotes so standard RFC 8259 JSON decoding succeeds while preserving native typed expansion.
func normalizeUnquotedTemplatePlaceholders(raw string) string {
	var b strings.Builder
	b.Grow(len(raw) + 32)
	inString := false
	escaped := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if inString {
			b.WriteByte(ch)
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			b.WriteByte(ch)
			continue
		}

		if ch == '{' && i+1 < len(raw) && raw[i+1] == '{' {
			endRel := strings.Index(raw[i+2:], "}}")
			if endRel >= 0 {
				inner := raw[i+2 : i+2+endRel]
				quoted, err := json.Marshal("{{" + inner + "}}")
				if err == nil {
					b.Write(quoted)
					i = i + 2 + endRel + 1
					continue
				}
			}
		}

		b.WriteByte(ch)
	}
	return b.String()
}

func decodeOrderedJSONDocument(input string) (interface{}, error) {
	dec := json.NewDecoder(strings.NewReader(input))
	dec.UseNumber()
	val, err := decodeOrderedJSONTokenStream(dec)
	if err != nil {
		return nil, err
	}
	if extraTok, extraErr := dec.Token(); extraErr != io.EOF {
		if extraErr != nil {
			return nil, extraErr
		}
		return nil, fmt.Errorf("unexpected trailing token %v", extraTok)
	}
	return val, nil
}

func decodeOrderedJSONTokenStream(dec *json.Decoder) (interface{}, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return decodeOrderedJSONFromFirstToken(dec, tok)
}

func decodeOrderedJSONFromFirstToken(dec *json.Decoder, tok json.Token) (interface{}, error) {
	if delim, ok := tok.(json.Delim); ok {
		switch delim {
		case '{':
			fields := make(orderedJSONObject, 0)
			for dec.More() {
				kTok, kErr := dec.Token()
				if kErr != nil {
					return nil, kErr
				}
				keyStr, isStr := kTok.(string)
				if !isStr {
					return nil, fmt.Errorf("expected string key in JSON object")
				}
				vVal, vErr := decodeOrderedJSONTokenStream(dec)
				if vErr != nil {
					return nil, vErr
				}
				fields = append(fields, orderedJSONField{Key: keyStr, Value: vVal})
			}
			endTok, endErr := dec.Token()
			if endErr != nil {
				return nil, endErr
			}
			if d, ok := endTok.(json.Delim); !ok || d != '}' {
				return nil, fmt.Errorf("expected closing '}'")
			}
			return fields, nil

		case '[':
			items := make([]interface{}, 0)
			for dec.More() {
				elem, eErr := decodeOrderedJSONTokenStream(dec)
				if eErr != nil {
					return nil, eErr
				}
				items = append(items, elem)
			}
			endTok, endErr := dec.Token()
			if endErr != nil {
				return nil, endErr
			}
			if d, ok := endTok.(json.Delim); !ok || d != ']' {
				return nil, fmt.Errorf("expected closing ']'")
			}
			return items, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter %v", delim)
		}
	}
	return tok, nil
}

type customSchemaEvalContext struct {
	batch            *OutboundBatchData
	cfg              *model.DestinationPayloadConfig
	stationID        string
	deviceCode       string
	deviceName       string
	devicesSummary   string
	deviceCount      int
	intervalSeconds  int
	intervalMinutes  int
	tsVal            interface{}
	localBucketTime  time.Time
	localPeriodStart time.Time
	localPeriodEnd   time.Time
	valuesObj        orderedJSONObject
	recordsValuesArr []interface{}
	recordsFlatArr   []interface{}
	paramByKey       map[string]model.SourceAggregationValueItem
	extractVal       func(model.SourceAggregationValueItem) interface{}
}

func buildCustomSchemaEvalContext(
	batch *OutboundBatchData,
	cfg *model.DestinationPayloadConfig,
	matcher *parameterMappingMatcher,
	stationID string,
	tsVal interface{},
	rootBucketTime time.Time,
	extractVal func(model.SourceAggregationValueItem) interface{},
) (*customSchemaEvalContext, error) {
	loc := resolveTimezoneLocation(cfg.TimestampTimezone)
	localBucket := rootBucketTime
	localStart := batch.PeriodStart
	localEnd := batch.PeriodEnd
	if loc != nil {
		localBucket = localBucket.In(loc)
		if !localStart.IsZero() {
			localStart = localStart.In(loc)
		}
		if !localEnd.IsZero() {
			localEnd = localEnd.In(loc)
		}
	}

	devCode := strings.TrimSpace(batch.DeviceCode)
	devName := strings.TrimSpace(batch.DeviceName)
	if devCode == "" && len(batch.Devices) > 0 {
		devCode = batch.Devices[0].DeviceCode
		devName = batch.Devices[0].DeviceName
	}
	if devCode == "" {
		devCode = stationID
	}

	intervalMin := batch.IntervalSeconds / 60
	if intervalMin <= 0 {
		intervalMin = 1
	}

	deviceMode := strings.ToUpper(strings.TrimSpace(cfg.DeviceIdentityMode))
	conflictResolution := strings.ToUpper(strings.TrimSpace(cfg.ConflictResolution))
	if conflictResolution == "" {
		conflictResolution = "PRIORITY_PRIMARY"
	}

	// Build mapping preferred order from ParameterMappingsJSON if defined
	var mappingPreferredOrder []string
	if strings.TrimSpace(cfg.ParameterMappingsJSON) != "" {
		var mappings []model.ParameterMappingItem
		if err := json.Unmarshal([]byte(cfg.ParameterMappingsJSON), &mappings); err == nil {
			seenPref := make(map[string]bool)
			for _, m := range mappings {
				if !m.Enabled {
					continue
				}
				ext := strings.TrimSpace(m.ExternalName)
				if ext == "" {
					ext = strings.TrimSpace(m.ParameterCode)
				}
				if ext != "" && !seenPref[ext] {
					seenPref[ext] = true
					mappingPreferredOrder = append(mappingPreferredOrder, ext)
				}
			}
		}
	}

	// Group items chronologically by bucket timestamp AND track latest values
	type bucketAcc struct {
		rowTime       time.Time
		measurements  map[string]interface{}
		keyOrder      []string
		keySourceDev  map[string]string
		keyValuesList map[string][]float64
	}
	bucketMap := make(map[int64]*bucketAcc)
	var bucketKeys []int64

	paramByKey := make(map[string]model.SourceAggregationValueItem)
	paramLatestTs := make(map[string]int64)

	for _, item := range batch.Values {
		pStart := item.PeriodStart
		pEnd := item.PeriodEnd
		if pStart.IsZero() && pEnd.IsZero() {
			pStart = batch.PeriodStart
			pEnd = batch.PeriodEnd
		}
		rowTime := resolveBucketTimestamp(pStart, pEnd, cfg.TimestampSource)
		bKey := rowTime.UnixNano()

		// Register in direct parameter lookup (keeping newest bucket reading when multi-bucket)
		_, mappedAlias := matcher.ResolveForValuesCollection(item.DeviceID, item.DeviceCode, item.ParameterID, item.ParameterCode, item.ParameterName)
		lookupKeys := []string{
			strings.ToUpper(strings.TrimSpace(item.ParameterCode)),
		}
		if strings.TrimSpace(mappedAlias) != "" {
			lookupKeys = append(lookupKeys, strings.ToUpper(strings.TrimSpace(mappedAlias)))
		}
		if strings.TrimSpace(item.DeviceCode) != "" {
			lookupKeys = append(lookupKeys, strings.ToUpper(strings.TrimSpace(item.DeviceCode)+":"+strings.TrimSpace(item.ParameterCode)))
		}
		for _, lk := range lookupKeys {
			if lk == "" {
				continue
			}
			if prevTs, exists := paramLatestTs[lk]; !exists || bKey >= prevTs {
				paramLatestTs[lk] = bKey
				paramByKey[lk] = item
			}
		}

		// Populate {{values}} / {{records}} strictly from parameters enabled in Parameter External Aliases & Filters
		enabledInValues, targetFieldName := matcher.ResolveForValuesCollection(item.DeviceID, item.DeviceCode, item.ParameterID, item.ParameterCode, item.ParameterName)
		if !enabledInValues {
			continue
		}
		if deviceMode == "PREFIX" || conflictResolution == "PREFIX" {
			if item.DeviceCode != "" {
				targetFieldName = fmt.Sprintf("%s_%s", strings.ToLower(item.DeviceCode), targetFieldName)
			}
		}

		acc, exists := bucketMap[bKey]
		if !exists {
			acc = &bucketAcc{
				rowTime:       rowTime,
				measurements:  make(map[string]interface{}),
				keySourceDev:  make(map[string]string),
				keyValuesList: make(map[string][]float64),
			}
			bucketMap[bKey] = acc
			bucketKeys = append(bucketKeys, bKey)
		}

		val := extractVal(item)
		prevDev, conflict := acc.keySourceDev[targetFieldName]
		if conflict && prevDev != item.DeviceCode {
			switch conflictResolution {
			case "FAIL_ON_CONFLICT":
				return nil, fmt.Errorf("parameter key collision detected on %q between devices %q and %q", targetFieldName, prevDev, item.DeviceCode)
			case "PREFIX":
				prefKey := fmt.Sprintf("%s_%s", strings.ToLower(item.DeviceCode), targetFieldName)
				if _, seen := acc.measurements[prefKey]; !seen {
					acc.keyOrder = append(acc.keyOrder, prefKey)
				}
				acc.measurements[prefKey] = val
				acc.keySourceDev[prefKey] = item.DeviceCode
			case "ARRAY":
				switch existing := acc.measurements[targetFieldName].(type) {
				case []interface{}:
					acc.measurements[targetFieldName] = append(existing, val)
				default:
					acc.measurements[targetFieldName] = []interface{}{existing, val}
				}
			case "AVERAGE":
				if item.Value != nil {
					acc.keyValuesList[targetFieldName] = append(acc.keyValuesList[targetFieldName], *item.Value)
					sum := 0.0
					for _, v := range acc.keyValuesList[targetFieldName] {
						sum += v
					}
					avg := sum / float64(len(acc.keyValuesList[targetFieldName]))
					avgCopy := avg
					avgItem := item
					avgItem.Value = &avgCopy
					acc.measurements[targetFieldName] = extractVal(avgItem)
				}
			default:
				// PRIORITY_PRIMARY / CUSTOM_MAPPING keeps first
				continue
			}
		} else {
			if _, seen := acc.measurements[targetFieldName]; !seen {
				acc.keyOrder = append(acc.keyOrder, targetFieldName)
			}
			acc.measurements[targetFieldName] = val
			acc.keySourceDev[targetFieldName] = item.DeviceCode
			if item.Value != nil {
				acc.keyValuesList[targetFieldName] = []float64{*item.Value}
			}
		}
	}

	sort.Slice(bucketKeys, func(i, j int) bool { return bucketKeys[i] < bucketKeys[j] })

	tsKeyForFlatRecords := strings.TrimSpace(cfg.TimestampField)
	if tsKeyForFlatRecords == "" || strings.Contains(tsKeyForFlatRecords, ",") {
		tsKeyForFlatRecords = "timestamp"
	}

	recordsValuesArr := make([]interface{}, 0, len(bucketKeys))
	recordsFlatArr := make([]interface{}, 0, len(bucketKeys))
	var latestValuesFields []orderedJSONField

	for _, bKey := range bucketKeys {
		acc := bucketMap[bKey]
		rowTs, err := formatTimestamp(acc.rowTime, cfg.TimestampFormat, cfg.TimestampTimezone)
		if err != nil {
			return nil, err
		}
		measFields := buildOrderedMeasurementFieldsWithOrder(acc.measurements, acc.keyOrder, mappingPreferredOrder)
		latestValuesFields = measFields

		recObj := orderedJSONObject{
			{Key: "timestamp", Value: rowTs},
			{Key: "values", Value: orderedJSONObject(measFields)},
		}
		recordsValuesArr = append(recordsValuesArr, recObj)

		flatRec := make(orderedJSONObject, 0, 1+len(measFields))
		flatRec = append(flatRec, orderedJSONField{Key: tsKeyForFlatRecords, Value: rowTs})
		flatRec = append(flatRec, measFields...)
		recordsFlatArr = append(recordsFlatArr, flatRec)
	}

	if latestValuesFields == nil {
		latestValuesFields = make([]orderedJSONField, 0)
	}

	return &customSchemaEvalContext{
		batch:            batch,
		cfg:              cfg,
		stationID:        stationID,
		deviceCode:       devCode,
		deviceName:       devName,
		devicesSummary:   batch.DevicesSummary,
		deviceCount:      batch.DeviceCount,
		intervalSeconds:  batch.IntervalSeconds,
		intervalMinutes:  intervalMin,
		tsVal:            tsVal,
		localBucketTime:  localBucket,
		localPeriodStart: localStart,
		localPeriodEnd:   localEnd,
		valuesObj:        orderedJSONObject(latestValuesFields),
		recordsValuesArr: recordsValuesArr,
		recordsFlatArr:   recordsFlatArr,
		paramByKey:       paramByKey,
		extractVal:       extractVal,
	}, nil
}

func (c *customSchemaEvalContext) resolveTokenValue(rawToken string) (interface{}, bool) {
	trimmed := strings.TrimSpace(rawToken)
	if trimmed == "" {
		return "", false
	}

	forceString := false
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "str:") {
		forceString = true
		trimmed = strings.TrimSpace(trimmed[4:])
		lower = strings.ToLower(trimmed)
	}

	var resolved interface{}
	matched := true

	switch lower {
	case "timestamp", "waktu":
		resolved = c.tsVal
	case "iso8601":
		resolved = c.localBucketTime.Format(time.RFC3339)
	case "datetime":
		resolved = c.localBucketTime.Format("2006-01-02 15:04:05")
	case "date", "tanggal", "date_ymd":
		resolved = c.localBucketTime.Format("2006-01-02")
	case "time", "jam", "time_hms":
		resolved = c.localBucketTime.Format("15:04:05")
	case "unix_sec", "epoch":
		resolved = c.localBucketTime.Unix()
	case "unix_ms", "epoch_ms":
		resolved = c.localBucketTime.UnixNano() / int64(time.Millisecond)
	case "period_start":
		resolved = c.localPeriodStart.Format(time.RFC3339)
	case "period_end":
		resolved = c.localPeriodEnd.Format(time.RFC3339)
	case "station_id", "uid", "site_code":
		resolved = c.stationID
	case "device_code":
		resolved = c.deviceCode
	case "device_name":
		resolved = c.deviceName
	case "devices_summary", "devices":
		resolved = c.devicesSummary
	case "device_count":
		resolved = c.deviceCount
	case "interval_minutes":
		resolved = c.intervalMinutes
	case "interval_seconds":
		resolved = c.intervalSeconds
	case "values", "parameters", "all_parameters":
		resolved = c.valuesObj
	case "records", "records_values":
		resolved = c.recordsValuesArr
	case "records_flat":
		resolved = c.recordsFlatArr
	default:
		if strings.HasPrefix(lower, "param:") {
			expr := strings.TrimSpace(trimmed[6:])
			resolved = c.lookupSingleParamValue(expr)
		} else if strings.HasPrefix(lower, "unit:") {
			key := strings.ToUpper(parseParamRefKey(trimmed[5:]))
			if item, ok := c.paramByKey[key]; ok {
				resolved = item.Unit
			} else {
				resolved = ""
			}
		} else if strings.HasPrefix(lower, "quality:") {
			key := strings.ToUpper(parseParamRefKey(trimmed[8:]))
			if item, ok := c.paramByKey[key]; ok {
				resolved = item.Quality
			} else {
				resolved = "BAD"
			}
		} else {
			matched = false
		}
	}

	if !matched {
		return nil, false
	}
	if forceString {
		return stringifyTemplateValue(resolved), true
	}
	return resolved, true
}

func (c *customSchemaEvalContext) lookupSingleParamValue(expr string) interface{} {
	trimmed := strings.TrimSpace(expr)
	var fallbackLiteral string
	hasFallback := false
	if pipeIdx := strings.Index(trimmed, "|"); pipeIdx >= 0 {
		fallbackLiteral = strings.TrimSpace(trimmed[pipeIdx+1:])
		trimmed = strings.TrimSpace(trimmed[:pipeIdx])
		hasFallback = true
	}

	lookupKey := strings.ToUpper(trimmed)
	if item, ok := c.paramByKey[lookupKey]; ok && item.Value != nil {
		return c.extractVal(item)
	}
	// Secondary lookup by parameter code suffix if "DEV:code" was supplied
	if colonIdx := strings.LastIndex(lookupKey, ":"); colonIdx >= 0 {
		shortKey := strings.TrimSpace(lookupKey[colonIdx+1:])
		if item, ok := c.paramByKey[shortKey]; ok && item.Value != nil {
			return c.extractVal(item)
		}
	}

	if hasFallback {
		if strings.EqualFold(fallbackLiteral, "null") || fallbackLiteral == "" {
			return nil
		}
		if num, err := strconv.ParseFloat(fallbackLiteral, 64); err == nil {
			return formatScalarOutboundNumber(num, c.cfg.NumericPrecisionMode)
		}
		return fallbackLiteral
	}
	if c.cfg.FillMissingPresetFields {
		return formatScalarOutboundNumber(0, c.cfg.NumericPrecisionMode)
	}
	return nil
}

func stringifyTemplateValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case fixed2Float:
		b, _ := val.MarshalJSON()
		return string(b)
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		s := string(b)
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			var unq string
			if json.Unmarshal(b, &unq) == nil {
				return unq
			}
		}
		return s
	}
}

func (c *customSchemaEvalContext) interpolateString(str string) interface{} {
	trimmed := strings.TrimSpace(str)
	// Exact single placeholder "{{token}}" -> preserve native JSON type!
	if strings.HasPrefix(trimmed, "{{") && strings.HasSuffix(trimmed, "}}") {
		inner := trimmed[2 : len(trimmed)-2]
		if !strings.Contains(inner, "{{") && !strings.Contains(inner, "}}") {
			if val, ok := c.resolveTokenValue(inner); ok {
				return val
			}
		}
	}

	// Mixed string interpolation (e.g. "Station {{station_id}} at {{date}}")
	if !strings.Contains(str, "{{") {
		return str
	}
	var out strings.Builder
	idx := 0
	for idx < len(str) {
		start := strings.Index(str[idx:], "{{")
		if start < 0 {
			out.WriteString(str[idx:])
			break
		}
		start += idx
		out.WriteString(str[idx:start])
		end := strings.Index(str[start+2:], "}}")
		if end < 0 {
			out.WriteString(str[start:])
			break
		}
		inner := str[start+2 : start+2+end]
		if val, ok := c.resolveTokenValue(inner); ok {
			out.WriteString(stringifyTemplateValue(val))
		} else {
			out.WriteString(str[start : start+2+end+2])
		}
		idx = start + 2 + end + 2
	}
	return out.String()
}

func (c *customSchemaEvalContext) evalOrderedAST(node interface{}) interface{} {
	switch typed := node.(type) {
	case orderedJSONObject:
		result := make(orderedJSONObject, 0, len(typed))
		for _, kv := range typed {
			evalKey := kv.Key
			if strings.Contains(evalKey, "{{") {
				evalKey = stringifyTemplateValue(c.interpolateString(evalKey))
			}
			evalVal := c.evalOrderedAST(kv.Value)

			if evalKey == "..." || evalKey == "@spread" || evalKey == "__spread__" || evalKey == "@root" {
				if spreadObj, ok := evalVal.(orderedJSONObject); ok {
					result = append(result, spreadObj...)
					continue
				}
			}
			result = append(result, orderedJSONField{Key: evalKey, Value: evalVal})
		}
		return result

	case []interface{}:
		arr := make([]interface{}, 0, len(typed))
		for _, elem := range typed {
			arr = append(arr, c.evalOrderedAST(elem))
		}
		return arr

	case string:
		return c.interpolateString(typed)

	default:
		return typed
	}
}

// orderedTreeNode builds nested ordered JSON objects from dot-path keys in VISUAL_FIELDS mode.
type orderedTreeNode struct {
	keys     []string
	values   map[string]interface{} // either *orderedTreeNode or leaf value
	isObject bool
}

func newOrderedTreeObject() *orderedTreeNode {
	return &orderedTreeNode{
		values:   make(map[string]interface{}),
		isObject: true,
	}
}

func (n *orderedTreeNode) setKey(key string, val interface{}) {
	if objVal, isOrdObj := val.(orderedJSONObject); isOrdObj {
		if existing, exists := n.values[key]; exists {
			if subTree, ok := existing.(*orderedTreeNode); ok && subTree.isObject {
				for _, kv := range objVal {
					subTree.setKey(kv.Key, kv.Value)
				}
				return
			}
		}
		sub := newOrderedTreeObject()
		for _, kv := range objVal {
			sub.setKey(kv.Key, kv.Value)
		}
		if _, exists := n.values[key]; !exists {
			n.keys = append(n.keys, key)
		}
		n.values[key] = sub
		return
	}

	if _, exists := n.values[key]; !exists {
		n.keys = append(n.keys, key)
	}
	n.values[key] = val
}

func (n *orderedTreeNode) setPath(segments []string, val interface{}) {
	if len(segments) == 0 {
		if spreadObj, ok := val.(orderedJSONObject); ok {
			for _, kv := range spreadObj {
				n.setKey(kv.Key, kv.Value)
			}
		}
		return
	}
	head := segments[0]
	if head == "@spread" || head == "..." || head == "@root" || head == "*" {
		if len(segments) == 1 {
			if spreadObj, ok := val.(orderedJSONObject); ok {
				for _, kv := range spreadObj {
					n.setKey(kv.Key, kv.Value)
				}
				return
			}
		}
	}
	if len(segments) == 1 {
		n.setKey(head, val)
		return
	}
	var child *orderedTreeNode
	if existing, exists := n.values[head]; exists {
		if existingNode, ok := existing.(*orderedTreeNode); ok && existingNode.isObject {
			child = existingNode
		}
	}
	if child == nil {
		child = newOrderedTreeObject()
		if _, exists := n.values[head]; !exists {
			n.keys = append(n.keys, head)
		}
		n.values[head] = child
	}
	child.setPath(segments[1:], val)
}

func (n *orderedTreeNode) toOrderedJSONObject() orderedJSONObject {
	out := make(orderedJSONObject, 0, len(n.keys))
	for _, k := range n.keys {
		v := n.values[k]
		if child, ok := v.(*orderedTreeNode); ok && child.isObject {
			out = append(out, orderedJSONField{Key: k, Value: child.toOrderedJSONObject()})
		} else {
			out = append(out, orderedJSONField{Key: k, Value: v})
		}
	}
	return out
}

func splitCustomFieldDotPath(rawKey string) []string {
	trimmed := strings.TrimSpace(rawKey)
	if trimmed == "" || trimmed == "@root" || trimmed == "*" || trimmed == "..." || trimmed == "@spread" {
		return nil
	}
	// Support escaped literal dot "\." when user wants a single key containing a dot (e.g. "PM2\.5")
	const dotPlaceholder = "\x00DOT\x00"
	safe := strings.ReplaceAll(trimmed, `\.`, dotPlaceholder)
	parts := strings.Split(safe, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		seg := strings.TrimSpace(strings.ReplaceAll(p, dotPlaceholder, "."))
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

func formatCustomSchemaPayload(
	mode string,
	batch *OutboundBatchData,
	cfg *model.DestinationPayloadConfig,
	matcher *parameterMappingMatcher,
	stationID string,
	tsVal interface{},
	rootBucketTime time.Time,
	extractVal func(model.SourceAggregationValueItem) interface{},
) ([]byte, error) {
	evalCtx, err := buildCustomSchemaEvalContext(batch, cfg, matcher, stationID, tsVal, rootBucketTime, extractVal)
	if err != nil {
		return nil, err
	}

	switch mode {
	case "JSON_TEMPLATE":
		normalized := normalizeUnquotedTemplatePlaceholders(strings.TrimSpace(cfg.CustomTemplateJSON))
		ast, decErr := decodeOrderedJSONDocument(normalized)
		if decErr != nil {
			return nil, fmt.Errorf("invalid JSON template: %w", decErr)
		}
		evaluated := evalCtx.evalOrderedAST(ast)
		return json.Marshal(evaluated)

	case "VISUAL_FIELDS":
		var fields []model.CustomPayloadFieldItem
		if err := json.Unmarshal([]byte(cfg.CustomFieldsJSON), &fields); err != nil {
			return nil, fmt.Errorf("invalid custom_fields_json: %w", err)
		}
		rootTree := newOrderedTreeObject()
		for _, f := range fields {
			if !f.Enabled {
				continue
			}
			src := strings.ToUpper(strings.TrimSpace(f.Source))
			var val interface{}
			switch src {
			case "STATIC_STRING", "":
				val = evalCtx.interpolateString(f.Value)
			case "STATIC_NUMBER":
				rawNum := strings.TrimSpace(f.Value)
				if iVal, iErr := strconv.ParseInt(rawNum, 10, 64); iErr == nil && !strings.Contains(rawNum, ".") {
					val = iVal
				} else if fVal, fErr := strconv.ParseFloat(rawNum, 64); fErr == nil {
					val = fVal
				} else {
					return nil, fmt.Errorf("invalid STATIC_NUMBER %q for field %q", f.Value, f.Key)
				}
			case "STATIC_BOOL":
				vLower := strings.ToLower(strings.TrimSpace(f.Value))
				val = vLower == "true" || vLower == "1" || vLower == "yes"
			case "STATIC_NULL":
				val = nil
			case "STATIC_JSON":
				parsedStatic, sErr := decodeOrderedJSONDocument(strings.TrimSpace(f.Value))
				if sErr != nil {
					return nil, fmt.Errorf("invalid STATIC_JSON for field %q: %w", f.Key, sErr)
				}
				val = evalCtx.evalOrderedAST(parsedStatic)
			case "TIMESTAMP":
				val = evalCtx.tsVal
			case "ISO8601":
				val = evalCtx.localBucketTime.Format(time.RFC3339)
			case "DATETIME":
				val = evalCtx.localBucketTime.Format("2006-01-02 15:04:05")
			case "DATE_YMD":
				val = evalCtx.localBucketTime.Format("2006-01-02")
			case "TIME_HMS":
				val = evalCtx.localBucketTime.Format("15:04:05")
			case "UNIX_SEC":
				val = evalCtx.localBucketTime.Unix()
			case "UNIX_MS":
				val = evalCtx.localBucketTime.UnixNano() / int64(time.Millisecond)
			case "PERIOD_START":
				val = evalCtx.localPeriodStart.Format(time.RFC3339)
			case "PERIOD_END":
				val = evalCtx.localPeriodEnd.Format(time.RFC3339)
			case "STATION_ID":
				val = evalCtx.stationID
			case "DEVICE_CODE":
				val = evalCtx.deviceCode
			case "DEVICE_NAME":
				val = evalCtx.deviceName
			case "INTERVAL_MINUTES":
				val = evalCtx.intervalMinutes
			case "INTERVAL_SECONDS":
				val = evalCtx.intervalSeconds
			case "ALL_PARAMETERS":
				val = evalCtx.valuesObj
			case "RECORDS_VALUES_ARRAY":
				val = evalCtx.recordsValuesArr
			case "RECORDS_FLAT_ARRAY":
				val = evalCtx.recordsFlatArr
			case "PARAMETER":
				val = evalCtx.lookupSingleParamValue(f.Value)
			default:
				val = f.Value
			}

			segments := splitCustomFieldDotPath(f.Key)
			rootTree.setPath(segments, val)
		}
		return json.Marshal(rootTree.toOrderedJSONObject())
	}

	return nil, fmt.Errorf("unsupported custom_schema_mode %q", mode)
}


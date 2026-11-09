package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"
	"datalogger/pkg/formula"
	"datalogger/pkg/tenantctx"
)

// IngestHTTPError carries an explicit HTTP status code and optional structured details
// for standardized multi-tenant telemetry ingestion error responses.
type IngestHTTPError struct {
	StatusCode int         `json:"-"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
}

func (e *IngestHTTPError) Error() string {
	return e.Message
}

type IngestService struct {
	tenantService    *TenantService
	deviceRepo       *repository.DeviceRepository
	categoryRepo     *repository.CategoryRepository
	telemetryService *TelemetryService
	hub              *websocket.Hub
	ingestMu         sync.Mutex
}

func NewIngestService(
	tenantService *TenantService,
	deviceRepo *repository.DeviceRepository,
	categoryRepo *repository.CategoryRepository,
	telemetryService *TelemetryService,
) *IngestService {
	return &IngestService{
		tenantService:    tenantService,
		deviceRepo:       deviceRepo,
		categoryRepo:     categoryRepo,
		telemetryService: telemetryService,
	}
}

// SetHub attaches the WebSocket hub so passive HTTP Push & MQTT receivers broadcast
// live device.connection.changed events to OperatorCommandCenterView identically to physical polling.
func (s *IngestService) SetHub(h *websocket.Hub) {
	s.hub = h
}

type EdgeIngestResponse struct {
	Accepted               bool     `json:"accepted"`
	TenantID               uint     `json:"tenant_id"`
	SiteID                 uint     `json:"site_id"`
	SiteCode               string   `json:"site_code"`
	DeviceCode             string   `json:"device_code,omitempty"`
	LoggerType             string   `json:"logger_type,omitempty"`
	Source                 string   `json:"source,omitempty"`
	DevicesProcessed       int      `json:"devices_processed"`
	RecordsProcessed       int      `json:"records_processed"`
	ParametersIngested     int      `json:"parameters_ingested"`
	IngestedParameters     []string `json:"ingested_parameters,omitempty"`
	ParametersSkipped      int      `json:"parameters_skipped"`
	UnregisteredParameters []string `json:"unregistered_parameters,omitempty"`
	ReceivedAt             string   `json:"received_at"`
}

type edgeReading struct {
	Code  string
	Name  string
	Unit  string
	Value float64
}

type edgeDeviceBatch struct {
	LoggerType string
	SiteCode   string
	DeviceCode string
	DeviceName string
	RawTS      interface{}
	Timestamp  *time.Time
	Readings   []edgeReading
}

var reservedEnvelopeKeys = map[string]bool{
	"site_id": true, "site_code": true, "station_code": true, "id_stasiun": true, "idstasiun": true,
	"tenant_id": true, "tenant_code": true,
	"device_id": true, "device_code": true, "device_name": true, "module_id": true,
	"logger_type": true, "domain": true, "monitoring_domain": true,
	"timestamp": true, "datetime": true, "waktu": true, "tanggal": true, "jam": true,
	"period_start": true, "period_end": true, "interval_seconds": true,
	"apikey": true, "apisecret": true, "token": true, "uid": true,
	"topic": true, "qos": true, "retain": true, "client_id": true,
	"devices": true, "records": true, "readings": true, "parameters": true, "data": true, "values": true,
}

func isValidLoggerType(lt string) bool {
	switch strings.ToUpper(strings.TrimSpace(lt)) {
	case "", "AQMS", "SPARING", "ONLIMO", "WQMS", "CUSTOM":
		return true
	default:
		return false
	}
}

func parseIngestTimestampInLocation(raw interface{}, loc *time.Location) *time.Time {
	if raw == nil {
		return nil
	}
	if loc == nil {
		loc = time.UTC
	}
	switch v := raw.(type) {
	case float64:
		if v > 1e12 {
			t := time.UnixMilli(int64(v)).UTC()
			return &t
		}
		if v > 1e9 {
			t := time.Unix(int64(v), 0).UTC()
			return &t
		}
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil
		}
		if unixSec, err := strconv.ParseInt(s, 10, 64); err == nil {
			if unixSec > 1e12 {
				t := time.UnixMilli(unixSec).UTC()
				return &t
			}
			if unixSec > 1e9 {
				t := time.Unix(unixSec, 0).UTC()
				return &t
			}
		}
		// Explicit timezone offset formats
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05-07:00", "2006-01-02 15:04:05Z07:00"} {
			if t, err := time.Parse(layout, s); err == nil {
				utc := t.UTC()
				return &utc
			}
		}
		// Naive local timestamp formats interpreted in station's configured timezone
		for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02 15:04"} {
			if t, err := time.ParseInLocation(layout, s, loc); err == nil {
				utc := t.UTC()
				return &utc
			}
		}
	}
	return nil
}

func loadStationLocation(tz string) *time.Location {
	clean := strings.TrimSpace(tz)
	if clean != "" {
		if loc, err := time.LoadLocation(clean); err == nil {
			return loc
		}
	}
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*3600)
}

func parseNumericValue(v interface{}) (float64, string, string, bool) {
	switch val := v.(type) {
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0, "", "", false
		}
		return val, "", "", true
	case bool:
		if val {
			return 1.0, "", "", true
		}
		return 0.0, "", "", true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, "", "", false
		}
		return f, "", "", true
	case map[string]interface{}:
		rawVal, exists := val["value"]
		if !exists {
			rawVal, exists = val["processed_value"]
		}
		if !exists {
			return 0, "", "", false
		}
		f, _, _, ok := parseNumericValue(rawVal)
		if !ok {
			return 0, "", "", false
		}
		name, _ := val["name"].(string)
		if name == "" {
			name, _ = val["parameter_name"].(string)
		}
		unit, _ := val["unit"].(string)
		return f, strings.TrimSpace(name), strings.TrimSpace(unit), true
	}
	return 0, "", "", false
}

func extractReadingsFromAny(container interface{}) []edgeReading {
	var out []edgeReading
	switch c := container.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(c))
		for k := range c {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := c[k]
			code := strings.TrimSpace(k)
			if code == "" || reservedEnvelopeKeys[strings.ToLower(code)] {
				continue
			}
			if f, name, unit, ok := parseNumericValue(v); ok {
				out = append(out, edgeReading{
					Code:  code,
					Name:  name,
					Unit:  unit,
					Value: f,
				})
			}
		}
	case []interface{}:
		for _, elem := range c {
			m, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}
			code, _ := m["code"].(string)
			if code == "" {
				code, _ = m["parameter_code"].(string)
			}
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			f, name, unit, ok := parseNumericValue(m)
			if !ok {
				continue
			}
			out = append(out, edgeReading{
				Code:  code,
				Name:  name,
				Unit:  unit,
				Value: f,
			})
		}
	}
	return out
}

func stringField(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if raw, ok := m[k]; ok && raw != nil {
			if s, isStr := raw.(string); isStr && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func firstNonNilField(m map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if raw, ok := m[k]; ok && raw != nil {
			if s, isStr := raw.(string); isStr && strings.TrimSpace(s) == "" {
				continue
			}
			return raw
		}
	}
	return nil
}

// unwrapMQTTWebhookEnvelope transparently extracts {"topic":"datalogger/ingest/<site>/<device>","payload":{...}}
// when an MQTT broker (EMQX / Mosquitto Bridge / Node-RED) forwards messages via HTTP webhook.
func unwrapMQTTWebhookEnvelope(root map[string]interface{}) map[string]interface{} {
	rawPayload, hasPayload := root["payload"]
	topic := stringField(root, "topic")
	if !hasPayload || topic == "" {
		return root
	}
	var inner map[string]interface{}
	switch p := rawPayload.(type) {
	case map[string]interface{}:
		inner = p
	case string:
		if err := json.Unmarshal([]byte(strings.TrimSpace(p)), &inner); err != nil {
			return root
		}
	default:
		return root
	}
	if inner == nil {
		return root
	}
	// Extract optional site_code / device_code hints from topic: datalogger/ingest/<site_code>[/<device_code>]
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if strings.EqualFold(parts[i], "ingest") {
			if stringField(inner, "site_code", "station_code") == "" && i+1 < len(parts) {
				inner["site_code"] = parts[i+1]
			}
			if stringField(inner, "device_code", "code") == "" && i+2 < len(parts) {
				inner["device_code"] = parts[i+2]
			}
			break
		}
	}
	return inner
}

func parseEdgeBatches(rawJSON []byte) ([]edgeDeviceBatch, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(rawJSON, &root); err != nil {
		return nil, &IngestHTTPError{
			StatusCode: 400,
			Message:    "Invalid JSON payload: " + err.Error(),
		}
	}
	root = unwrapMQTTWebhookEnvelope(root)

	rootLoggerType := strings.ToUpper(stringField(root, "logger_type", "domain", "monitoring_domain"))
	if !isValidLoggerType(rootLoggerType) {
		return nil, &IngestHTTPError{
			StatusCode: 400,
			Message:    fmt.Sprintf("Invalid logger_type '%s' (supported: AQMS, SPARING, ONLIMO)", rootLoggerType),
		}
	}

	rootSiteCode := stringField(root, "site_code", "station_code", "id_stasiun", "IDStasiun")
	rootDevCode := stringField(root, "device_code", "code", "module_id")
	rootDevName := stringField(root, "device_name", "name")
	rootRawTS := firstNonNilField(root, "timestamp", "datetime", "waktu", "period_end")

	// 1. Mode B — Standardized Batch / Offline Backfill format: {"records": [ {"timestamp": "...", "values": {...}}, ... ]}
	if rawRecs, ok := root["records"].([]interface{}); ok {
		if len(rawRecs) == 0 {
			return nil, &IngestHTTPError{
				StatusCode: 400,
				Message:    "Batch 'records' array cannot be empty",
			}
		}
		batches := make([]edgeDeviceBatch, 0, len(rawRecs))
		for idx, item := range rawRecs {
			rm, ok := item.(map[string]interface{})
			if !ok {
				return nil, &IngestHTTPError{
					StatusCode: 400,
					Message:    fmt.Sprintf("Invalid JSON object at records[%d]", idx),
				}
			}
			recRawTS := firstNonNilField(rm, "timestamp", "datetime", "waktu", "period_end")
			if recRawTS == nil {
				recRawTS = rootRawTS
			}
			var readings []edgeReading
			for _, key := range []string{"values", "readings", "parameters", "data"} {
				if sub, exists := rm[key]; exists {
					readings = append(readings, extractReadingsFromAny(sub)...)
				}
			}
			if len(readings) == 0 {
				readings = extractReadingsFromAny(rm)
			}
			if len(readings) == 0 {
				return nil, &IngestHTTPError{
					StatusCode: 400,
					Message:    fmt.Sprintf("records[%d] contains no numeric telemetry values", idx),
				}
			}
			sCode := stringField(rm, "site_code", "station_code")
			if sCode == "" {
				sCode = rootSiteCode
			}
			dCode := stringField(rm, "device_code", "code", "module_id")
			if dCode == "" {
				dCode = rootDevCode
			}
			lt := strings.ToUpper(stringField(rm, "logger_type", "domain"))
			if lt == "" {
				lt = rootLoggerType
			}
			batches = append(batches, edgeDeviceBatch{
				LoggerType: lt,
				SiteCode:   sCode,
				DeviceCode: dCode,
				DeviceName: rootDevName,
				RawTS:      recRawTS,
				Readings:   readings,
			})
		}
		return batches, nil
	}

	// 2. Multi-device Outbound Sync format: {"devices": [ ... ]}
	if rawDevs, ok := root["devices"].([]interface{}); ok && len(rawDevs) > 0 {
		var batches []edgeDeviceBatch
		for _, item := range rawDevs {
			dm, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			dCode := stringField(dm, "device_code", "code", "module_id")
			if dCode == "" {
				dCode = rootDevCode
			}
			dName := stringField(dm, "device_name", "name")
			rawTS := firstNonNilField(dm, "timestamp", "datetime", "waktu")
			if rawTS == nil {
				rawTS = rootRawTS
			}
			var readings []edgeReading
			for _, key := range []string{"values", "readings", "parameters", "data"} {
				if sub, exists := dm[key]; exists {
					readings = append(readings, extractReadingsFromAny(sub)...)
				}
			}
			if len(readings) == 0 {
				readings = extractReadingsFromAny(dm)
			}
			if len(readings) > 0 {
				sCode := stringField(dm, "site_code", "station_code")
				if sCode == "" {
					sCode = rootSiteCode
				}
				lt := strings.ToUpper(stringField(dm, "logger_type", "domain"))
				if lt == "" {
					lt = rootLoggerType
				}
				batches = append(batches, edgeDeviceBatch{
					LoggerType: lt,
					SiteCode:   sCode,
					DeviceCode: dCode,
					DeviceName: dName,
					RawTS:      rawTS,
					Readings:   readings,
				})
			}
		}
		if len(batches) > 0 {
			return batches, nil
		}
	}

	// 3. Mode A — Standardized Single Reading (`values` / `readings` / `parameters` / `data` / flat JSON)
	var readings []edgeReading
	for _, key := range []string{"values", "readings", "parameters", "data"} {
		if sub, exists := root[key]; exists {
			if subMap, isMap := sub.(map[string]interface{}); isMap {
				if rootSiteCode == "" {
					rootSiteCode = stringField(subMap, "site_code", "IDStasiun", "id_stasiun")
				}
				if rootDevCode == "" {
					rootDevCode = stringField(subMap, "device_code", "IDStasiun", "uid")
				}
			}
			readings = append(readings, extractReadingsFromAny(sub)...)
		}
	}
	if len(readings) == 0 {
		readings = extractReadingsFromAny(root)
	}
	if len(readings) == 0 {
		return nil, &IngestHTTPError{
			StatusCode: 400,
			Message:    "Payload contains no numeric telemetry values (provide 'values' object with numeric readings)",
		}
	}

	return []edgeDeviceBatch{
		{
			LoggerType: rootLoggerType,
			SiteCode:   rootSiteCode,
			DeviceCode: rootDevCode,
			DeviceName: rootDevName,
			RawTS:      rootRawTS,
			Readings:   readings,
		},
	}, nil
}

// normalizeParamMatchKey strips non-alphanumeric separators (e.g. "PM2.5" -> "pm25", "NH3-N" -> "nh3n")
// for unambiguous fallback matching against manually created parameter codes.
func normalizeParamMatchKey(code string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(code)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func roundIngestPrecision(val float64, prec int) float64 {
	if prec < 0 || prec > 10 {
		prec = 4
	}
	pow := math.Pow10(prec)
	return math.Round(val*pow) / pow
}

// IngestPayload authenticates a Bearer Token / API Key and ingests AQMS, SPARING, or ONLIMO JSON telemetry
// strictly into manually registered Sites, Devices, and Parameters (no auto-provisioning).
func (s *IngestService) IngestPayload(rawKey string, rawJSON []byte, clientIP string) (*EdgeIngestResponse, error) {
	return s.IngestPayloadWithSource(rawKey, rawJSON, clientIP, string(model.ProtocolEdgePush))
}

// IngestPayloadWithSource processes standardized telemetry via HTTP Push ("EDGE_PUSH") or MQTT ("MQTT").
func (s *IngestService) IngestPayloadWithSource(rawKey string, rawJSON []byte, clientIP string, source string) (*EdgeIngestResponse, error) {
	apiKey, err := s.tenantService.ValidateIngestKey(rawKey)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "suspended") {
			return nil, &IngestHTTPError{StatusCode: 403, Message: err.Error()}
		}
		return nil, &IngestHTTPError{StatusCode: 401, Message: err.Error()}
	}
	tenantID := apiKey.TenantID

	batches, err := parseEdgeBatches(rawJSON)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(source) == "" {
		source = string(model.ProtocolEdgePush)
	}
	source = strings.ToUpper(strings.TrimSpace(source))

	now := time.Now().UTC()

	s.ingestMu.Lock()
	defer s.ingestMu.Unlock()

	type resolvedBatch struct {
		batch *edgeDeviceBatch
		site  *model.Site
		dev   *model.Device
		ts    time.Time
	}

	resolved := make([]resolvedBatch, 0, len(batches))
	uniqueDevices := make(map[uint]*model.Device)

	for i := range batches {
		batch := &batches[i]

		// 1. Strict Manual Site Resolution (NO Auto-Create)
		var site *model.Site
		if apiKey.SiteID != nil && *apiKey.SiteID > 0 {
			site, _ = s.tenantService.GetRepo().GetSiteByID(*apiKey.SiteID)
			if site == nil || site.TenantID != tenantID {
				return nil, &IngestHTTPError{
					StatusCode: 403,
					Message:    "Token station binding is invalid or inaccessible",
				}
			}
			if batch.SiteCode != "" {
				reqSlug := model.NormalizeSlugCode(batch.SiteCode)
				if !strings.EqualFold(batch.SiteCode, site.Code) && reqSlug != site.Code {
					return nil, &IngestHTTPError{
						StatusCode: 403,
						Message:    fmt.Sprintf("Bearer token is locked to station '%s' and cannot push to '%s'", site.Code, batch.SiteCode),
					}
				}
			}
		} else if batch.SiteCode != "" {
			site, _ = s.tenantService.GetRepo().GetSiteByCode(tenantID, model.NormalizeSlugCode(batch.SiteCode))
			if site == nil {
				site, _ = s.tenantService.GetRepo().GetSiteByCode(tenantID, strings.TrimSpace(batch.SiteCode))
			}
			if site == nil {
				return nil, &IngestHTTPError{
					StatusCode: 404,
					Message:    fmt.Sprintf("Station '%s' is not registered in this tenant (please create station manually first)", batch.SiteCode),
				}
			}
		} else {
			site, _ = s.tenantService.GetRepo().GetDefaultSiteForTenant(tenantID)
			if site == nil {
				return nil, &IngestHTTPError{
					StatusCode: 404,
					Message:    "No registered station found in tenant (please create station manually first)",
				}
			}
		}

		if site.TenantID != tenantID || !site.Enabled {
			return nil, &IngestHTTPError{
				StatusCode: 403,
				Message:    fmt.Sprintf("Station '%s' is disabled or inaccessible", site.Code),
			}
		}

		// 2. Strict Manual Device Resolution (NO Auto-Create)
		var dev *model.Device
		devCode := strings.TrimSpace(batch.DeviceCode)
		if devCode != "" {
			dev, _ = s.deviceRepo.GetByCodeInTenant(tenantID, devCode)
			if dev == nil {
				return nil, &IngestHTTPError{
					StatusCode: 404,
					Message:    fmt.Sprintf("Device '%s' is not registered in tenant (please create device manually first)", devCode),
				}
			}
			if dev.SiteID != site.ID {
				return nil, &IngestHTTPError{
					StatusCode: 403,
					Message:    fmt.Sprintf("Device '%s' is registered under a different station, not '%s'", dev.DeviceCode, site.Code),
				}
			}
		} else {
			enabledFlag := true
			siteDevs, _, _ := s.deviceRepo.List(repository.DeviceFilterParams{
				TenantID: tenantID,
				SiteID:   site.ID,
				Enabled:  &enabledFlag,
				PageSize: 10,
			})
			if len(siteDevs) == 1 {
				dev = &siteDevs[0]
			} else if len(siteDevs) == 0 {
				return nil, &IngestHTTPError{
					StatusCode: 404,
					Message:    fmt.Sprintf("No active device registered in station '%s' (please create device manually first)", site.Code),
				}
			} else {
				return nil, &IngestHTTPError{
					StatusCode: 400,
					Message:    fmt.Sprintf("Station '%s' has %d active devices; please specify 'device_code' in JSON payload", site.Code, len(siteDevs)),
				}
			}
		}

		if !dev.Enabled || dev.Status == model.DeviceStatusDisabled || dev.Status == model.DeviceStatusInactive {
			return nil, &IngestHTTPError{
				StatusCode: 422,
				Message:    fmt.Sprintf("Device '%s' is currently disabled or inactive", dev.DeviceCode),
			}
		}

		// Ensure device parameters are loaded
		if cachedDev, exists := uniqueDevices[dev.ID]; exists {
			dev = cachedDev
		} else {
			if len(dev.Parameters) == 0 {
				if params, pErr := s.deviceRepo.ListParameters(dev.ID); pErr == nil {
					dev.Parameters = params
				}
			}
			uniqueDevices[dev.ID] = dev
		}

		// Parse timestamp in station's configured timezone
		loc := loadStationLocation(site.Timezone)
		parsedTS := parseIngestTimestampInLocation(batch.RawTS, loc)
		effectiveTS := now
		if parsedTS != nil {
			effectiveTS = *parsedTS
		}
		batch.Timestamp = &effectiveTS

		resolved = append(resolved, resolvedBatch{
			batch: batch,
			site:  site,
			dev:   dev,
			ts:    effectiveTS,
		})
	}

	// Sort chronologically (oldest -> newest) so historical backfill updates latestCache & Parameter.current_value with the newest record
	sort.SliceStable(resolved, func(i, j int) bool {
		return resolved[i].ts.Before(resolved[j].ts)
	})

	type paramMapBundle struct {
		exact map[string]*model.Parameter
		norm  map[string]*model.Parameter
	}
	paramMapsByDev := make(map[uint]*paramMapBundle, len(uniqueDevices))
	for devID, dev := range uniqueDevices {
		exact := make(map[string]*model.Parameter, len(dev.Parameters)*2)
		normCounts := make(map[string]int, len(dev.Parameters))
		normTemp := make(map[string]*model.Parameter, len(dev.Parameters))
		for i := range dev.Parameters {
			p := &dev.Parameters[i]
			k1 := strings.ToUpper(strings.TrimSpace(p.ParameterCode))
			if k1 != "" {
				exact[k1] = p
			}
			k2 := strings.ToUpper(strings.TrimSpace(p.Code))
			if k2 != "" {
				exact[k2] = p
			}
			nk := normalizeParamMatchKey(p.ParameterCode)
			if nk != "" {
				normCounts[nk]++
				normTemp[nk] = p
			}
		}
		norm := make(map[string]*model.Parameter, len(normTemp))
		for nk, p := range normTemp {
			if normCounts[nk] == 1 {
				norm[nk] = p
			}
		}
		paramMapsByDev[devID] = &paramMapBundle{exact: exact, norm: norm}
	}

	type latestParamReading struct {
		uncalVal   float64
		finalVal   float64
		formulaVal *float64
		ts         time.Time
	}
	latestByParam := make(map[uint]latestParamReading)
	touchedDevices := make(map[uint]time.Time)

	var (
		totalIngested    int
		totalSkipped     int
		ingestedSet      = make(map[string]bool)
		ingestedList     []string
		unregisteredSet  = make(map[string]bool)
		unregisteredList []string
		lastSiteID       uint
		lastSiteCode     string
		lastDevCode      string
		lastLoggerType   string
	)

	for _, item := range resolved {
		lastSiteID = item.site.ID
		lastSiteCode = item.site.Code
		lastDevCode = item.dev.DeviceCode
		if item.batch.LoggerType != "" {
			lastLoggerType = item.batch.LoggerType
		} else if lastLoggerType == "" {
			lastLoggerType = item.site.MonitoringDomain
		}

		bundle := paramMapsByDev[item.dev.ID]
		recTS := item.ts

		for _, r := range item.batch.Readings {
			rawCode := strings.TrimSpace(r.Code)
			upperKey := strings.ToUpper(rawCode)
			param := bundle.exact[upperKey]
			if param == nil {
				if nk := normalizeParamMatchKey(rawCode); nk != "" {
					param = bundle.norm[nk]
				}
			}

			// Strict Manual Parameter Enforcement: NEVER auto-create unregistered parameters
			if param == nil {
				totalSkipped++
				if rawCode != "" && !unregisteredSet[rawCode] {
					unregisteredSet[rawCode] = true
					unregisteredList = append(unregisteredList, rawCode)
				}
				continue
			}

			if !param.Enabled {
				totalSkipped++
				continue
			}

			// Apply parameter engineering Scale, Offset, Calibration Curve, and Custom Formula
			scale := param.Scale
			if scale == 0 {
				scale = 1.0
			}
			prec := param.Precision
			if prec <= 0 {
				prec = 2
			}
			uncalVal := roundIngestPrecision(r.Value*scale+param.Offset, prec)
			engVal := param.ApplyCalibration(uncalVal)

			effectiveFormula := strings.TrimSpace(param.Formula)
			if param.CalibrationEnabled || param.CalibrationSessionActive {
				effectiveFormula = ""
			}
			var formulaVal *float64
			if effectiveFormula != "" {
				if evalRes, fErr := formula.EvalWithXRaw(effectiveFormula, engVal, r.Value); fErr == nil && !math.IsNaN(evalRes) && !math.IsInf(evalRes, 0) {
					roundedF := roundIngestPrecision(evalRes, prec)
					formulaVal = &roundedF
				}
			}

			finalVal := engVal
			if formulaVal != nil {
				finalVal = *formulaVal
			}

			if s.telemetryService != nil {
				tsCopy := recTS
				_ = s.telemetryService.Ingest(&TelemetryIngestPayload{
					TenantID:          tenantID,
					SiteID:            item.site.ID,
					DeviceID:          item.dev.ID,
					DeviceCode:        item.dev.DeviceCode,
					DeviceName:        item.dev.DeviceName,
					ParameterID:       param.ID,
					ParameterCode:     param.ParameterCode,
					ParameterName:     param.ParameterName,
					Unit:              param.Unit,
					DataType:          param.DataType,
					RawValue:          r.Value,
					UncalibratedValue: &uncalVal,
					Value:             finalVal,
					ProcessedValue:    finalVal,
					FormulaValue:      formulaVal,
					Formula:           effectiveFormula,
					Quality:           model.QualityGood,
					QualityReason:     model.ReasonNone,
					Source:            source,
					DeviceTimestamp:   &tsCopy,
					ReceivedAt:        tsCopy,
					Parameter:         param,
				})
			}

			if prev, exists := latestByParam[param.ID]; !exists || !recTS.Before(prev.ts) {
				latestByParam[param.ID] = latestParamReading{
					uncalVal:   uncalVal,
					finalVal:   finalVal,
					formulaVal: formulaVal,
					ts:         recTS,
				}
			}
			if prevDevTS, exists := touchedDevices[item.dev.ID]; !exists || recTS.After(prevDevTS) {
				touchedDevices[item.dev.ID] = recTS
			}

			if !ingestedSet[param.ParameterCode] {
				ingestedSet[param.ParameterCode] = true
				ingestedList = append(ingestedList, param.ParameterCode)
			}
			totalIngested++
		}
	}

	resp := &EdgeIngestResponse{
		Accepted:               totalIngested > 0,
		TenantID:               tenantID,
		SiteID:                 lastSiteID,
		SiteCode:               lastSiteCode,
		DeviceCode:             lastDevCode,
		LoggerType:             lastLoggerType,
		Source:                 source,
		DevicesProcessed:       len(touchedDevices),
		RecordsProcessed:       len(resolved),
		ParametersIngested:     totalIngested,
		IngestedParameters:     ingestedList,
		ParametersSkipped:      totalSkipped,
		UnregisteredParameters: unregisteredList,
		ReceivedAt:             now.Format(time.RFC3339),
	}

	if totalIngested == 0 {
		msg := fmt.Sprintf("None of the submitted parameters are registered/enabled on device '%s' (create parameters manually first)", lastDevCode)
		if len(unregisteredList) > 0 {
			msg = fmt.Sprintf("No matching registered parameters on device '%s'. Unregistered parameter codes: %s", lastDevCode, strings.Join(unregisteredList, ", "))
		}
		return nil, &IngestHTTPError{
			StatusCode: 422,
			Message:    msg,
			Details:    resp,
		}
	}

	// Persist latest parameter values & device ONLINE state so Operator Dashboard reflects pushed values immediately
	for paramID, lr := range latestByParam {
		u := lr.uncalVal
		_ = s.deviceRepo.UpdateParameterCurrentReadings(paramID, &u, lr.finalVal, lr.formulaVal, false, lr.ts)
	}

	for devID, lastTS := range touchedDevices {
		dev := uniqueDevices[devID]
		seenAt := now
		if lastTS.After(seenAt) {
			seenAt = lastTS
		}
		_ = s.deviceRepo.UpdateLastSeen(devID, seenAt)
		_ = s.deviceRepo.UpdateLastData(devID, lastTS)
		_ = s.deviceRepo.SetConnectionStatus(devID, model.DeviceConnOnline)
		_ = s.deviceRepo.RecordCommunicationResult(devID, true, 1)

		if s.hub != nil && dev != nil {
			s.hub.Broadcast(websocket.WSMessage{
				Type:      "device.connection.changed",
				Timestamp: now,
				TenantID:  dev.TenantID,
				SiteID:    dev.SiteID,
				Data: map[string]interface{}{
					"device_id":         dev.ID,
					"id":                dev.ID,
					"device_code":       dev.DeviceCode,
					"connection_status": "ONLINE",
					"status":            "ONLINE",
					"source":            source,
				},
			})
		}
	}

	if s.tenantService.GetRepo() != nil {
		s.tenantService.GetRepo().RecordAPIKeyUsage(apiKey.ID, clientIP, int64(totalIngested))
	}

	return resp, nil
}

// RegisteredParamSummary describes a manually registered parameter for the interactive Payload Builder.
type RegisteredParamSummary struct {
	ParameterCode string  `json:"parameter_code"`
	ParameterName string  `json:"parameter_name"`
	Unit          string  `json:"unit"`
	DataType      string  `json:"data_type"`
	SampleValue   float64 `json:"sample_value"`
}

// PayloadTemplateResponse provides ready-to-use AQMS, SPARING, and ONLIMO JSON payloads + Bearer Token cURL & MQTT examples.
type PayloadTemplateResponse struct {
	LoggerType           string                   `json:"logger_type"`
	SiteCode             string                   `json:"site_code"`
	DeviceCode           string                   `json:"device_code"`
	EndpointURL          string                   `json:"endpoint_url"`
	MQTTEndpointURL      string                   `json:"mqtt_endpoint_url"`
	MQTTTopic            string                   `json:"mqtt_topic"`
	RegisteredParameters []RegisteredParamSummary `json:"registered_parameters"`
	UsingActualParams    bool                     `json:"using_actual_params"`
	SinglePayload        map[string]interface{}   `json:"single_payload"`
	BatchPayload         map[string]interface{}   `json:"batch_payload"`
	CurlBearerCommand    string                   `json:"curl_bearer_command"`
}

func sampleValueForCode(code, loggerType string) float64 {
	c := normalizeParamMatchKey(code)
	samples := map[string]float64{
		"pm25": 34.5, "pm10": 62.1, "so2": 12.4, "co": 1150.0, "o3": 41.2, "no2": 18.9, "hc": 8.4,
		"ws": 2.8, "wd": 185.0, "temp": 28.4, "temperature": 28.4, "watertemp": 27.6,
		"humidity": 71.2, "rh": 71.2, "pressure": 1009.5, "sr": 540.0, "rain": 0.0,
		"ph": 7.18, "cod": 45.2, "bod": 18.5, "tss": 31.4, "nh3n": 0.42, "nh4": 0.48,
		"do": 6.45, "tds": 312.0, "turbidity": 14.8, "no3": 3.2, "orp": 215.0, "conductivity": 480.0,
		"debit": 124.8, "flow": 124.8, "waterlevel": 1.42, "level": 1.42,
	}
	if v, ok := samples[c]; ok {
		return v
	}
	if strings.EqualFold(loggerType, "SPARING") || strings.EqualFold(loggerType, "ONLIMO") {
		return 12.5
	}
	return 25.0
}

func canonicalValuesForLoggerType(loggerType string) map[string]float64 {
	switch strings.ToUpper(strings.TrimSpace(loggerType)) {
	case "SPARING":
		return map[string]float64{
			"ph":    7.18,
			"cod":   45.2,
			"tss":   31.4,
			"nh3n":  0.42,
			"debit": 124.8,
		}
	case "ONLIMO":
		return map[string]float64{
			"ph":          7.05,
			"do":          6.45,
			"bod":         14.2,
			"cod":         32.6,
			"tss":         24.0,
			"nh3n":        0.31,
			"no3":         2.85,
			"tds":         312.0,
			"turbidity":   14.8,
			"temperature": 27.6,
			"water_level": 1.42,
		}
	default: // AQMS
		return map[string]float64{
			"pm25":        34.5,
			"pm10":        62.1,
			"so2":         12.4,
			"co":          1150.0,
			"o3":          41.2,
			"no2":         18.9,
			"hc":          8.4,
			"ws":          2.8,
			"wd":          185.0,
			"temperature": 28.4,
			"humidity":    71.2,
			"pressure":    1009.5,
			"sr":          540.0,
			"rain":        0.0,
		}
	}
}

func (s *IngestService) BuildPayloadTemplate(scope tenantctx.Scope, siteID, deviceID uint, loggerType, baseURL string) (*PayloadTemplateResponse, error) {
	tenantID := scope.TenantID
	if tenantID == 0 {
		tenantID = tenantctx.DefaultTenantID
	}

	var site *model.Site
	if siteID > 0 && s.tenantService != nil && s.tenantService.GetRepo() != nil {
		if st, err := s.tenantService.GetRepo().GetSiteByID(siteID); err == nil && st != nil && st.TenantID == tenantID {
			site = st
		}
	}
	if site == nil && s.tenantService != nil {
		site, _ = s.tenantService.ResolveEffectiveSite(scope)
	}

	lt := strings.ToUpper(strings.TrimSpace(loggerType))
	if lt == "" && site != nil {
		lt = strings.ToUpper(strings.TrimSpace(site.MonitoringDomain))
	}
	if lt == "" || lt == "WQMS" && loggerType == "" {
		if lt == "" {
			lt = "AQMS"
		}
	}

	siteCode := "stasiun-" + strings.ToLower(lt) + "-01"
	if site != nil && site.Code != "" {
		siteCode = site.Code
	}

	var dev *model.Device
	if deviceID > 0 && s.deviceRepo != nil {
		if d, err := s.deviceRepo.GetByID(deviceID); err == nil && d != nil && d.TenantID == tenantID {
			dev = d
		}
	}
	if dev == nil && site != nil && s.deviceRepo != nil {
		enabledFlag := true
		devs, _, _ := s.deviceRepo.List(repository.DeviceFilterParams{
			TenantID: tenantID,
			SiteID:   site.ID,
			Enabled:  &enabledFlag,
			PageSize: 1,
		})
		if len(devs) > 0 {
			dev = &devs[0]
		}
	}

	devCode := lt + "-LOGGER-01"
	var regParams []RegisteredParamSummary
	values1 := make(map[string]float64)
	values2 := make(map[string]float64)
	usingActual := false

	if dev != nil {
		devCode = dev.DeviceCode
		for _, p := range dev.Parameters {
			if !p.Enabled {
				continue
			}
			sv := sampleValueForCode(p.ParameterCode, lt)
			regParams = append(regParams, RegisteredParamSummary{
				ParameterCode: p.ParameterCode,
				ParameterName: p.ParameterName,
				Unit:          p.Unit,
				DataType:      string(p.DataType),
				SampleValue:   sv,
			})
			values1[p.ParameterCode] = sv
			values2[p.ParameterCode] = roundIngestPrecision(sv*1.03, 2)
		}
		if len(values1) > 0 {
			usingActual = true
		}
	}

	if !usingActual {
		canon := canonicalValuesForLoggerType(lt)
		for k, v := range canon {
			values1[k] = v
			values2[k] = roundIngestPrecision(v*1.02, 2)
		}
	}

	loc := time.FixedZone("WIB", 7*3600)
	if site != nil {
		loc = loadStationLocation(site.Timezone)
	}
	nowLocal := time.Now().In(loc).Truncate(time.Minute)
	prevLocal := nowLocal.Add(-5 * time.Minute)

	singlePayload := map[string]interface{}{
		"logger_type": lt,
		"site_code":   siteCode,
		"device_code": devCode,
		"timestamp":   nowLocal.Format(time.RFC3339),
		"values":      values1,
	}

	batchPayload := map[string]interface{}{
		"logger_type": lt,
		"site_code":   siteCode,
		"device_code": devCode,
		"records": []map[string]interface{}{
			{
				"timestamp": prevLocal.Format(time.RFC3339),
				"values":    values1,
			},
			{
				"timestamp": nowLocal.Format(time.RFC3339),
				"values":    values2,
			},
		},
	}

	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://central.example.com"
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	endpointURL := baseURL + "/api/v1/ingest/telemetry"
	mqttEndpointURL := baseURL + "/api/v1/ingest/mqtt"
	mqttTopic := fmt.Sprintf("datalogger/ingest/%s/%s", siteCode, devCode)

	prettySingle, _ := json.MarshalIndent(singlePayload, "", "  ")
	curlCmd := fmt.Sprintf(
		"curl -X POST \"%s\" \\\n  -H \"Authorization: Bearer dlg_live_YOUR_SECRET_TOKEN\" \\\n  -H \"Content-Type: application/json\" \\\n  -d '%s'",
		endpointURL, string(prettySingle),
	)

	return &PayloadTemplateResponse{
		LoggerType:           lt,
		SiteCode:             siteCode,
		DeviceCode:           devCode,
		EndpointURL:          endpointURL,
		MQTTEndpointURL:      mqttEndpointURL,
		MQTTTopic:            mqttTopic,
		RegisteredParameters: regParams,
		UsingActualParams:    usingActual,
		SinglePayload:        singlePayload,
		BatchPayload:         batchPayload,
		CurlBearerCommand:    curlCmd,
	}, nil
}

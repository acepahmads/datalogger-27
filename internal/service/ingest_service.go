package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
)

type IngestService struct {
	tenantService    *TenantService
	deviceRepo       *repository.DeviceRepository
	categoryRepo     *repository.CategoryRepository
	telemetryService *TelemetryService
	provisionMu      sync.Mutex
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

type EdgeIngestResponse struct {
	Accepted              bool     `json:"accepted"`
	TenantID              uint     `json:"tenant_id"`
	SiteID                uint     `json:"site_id"`
	SiteCode              string   `json:"site_code"`
	DevicesProcessed      int      `json:"devices_processed"`
	ParametersIngested    int      `json:"parameters_ingested"`
	ParametersSkipped     int      `json:"parameters_skipped,omitempty"`
	AutoProvisionedDevs   []string `json:"auto_provisioned_devices,omitempty"`
	AutoProvisionedParams []string `json:"auto_provisioned_parameters,omitempty"`
	ReceivedAt            string   `json:"received_at"`
}

type edgeReading struct {
	Code  string
	Name  string
	Unit  string
	Value float64
}

type edgeDeviceBatch struct {
	SiteCode   string
	DeviceCode string
	DeviceName string
	Timestamp  *time.Time
	Readings   []edgeReading
}

var reservedEnvelopeKeys = map[string]bool{
	"site_id": true, "site_code": true, "station_code": true, "id_stasiun": true, "idstasiun": true,
	"tenant_id": true, "tenant_code": true,
	"device_id": true, "device_code": true, "device_name": true, "module_id": true,
	"timestamp": true, "datetime": true, "waktu": true, "tanggal": true, "jam": true,
	"period_start": true, "period_end": true, "interval_seconds": true,
	"apikey": true, "apisecret": true, "token": true, "uid": true,
	"devices": true, "readings": true, "parameters": true, "data": true, "values": true,
}

func parseIngestTimestamp(raw interface{}) *time.Time {
	if raw == nil {
		return nil
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
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, s); err == nil {
				utc := t.UTC()
				return &utc
			}
		}
	}
	return nil
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
		for k, v := range c {
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

func parseEdgeBatches(rawJSON []byte) ([]edgeDeviceBatch, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(rawJSON, &root); err != nil {
		return nil, errors.New("invalid JSON payload: " + err.Error())
	}

	rootSiteCode := stringField(root, "site_code", "station_code")
	rootDevCode := stringField(root, "device_code", "code", "module_id")
	rootDevName := stringField(root, "device_name", "name")
	rootTS := parseIngestTimestamp(root["timestamp"])
	if rootTS == nil {
		rootTS = parseIngestTimestamp(root["datetime"])
	}
	if rootTS == nil {
		rootTS = parseIngestTimestamp(root["period_end"])
	}

	// 1. Multi-device Outbound Sync format: {"devices": [ ... ]}
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
			ts := parseIngestTimestamp(dm["timestamp"])
			if ts == nil {
				ts = rootTS
			}
			var readings []edgeReading
			for _, key := range []string{"readings", "parameters", "data", "values"} {
				if sub, exists := dm[key]; exists {
					readings = append(readings, extractReadingsFromAny(sub)...)
				}
			}
			if len(readings) == 0 {
				readings = extractReadingsFromAny(dm)
			}
			if len(readings) > 0 {
				batches = append(batches, edgeDeviceBatch{
					SiteCode:   stringField(dm, "site_code", "station_code"),
					DeviceCode: dCode,
					DeviceName: dName,
					Timestamp:  ts,
					Readings:   readings,
				})
			}
		}
		if len(batches) > 0 {
			for i := range batches {
				if batches[i].SiteCode == "" {
					batches[i].SiteCode = rootSiteCode
				}
			}
			return batches, nil
		}
	}

	// 2. Single-device Readings / Parameters / Data envelope or Flat JSON object
	var readings []edgeReading
	for _, key := range []string{"readings", "parameters", "data", "values"} {
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
		return nil, errors.New("payload contains no numeric telemetry readings")
	}

	return []edgeDeviceBatch{
		{
			SiteCode:   rootSiteCode,
			DeviceCode: rootDevCode,
			DeviceName: rootDevName,
			Timestamp:  rootTS,
			Readings:   readings,
		},
	}, nil
}

func inferDefaultUnitAndCategory(code string, categories []model.ParameterCategory) (string, *uint) {
	upper := strings.ToUpper(strings.TrimSpace(code))
	defaultUnits := map[string]string{
		"PM25": "µg/m³", "PM2.5": "µg/m³", "PM10": "µg/m³",
		"SO2": "µg/m³", "CO": "µg/m³", "O3": "µg/m³", "NO2": "µg/m³", "HC": "µg/m³",
		"TEMP": "°C", "TEMPERATURE": "°C", "HUMIDITY": "%", "RH": "%",
		"WS": "m/s", "WD": "°", "PRESSURE": "hPa", "SR": "W/m²", "RAIN": "mm",
		"PH": "-", "COD": "mg/L", "BOD": "mg/L", "TSS": "mg/L", "NH3N": "mg/L", "NH4": "mg/L", "NO3": "mg/L",
		"DO": "mg/L", "DEBIT": "m³/s", "FLOW": "m³/s", "TDS": "mg/L", "TURBIDITY": "NTU",
	}
	unit := defaultUnits[upper]

	cleanKey := strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(code), ".", ""), "_", "")
	for _, cat := range categories {
		catKey := strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(cat.Code), ".", ""), "_", "")
		if catKey != "" && (catKey == cleanKey || strings.EqualFold(cat.Code, code)) {
			cid := cat.ID
			return unit, &cid
		}
	}
	return unit, nil
}

// IngestPayload authenticates an Edge Push API Key and ingests single or multi-device telemetry into the Tenant's workspace.
func (s *IngestService) IngestPayload(rawKey string, rawJSON []byte, clientIP string) (*EdgeIngestResponse, error) {
	apiKey, err := s.tenantService.ValidateIngestKey(rawKey)
	if err != nil {
		return nil, err
	}
	tenantID := apiKey.TenantID

	batches, err := parseEdgeBatches(rawJSON)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var (
		totalIngested    int
		totalSkipped     int
		autoDevs         []string
		autoParams       []string
		resolvedSiteID   uint
		resolvedSiteCode string
	)

	s.provisionMu.Lock()
	var tenantCategories []model.ParameterCategory
	if s.categoryRepo != nil {
		tenantCategories, _ = s.categoryRepo.ListForTenant(tenantID, true, "")
	}

	for _, batch := range batches {
		// 1. Resolve Site for this batch (strictly inside apiKey.TenantID)
		var site *model.Site
		if apiKey.SiteID != nil && *apiKey.SiteID > 0 {
			site, _ = s.tenantService.GetRepo().GetSiteByID(*apiKey.SiteID)
		}
		if site == nil && batch.SiteCode != "" {
			site, _ = s.tenantService.GetRepo().GetSiteByCode(tenantID, model.NormalizeSlugCode(batch.SiteCode))
			if site == nil && apiKey.AutoProvision {
				t, _ := s.tenantService.GetTenantByID(tenantID)
				siteCount, _ := s.tenantService.GetRepo().CountSitesByTenant(tenantID)
				if t != nil && (t.MaxSites <= 0 || int(siteCount) < t.MaxSites) {
					slug := model.NormalizeSlugCode(batch.SiteCode)
					if slug != "" {
						newSite := &model.Site{
							TenantID:         tenantID,
							Code:             slug,
							Name:             batch.SiteCode,
							MonitoringDomain: "AQMS",
							StationTypeMode:  "MANUAL",
							LocationSource:   model.LocationSourceUnconfigured,
							Timezone:         "Asia/Jakarta",
							Enabled:          true,
							UpdatedBy:        "edge_ingest",
						}
						if s.tenantService.GetRepo().CreateSite(newSite) == nil {
							site = newSite
						}
					}
				}
			}
		}
		if site == nil {
			site, _ = s.tenantService.GetRepo().GetDefaultSiteForTenant(tenantID)
		}
		if site == nil || site.TenantID != tenantID {
			s.provisionMu.Unlock()
			return nil, errors.New("unable to resolve valid station for tenant")
		}
		resolvedSiteID = site.ID
		resolvedSiteCode = site.Code

		// 2. Resolve or Auto-Provision Device inside Tenant
		devCode := strings.TrimSpace(batch.DeviceCode)
		if devCode == "" {
			devCode = "EDGE-" + strings.ToUpper(site.Code)
		}
		dev, _ := s.deviceRepo.GetByCodeInTenant(tenantID, devCode)
		if dev == nil {
			if !apiKey.AutoProvision {
				s.provisionMu.Unlock()
				return nil, fmt.Errorf("device '%s' is not registered in tenant and auto_provision is disabled", devCode)
			}
			if qErr := s.tenantService.CheckDeviceQuota(tenantID, 1); qErr != nil {
				s.provisionMu.Unlock()
				return nil, qErr
			}
			devName := strings.TrimSpace(batch.DeviceName)
			if devName == "" {
				devName = devCode
			}
			newDev := &model.Device{
				TenantID:          tenantID,
				SiteID:            site.ID,
				DeviceCode:        devCode,
				Code:              devCode,
				DeviceName:        devName,
				Name:              devName,
				DeviceType:        string(model.ProtocolEdgePush),
				Timezone:          site.Timezone,
				Status:            model.DeviceStatusActive,
				Enabled:           true,
				ConnectionStatus:  model.DeviceConnOnline,
				LastSeenAt:        &now,
				LastDataAt:        &now,
				LastCommunication: &now,
				Connection: &model.DeviceConnection{
					Protocol:       model.ProtocolEdgePush,
					ConnectionType: "HTTPS_PUSH",
					Enabled:        true,
				},
			}
			if cErr := s.deviceRepo.Create(newDev); cErr != nil {
				s.provisionMu.Unlock()
				return nil, fmt.Errorf("failed to auto-provision device '%s': %w", devCode, cErr)
			}
			dev = newDev
			autoDevs = append(autoDevs, devCode)
		} else if !dev.Enabled {
			continue
		}

		paramByCode := make(map[string]*model.Parameter, len(dev.Parameters))
		for i := range dev.Parameters {
			p := &dev.Parameters[i]
			paramByCode[strings.ToUpper(strings.TrimSpace(p.ParameterCode))] = p
			if p.Code != "" {
				paramByCode[strings.ToUpper(strings.TrimSpace(p.Code))] = p
			}
		}

		devTS := batch.Timestamp
		if devTS == nil {
			devTS = &now
		}

		for _, r := range batch.Readings {
			pKey := strings.ToUpper(strings.TrimSpace(r.Code))
			param := paramByCode[pKey]
			if param == nil {
				if existingP, gErr := s.deviceRepo.GetParameterByCode(dev.ID, r.Code); gErr == nil && existingP != nil {
					param = existingP
					paramByCode[pKey] = param
				}
			}
			if param == nil {
				if !apiKey.AutoProvision {
					totalSkipped++
					continue
				}
				pName := r.Name
				if pName == "" {
					pName = r.Code
				}
				inferredUnit, catID := inferDefaultUnitAndCategory(r.Code, tenantCategories)
				pUnit := r.Unit
				if pUnit == "" {
					pUnit = inferredUnit
				}
				newParam := &model.Parameter{
					DeviceID:                 dev.ID,
					ParameterCode:            r.Code,
					Code:                     r.Code,
					ParameterName:            pName,
					Name:                     pName,
					DataType:                 model.DataTypeFloat32,
					Unit:                     pUnit,
					CategoryID:               catID,
					Precision:                2,
					Scale:                    1.0,
					ScaleFactor:              1.0,
					Enabled:                  true,
					QualityValidationEnabled: true,
					ProcessingEnabled:        true,
					StaleTimeoutSeconds:      300,
				}
				if pErr := s.deviceRepo.CreateParameter(newParam); pErr != nil {
					totalSkipped++
					continue
				}
				param = newParam
				paramByCode[pKey] = param
				autoParams = append(autoParams, fmt.Sprintf("%s:%s", dev.DeviceCode, param.ParameterCode))
			}

			if !param.Enabled {
				totalSkipped++
				continue
			}

			uncal := r.Value
			if s.telemetryService != nil {
				_ = s.telemetryService.Ingest(&TelemetryIngestPayload{
					TenantID:          tenantID,
					SiteID:            dev.SiteID,
					DeviceID:          dev.ID,
					DeviceCode:        dev.DeviceCode,
					DeviceName:        dev.DeviceName,
					ParameterID:       param.ID,
					ParameterCode:     param.ParameterCode,
					ParameterName:     param.ParameterName,
					Unit:              param.Unit,
					DataType:          param.DataType,
					RawValue:          r.Value,
					UncalibratedValue: &uncal,
					Value:             r.Value,
					ProcessedValue:    r.Value,
					Quality:           model.QualityGood,
					QualityReason:     model.ReasonNone,
					Source:            "EDGE_PUSH",
					DeviceTimestamp:   devTS,
					ReceivedAt:        now,
					Parameter:         param,
				})
			}
			totalIngested++
		}

		_ = s.deviceRepo.UpdateLastSeen(dev.ID, now)
		_ = s.deviceRepo.UpdateLastData(dev.ID, now)
		_ = s.deviceRepo.SetConnectionStatus(dev.ID, model.DeviceConnOnline)
	}
	s.provisionMu.Unlock()

	if s.tenantService.GetRepo() != nil && totalIngested > 0 {
		s.tenantService.GetRepo().RecordAPIKeyUsage(apiKey.ID, clientIP, int64(totalIngested))
	}

	return &EdgeIngestResponse{
		Accepted:              true,
		TenantID:              tenantID,
		SiteID:                resolvedSiteID,
		SiteCode:              resolvedSiteCode,
		DevicesProcessed:      len(batches),
		ParametersIngested:    totalIngested,
		ParametersSkipped:     totalSkipped,
		AutoProvisionedDevs:   autoDevs,
		AutoProvisionedParams: autoParams,
		ReceivedAt:            now.Format(time.RFC3339),
	}, nil
}

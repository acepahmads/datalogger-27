package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"datalogger/internal/calibration"
	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/pkg/tenantctx"
)

// CalibratableParameterDTO enriches a parameter with parent device info, active calibration curve, and live telemetry
type CalibratableParameterDTO struct {
	ParameterID      uint    `json:"parameter_id"`
	DeviceID         uint    `json:"device_id"`
	DeviceCode       string  `json:"device_code"`
	DeviceName       string  `json:"device_name"`
	Protocol         string  `json:"protocol"`
	DeviceOnline     bool    `json:"device_online"`
	ParameterCode    string  `json:"parameter_code"`
	ParameterName    string  `json:"parameter_name"`
	Unit             string  `json:"unit"`
	Precision        int     `json:"precision"`
	RegisterAddress  int     `json:"register_address"`
	RegisterType     string  `json:"register_type"`
	RegisterScale    float64 `json:"register_scale"`
	RegisterOffset   float64 `json:"register_offset"`
	Formula          string  `json:"formula,omitempty"`
	Enabled          bool    `json:"enabled"`
	EffectiveDomain  string  `json:"effective_domain"`
	ExplicitDomain   string  `json:"explicit_domain"`
	CategoryName     string  `json:"category_name,omitempty"`

	CalibrationEnabled          bool                            `json:"calibration_enabled"`
	CalibrationModel            string                          `json:"calibration_model"`
	CalC2                       float64                         `json:"cal_c2"`
	CalC1                       float64                         `json:"cal_c1"`
	CalC0                       float64                         `json:"cal_c0"`
	CalMinR2                    float64                         `json:"cal_min_r2"`
	CalR2                       *float64                        `json:"cal_r2,omitempty"`
	CalRMSE                     *float64                        `json:"cal_rmse,omitempty"`
	CalPointCount               int                             `json:"cal_point_count"`
	CalPoints                   []calibration.CalibrationPoint  `json:"cal_points"`
	CalStandardNote             string                          `json:"cal_standard_note,omitempty"`
	CalEquation                 string                          `json:"cal_equation"`
	CalUpdatedAt                *time.Time                      `json:"cal_updated_at,omitempty"`
	CalUpdatedBy                string                          `json:"cal_updated_by,omitempty"`
	CalibrationSessionActive    bool                            `json:"calibration_session_active"`
	CalibrationSessionStartedAt *time.Time                      `json:"calibration_session_started_at,omitempty"`

	LiveUncalibratedValue *float64   `json:"live_uncalibrated_value,omitempty"`
	LiveCalibratedValue   *float64   `json:"live_calibrated_value,omitempty"`
	LiveRawValue          *float64   `json:"live_raw_value,omitempty"`
	LiveQuality           string     `json:"live_quality,omitempty"`
	LiveReceivedAt        *time.Time `json:"live_received_at,omitempty"`
}

type CalculateCalibrationRequest struct {
	Model  string                         `json:"model"`
	Points []calibration.CalibrationPoint `json:"points"`
	MinR2  float64                        `json:"min_r2"`
}

type ApplyCalibrationRequest struct {
	Model         string                         `json:"model"`
	Domain        string                         `json:"domain"`
	Points        []calibration.CalibrationPoint `json:"points"`
	MinR2         float64                        `json:"min_r2"`
	StandardNote  string                         `json:"standard_note"`
	EndSession    bool                           `json:"end_session"`
	ForceOverride bool                           `json:"force_override"`
}

type UpdateCalibrationConfigRequest struct {
	CalMinR2     *float64 `json:"cal_min_r2"`
	Domain       *string  `json:"domain"`
	ClearFormula bool     `json:"clear_formula"`
}

type ToggleCalibrationSessionRequest struct {
	Active bool `json:"active"`
}

type ResetCalibrationRequest struct {
	ReasonNote string `json:"reason_note"`
}

type LiveCalibrationReadingDTO struct {
	ParameterID              uint       `json:"parameter_id"`
	DeviceID                 uint       `json:"device_id"`
	ParameterCode            string     `json:"parameter_code"`
	ParameterName            string     `json:"parameter_name"`
	Unit                     string     `json:"unit"`
	UncalibratedValue        *float64   `json:"uncalibrated_value"`
	CalibratedValue          *float64   `json:"calibrated_value"`
	RawValue                 *float64   `json:"raw_value,omitempty"`
	Quality                  string     `json:"quality"`
	CalibrationEnabled       bool       `json:"calibration_enabled"`
	CalibrationSessionActive bool       `json:"calibration_session_active"`
	CalEquation              string     `json:"cal_equation"`
	ReceivedAt               *time.Time `json:"received_at,omitempty"`
	PolledOnDemand           bool       `json:"polled_on_demand"`
}

type OnDemandParameterPollFunc func(ctx context.Context, deviceID, paramID uint) (uncalVal, calVal, rawVal float64, err error)

type CalibrationService struct {
	repo               *repository.CalibrationRepository
	deviceRepo         *repository.DeviceRepository
	systemRepo         *repository.SystemRepository
	telemetryService   *TelemetryService
	tenantService      *TenantService
	onDeviceReload     func(deviceID uint)
	onDemandPoll       OnDemandParameterPollFunc
	onClearParamAlarms func(ctx context.Context, deviceID, paramID uint)
}

func NewCalibrationService(
	repo *repository.CalibrationRepository,
	deviceRepo *repository.DeviceRepository,
	systemRepo *repository.SystemRepository,
	telemetryService *TelemetryService,
) *CalibrationService {
	return &CalibrationService{
		repo:             repo,
		deviceRepo:       deviceRepo,
		systemRepo:       systemRepo,
		telemetryService: telemetryService,
	}
}

func (s *CalibrationService) SetTenantService(ts *TenantService) {
	s.tenantService = ts
}

func (s *CalibrationService) SetOnDeviceReload(fn func(deviceID uint)) {
	s.onDeviceReload = fn
}

func (s *CalibrationService) SetOnDemandPoll(fn OnDemandParameterPollFunc) {
	s.onDemandPoll = fn
}

func (s *CalibrationService) SetOnClearParamAlarms(fn func(ctx context.Context, deviceID, paramID uint)) {
	s.onClearParamAlarms = fn
}

func (s *CalibrationService) resolveStationDomainCtx(ctx context.Context) string {
	scope := tenantctx.FromContext(ctx)
	if s.tenantService != nil && (scope.TenantID > 0 || scope.SiteID > 0) {
		if site, err := s.tenantService.ResolveEffectiveSite(scope); err == nil && site != nil {
			sd := strings.TrimSpace(site.MonitoringDomain)
			if sd != "" && !strings.EqualFold(sd, "CUSTOM") {
				return sd
			}
		}
	}
	return s.resolveStationDomain()
}

func (s *CalibrationService) checkDeviceScope(ctx context.Context, dev *model.Device) error {
	if dev == nil {
		return nil
	}
	scope := tenantctx.FromContext(ctx)
	if scope.Authenticated && !scope.CanAccessRecord(dev.TenantID, dev.SiteID) {
		return errors.New("akses ditolak: parameter milik tenant atau stasiun lain")
	}
	return nil
}

func (s *CalibrationService) resolveStationDomain() string {
	if globalCfg := config.Get(); globalCfg != nil {
		gd := strings.TrimSpace(globalCfg.MonitoringDomain)
		if gd != "" && !strings.EqualFold(gd, "CUSTOM") {
			return gd
		}
	}
	if s.systemRepo == nil {
		return ""
	}
	branding, err := s.systemRepo.GetBranding()
	if err != nil || branding == nil {
		return ""
	}
	return branding.CustomerSubtitle
}

func isSharedWaterParameter(paramCode, paramName, unit string) bool {
	tok := strings.ToUpper(paramCode + " " + paramName + " " + unit)
	for _, kw := range []string{"PH", "COD", "TSS", "NH3", "AMMONIA", "AMONIA", "DEBIT", "FLOW", "SUHU AIR", "WATER TEMP"} {
		if strings.Contains(tok, kw) {
			return true
		}
	}
	return false
}

func matchesRequestedDomain(reqDomain string, param *model.Parameter, effectiveDomain string) bool {
	normReq := model.NormalizeCalibrationDomain(reqDomain)
	if normReq == "" {
		return true
	}
	if effectiveDomain == normReq {
		return true
	}
	// If parameter does not have a rigid explicit domain override, allow shared water parameters (pH, COD, TSS, NH3-N, Debit)
	// to appear seamlessly under both SPARING and WQMS_ONLIMO tabs.
	if param != nil && strings.TrimSpace(param.CalibrationDomain) == "" {
		if (normReq == model.CalibrationDomainSparing || normReq == model.CalibrationDomainWQMSOnlimo) &&
			isSharedWaterParameter(param.ParameterCode, param.ParameterName, param.Unit) {
			return true
		}
	}
	return false
}

func (s *CalibrationService) buildParameterDTO(ctx context.Context, p *model.Parameter, dev *model.Device, stationDomain string) CalibratableParameterDTO {
	effDomain := model.InferCalibrationDomain(p.CalibrationDomain, p.ParameterCode, p.ParameterName, p.Unit, stationDomain)
	var pts []calibration.CalibrationPoint
	if strings.TrimSpace(p.CalPointsJSON) != "" {
		_ = json.Unmarshal([]byte(p.CalPointsJSON), &pts)
	}
	if pts == nil {
		pts = []calibration.CalibrationPoint{}
	}

	proto := string(dev.DeviceType)
	if dev.Connection != nil && dev.Connection.Protocol != "" {
		proto = string(dev.Connection.Protocol)
	}
	catName := ""
	if p.Category != nil {
		catName = p.Category.Name
	}

	dto := CalibratableParameterDTO{
		ParameterID:                 p.ID,
		DeviceID:                    p.DeviceID,
		DeviceCode:                  dev.DeviceCode,
		DeviceName:                  dev.DeviceName,
		Protocol:                    proto,
		DeviceOnline:                dev.ConnectionStatus == model.DeviceConnOnline,
		ParameterCode:               p.ParameterCode,
		ParameterName:               p.ParameterName,
		Unit:                        p.Unit,
		Precision:                   p.Precision,
		RegisterAddress:             p.RegisterAddress,
		RegisterType:                p.RegisterType,
		RegisterScale:               p.Scale,
		RegisterOffset:              p.Offset,
		Formula:                     p.Formula,
		Enabled:                     p.Enabled,
		EffectiveDomain:             effDomain,
		ExplicitDomain:              p.CalibrationDomain,
		CategoryName:                catName,
		CalibrationEnabled:          p.CalibrationEnabled,
		CalibrationModel:            calibration.NormalizeModel(p.CalibrationModel),
		CalC2:                       p.CalC2,
		CalC1:                       p.CalC1,
		CalC0:                       p.CalC0,
		CalMinR2:                    calibration.NormalizeMinR2(p.CalMinR2),
		CalR2:                       p.CalR2,
		CalRMSE:                     p.CalRMSE,
		CalPointCount:               p.CalPointCount,
		CalPoints:                   pts,
		CalStandardNote:             p.CalStandardNote,
		CalEquation:                 p.CalibrationEquation(),
		CalUpdatedAt:                p.CalUpdatedAt,
		CalUpdatedBy:                p.CalUpdatedBy,
		CalibrationSessionActive:    p.CalibrationSessionActive,
		CalibrationSessionStartedAt: p.CalibrationSessionStartedAt,
	}

	// Hydrate live readings from TelemetryService cache or Parameter DB state
	if s.telemetryService != nil {
		if latest, err := s.telemetryService.GetLatestForParameter(ctx, p.DeviceID, p.ID); err == nil && latest != nil {
			calV := latest.ProcessedValue
			dto.LiveCalibratedValue = &calV
			if latest.UncalibratedValue != nil {
				u := *latest.UncalibratedValue
				dto.LiveUncalibratedValue = &u
			} else if p.CurrentUncalibratedValue != nil {
				u := *p.CurrentUncalibratedValue
				dto.LiveUncalibratedValue = &u
			} else if !p.CalibrationEnabled {
				u := calV
				dto.LiveUncalibratedValue = &u
			}
			rawV := latest.RawValue
			dto.LiveRawValue = &rawV
			dto.LiveQuality = string(latest.Quality)
			if !latest.ReceivedAt.IsZero() {
				rec := latest.ReceivedAt
				dto.LiveReceivedAt = &rec
			}
		}
	}
	if dto.LiveUncalibratedValue == nil && p.CurrentUncalibratedValue != nil {
		u := *p.CurrentUncalibratedValue
		dto.LiveUncalibratedValue = &u
		calV := p.ApplyCalibration(u)
		dto.LiveCalibratedValue = &calV
	} else if dto.LiveCalibratedValue == nil && p.CurrentValue != nil {
		calV := *p.CurrentValue
		dto.LiveCalibratedValue = &calV
		if !p.CalibrationEnabled {
			u := calV
			dto.LiveUncalibratedValue = &u
		}
	}

	return dto
}

// ListCalibratableParameters returns all analog/numeric parameters enriched with live calibration state
func (s *CalibrationService) ListCalibratableParameters(ctx context.Context, deviceID uint, domainFilter string) ([]CalibratableParameterDTO, string, error) {
	params, devMap, err := s.repo.ListCalibratableParameters(ctx, deviceID)
	if err != nil {
		return nil, "", err
	}
	stationDomain := s.resolveStationDomainCtx(ctx)
	normStation := model.NormalizeCalibrationDomain(stationDomain)
	if normStation == "" {
		normStation = model.CalibrationDomainSparing
	}

	out := make([]CalibratableParameterDTO, 0, len(params))
	for i := range params {
		p := &params[i]
		dev, ok := devMap[p.DeviceID]
		if !ok {
			continue
		}
		dto := s.buildParameterDTO(ctx, p, &dev, stationDomain)
		if !matchesRequestedDomain(domainFilter, p, dto.EffectiveDomain) {
			continue
		}
		out = append(out, dto)
	}
	return out, normStation, nil
}

// GetLiveParameterReading retrieves instantaneous sensor reading (and optionally executes an on-demand Modbus poll)
func (s *CalibrationService) GetLiveParameterReading(ctx context.Context, paramID uint, pollNow bool) (*LiveCalibrationReadingDTO, error) {
	param, dev, err := s.repo.GetParameterWithDevice(ctx, paramID)
	if err != nil {
		return nil, errors.New("parameter tidak ditemukan")
	}
	if err := s.checkDeviceScope(ctx, dev); err != nil {
		return nil, err
	}
	if err := s.checkDeviceScope(ctx, dev); err != nil {
		return nil, err
	}

	res := &LiveCalibrationReadingDTO{
		ParameterID:              param.ID,
		DeviceID:                 param.DeviceID,
		ParameterCode:            param.ParameterCode,
		ParameterName:            param.ParameterName,
		Unit:                     param.Unit,
		Quality:                  string(model.QualityUnknown),
		CalibrationEnabled:       param.CalibrationEnabled,
		CalibrationSessionActive: param.CalibrationSessionActive,
		CalEquation:              param.CalibrationEquation(),
	}

	if pollNow && s.onDemandPoll != nil && dev != nil {
		if uncal, cal, raw, pollErr := s.onDemandPoll(ctx, dev.ID, param.ID); pollErr == nil {
			now := time.Now().UTC()
			res.UncalibratedValue = &uncal
			res.CalibratedValue = &cal
			res.RawValue = &raw
			res.Quality = string(model.QualityGood)
			if param.CalibrationSessionActive {
				res.Quality = string(model.QualityUncertain)
			}
			res.ReceivedAt = &now
			res.PolledOnDemand = true
			return res, nil
		}
	}

	dto := s.buildParameterDTO(ctx, param, dev, "")
	res.UncalibratedValue = dto.LiveUncalibratedValue
	res.CalibratedValue = dto.LiveCalibratedValue
	res.RawValue = dto.LiveRawValue
	if dto.LiveQuality != "" {
		res.Quality = dto.LiveQuality
	}
	res.ReceivedAt = dto.LiveReceivedAt
	return res, nil
}

// CalculatePreview computes 1–10 point regression coefficients and diagnostics without modifying DB state
func (s *CalibrationService) CalculatePreview(req *CalculateCalibrationRequest) (*calibration.RegressionResult, error) {
	if req == nil {
		return nil, errors.New("payload perhitungan kosong")
	}
	return calibration.CalculateRegression(req.Model, req.Points, req.MinR2)
}

// UpdateParameterConfig updates per-parameter minimum R2 threshold and/or explicit domain tag
func (s *CalibrationService) UpdateParameterConfig(
	ctx context.Context,
	paramID uint,
	req *UpdateCalibrationConfigRequest,
	username, role, ip, ua string,
) (*CalibratableParameterDTO, error) {
	if req == nil {
		return nil, errors.New("payload konfigurasi kosong")
	}
	param, dev, err := s.repo.GetParameterWithDevice(ctx, paramID)
	if err != nil {
		return nil, errors.New("parameter tidak ditemukan")
	}
	if err := s.checkDeviceScope(ctx, dev); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.CalMinR2 != nil {
		v := *req.CalMinR2
		if v < 0.5 || v > 1.0 {
			return nil, errors.New("batas minimum R² harus berada di rentang 0.5000 hingga 1.0000")
		}
		updates["cal_min_r2"] = v
		param.CalMinR2 = v
	}
	if req.Domain != nil {
		normDom := model.NormalizeCalibrationDomain(*req.Domain)
		updates["calibration_domain"] = normDom
		param.CalibrationDomain = normDom
	}
	prevFormula := strings.TrimSpace(param.Formula)
	if req.ClearFormula && prevFormula != "" {
		updates["formula"] = ""
		updates["current_formula_value"] = nil
		param.Formula = ""
		param.CurrentFormulaValue = nil
		if param.CurrentUncalibratedValue != nil {
			cleanCur := param.ApplyCalibration(*param.CurrentUncalibratedValue)
			updates["current_value"] = &cleanCur
			param.CurrentValue = &cleanCur
		}
	}

	if len(updates) > 0 {
		if err := s.repo.UpdateParameterCalibrationFields(ctx, param.ID, updates); err != nil {
			return nil, fmt.Errorf("gagal menyimpan pengaturan parameter kalibrasi: %w", err)
		}
		s.recordAudit(username, "UPDATE_CALIBRATION_CONFIG", fmt.Sprintf("parameter:%d", param.ID), map[string]interface{}{
			"parameter_code":           param.ParameterCode,
			"device_id":                param.DeviceID,
			"updates":                  updates,
			"cleared_previous_formula": prevFormula,
			"role":                     role,
		}, ip, ua)
	}

	if s.telemetryService != nil {
		s.telemetryService.SyncParameterCalibrationCache(param)
	}

	if s.onDeviceReload != nil {
		s.onDeviceReload(param.DeviceID)
	}

	dto := s.buildParameterDTO(ctx, param, dev, s.resolveStationDomain())
	return &dto, nil
}

// ToggleCalibrationSession activates or deactivates calibration maintenance mode on a parameter
func (s *CalibrationService) ToggleCalibrationSession(
	ctx context.Context,
	paramID uint,
	active bool,
	username, role, ip, ua string,
) (*CalibratableParameterDTO, error) {
	param, dev, err := s.repo.GetParameterWithDevice(ctx, paramID)
	if err != nil {
		return nil, errors.New("parameter tidak ditemukan")
	}
	if err := s.checkDeviceScope(ctx, dev); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var startedAt *time.Time
	if active {
		startedAt = &now
	}

	updates := map[string]interface{}{
		"calibration_session_active":     active,
		"calibration_session_started_at": startedAt,
	}
	clearedFormula := ""
	if active && strings.TrimSpace(param.Formula) != "" {
		clearedFormula = strings.TrimSpace(param.Formula)
		updates["formula"] = ""
		updates["current_formula_value"] = nil
		param.Formula = ""
		param.CurrentFormulaValue = nil
		if param.CurrentUncalibratedValue != nil {
			cleanCur := param.ApplyCalibration(*param.CurrentUncalibratedValue)
			updates["current_value"] = &cleanCur
			param.CurrentValue = &cleanCur
		}
	}
	if err := s.repo.UpdateParameterCalibrationFields(ctx, param.ID, updates); err != nil {
		return nil, fmt.Errorf("gagal mengubah status sesi kalibrasi: %w", err)
	}

	param.CalibrationSessionActive = active
	param.CalibrationSessionStartedAt = startedAt

	if active {
		if s.onClearParamAlarms != nil {
			s.onClearParamAlarms(ctx, param.DeviceID, param.ID)
		}
		if s.systemRepo != nil {
			_ = s.systemRepo.ClearActiveAlarmsByParameter(param.ID)
		}
	}

	if s.telemetryService != nil {
		s.telemetryService.SyncParameterCalibrationCache(param)
	}
	if s.onDeviceReload != nil {
		s.onDeviceReload(param.DeviceID)
	}

	action := "STOP_CALIBRATION_SESSION"
	if active {
		action = "START_CALIBRATION_SESSION"
	}
	s.recordAudit(username, action, fmt.Sprintf("parameter:%d", param.ID), map[string]interface{}{
		"parameter_code":           param.ParameterCode,
		"parameter_name":           param.ParameterName,
		"device_id":                param.DeviceID,
		"active":                   active,
		"cleared_previous_formula": clearedFormula,
		"role":                     role,
	}, ip, ua)

	dto := s.buildParameterDTO(ctx, param, dev, s.resolveStationDomain())
	return &dto, nil
}

// ApplyCalibration computes, validates, persists, and hot-activates a 1–10 point calibration curve
func (s *CalibrationService) ApplyCalibration(
	ctx context.Context,
	paramID uint,
	req *ApplyCalibrationRequest,
	username, role, ip, ua string,
) (*CalibratableParameterDTO, *model.SensorCalibrationLog, error) {
	if req == nil {
		return nil, nil, errors.New("payload kalibrasi kosong")
	}

	param, dev, err := s.repo.GetParameterWithDevice(ctx, paramID)
	if err != nil {
		return nil, nil, errors.New("parameter tidak ditemukan")
	}
	if err := s.checkDeviceScope(ctx, dev); err != nil {
		return nil, nil, err
	}
	if param.IsDiscreteStatus() {
		return nil, nil, errors.New("parameter diskrit/status tidak mendukung kalibrasi regresi")
	}

	effMinR2 := param.CalMinR2
	if req.MinR2 > 0 {
		if req.MinR2 < 0.5 || req.MinR2 > 1.0 {
			return nil, nil, errors.New("batas minimum R² harus berada di rentang 0.5000 hingga 1.0000")
		}
		effMinR2 = req.MinR2
	} else {
		effMinR2 = calibration.NormalizeMinR2(effMinR2)
	}

	regResult, err := calibration.CalculateRegression(req.Model, req.Points, effMinR2)
	if err != nil {
		return nil, nil, err
	}

	if !regResult.PassedR2 && !req.ForceOverride {
		r2Val := 0.0
		if regResult.R2 != nil {
			r2Val = *regResult.R2
		}
		return nil, nil, fmt.Errorf("skor R² (%.5f) di bawah batas lulus parameter (%.4f). Perbaiki titik outlier atau sesuaikan batas R² parameter", r2Val, effMinR2)
	}

	stationDomain := s.resolveStationDomain()
	effDomain := model.InferCalibrationDomain(req.Domain, param.ParameterCode, param.ParameterName, param.Unit, stationDomain)
	if req.Domain == "" && param.CalibrationDomain != "" {
		effDomain = model.InferCalibrationDomain(param.CalibrationDomain, param.ParameterCode, param.ParameterName, param.Unit, stationDomain)
	}

	ptsBytes, _ := json.Marshal(regResult.Points)
	ptsJSON := string(ptsBytes)
	now := time.Now().UTC()
	note := strings.TrimSpace(req.StandardNote)
	rmseVal := regResult.RMSE

	prevEnabled := param.CalibrationEnabled
	prevModel := param.CalibrationModel
	prevC2 := param.CalC2
	prevC1 := param.CalC1
	prevC0 := param.CalC0
	prevEq := param.CalibrationEquation()
	prevPts := param.CalPointsJSON
	prevNote := param.CalStandardNote
	clearedFormula := strings.TrimSpace(param.Formula)

	updates := map[string]interface{}{
		"calibration_enabled":   true,
		"calibration_model":     regResult.Model,
		"cal_c2":                regResult.C2,
		"cal_c1":                regResult.C1,
		"cal_c0":                regResult.C0,
		"cal_min_r2":            effMinR2,
		"cal_r2":                regResult.R2,
		"cal_rmse":              &rmseVal,
		"cal_point_count":       regResult.ActiveCount,
		"cal_points_json":       ptsJSON,
		"cal_standard_note":     note,
		"cal_updated_at":        &now,
		"cal_updated_by":        username,
		"formula":               "",
		"current_formula_value": nil,
	}
	if normDom := model.NormalizeCalibrationDomain(req.Domain); normDom != "" {
		updates["calibration_domain"] = normDom
		param.CalibrationDomain = normDom
	}
	if req.EndSession && param.CalibrationSessionActive {
		updates["calibration_session_active"] = false
		updates["calibration_session_started_at"] = nil
		param.CalibrationSessionActive = false
		param.CalibrationSessionStartedAt = nil
	}

	// Update current_value immediately if current_uncalibrated_value is known
	if param.CurrentUncalibratedValue != nil {
		param.CalibrationEnabled = true
		param.CalibrationModel = regResult.Model
		param.CalC2 = regResult.C2
		param.CalC1 = regResult.C1
		param.CalC0 = regResult.C0
		newCur := param.ApplyCalibration(*param.CurrentUncalibratedValue)
		updates["current_value"] = &newCur
		param.CurrentValue = &newCur
	}

	if err := s.repo.UpdateParameterCalibrationFields(ctx, param.ID, updates); err != nil {
		return nil, nil, fmt.Errorf("gagal menyimpan kurva kalibrasi: %w", err)
	}

	param.CalibrationEnabled = true
	param.CalibrationModel = regResult.Model
	param.CalC2 = regResult.C2
	param.CalC1 = regResult.C1
	param.CalC0 = regResult.C0
	param.CalMinR2 = effMinR2
	param.CalR2 = regResult.R2
	param.CalRMSE = &rmseVal
	param.CalPointCount = regResult.ActiveCount
	param.CalPointsJSON = ptsJSON
	param.CalStandardNote = note
	param.CalUpdatedAt = &now
	param.CalUpdatedBy = username
	param.Formula = ""
	param.CurrentFormulaValue = nil

	logEntry := &model.SensorCalibrationLog{
		TenantID:               dev.TenantID,
		SiteID:                 dev.SiteID,
		DeviceID:               param.DeviceID,
		DeviceCode:             dev.DeviceCode,
		DeviceName:             dev.DeviceName,
		ParameterID:            param.ID,
		ParameterCode:          param.ParameterCode,
		ParameterName:          param.ParameterName,
		Unit:                   param.Unit,
		Domain:                 effDomain,
		Action:                 model.CalibrationActionApply,
		CalibrationModel:       regResult.Model,
		PointCount:             regResult.ActiveCount,
		TotalPoints:            regResult.TotalCount,
		C2:                     regResult.C2,
		C1:                     regResult.C1,
		C0:                     regResult.C0,
		Equation:               regResult.Equation,
		PrevCalibrationEnabled: prevEnabled,
		PrevCalibrationModel:   prevModel,
		PrevC2:                 prevC2,
		PrevC1:                 prevC1,
		PrevC0:                 prevC0,
		PrevEquation:           prevEq,
		PrevPointsJSON:         prevPts,
		PrevStandardNote:       prevNote,
		R2:                     regResult.R2,
		MinR2:                  effMinR2,
		PassedR2:               regResult.PassedR2,
		RMSE:                   &rmseVal,
		PointsJSON:             ptsJSON,
		StandardNote:           note,
		PerformedBy:            username,
		PerformedRole:          role,
		IPAddress:              ip,
		CreatedAt:              now,
	}
	_ = s.repo.CreateCalibrationLog(ctx, logEntry)

	s.recordAudit(username, "APPLY_SENSOR_CALIBRATION", fmt.Sprintf("parameter:%d", param.ID), map[string]interface{}{
		"parameter_code":           param.ParameterCode,
		"domain":                   effDomain,
		"model":                    regResult.Model,
		"points":                   regResult.ActiveCount,
		"equation":                 regResult.Equation,
		"prev_equation":            prevEq,
		"r2":                       regResult.R2,
		"min_r2":                   effMinR2,
		"standard_note":            note,
		"cleared_previous_formula": clearedFormula,
	}, ip, ua)

	if s.telemetryService != nil {
		s.telemetryService.SyncParameterCalibrationCache(param)
	}
	if s.onDeviceReload != nil {
		s.onDeviceReload(param.DeviceID)
	}

	dto := s.buildParameterDTO(ctx, param, dev, stationDomain)
	return &dto, logEntry, nil
}

// ResetCalibration restores a parameter to identity calibration (y = x) without modifying Modbus register Scale & Offset
func (s *CalibrationService) ResetCalibration(
	ctx context.Context,
	paramID uint,
	reasonNote, username, role, ip, ua string,
) (*CalibratableParameterDTO, *model.SensorCalibrationLog, error) {
	param, dev, err := s.repo.GetParameterWithDevice(ctx, paramID)
	if err != nil {
		return nil, nil, errors.New("parameter tidak ditemukan")
	}
	if err := s.checkDeviceScope(ctx, dev); err != nil {
		return nil, nil, err
	}

	stationDomain := s.resolveStationDomain()
	effDomain := model.InferCalibrationDomain(param.CalibrationDomain, param.ParameterCode, param.ParameterName, param.Unit, stationDomain)

	prevEnabled := param.CalibrationEnabled
	prevModel := param.CalibrationModel
	prevC2 := param.CalC2
	prevC1 := param.CalC1
	prevC0 := param.CalC0
	prevEq := param.CalibrationEquation()
	prevPts := param.CalPointsJSON
	prevNote := param.CalStandardNote

	now := time.Now().UTC()
	note := strings.TrimSpace(reasonNote)
	if note == "" {
		note = "Reset ke kurva standar default (y = x)"
	}

	updates := map[string]interface{}{
		"calibration_enabled": false,
		"calibration_model":   calibration.ModelLinear,
		"cal_c2":              0.0,
		"cal_c1":              1.0,
		"cal_c0":              0.0,
		"cal_r2":              nil,
		"cal_rmse":            nil,
		"cal_point_count":     0,
		"cal_points_json":     "",
		"cal_standard_note":   note,
		"cal_updated_at":      &now,
		"cal_updated_by":      username,
	}
	if param.CurrentUncalibratedValue != nil {
		u := *param.CurrentUncalibratedValue
		updates["current_value"] = &u
		param.CurrentValue = &u
	}

	if err := s.repo.UpdateParameterCalibrationFields(ctx, param.ID, updates); err != nil {
		return nil, nil, fmt.Errorf("gagal mereset kurva kalibrasi: %w", err)
	}

	param.CalibrationEnabled = false
	param.CalibrationModel = calibration.ModelLinear
	param.CalC2 = 0
	param.CalC1 = 1
	param.CalC0 = 0
	param.CalR2 = nil
	param.CalRMSE = nil
	param.CalPointCount = 0
	param.CalPointsJSON = ""
	param.CalStandardNote = note
	param.CalUpdatedAt = &now
	param.CalUpdatedBy = username

	logEntry := &model.SensorCalibrationLog{
		TenantID:               dev.TenantID,
		SiteID:                 dev.SiteID,
		DeviceID:               param.DeviceID,
		DeviceCode:             dev.DeviceCode,
		DeviceName:             dev.DeviceName,
		ParameterID:            param.ID,
		ParameterCode:          param.ParameterCode,
		ParameterName:          param.ParameterName,
		Unit:                   param.Unit,
		Domain:                 effDomain,
		Action:                 model.CalibrationActionReset,
		CalibrationModel:       calibration.ModelNone,
		PointCount:             0,
		TotalPoints:            0,
		C2:                     0,
		C1:                     1,
		C0:                     0,
		Equation:               "y = x (Default)",
		PrevCalibrationEnabled: prevEnabled,
		PrevCalibrationModel:   prevModel,
		PrevC2:                 prevC2,
		PrevC1:                 prevC1,
		PrevC0:                 prevC0,
		PrevEquation:           prevEq,
		PrevPointsJSON:         prevPts,
		PrevStandardNote:       prevNote,
		MinR2:                  param.CalMinR2,
		PassedR2:               true,
		StandardNote:           note,
		PerformedBy:            username,
		PerformedRole:          role,
		IPAddress:              ip,
		CreatedAt:              now,
	}
	_ = s.repo.CreateCalibrationLog(ctx, logEntry)

	s.recordAudit(username, "RESET_SENSOR_CALIBRATION", fmt.Sprintf("parameter:%d", param.ID), map[string]interface{}{
		"parameter_code": param.ParameterCode,
		"prev_equation":  prevEq,
		"reason_note":    note,
	}, ip, ua)

	if s.telemetryService != nil {
		s.telemetryService.SyncParameterCalibrationCache(param)
	}
	if s.onDeviceReload != nil {
		s.onDeviceReload(param.DeviceID)
	}

	dto := s.buildParameterDTO(ctx, param, dev, stationDomain)
	return &dto, logEntry, nil
}

// RollbackCalibration restores the previous calibration state captured in a specific SensorCalibrationLog entry
func (s *CalibrationService) RollbackCalibration(
	ctx context.Context,
	logID uint,
	username, role, ip, ua string,
) (*CalibratableParameterDTO, *model.SensorCalibrationLog, error) {
	targetLog, err := s.repo.GetCalibrationLogByID(ctx, logID)
	if err != nil {
		return nil, nil, errors.New("riwayat log kalibrasi tidak ditemukan")
	}
	scope := tenantctx.FromContext(ctx)
	if scope.Authenticated && !scope.CanAccessRecord(targetLog.TenantID, targetLog.SiteID) {
		return nil, nil, errors.New("akses ditolak: log kalibrasi milik tenant atau stasiun lain")
	}

	param, dev, err := s.repo.GetParameterWithDevice(ctx, targetLog.ParameterID)
	if err != nil {
		return nil, nil, errors.New("parameter target rollback tidak ditemukan")
	}
	if err := s.checkDeviceScope(ctx, dev); err != nil {
		return nil, nil, err
	}

	prevEnabled := param.CalibrationEnabled
	prevModel := param.CalibrationModel
	prevC2 := param.CalC2
	prevC1 := param.CalC1
	prevC0 := param.CalC0
	prevEq := param.CalibrationEquation()
	prevPts := param.CalPointsJSON
	prevNote := param.CalStandardNote

	// Determine what state to restore:
	// - If rolling back onto an APPLY entry that differs from current state, restore that APPLY entry's curve;
	// - If rolling back the currently active entry itself, restore its Prev* state!
	restoreEnabled := targetLog.CalibrationEnabledFromAction()
	restoreModel := targetLog.CalibrationModel
	restoreC2 := targetLog.C2
	restoreC1 := targetLog.C1
	restoreC0 := targetLog.C0
	restorePts := targetLog.PointsJSON
	restoreNote := fmt.Sprintf("Rollback ke log #%d (%s)", targetLog.ID, targetLog.StandardNote)
	restoreR2 := targetLog.R2
	restoreRMSE := targetLog.RMSE
	restorePtCount := targetLog.PointCount

	isSameAsCurrent := (param.CalibrationEnabled == restoreEnabled) &&
		(param.CalC2 == restoreC2) && (param.CalC1 == restoreC1) && (param.CalC0 == restoreC0)

	if isSameAsCurrent {
		restoreEnabled = targetLog.PrevCalibrationEnabled
		restoreModel = targetLog.PrevCalibrationModel
		if restoreModel == "" || restoreModel == calibration.ModelNone {
			restoreModel = calibration.ModelLinear
		}
		restoreC2 = targetLog.PrevC2
		restoreC1 = targetLog.PrevC1
		if !restoreEnabled && restoreC1 == 0 && restoreC2 == 0 && targetLog.PrevC0 == 0 {
			restoreC1 = 1.0
		}
		restoreC0 = targetLog.PrevC0
		restorePts = targetLog.PrevPointsJSON
		restoreNote = fmt.Sprintf("Rollback log #%d ke keadaan sebelumnya (%s)", targetLog.ID, targetLog.PrevEquation)
		restoreR2 = nil
		restoreRMSE = nil
		var parsedPts []calibration.CalibrationPoint
		if restorePts != "" && json.Unmarshal([]byte(restorePts), &parsedPts) == nil {
			cnt := 0
			for _, p := range parsedPts {
				if p.Active {
					cnt++
				}
			}
			restorePtCount = cnt
		} else {
			restorePtCount = 0
		}
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"calibration_enabled": restoreEnabled,
		"calibration_model":   restoreModel,
		"cal_c2":              restoreC2,
		"cal_c1":              restoreC1,
		"cal_c0":              restoreC0,
		"cal_r2":              restoreR2,
		"cal_rmse":            restoreRMSE,
		"cal_point_count":     restorePtCount,
		"cal_points_json":     restorePts,
		"cal_standard_note":   restoreNote,
		"cal_updated_at":      &now,
		"cal_updated_by":      username,
	}

	if restoreEnabled {
		updates["formula"] = ""
		updates["current_formula_value"] = nil
		param.Formula = ""
		param.CurrentFormulaValue = nil
	}

	param.CalibrationEnabled = restoreEnabled
	param.CalibrationModel = restoreModel
	param.CalC2 = restoreC2
	param.CalC1 = restoreC1
	param.CalC0 = restoreC0
	param.CalR2 = restoreR2
	param.CalRMSE = restoreRMSE
	param.CalPointCount = restorePtCount
	param.CalPointsJSON = restorePts
	param.CalStandardNote = restoreNote
	param.CalUpdatedAt = &now
	param.CalUpdatedBy = username

	if param.CurrentUncalibratedValue != nil {
		newCur := param.ApplyCalibration(*param.CurrentUncalibratedValue)
		updates["current_value"] = &newCur
		param.CurrentValue = &newCur
	}

	if err := s.repo.UpdateParameterCalibrationFields(ctx, param.ID, updates); err != nil {
		return nil, nil, fmt.Errorf("gagal melakukan rollback kalibrasi: %w", err)
	}

	stationDomain := s.resolveStationDomain()
	effDomain := model.InferCalibrationDomain(param.CalibrationDomain, param.ParameterCode, param.ParameterName, param.Unit, stationDomain)
	newEq := param.CalibrationEquation()

	rbLog := &model.SensorCalibrationLog{
		TenantID:               dev.TenantID,
		SiteID:                 dev.SiteID,
		DeviceID:               param.DeviceID,
		DeviceCode:             dev.DeviceCode,
		DeviceName:             dev.DeviceName,
		ParameterID:            param.ID,
		ParameterCode:          param.ParameterCode,
		ParameterName:          param.ParameterName,
		Unit:                   param.Unit,
		Domain:                 effDomain,
		Action:                 model.CalibrationActionRollback,
		CalibrationModel:       restoreModel,
		PointCount:             restorePtCount,
		TotalPoints:            restorePtCount,
		C2:                     restoreC2,
		C1:                     restoreC1,
		C0:                     restoreC0,
		Equation:               newEq,
		PrevCalibrationEnabled: prevEnabled,
		PrevCalibrationModel:   prevModel,
		PrevC2:                 prevC2,
		PrevC1:                 prevC1,
		PrevC0:                 prevC0,
		PrevEquation:           prevEq,
		PrevPointsJSON:         prevPts,
		PrevStandardNote:       prevNote,
		R2:                     restoreR2,
		MinR2:                  param.CalMinR2,
		PassedR2:               true,
		RMSE:                   restoreRMSE,
		PointsJSON:             restorePts,
		StandardNote:           restoreNote,
		PerformedBy:            username,
		PerformedRole:          role,
		IPAddress:              ip,
		CreatedAt:              now,
	}
	_ = s.repo.CreateCalibrationLog(ctx, rbLog)

	s.recordAudit(username, "ROLLBACK_SENSOR_CALIBRATION", fmt.Sprintf("parameter:%d", param.ID), map[string]interface{}{
		"parameter_code": param.ParameterCode,
		"source_log_id":  logID,
		"new_equation":   newEq,
		"prev_equation":  prevEq,
	}, ip, ua)

	if s.telemetryService != nil {
		s.telemetryService.SyncParameterCalibrationCache(param)
	}
	if s.onDeviceReload != nil {
		s.onDeviceReload(param.DeviceID)
	}

	dto := s.buildParameterDTO(ctx, param, dev, stationDomain)
	return &dto, rbLog, nil
}

func (s *CalibrationService) ListCalibrationLogs(ctx context.Context, filter repository.CalibrationLogFilter) ([]model.SensorCalibrationLog, int64, error) {
	return s.repo.ListCalibrationLogs(ctx, filter)
}

func (s *CalibrationService) recordAudit(username, action, resource string, detailObj interface{}, ip, ua string) {
	if s.systemRepo == nil {
		return
	}
	if strings.TrimSpace(username) == "" {
		username = "system"
	}
	raw, _ := json.Marshal(detailObj)
	_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
		Username:  username,
		Action:    action,
		Resource:  resource,
		Details:   string(raw),
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now().UTC(),
	})
}


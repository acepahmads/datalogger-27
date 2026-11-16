package service

import (
	"math"
	"sort"
	"strings"
	"time"

	"datalogger/internal/model"
)

// FleetSiteAlarmItem represents a compact active alarm summary attached to a Site on the Superadmin Fleet Map.
type FleetSiteAlarmItem struct {
	ID            uint      `json:"id"`
	AlarmCode     string    `json:"alarm_code"`
	Severity      string    `json:"severity"`
	Status        string    `json:"status"`
	DeviceID      *uint     `json:"device_id,omitempty"`
	DeviceCode    string    `json:"device_code,omitempty"`
	ParameterID   *uint     `json:"parameter_id,omitempty"`
	ParameterCode string    `json:"parameter_code,omitempty"`
	ParameterName string    `json:"parameter_name,omitempty"`
	Message       string    `json:"message"`
	TriggeredAt   time.Time `json:"triggered_at"`
}

// FleetSiteDeviceBrief represents a compact per-device summary inside a Site on the Superadmin Fleet Map.
type FleetSiteDeviceBrief struct {
	ID               uint       `json:"id"`
	DeviceCode       string     `json:"device_code"`
	DeviceName       string     `json:"device_name"`
	Protocol         string     `json:"protocol"`
	ConnectionStatus string     `json:"connection_status"`
	Enabled          bool       `json:"enabled"`
	LastDataAt       *time.Time `json:"last_data_at"`
	LastSeenAt       *time.Time `json:"last_seen_at"`
}

// FleetEnvParamMetric represents a compact per-pollutant (ISPU) or per-water-parameter (WQMS) breakdown metric.
type FleetEnvParamMetric struct {
	Code          string  `json:"code"`
	Concentration float64 `json:"concentration"`
	Unit          string  `json:"unit"`
	SubIndex      float64 `json:"sub_index"`
	Category      string  `json:"category,omitempty"`
}

// FleetSiteISPUSummary holds the latest Air Pollution Standard Index (Permen LHK No. 14/2020) for an AQMS station.
type FleetSiteISPUSummary struct {
	Available           bool                  `json:"available"`
	Score               int                   `json:"score"`
	RawScore            float64               `json:"raw_score"`
	Category            string                `json:"category"`
	CategoryLabel       string                `json:"category_label"`
	DominantPollutant   string                `json:"dominant_pollutant"`
	DominantConc        float64               `json:"dominant_concentration"`
	DominantUnit        string                `json:"dominant_unit"`
	CompletenessPercent float64               `json:"completeness_percent"`
	CalculationMode     string                `json:"calculation_mode"`
	CalculatedAt        *time.Time            `json:"calculated_at,omitempty"`
	Parameters          []FleetEnvParamMetric `json:"parameters,omitempty"`
}

// FleetSiteWQMSSummary holds the latest Pollution Index (IP - Kepmen LH 115/2003) & Water Quality Index (IKA - Permen LHK 27/2021) for a WQMS/ONLIMO station.
type FleetSiteWQMSSummary struct {
	Available         bool                  `json:"available"`
	WaterClass        string                `json:"water_class"`
	IPScore           float64               `json:"ip_score"`
	IPCategory        string                `json:"ip_category"`
	IPCategoryLabel   string                `json:"ip_category_label"`
	DominantParameter string                `json:"dominant_parameter"`
	DominantConc      float64               `json:"dominant_concentration"`
	DominantUnit      string                `json:"dominant_unit"`
	IKAScore          float64               `json:"ika_score"`
	IKACategory       string                `json:"ika_category"`
	IKACategoryLabel  string                `json:"ika_category_label"`
	CalculationMode   string                `json:"calculation_mode"`
	CalculatedAt      *time.Time            `json:"calculated_at,omitempty"`
	Parameters        []FleetEnvParamMetric `json:"parameters,omitempty"`
}

// FleetSiteItem represents a single Tenant Station / Site with coordinates, freshness, devices, alarms, and environmental indices.
type FleetSiteItem struct {
	SiteID              uint                   `json:"site_id"`
	SiteCode            string                 `json:"site_code"`
	SiteName            string                 `json:"site_name"`
	TenantID            uint                   `json:"tenant_id"`
	TenantCode          string                 `json:"tenant_code"`
	TenantName          string                 `json:"tenant_name"`
	TenantStatus        string                 `json:"tenant_status"`
	CustomerName        string                 `json:"customer_name"`
	ProjectYear         int                    `json:"project_year"`
	MonitoringDomain    string                 `json:"monitoring_domain"`
	Enabled             bool                   `json:"enabled"`
	Address             string                 `json:"address"`
	Timezone            string                 `json:"timezone"`
	HasCoordinates      bool                   `json:"has_coordinates"`
	Latitude            *float64               `json:"latitude"`
	Longitude           *float64               `json:"longitude"`
	LocationSource      string                 `json:"location_source"`
	DeviceCount         int                    `json:"device_count"`
	EnabledDeviceCount  int                    `json:"enabled_device_count"`
	OnlineDeviceCount   int                    `json:"online_device_count"`
	OfflineDeviceCount  int                    `json:"offline_device_count"`
	PrimaryProtocol     string                 `json:"primary_protocol"`
	EffectiveTimeoutSec int                    `json:"effective_timeout_seconds"`
	LastDataAt          *time.Time             `json:"last_data_at"`
	LastSeenAt          *time.Time             `json:"last_seen_at"`
	LastTransmissionAt  *time.Time             `json:"last_transmission_at"`
	DataAgeSeconds      *int64                 `json:"data_age_seconds"`
	DeliveryStatus      string                 `json:"delivery_status"` // ONLINE | DELAYED | OFFLINE | INACTIVE
	SeverityState       string                 `json:"severity_state"`  // CRITICAL | WARNING | NORMAL | INACTIVE
	CriticalAlarmCount  int                    `json:"critical_alarm_count"`
	WarningAlarmCount   int                    `json:"warning_alarm_count"`
	TotalActiveAlarms   int                    `json:"total_active_alarms"`
	ActiveAlarms        []FleetSiteAlarmItem   `json:"active_alarms"`
	Devices             []FleetSiteDeviceBrief `json:"devices"`
	ISPU                *FleetSiteISPUSummary  `json:"ispu,omitempty"`
	WQMS                *FleetSiteWQMSSummary  `json:"wqms,omitempty"`
}

// FleetTenantOption represents a lightweight tenant summary for filter dropdowns on the Fleet Map.
type FleetTenantOption struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	ProjectYear int    `json:"project_year"`
	SiteCount   int    `json:"site_count"`
	OnlineSites int    `json:"online_sites"`
	AlarmSites  int    `json:"alarm_sites"`
}

// FleetOverviewSummary aggregates top-level KPI counters for the 1-row Superadmin Command Ribbon.
type FleetOverviewSummary struct {
	TotalTenants       int `json:"total_tenants"`
	ActiveTenants      int `json:"active_tenants"`
	TotalSites         int `json:"total_sites"`
	MappedSites        int `json:"mapped_sites"`
	NoGPSSites         int `json:"no_gps_sites"`
	OnlineSites        int `json:"online_sites"`
	DelayedSites       int `json:"delayed_sites"`
	OfflineSites       int `json:"offline_sites"`
	InactiveSites      int `json:"inactive_sites"`
	SitesWithAlarms    int `json:"sites_with_alarms"`
	TotalActiveAlarms  int `json:"total_active_alarms"`
	CriticalAlarms     int `json:"critical_alarms"`
	WarningAlarms      int `json:"warning_alarms"`
	TotalDevices       int `json:"total_devices"`
	OnlineDevices      int `json:"online_devices"`
}

// FleetOverviewResponse is the payload returned by GET /api/tenants/fleet-overview.
type FleetOverviewResponse struct {
	GeneratedAt  time.Time            `json:"generated_at"`
	Summary      FleetOverviewSummary `json:"summary"`
	ProjectYears []int                `json:"project_years"`
	Tenants      []FleetTenantOption  `json:"tenants"`
	Sites        []FleetSiteItem      `json:"sites"`
}

// isValidFleetCoordinate returns true iff lat/lng are finite, within valid WGS84 bounds, and NOT (0, 0) Null-Island.
func isValidFleetCoordinate(lat, lng *float64) bool {
	if lat == nil || lng == nil {
		return false
	}
	la, ln := *lat, *lng
	if math.IsNaN(la) || math.IsNaN(ln) || math.IsInf(la, 0) || math.IsInf(ln, 0) {
		return false
	}
	if la < -90 || la > 90 || ln < -180 || ln > 180 {
		return false
	}
	if math.Abs(la) < 0.0001 && math.Abs(ln) < 0.0001 {
		return false
	}
	return true
}

func isDeviceAdminDisabled(dev *model.Device) bool {
	if dev == nil {
		return true
	}
	if !dev.Enabled {
		return true
	}
	st := strings.ToUpper(strings.TrimSpace(string(dev.Status)))
	return st == "DISABLED" || st == "INACTIVE"
}

func resolveFleetDeviceProtocol(dev *model.Device) string {
	if dev == nil {
		return ""
	}
	if dev.Connection != nil && strings.TrimSpace(string(dev.Connection.Protocol)) != "" {
		return strings.ToUpper(strings.TrimSpace(string(dev.Connection.Protocol)))
	}
	return strings.ToUpper(strings.TrimSpace(dev.DeviceType))
}

func resolveFleetDeviceTimeoutSeconds(dev *model.Device) int {
	if dev == nil {
		return 300
	}
	proto := resolveFleetDeviceProtocol(dev)
	isPassive := proto == string(model.ProtocolEdgePush) || proto == string(model.ProtocolMQTT) || proto == "HTTP" || proto == "HTTPS"
	timeoutSec := 180
	if isPassive {
		timeoutSec = 300
		if dev.Connection != nil && dev.Connection.Timeout >= 10 && dev.Connection.Timeout != 1000 {
			if dev.Connection.Timeout > 86400 {
				timeoutSec = dev.Connection.Timeout / 1000
			} else {
				timeoutSec = dev.Connection.Timeout
			}
		}
	} else if dev.Connection != nil {
		pollMs := dev.Connection.PollingInterval
		if pollMs <= 0 {
			pollMs = dev.Connection.PollIntervalMs
		}
		if pollMs > 0 {
			pollSec := (pollMs + 999) / 1000
			if pollSec*3 > timeoutSec {
				timeoutSec = pollSec * 3
			}
		}
	}
	if timeoutSec < 30 {
		timeoutSec = 30
	}
	return timeoutSec
}

// GetFleetOverview returns the cross-tenant global station map overview for SUPERADMIN.
func (s *TenantService) GetFleetOverview() (*FleetOverviewResponse, error) {
	now := time.Now().UTC()
	out := &FleetOverviewResponse{
		GeneratedAt:  now,
		ProjectYears: make([]int, 0),
		Tenants:      make([]FleetTenantOption, 0),
		Sites:        make([]FleetSiteItem, 0),
	}
	if s == nil || s.repo == nil || s.repo.GetDB() == nil {
		return out, nil
	}
	db := s.repo.GetDB()

	var tenants []model.Tenant
	if err := db.Order("id ASC").Find(&tenants).Error; err != nil {
		return nil, err
	}
	tenantMap := make(map[uint]model.Tenant, len(tenants))
	for _, t := range tenants {
		tenantMap[t.ID] = t
	}

	var sites []model.Site
	if err := db.Order("tenant_id ASC, id ASC").Find(&sites).Error; err != nil {
		return nil, err
	}
	defaultSiteByTenant := make(map[uint]uint, len(tenants))
	siteExists := make(map[uint]bool, len(sites))
	for _, st := range sites {
		siteExists[st.ID] = true
		if _, exists := defaultSiteByTenant[st.TenantID]; !exists {
			defaultSiteByTenant[st.TenantID] = st.ID
		}
	}

	var devices []model.Device
	if err := db.Preload("Connection").Preload("Parameters", "deleted_at IS NULL").Order("id ASC").Find(&devices).Error; err != nil {
		return nil, err
	}
	deviceByID := make(map[uint]*model.Device, len(devices))
	devicesBySite := make(map[uint][]*model.Device, len(sites))
	for i := range devices {
		dev := &devices[i]
		deviceByID[dev.ID] = dev
		sid := dev.SiteID
		if !siteExists[sid] {
			if fallbackSid, ok := defaultSiteByTenant[dev.TenantID]; ok {
				sid = fallbackSid
			}
		}
		devicesBySite[sid] = append(devicesBySite[sid], dev)
	}

	// Fetch operational active / acknowledged alarms across all tenants
	var rawAlarms []model.Alarm
	_ = db.Preload("Device").Preload("Parameter").
		Where("status IN ?", []string{string(model.AlarmActive), string(model.AlarmAcked)}).
		Order("triggered_at DESC").
		Find(&rawAlarms).Error

	alarmsBySite := make(map[uint][]model.Alarm, len(sites))
	for _, al := range rawAlarms {
		// Mirror frontend operationalActiveAlarms hygiene: skip alarms on deleted/disabled devices or disabled parameters
		if al.DeviceID != nil {
			dev, ok := deviceByID[*al.DeviceID]
			if !ok || isDeviceAdminDisabled(dev) {
				continue
			}
		} else if al.Device != nil && isDeviceAdminDisabled(al.Device) {
			continue
		}
		if al.Parameter != nil && !al.Parameter.Enabled {
			continue
		}

		sid := al.SiteID
		if al.DeviceID != nil {
			if dev, ok := deviceByID[*al.DeviceID]; ok && siteExists[dev.SiteID] {
				sid = dev.SiteID
			}
		}
		if !siteExists[sid] {
			if fallbackSid, ok := defaultSiteByTenant[al.TenantID]; ok {
				sid = fallbackSid
			}
		}
		if siteExists[sid] {
			alarmsBySite[sid] = append(alarmsBySite[sid], al)
		}
	}

	resolveResultSiteID := func(tenantID, siteID, deviceID uint) uint {
		if siteID > 0 && siteExists[siteID] {
			return siteID
		}
		if deviceID > 0 {
			if dev, ok := deviceByID[deviceID]; ok && siteExists[dev.SiteID] {
				return dev.SiteID
			}
		}
		if fallbackSid, ok := defaultSiteByTenant[tenantID]; ok && siteExists[fallbackSid] {
			return fallbackSid
		}
		return 0
	}

	// Bulk-load latest ISPU & WQMS/IKA environmental index evaluations across all sites (zero N+1 queries)
	latestISPUBySite := make(map[uint]model.ISPUCalculationResult)
	if db.Migrator().HasTable(&model.ISPUCalculationResult{}) {
		var ispuRows []model.ISPUCalculationResult
		_ = db.Preload("ParameterResults").
			Where("calculated_at >= ? AND overall_category <> ?", now.Add(-48*time.Hour), string(model.CategoryUnavailable)).
			Order("calculated_at DESC, id DESC").
			Limit(500).
			Find(&ispuRows).Error
		for _, r := range ispuRows {
			sid := resolveResultSiteID(r.TenantID, r.SiteID, r.DeviceID)
			if sid > 0 {
				if _, exists := latestISPUBySite[sid]; !exists {
					latestISPUBySite[sid] = r
				}
			}
		}
	}

	waterClassBySite := make(map[uint]model.WQMSWaterClass)
	if db.Migrator().HasTable(&model.WQMSStationConfig{}) {
		var wqmsConfigs []model.WQMSStationConfig
		_ = db.Order("id DESC").Find(&wqmsConfigs).Error
		for _, cfg := range wqmsConfigs {
			sid := resolveResultSiteID(cfg.TenantID, cfg.SiteID, cfg.DeviceID)
			if sid > 0 {
				if _, exists := waterClassBySite[sid]; !exists && cfg.WaterClass != "" {
					waterClassBySite[sid] = NormalizeWaterClass(string(cfg.WaterClass))
				}
			}
		}
	}

	latestWQMSBySite := make(map[uint]model.WQMSCalculationResult)
	ikaCategoriesBySite := make(map[uint][]model.WQMSIPCategory)
	if db.Migrator().HasTable(&model.WQMSCalculationResult{}) {
		var wqmsRows []model.WQMSCalculationResult
		_ = db.Preload("ParameterResults").
			Where("calculated_at >= ? AND overall_category <> ?", now.Add(-48*time.Hour), string(model.WQMSCategoryUnavailable)).
			Order("calculated_at DESC, id DESC").
			Limit(500).
			Find(&wqmsRows).Error
		for _, r := range wqmsRows {
			sid := resolveResultSiteID(r.TenantID, r.SiteID, r.DeviceID)
			if sid > 0 {
				if _, exists := latestWQMSBySite[sid]; !exists {
					latestWQMSBySite[sid] = r
				}
			}
		}

		type wqmsCatRow struct {
			TenantID        uint                 `gorm:"column:tenant_id"`
			SiteID          uint                 `gorm:"column:site_id"`
			DeviceID        uint                 `gorm:"column:device_id"`
			OverallCategory model.WQMSIPCategory `gorm:"column:overall_category"`
		}
		var histRows []wqmsCatRow
		_ = db.Model(&model.WQMSCalculationResult{}).
			Select("tenant_id, site_id, device_id, overall_category").
			Where("calculated_at >= ? AND overall_category <> ?", now.Add(-30*24*time.Hour), string(model.WQMSCategoryUnavailable)).
			Order("calculated_at DESC").
			Limit(5000).
			Scan(&histRows).Error
		for _, hr := range histRows {
			sid := resolveResultSiteID(hr.TenantID, hr.SiteID, hr.DeviceID)
			if sid > 0 && len(ikaCategoriesBySite[sid]) < 720 {
				ikaCategoriesBySite[sid] = append(ikaCategoriesBySite[sid], hr.OverallCategory)
			}
		}
	}

	ispuCalc := NewISPUCalculator()
	wqmsCalc := NewWQMSCalculator()

	yearSet := make(map[int]struct{})
	tenantStats := make(map[uint]*FleetTenantOption, len(tenants))
	for _, t := range tenants {
		opt := FleetTenantOption{
			ID:          t.ID,
			Code:        t.Code,
			Name:        t.Name,
			Status:      t.Status,
			ProjectYear: t.ProjectYear,
		}
		if t.ProjectYear > 0 {
			yearSet[t.ProjectYear] = struct{}{}
		}
		out.Tenants = append(out.Tenants, opt)
		if !strings.EqualFold(t.Status, model.TenantStatusSuspended) {
			out.Summary.ActiveTenants++
		}
	}
	out.Summary.TotalTenants = len(tenants)
	for i := range out.Tenants {
		tenantStats[out.Tenants[i].ID] = &out.Tenants[i]
	}

	for _, st := range sites {
		ten, hasTenant := tenantMap[st.TenantID]
		if !hasTenant {
			continue
		}

		// 1. Resolve coordinates (Site GPS -> fallback to any valid Device GPS -> No-GPS)
		var latPtr, lngPtr *float64
		hasCoords := false
		locSource := strings.ToUpper(strings.TrimSpace(st.LocationSource))
		if isValidFleetCoordinate(st.Latitude, st.Longitude) {
			la, ln := *st.Latitude, *st.Longitude
			latPtr, lngPtr = &la, &ln
			hasCoords = true
			if locSource == "" || locSource == model.LocationSourceUnconfigured {
				locSource = model.LocationSourceManual
			}
		} else {
			for _, dev := range devicesBySite[st.ID] {
				if isValidFleetCoordinate(dev.Latitude, dev.Longitude) {
					la, ln := *dev.Latitude, *dev.Longitude
					latPtr, lngPtr = &la, &ln
					hasCoords = true
					locSource = "DEVICE_FALLBACK"
					break
				}
			}
			if !hasCoords {
				locSource = model.LocationSourceUnconfigured
			}
		}

		// 2. Aggregate Devices & Freshness
		siteDevs := devicesBySite[st.ID]
		devBriefs := make([]FleetSiteDeviceBrief, 0, len(siteDevs))
		totalDevs := len(siteDevs)
		enabledDevs := 0
		onlineDevs := 0
		explicitOfflineDevs := 0
		var latestDataAt *time.Time
		var latestSeenAt *time.Time
		primaryProto := ""
		protoFreq := make(map[string]int)
		effectiveTimeoutSec := 0

		for _, dev := range siteDevs {
			adminDisabled := isDeviceAdminDisabled(dev)
			proto := resolveFleetDeviceProtocol(dev)
			connSt := strings.ToUpper(strings.TrimSpace(string(dev.ConnectionStatus)))
			if connSt == "" {
				connSt = string(model.DeviceConnUnknown)
			}

			var devSeen *time.Time
			if dev.LastSeenAt != nil && !dev.LastSeenAt.IsZero() {
				devSeen = dev.LastSeenAt
			} else if dev.LastCommunication != nil && !dev.LastCommunication.IsZero() {
				devSeen = dev.LastCommunication
			}
			var devData *time.Time
			if dev.LastDataAt != nil && !dev.LastDataAt.IsZero() {
				devData = dev.LastDataAt
			}

			devBriefs = append(devBriefs, FleetSiteDeviceBrief{
				ID:               dev.ID,
				DeviceCode:       dev.DeviceCode,
				DeviceName:       dev.DeviceName,
				Protocol:         proto,
				ConnectionStatus: connSt,
				Enabled:          !adminDisabled,
				LastDataAt:       devData,
				LastSeenAt:       devSeen,
			})

			if adminDisabled {
				continue
			}
			enabledDevs++
			if proto != "" {
				protoFreq[proto]++
				if primaryProto == "" || protoFreq[proto] > protoFreq[primaryProto] {
					primaryProto = proto
				}
			}
			tSec := resolveFleetDeviceTimeoutSeconds(dev)
			if effectiveTimeoutSec == 0 || tSec > effectiveTimeoutSec {
				effectiveTimeoutSec = tSec
			}

			if devData != nil && (latestDataAt == nil || devData.After(*latestDataAt)) {
				tCopy := devData.UTC()
				latestDataAt = &tCopy
			}
			if devSeen != nil && (latestSeenAt == nil || devSeen.After(*latestSeenAt)) {
				tCopy := devSeen.UTC()
				latestSeenAt = &tCopy
			}

			// Determine per-device online state factoring both ConnectionStatus and data timeout
			var devLast *time.Time
			if devData != nil {
				devLast = devData
			} else if devSeen != nil {
				devLast = devSeen
			}
			offlineThreshold := time.Duration(maxInt(tSec*3, 1800)) * time.Second
			isTimedOutHard := devLast != nil && now.Sub(*devLast) > offlineThreshold
			if (connSt == "OFFLINE" || connSt == "ERROR") || isTimedOutHard {
				explicitOfflineDevs++
			} else if connSt == "ONLINE" || (devLast != nil && now.Sub(*devLast) <= time.Duration(tSec)*time.Second) {
				onlineDevs++
			}
		}

		if effectiveTimeoutSec == 0 {
			effectiveTimeoutSec = 300
		}
		offlineDevs := enabledDevs - onlineDevs
		if offlineDevs < 0 {
			offlineDevs = 0
		}

		var lastTrans *time.Time
		if latestDataAt != nil {
			lastTrans = latestDataAt
		} else if latestSeenAt != nil {
			lastTrans = latestSeenAt
		}

		var ageSecPtr *int64
		if lastTrans != nil {
			age := int64(now.Sub(*lastTrans).Seconds())
			if age < 0 {
				age = 0
			}
			ageSecPtr = &age
		}

		// 3. Compute Delivery Status: ONLINE | DELAYED | OFFLINE | INACTIVE
		tenantSuspended := strings.EqualFold(ten.Status, model.TenantStatusSuspended)
		deliveryStatus := "ONLINE"
		if tenantSuspended || !st.Enabled || enabledDevs == 0 {
			deliveryStatus = "INACTIVE"
		} else if lastTrans == nil {
			if onlineDevs > 0 {
				deliveryStatus = "DELAYED"
			} else {
				deliveryStatus = "OFFLINE"
			}
		} else {
			ageSec := *ageSecPtr
			delayedThreshold := int64(effectiveTimeoutSec)
			offlineThreshold := int64(maxInt(effectiveTimeoutSec*3, 1800))

			if onlineDevs == 0 && explicitOfflineDevs == enabledDevs {
				deliveryStatus = "OFFLINE"
			} else if ageSec > offlineThreshold {
				deliveryStatus = "OFFLINE"
			} else if ageSec > delayedThreshold || (onlineDevs > 0 && onlineDevs < enabledDevs) {
				deliveryStatus = "DELAYED"
			} else {
				deliveryStatus = "ONLINE"
			}
		}

		// 4. Aggregate Active Alarms for this Site
		siteAlarms := alarmsBySite[st.ID]
		// Sort site alarms: CRITICAL first, then newest TriggeredAt
		sort.SliceStable(siteAlarms, func(i, j int) bool {
			si := strings.EqualFold(string(siteAlarms[i].Severity), string(model.SeverityCritical))
			sj := strings.EqualFold(string(siteAlarms[j].Severity), string(model.SeverityCritical))
			if si != sj {
				return si
			}
			return siteAlarms[i].TriggeredAt.After(siteAlarms[j].TriggeredAt)
		})

		critCount := 0
		warnCount := 0
		alarmItems := make([]FleetSiteAlarmItem, 0, minInt(len(siteAlarms), 5))
		for idx, al := range siteAlarms {
			sev := strings.ToUpper(strings.TrimSpace(string(al.Severity)))
			if sev == string(model.SeverityCritical) {
				critCount++
			} else {
				warnCount++
			}
			if idx < 5 {
				devCode := ""
				if al.DeviceID != nil {
					if d, ok := deviceByID[*al.DeviceID]; ok {
						devCode = d.DeviceCode
					}
				} else if al.Device != nil {
					devCode = al.Device.DeviceCode
				}
				pCode, pName := "", ""
				if al.Parameter != nil {
					pCode = al.Parameter.ParameterCode
					pName = al.Parameter.ParameterName
				}
				alarmItems = append(alarmItems, FleetSiteAlarmItem{
					ID:            al.ID,
					AlarmCode:     al.AlarmCode,
					Severity:      sev,
					Status:        string(al.Status),
					DeviceID:      al.DeviceID,
					DeviceCode:    devCode,
					ParameterID:   al.ParameterID,
					ParameterCode: pCode,
					ParameterName: pName,
					Message:       al.Message,
					TriggeredAt:   al.TriggeredAt,
				})
			}
		}
		totalSiteAlarms := critCount + warnCount

		// 5. Compute Overall Severity State for Pin Color & Priority Sorting
		severityState := "NORMAL"
		if deliveryStatus == "INACTIVE" {
			severityState = "INACTIVE"
		} else if critCount > 0 || deliveryStatus == "OFFLINE" {
			severityState = "CRITICAL"
		} else if totalSiteAlarms > 0 || deliveryStatus == "DELAYED" {
			severityState = "WARNING"
		}

		dom := strings.ToUpper(strings.TrimSpace(st.MonitoringDomain))
		if dom == "" {
			dom = "AQMS"
		}

		effectiveYear := st.ProjectYear
		if effectiveYear <= 0 {
			effectiveYear = ten.ProjectYear
		}
		if effectiveYear > 0 {
			yearSet[effectiveYear] = struct{}{}
		}

		var siteISPUSummary *FleetSiteISPUSummary
		var siteWQMSSummary *FleetSiteWQMSSummary
		if dom == "AQMS" {
			siteISPUSummary = buildFleetSiteISPUSummary(st.ID, siteDevs, latestISPUBySite, ispuCalc)
		} else if dom == "WQMS" || dom == "ONLIMO" {
			wc := waterClassBySite[st.ID]
			if wc == "" {
				wc = model.WaterClass2
			}
			siteWQMSSummary = buildFleetSiteWQMSSummary(st.ID, siteDevs, wc, latestWQMSBySite, ikaCategoriesBySite[st.ID], wqmsCalc)
		}

		item := FleetSiteItem{
			SiteID:              st.ID,
			SiteCode:            st.Code,
			SiteName:            st.Name,
			TenantID:            ten.ID,
			TenantCode:          ten.Code,
			TenantName:          ten.Name,
			TenantStatus:        ten.Status,
			CustomerName:        ten.CustomerName,
			ProjectYear:         effectiveYear,
			MonitoringDomain:    dom,
			Enabled:             st.Enabled,
			Address:             st.Address,
			Timezone:            st.Timezone,
			HasCoordinates:      hasCoords,
			Latitude:            latPtr,
			Longitude:           lngPtr,
			LocationSource:      locSource,
			DeviceCount:         totalDevs,
			EnabledDeviceCount:  enabledDevs,
			OnlineDeviceCount:   onlineDevs,
			OfflineDeviceCount:  offlineDevs,
			PrimaryProtocol:     primaryProto,
			EffectiveTimeoutSec: effectiveTimeoutSec,
			LastDataAt:          latestDataAt,
			LastSeenAt:          latestSeenAt,
			LastTransmissionAt:  lastTrans,
			DataAgeSeconds:      ageSecPtr,
			DeliveryStatus:      deliveryStatus,
			SeverityState:       severityState,
			CriticalAlarmCount:  critCount,
			WarningAlarmCount:   warnCount,
			TotalActiveAlarms:   totalSiteAlarms,
			ActiveAlarms:        alarmItems,
			Devices:             devBriefs,
			ISPU:                siteISPUSummary,
			WQMS:                siteWQMSSummary,
		}
		out.Sites = append(out.Sites, item)

		// Update Global Summary Counters
		out.Summary.TotalSites++
		if hasCoords {
			out.Summary.MappedSites++
		} else {
			out.Summary.NoGPSSites++
		}
		switch deliveryStatus {
		case "ONLINE":
			out.Summary.OnlineSites++
		case "DELAYED":
			out.Summary.DelayedSites++
		case "OFFLINE":
			out.Summary.OfflineSites++
		default:
			out.Summary.InactiveSites++
		}
		if totalSiteAlarms > 0 {
			out.Summary.SitesWithAlarms++
			out.Summary.TotalActiveAlarms += totalSiteAlarms
			out.Summary.CriticalAlarms += critCount
			out.Summary.WarningAlarms += warnCount
		}
		out.Summary.TotalDevices += enabledDevs
		out.Summary.OnlineDevices += onlineDevs

		if tOpt, ok := tenantStats[ten.ID]; ok {
			tOpt.SiteCount++
			if deliveryStatus == "ONLINE" {
				tOpt.OnlineSites++
			}
			if totalSiteAlarms > 0 {
				tOpt.AlarmSites++
			}
		}
	}

	// Sort Sites by Urgency: CRITICAL -> WARNING -> NORMAL -> INACTIVE
	sort.SliceStable(out.Sites, func(i, j int) bool {
		a, b := out.Sites[i], out.Sites[j]
		ra, rb := fleetSeverityRank(a.SeverityState), fleetSeverityRank(b.SeverityState)
		if ra != rb {
			return ra < rb
		}
		if a.CriticalAlarmCount != b.CriticalAlarmCount {
			return a.CriticalAlarmCount > b.CriticalAlarmCount
		}
		da, dbRank := fleetDeliveryRank(a.DeliveryStatus), fleetDeliveryRank(b.DeliveryStatus)
		if da != dbRank {
			return da < dbRank
		}
		if a.TotalActiveAlarms != b.TotalActiveAlarms {
			return a.TotalActiveAlarms > b.TotalActiveAlarms
		}
		if !strings.EqualFold(a.TenantName, b.TenantName) {
			return strings.ToLower(a.TenantName) < strings.ToLower(b.TenantName)
		}
		return strings.ToLower(a.SiteName) < strings.ToLower(b.SiteName)
	})

	for y := range yearSet {
		out.ProjectYears = append(out.ProjectYears, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(out.ProjectYears)))

	return out, nil
}

func fleetSeverityRank(state string) int {
	switch state {
	case "CRITICAL":
		return 0
	case "WARNING":
		return 1
	case "NORMAL":
		return 2
	default:
		return 3
	}
}

func fleetDeliveryRank(status string) int {
	switch status {
	case "OFFLINE":
		return 0
	case "DELAYED":
		return 1
	case "ONLINE":
		return 2
	default:
		return 3
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatISPUCategoryLabel(cat model.ISPUCategory) string {
	switch cat {
	case model.CategoryBaik:
		return "Baik"
	case model.CategorySedang:
		return "Sedang"
	case model.CategoryTidakSehat:
		return "Tidak Sehat"
	case model.CategorySangatTidakSehat:
		return "Sangat Tidak Sehat"
	case model.CategoryBerbahaya:
		return "Berbahaya"
	default:
		return "Tidak Tersedia"
	}
}

func formatWQMSIPCategoryLabel(cat model.WQMSIPCategory) string {
	switch cat {
	case model.WQMSCategoryMemenuhi:
		return "Memenuhi Baku Mutu"
	case model.WQMSCategoryCemarRingan:
		return "Cemar Ringan"
	case model.WQMSCategoryCemarSedang:
		return "Cemar Sedang"
	case model.WQMSCategoryCemarBerat:
		return "Cemar Berat"
	default:
		return "Tidak Tersedia"
	}
}

func resolveLiveParamValue(p *model.Parameter) (float64, bool) {
	if p == nil || !p.Enabled {
		return 0, false
	}
	if p.CurrentFormulaValue != nil && !math.IsNaN(*p.CurrentFormulaValue) && !math.IsInf(*p.CurrentFormulaValue, 0) {
		return *p.CurrentFormulaValue, true
	}
	if p.CurrentValue != nil && !math.IsNaN(*p.CurrentValue) && !math.IsInf(*p.CurrentValue, 0) {
		return *p.CurrentValue, true
	}
	return 0, false
}

func buildFleetSiteISPUSummary(
	siteID uint,
	siteDevs []*model.Device,
	latestBySite map[uint]model.ISPUCalculationResult,
	calc *ISPUCalculator,
) *FleetSiteISPUSummary {
	if r, ok := latestBySite[siteID]; ok && r.OverallCategory != model.CategoryUnavailable {
		rawScore := r.OverallISPU
		cat := r.OverallCategory
		domPoll := r.DominantPollutant
		if r.CalculationMode == model.CalculationModeOperational && r.AggregationISPU > 0 && r.AggregationCategory != "" && r.AggregationCategory != model.CategoryUnavailable {
			rawScore = r.AggregationISPU
			cat = r.AggregationCategory
			if r.AggregationDominant != "" {
				domPoll = r.AggregationDominant
			}
		}
		calcAt := r.CalculatedAt.UTC()
		metrics := make([]FleetEnvParamMetric, 0, len(r.ParameterResults))
		var domConc float64
		domUnit := "µg/m³"
		for _, pr := range r.ParameterResults {
			subIdx := pr.CalculatedISPU
			conc := pr.NormalizedConcentration
			pCat := pr.Category
			if r.CalculationMode == model.CalculationModeOperational && pr.AggregationISPU > 0 {
				subIdx = pr.AggregationISPU
				conc = pr.AggregationConcentration
				if pr.AggregationCategory != "" && pr.AggregationCategory != model.CategoryUnavailable {
					pCat = pr.AggregationCategory
				}
			}
			u := pr.NormalizedUnit
			if u == "" {
				u = "µg/m³"
			}
			if pr.PollutantCode == domPoll {
				domConc = roundTo(conc, 2)
				domUnit = u
			}
			metrics = append(metrics, FleetEnvParamMetric{
				Code:          string(pr.PollutantCode),
				Concentration: roundTo(conc, 2),
				Unit:          u,
				SubIndex:      math.Round(subIdx),
				Category:      string(pCat),
			})
		}
		sort.SliceStable(metrics, func(i, j int) bool {
			return metrics[i].SubIndex > metrics[j].SubIndex
		})
		return &FleetSiteISPUSummary{
			Available:           true,
			Score:               int(math.Round(rawScore)),
			RawScore:            roundTo(rawScore, 1),
			Category:            string(cat),
			CategoryLabel:       formatISPUCategoryLabel(cat),
			DominantPollutant:   string(domPoll),
			DominantConc:        domConc,
			DominantUnit:        domUnit,
			CompletenessPercent: roundTo(r.CompletenessPercent, 1),
			CalculationMode:     string(r.CalculationMode),
			CalculatedAt:        &calcAt,
			Parameters:          metrics,
		}
	}

	// Instant live-sensor fallback for AQMS stations with active telemetry parameters
	if calc == nil {
		return &FleetSiteISPUSummary{Available: false, Category: string(model.CategoryUnavailable), CategoryLabel: "Tidak Tersedia"}
	}
	bestByPollutant := make(map[model.ISPUParameter]ParameterCalculationOutput)
	var latestParamTime *time.Time
	for _, dev := range siteDevs {
		if isDeviceAdminDisabled(dev) {
			continue
		}
		for i := range dev.Parameters {
			p := &dev.Parameters[i]
			val, ok := resolveLiveParamValue(p)
			if !ok || val < 0 {
				continue
			}
			pCode := p.ParameterCode
			if pCode == "" {
				pCode = p.Code
			}
			pName := p.ParameterName
			if pName == "" {
				pName = p.Name
			}
			poll, unit, factor := DetectOfficialPollutant(pCode, pName, p.Unit)
			if poll == "" {
				continue
			}
			eval := calc.CalculateParameter(ParameterCalculationInput{
				PollutantCode:    poll,
				ParameterID:      p.ID,
				RawConcentration: val,
				SourceUnit:       unit,
				ConversionFactor: factor,
				ValidSamples:     1,
				ExpectedSamples:  1,
				MinCompleteness:  75.0,
			})
			if !eval.IsEligible || eval.ErrorMessage != "" {
				continue
			}
			if prev, exists := bestByPollutant[poll]; !exists || eval.CalculatedISPU > prev.CalculatedISPU {
				bestByPollutant[poll] = eval
			}
			if p.LastUpdated != nil && (latestParamTime == nil || p.LastUpdated.After(*latestParamTime)) {
				tCopy := p.LastUpdated.UTC()
				latestParamTime = &tCopy
			}
		}
	}

	if len(bestByPollutant) == 0 {
		return &FleetSiteISPUSummary{Available: false, Category: string(model.CategoryUnavailable), CategoryLabel: "Tidak Tersedia"}
	}

	paramOuts := make([]ParameterCalculationOutput, 0, len(bestByPollutant))
	for _, po := range bestByPollutant {
		paramOuts = append(paramOuts, po)
	}
	overall := calc.CalculateOverall(paramOuts, 75.0)
	if overall.Category == model.CategoryUnavailable {
		return &FleetSiteISPUSummary{Available: false, Category: string(model.CategoryUnavailable), CategoryLabel: "Tidak Tersedia"}
	}

	metrics := make([]FleetEnvParamMetric, 0, len(paramOuts))
	var domConc float64
	domUnit := "µg/m³"
	for _, po := range paramOuts {
		if po.PollutantCode == overall.DominantPollutant {
			domConc = roundTo(po.NormalizedConcentration, 2)
			if po.NormalizedUnit != "" {
				domUnit = po.NormalizedUnit
			}
		}
		metrics = append(metrics, FleetEnvParamMetric{
			Code:          string(po.PollutantCode),
			Concentration: roundTo(po.NormalizedConcentration, 2),
			Unit:          po.NormalizedUnit,
			SubIndex:      float64(po.DisplayISPU),
			Category:      string(po.Category),
		})
	}
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].SubIndex > metrics[j].SubIndex
	})

	return &FleetSiteISPUSummary{
		Available:           true,
		Score:               overall.DisplayISPU,
		RawScore:            roundTo(overall.OverallISPU, 1),
		Category:            string(overall.Category),
		CategoryLabel:       formatISPUCategoryLabel(overall.Category),
		DominantPollutant:   string(overall.DominantPollutant),
		DominantConc:        domConc,
		DominantUnit:        domUnit,
		CompletenessPercent: roundTo(overall.CompletenessPercent, 1),
		CalculationMode:     "LIVE_SENSOR",
		CalculatedAt:        latestParamTime,
		Parameters:          metrics,
	}
}

func buildFleetSiteWQMSSummary(
	siteID uint,
	siteDevs []*model.Device,
	waterClass model.WQMSWaterClass,
	latestBySite map[uint]model.WQMSCalculationResult,
	histCats []model.WQMSIPCategory,
	calc *WQMSCalculator,
) *FleetSiteWQMSSummary {
	wc := NormalizeWaterClass(string(waterClass))
	if r, ok := latestBySite[siteID]; ok && r.OverallCategory != model.WQMSCategoryUnavailable {
		if r.WaterClass != "" {
			wc = NormalizeWaterClass(string(r.WaterClass))
		}
		cats := histCats
		if len(cats) == 0 {
			cats = []model.WQMSIPCategory{r.OverallCategory}
		}
		var ika model.WQMSIKASummary
		if calc != nil {
			ika = calc.ComputeIKASummary(cats, wc, 30, string(r.CalculationMode))
		}
		calcAt := r.CalculatedAt.UTC()
		metrics := make([]FleetEnvParamMetric, 0, len(r.ParameterResults))
		for _, pr := range r.ParameterResults {
			metrics = append(metrics, FleetEnvParamMetric{
				Code:          string(pr.ParamCode),
				Concentration: roundTo(pr.NormalizedMeasured, 2),
				Unit:          pr.NormalizedUnit,
				SubIndex:      roundTo(pr.AdjustedRatio, 2),
			})
		}
		sort.SliceStable(metrics, func(i, j int) bool {
			return metrics[i].SubIndex > metrics[j].SubIndex
		})
		return &FleetSiteWQMSSummary{
			Available:         true,
			WaterClass:        string(wc),
			IPScore:           roundTo(r.OverallIP, 2),
			IPCategory:        string(r.OverallCategory),
			IPCategoryLabel:   formatWQMSIPCategoryLabel(r.OverallCategory),
			DominantParameter: string(r.DominantParameter),
			DominantConc:      roundTo(r.DominantConcentration, 2),
			DominantUnit:      r.DominantUnit,
			IKAScore:          ika.Score,
			IKACategory:       string(ika.Category),
			IKACategoryLabel:  ika.CategoryLabel,
			CalculationMode:   string(r.CalculationMode),
			CalculatedAt:      &calcAt,
			Parameters:        metrics,
		}
	}

	// Instant live-sensor fallback for WQMS / ONLIMO stations with active water parameters
	if calc == nil {
		return &FleetSiteWQMSSummary{Available: false, WaterClass: string(wc), IPCategory: string(model.WQMSCategoryUnavailable), IKACategory: string(model.IKACategoryUnavailable)}
	}
	inputsByCode := make(map[model.WQMSParamCode]WQMSParameterInput)
	var latestParamTime *time.Time
	for _, dev := range siteDevs {
		if isDeviceAdminDisabled(dev) {
			continue
		}
		for i := range dev.Parameters {
			p := &dev.Parameters[i]
			val, ok := resolveLiveParamValue(p)
			if !ok || val < 0 {
				continue
			}
			pCode := p.ParameterCode
			if pCode == "" {
				pCode = p.Code
			}
			pName := p.ParameterName
			if pName == "" {
				pName = p.Name
			}
			wParam, unit, factor := DetectOfficialWQMSParameter(pCode, pName, p.Unit)
			if wParam == "" {
				continue
			}
			pid := p.ID
			inputsByCode[wParam] = WQMSParameterInput{
				ParamCode:        wParam,
				ParameterID:      &pid,
				SourceType:       "SENSOR",
				RawMeasured:      val,
				SourceUnit:       unit,
				ConversionFactor: factor,
				ValidSamples:     1,
				ExpectedSamples:  1,
				MinCompleteness:  75.0,
			}
			if p.LastUpdated != nil && (latestParamTime == nil || p.LastUpdated.After(*latestParamTime)) {
				tCopy := p.LastUpdated.UTC()
				latestParamTime = &tCopy
			}
		}
	}

	if len(inputsByCode) == 0 {
		return &FleetSiteWQMSSummary{Available: false, WaterClass: string(wc), IPCategory: string(model.WQMSCategoryUnavailable), IKACategory: string(model.IKACategoryUnavailable)}
	}

	inputs := make([]WQMSParameterInput, 0, len(inputsByCode))
	for _, in := range inputsByCode {
		inputs = append(inputs, in)
	}
	eval := calc.EvaluateStation(inputs, wc, 8.0)
	if eval.OverallCategory == model.WQMSCategoryUnavailable {
		return &FleetSiteWQMSSummary{Available: false, WaterClass: string(wc), IPCategory: string(model.WQMSCategoryUnavailable), IKACategory: string(model.IKACategoryUnavailable)}
	}
	cats := histCats
	if len(cats) == 0 {
		cats = []model.WQMSIPCategory{eval.OverallCategory}
	}
	ika := calc.ComputeIKASummary(cats, wc, 30, "LIVE_SENSOR")

	metrics := make([]FleetEnvParamMetric, 0, len(eval.ParameterResults))
	for _, pr := range eval.ParameterResults {
		metrics = append(metrics, FleetEnvParamMetric{
			Code:          string(pr.ParamCode),
			Concentration: roundTo(pr.NormalizedMeasured, 2),
			Unit:          pr.NormalizedUnit,
			SubIndex:      roundTo(pr.AdjustedRatio, 2),
		})
	}
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].SubIndex > metrics[j].SubIndex
	})

	return &FleetSiteWQMSSummary{
		Available:         true,
		WaterClass:        string(wc),
		IPScore:           eval.OverallIP,
		IPCategory:        string(eval.OverallCategory),
		IPCategoryLabel:   formatWQMSIPCategoryLabel(eval.OverallCategory),
		DominantParameter: string(eval.DominantParameter),
		DominantConc:      roundTo(eval.DominantConcentration, 2),
		DominantUnit:      eval.DominantUnit,
		IKAScore:          ika.Score,
		IKACategory:       string(ika.Category),
		IKACategoryLabel:  ika.CategoryLabel,
		CalculationMode:   "LIVE_SENSOR",
		CalculatedAt:      latestParamTime,
		Parameters:        metrics,
	}
}

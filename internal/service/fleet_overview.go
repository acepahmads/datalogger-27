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

// FleetSiteItem represents a single Tenant Station / Site with coordinates, freshness, devices, and alarms.
type FleetSiteItem struct {
	SiteID             uint                   `json:"site_id"`
	SiteCode           string                 `json:"site_code"`
	SiteName           string                 `json:"site_name"`
	TenantID           uint                   `json:"tenant_id"`
	TenantCode         string                 `json:"tenant_code"`
	TenantName         string                 `json:"tenant_name"`
	TenantStatus       string                 `json:"tenant_status"`
	CustomerName       string                 `json:"customer_name"`
	MonitoringDomain   string                 `json:"monitoring_domain"`
	Enabled            bool                   `json:"enabled"`
	Address            string                 `json:"address"`
	Timezone           string                 `json:"timezone"`
	HasCoordinates     bool                   `json:"has_coordinates"`
	Latitude           *float64               `json:"latitude"`
	Longitude          *float64               `json:"longitude"`
	LocationSource     string                 `json:"location_source"`
	DeviceCount        int                    `json:"device_count"`
	EnabledDeviceCount int                    `json:"enabled_device_count"`
	OnlineDeviceCount  int                    `json:"online_device_count"`
	OfflineDeviceCount int                    `json:"offline_device_count"`
	PrimaryProtocol    string                 `json:"primary_protocol"`
	EffectiveTimeoutSec int                   `json:"effective_timeout_seconds"`
	LastDataAt         *time.Time             `json:"last_data_at"`
	LastSeenAt         *time.Time             `json:"last_seen_at"`
	LastTransmissionAt *time.Time             `json:"last_transmission_at"`
	DataAgeSeconds     *int64                 `json:"data_age_seconds"`
	DeliveryStatus     string                 `json:"delivery_status"` // ONLINE | DELAYED | OFFLINE | INACTIVE
	SeverityState      string                 `json:"severity_state"`  // CRITICAL | WARNING | NORMAL | INACTIVE
	CriticalAlarmCount int                    `json:"critical_alarm_count"`
	WarningAlarmCount  int                    `json:"warning_alarm_count"`
	TotalActiveAlarms  int                    `json:"total_active_alarms"`
	ActiveAlarms       []FleetSiteAlarmItem   `json:"active_alarms"`
	Devices            []FleetSiteDeviceBrief `json:"devices"`
}

// FleetTenantOption represents a lightweight tenant summary for filter dropdowns on the Fleet Map.
type FleetTenantOption struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Status      string `json:"status"`
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
	GeneratedAt time.Time            `json:"generated_at"`
	Summary     FleetOverviewSummary `json:"summary"`
	Tenants     []FleetTenantOption  `json:"tenants"`
	Sites       []FleetSiteItem      `json:"sites"`
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
		GeneratedAt: now,
		Tenants:     make([]FleetTenantOption, 0),
		Sites:       make([]FleetSiteItem, 0),
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
	if err := db.Preload("Connection").Order("id ASC").Find(&devices).Error; err != nil {
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

	tenantStats := make(map[uint]*FleetTenantOption, len(tenants))
	for _, t := range tenants {
		opt := FleetTenantOption{
			ID:     t.ID,
			Code:   t.Code,
			Name:   t.Name,
			Status: t.Status,
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

		item := FleetSiteItem{
			SiteID:              st.ID,
			SiteCode:            st.Code,
			SiteName:            st.Name,
			TenantID:            ten.ID,
			TenantCode:          ten.Code,
			TenantName:          ten.Name,
			TenantStatus:        ten.Status,
			CustomerName:        ten.CustomerName,
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

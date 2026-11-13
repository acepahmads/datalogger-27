package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	TenantStatusActive    = "ACTIVE"
	TenantStatusSuspended = "SUSPENDED"

	IngestFormatAuto       = "AUTO"
	IngestFormatStandardV1 = "STANDARD_V1"
	IngestFormatCBILegacy  = "CBI_LEGACY"
)

// NormalizeIngestFormat normalizes an inbound telemetry format mode to AUTO, STANDARD_V1, or CBI_LEGACY.
func NormalizeIngestFormat(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case IngestFormatStandardV1, "STANDARD", "V1":
		return IngestFormatStandardV1
	case IngestFormatCBILegacy, "LEGACY", "CREATE_MULTI", "CBI_MULTI":
		return IngestFormatCBILegacy
	default:
		return IngestFormatAuto
	}
}

// Tenant represents an isolated organization, company, or environmental agency on the platform.
type Tenant struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Code             string         `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name             string         `gorm:"size:128;not null" json:"name"`
	Description      string         `gorm:"size:255" json:"description,omitempty"`
	Status           string         `gorm:"size:32;default:'ACTIVE';index" json:"status"` // ACTIVE | SUSPENDED
	MaxSites         int            `gorm:"default:0" json:"max_sites"`                   // 0 = unlimited
	MaxDevices       int            `gorm:"default:0" json:"max_devices"`                 // 0 = unlimited
	ContactName      string         `gorm:"size:128" json:"contact_name,omitempty"`
	ContactEmail     string         `gorm:"size:128" json:"contact_email,omitempty"`
	ContactPhone     string         `gorm:"size:64" json:"contact_phone,omitempty"`
	CustomerName     string         `gorm:"size:255" json:"customer_name"`
	CustomerSubtitle string         `gorm:"size:255" json:"customer_subtitle"`
	CustomerLogo     string         `gorm:"type:longtext" json:"customer_logo,omitempty"`
	CustomerHasLogo  bool           `gorm:"default:false" json:"customer_has_logo"`
	SiteCount        int64          `gorm:"-" json:"site_count"`
	DeviceCount      int64          `gorm:"-" json:"device_count"`
	UserCount        int64          `gorm:"-" json:"user_count"`
	Sites            []Site         `gorm:"foreignKey:TenantID" json:"sites,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Tenant) TableName() string {
	return "tenants"
}

func (t *Tenant) BeforeSave(tx *gorm.DB) error {
	t.Code = NormalizeSlugCode(t.Code)
	t.Name = strings.TrimSpace(t.Name)
	if t.Status == "" {
		t.Status = TenantStatusActive
	}
	if strings.TrimSpace(t.CustomerName) == "" && t.Name != "" {
		t.CustomerName = t.Name
	}
	t.CustomerHasLogo = strings.TrimSpace(t.CustomerLogo) != ""
	return nil
}

// Site represents a physical monitoring station / field site under a Tenant (AQMS, SPARING, WQMS).
type Site struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	TenantID         uint           `gorm:"not null;index;uniqueIndex:idx_tenant_site_code" json:"tenant_id"`
	Tenant           *Tenant        `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Code             string         `gorm:"size:64;not null;uniqueIndex:idx_tenant_site_code" json:"code"`
	Name             string         `gorm:"size:128;not null" json:"name"`
	Description      string         `gorm:"size:255" json:"description,omitempty"`
	MonitoringDomain string         `gorm:"size:32;not null;default:'AQMS';index" json:"monitoring_domain"` // AQMS | SPARING | WQMS
	StationTypeMode  string         `gorm:"size:32;default:'MANUAL'" json:"station_type_mode"`              // MANUAL | AUTO
	IngestFormat     string         `gorm:"size:32;default:'AUTO'" json:"ingest_format"`                    // AUTO | STANDARD_V1 | CBI_LEGACY
	LegacyStationID  string         `gorm:"size:64;index" json:"legacy_station_id,omitempty"`               // Legacy Cloud station_id (e.g. "202")
	Latitude         *float64       `json:"latitude"`
	Longitude        *float64       `json:"longitude"`
	LocationSource   string         `gorm:"size:32;default:'UNCONFIGURED'" json:"location_source"`
	AccuracyMeters   *float64       `json:"accuracy_meters,omitempty"`
	Address          string         `gorm:"size:255" json:"address,omitempty"`
	IsConfigured     bool           `gorm:"default:false" json:"is_configured"`
	Timezone         string         `gorm:"size:64;default:'Asia/Jakarta'" json:"timezone"`
	Enabled          bool           `gorm:"default:true;index" json:"enabled"`
	DeviceCount      int64          `gorm:"-" json:"device_count"`
	OnlineCount      int64          `gorm:"-" json:"online_count"`
	UpdatedBy        string         `gorm:"size:128" json:"updated_by,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Site) TableName() string {
	return "sites"
}

func (s *Site) BeforeSave(tx *gorm.DB) error {
	s.Code = NormalizeSlugCode(s.Code)
	s.Name = strings.TrimSpace(s.Name)
	dom := strings.ToUpper(strings.TrimSpace(s.MonitoringDomain))
	if dom != "AQMS" && dom != "SPARING" && dom != "ONLIMO" && dom != "WQMS" && dom != "CUSTOM" && dom != "CLOUD" {
		dom = "AQMS"
	}
	s.MonitoringDomain = dom
	if strings.TrimSpace(s.StationTypeMode) == "" {
		s.StationTypeMode = "MANUAL"
	}
	s.IngestFormat = NormalizeIngestFormat(s.IngestFormat)
	s.LegacyStationID = strings.TrimSpace(s.LegacyStationID)
	if strings.TrimSpace(s.Timezone) == "" {
		s.Timezone = "Asia/Jakarta"
	}
	if strings.TrimSpace(s.LocationSource) == "" {
		s.LocationSource = LocationSourceUnconfigured
	}
	if s.Latitude != nil && s.Longitude != nil {
		s.IsConfigured = true
	}
	return nil
}

// TenantAPIKey authenticates Multi-Tenant Telemetry Push requests (Authorization: Bearer dlg_live_...).
type TenantAPIKey struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	TenantID      uint       `gorm:"not null;index" json:"tenant_id"`
	Tenant        *Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	SiteID        *uint      `gorm:"index" json:"site_id,omitempty"` // nil = key allowed to push for any site of TenantID (resolved by site_code)
	Site          *Site      `gorm:"foreignKey:SiteID" json:"site,omitempty"`
	Name          string     `gorm:"size:128;not null" json:"name"`
	KeyPrefix     string     `gorm:"size:32;index;not null" json:"key_prefix"`
	KeyHash       string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	IngestFormat  string     `gorm:"size:32;default:'AUTO'" json:"ingest_format"` // AUTO | STANDARD_V1 | CBI_LEGACY
	AutoProvision bool       `gorm:"default:false" json:"auto_provision"`
	Enabled       bool       `gorm:"default:true;index" json:"enabled"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP    string     `gorm:"size:64" json:"last_used_ip,omitempty"`
	TotalIngested int64      `gorm:"default:0" json:"total_ingested"`
	CreatedBy     string     `gorm:"size:64" json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (TenantAPIKey) TableName() string {
	return "tenant_api_keys"
}

func (k *TenantAPIKey) BeforeSave(tx *gorm.DB) error {
	k.IngestFormat = NormalizeIngestFormat(k.IngestFormat)
	return nil
}

// NormalizeSlugCode converts an arbitrary code/slug into lowercase alphanumeric + hyphen format.
func NormalizeSlugCode(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if r == '-' || r == '_' || r == ' ' || r == '.' {
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}


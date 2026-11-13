package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"datalogger/internal/database"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/pkg/tenantctx"
)

type cachedIngestKey struct {
	key       *model.TenantAPIKey
	expiresAt time.Time
}

type TenantService struct {
	repo     *repository.TenantRepository
	sysRepo  *repository.SystemRepository
	keyMu    sync.RWMutex
	keyCache map[string]cachedIngestKey
}

func NewTenantService(repo *repository.TenantRepository, sysRepo *repository.SystemRepository) *TenantService {
	return &TenantService{
		repo:     repo,
		sysRepo:  sysRepo,
		keyCache: make(map[string]cachedIngestKey),
	}
}

func (s *TenantService) GetRepo() *repository.TenantRepository {
	return s.repo
}

// --- DTOs ---

type CreateTenantRequest struct {
	Code             string `json:"code"`
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description"`
	Status           string `json:"status"`
	MaxSites         int    `json:"max_sites"`
	MaxDevices       int    `json:"max_devices"`
	ContactName      string `json:"contact_name"`
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`
	CustomerName     string `json:"customer_name"`
	CustomerSubtitle string `json:"customer_subtitle"`
	CustomerLogo     string `json:"customer_logo"`
	DefaultSiteName  string `json:"default_site_name"`
	DefaultDomain    string `json:"default_domain"` // AQMS | SPARING | WQMS
}

type UpdateTenantRequest struct {
	Code             *string `json:"code"`
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	Status           *string `json:"status"`
	MaxSites         *int    `json:"max_sites"`
	MaxDevices       *int    `json:"max_devices"`
	ContactName      *string `json:"contact_name"`
	ContactEmail     *string `json:"contact_email"`
	ContactPhone     *string `json:"contact_phone"`
	CustomerName     *string `json:"customer_name"`
	CustomerSubtitle *string `json:"customer_subtitle"`
	CustomerLogo     *string `json:"customer_logo"`
	ClearCustomerLogo bool   `json:"clear_customer_logo"`
}

type CreateSiteRequest struct {
	TenantID         uint     `json:"tenant_id"`
	Code             string   `json:"code"`
	Name             string   `json:"name" binding:"required"`
	Description      string   `json:"description"`
	MonitoringDomain string   `json:"monitoring_domain"`
	StationTypeMode  string   `json:"station_type_mode"`
	IngestFormat     string   `json:"ingest_format"`
	LegacyStationID  string   `json:"legacy_station_id"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	LocationSource   string   `json:"location_source"`
	Address          string   `json:"address"`
	Timezone         string   `json:"timezone"`
	Enabled          *bool    `json:"enabled"`
}

type UpdateSiteRequest struct {
	Code             *string  `json:"code"`
	Name             *string  `json:"name"`
	Description      *string  `json:"description"`
	MonitoringDomain *string  `json:"monitoring_domain"`
	StationTypeMode  *string  `json:"station_type_mode"`
	IngestFormat     *string  `json:"ingest_format"`
	LegacyStationID  *string  `json:"legacy_station_id"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	LocationSource   *string  `json:"location_source"`
	Address          *string  `json:"address"`
	Timezone         *string  `json:"timezone"`
	Enabled          *bool    `json:"enabled"`
}

type CreateAPIKeyRequest struct {
	TenantID      uint   `json:"tenant_id"`
	SiteID        *uint  `json:"site_id"`
	IsGlobal      bool   `json:"is_global"`
	Name          string `json:"name" binding:"required"`
	IngestFormat  string `json:"ingest_format"`
	CustomToken   string `json:"custom_token"`
	AutoProvision *bool  `json:"auto_provision"`
}

type UpdateAPIKeyRequest struct {
	Enabled       *bool   `json:"enabled"`
	IngestFormat  *string `json:"ingest_format"`
	SiteID        *uint   `json:"site_id"`
	ClearSiteID   bool    `json:"clear_site_id"`
	IsGlobal      *bool   `json:"is_global"`
	Name          *string `json:"name"`
	CustomToken   *string `json:"custom_token"`
	AutoProvision *bool   `json:"auto_provision"`
}

// --- Tenant Operations ---

func (s *TenantService) ListTenants(scope tenantctx.Scope) ([]model.Tenant, error) {
	if scope.IsSuperAdmin {
		return s.repo.ListTenants(true)
	}
	tid := scope.TenantID
	if tid == 0 {
		tid = tenantctx.DefaultTenantID
	}
	t, err := s.repo.GetTenantByID(tid)
	if err != nil {
		return nil, err
	}
	if len(scope.AllowedSiteIDs) > 0 {
		filtered := make([]model.Site, 0, len(t.Sites))
		for _, st := range t.Sites {
			if scope.IsSiteAllowed(st.ID) {
				filtered = append(filtered, st)
			}
		}
		t.Sites = filtered
	}
	return []model.Tenant{*t}, nil
}

func (s *TenantService) GetTenantByID(id uint) (*model.Tenant, error) {
	return s.repo.GetTenantByID(id)
}

func (s *TenantService) CreateTenant(req CreateTenantRequest, actorUsername, ip string) (*model.Tenant, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("tenant name is required")
	}
	code := model.NormalizeSlugCode(req.Code)
	if code == "" {
		code = model.NormalizeSlugCode(name)
	}
	if code == "" {
		code = fmt.Sprintf("tenant-%d", time.Now().Unix())
	}
	if existing, _ := s.repo.GetTenantByCode(code); existing != nil {
		return nil, fmt.Errorf("tenant code '%s' already exists", code)
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != model.TenantStatusSuspended {
		status = model.TenantStatusActive
	}
	custName := strings.TrimSpace(req.CustomerName)
	if custName == "" {
		custName = name
	}
	custSub := strings.TrimSpace(req.CustomerSubtitle)
	if custSub == "" {
		custSub = strings.ToUpper(strings.TrimSpace(req.DefaultDomain))
		if custSub == "" {
			custSub = "AQMS"
		}
	}

	t := &model.Tenant{
		Code:             code,
		Name:             name,
		Description:      strings.TrimSpace(req.Description),
		Status:           status,
		MaxSites:         req.MaxSites,
		MaxDevices:       req.MaxDevices,
		ContactName:      strings.TrimSpace(req.ContactName),
		ContactEmail:     strings.TrimSpace(req.ContactEmail),
		ContactPhone:     strings.TrimSpace(req.ContactPhone),
		CustomerName:     custName,
		CustomerSubtitle: custSub,
		CustomerLogo:     req.CustomerLogo,
		CustomerHasLogo:  strings.TrimSpace(req.CustomerLogo) != "",
	}

	if err := s.repo.CreateTenant(t); err != nil {
		return nil, err
	}

	// Seed default SCADA categories & initial Site for this new Tenant
	if s.repo.GetDB() != nil {
		database.SeedTenantDefaultCategories(s.repo.GetDB(), t.ID)
	}

	defSiteName := strings.TrimSpace(req.DefaultSiteName)
	if defSiteName == "" {
		defSiteName = name + " - Stasiun Utama"
	}
	defDom := strings.ToUpper(strings.TrimSpace(req.DefaultDomain))
	if defDom != "AQMS" && defDom != "SPARING" && defDom != "ONLIMO" && defDom != "WQMS" {
		defDom = "AQMS"
	}
	_ = s.repo.CreateSite(&model.Site{
		TenantID:         t.ID,
		Code:             "main-station",
		Name:             defSiteName,
		MonitoringDomain: defDom,
		StationTypeMode:  "MANUAL",
		LocationSource:   model.LocationSourceUnconfigured,
		Timezone:         "Asia/Jakarta",
		Enabled:          true,
		UpdatedBy:        actorUsername,
	})

	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &t.ID,
			Username:  actorUsername,
			Action:    "TENANT_CREATE",
			Resource:  "TENANT",
			Details:   fmt.Sprintf("Created tenant '%s' (%s)", t.Name, t.Code),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}

	return s.repo.GetTenantByID(t.ID)
}

func (s *TenantService) UpdateTenant(id uint, req UpdateTenantRequest, scope tenantctx.Scope, ip string) (*model.Tenant, error) {
	t, err := s.repo.GetTenantByID(id)
	if err != nil || t == nil {
		return nil, errors.New("tenant not found")
	}

	if !scope.IsSuperAdmin {
		if scope.TenantID == 0 || scope.TenantID != id {
			return nil, errors.New("access denied: cannot modify another tenant")
		}
		// Tenant ADMIN can update organization profile & customer branding, NOT platform quotas or suspension status
		if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
			t.Name = strings.TrimSpace(*req.Name)
		}
		if req.Description != nil {
			t.Description = strings.TrimSpace(*req.Description)
		}
		if req.ContactName != nil {
			t.ContactName = strings.TrimSpace(*req.ContactName)
		}
		if req.ContactEmail != nil {
			t.ContactEmail = strings.TrimSpace(*req.ContactEmail)
		}
		if req.ContactPhone != nil {
			t.ContactPhone = strings.TrimSpace(*req.ContactPhone)
		}
		if req.CustomerName != nil {
			t.CustomerName = strings.TrimSpace(*req.CustomerName)
		}
		if req.CustomerSubtitle != nil {
			t.CustomerSubtitle = strings.TrimSpace(*req.CustomerSubtitle)
		}
		if req.ClearCustomerLogo {
			t.CustomerLogo = ""
			t.CustomerHasLogo = false
		} else if req.CustomerLogo != nil && strings.TrimSpace(*req.CustomerLogo) != "" {
			t.CustomerLogo = *req.CustomerLogo
			t.CustomerHasLogo = true
		}
	} else {
		if req.Code != nil && strings.TrimSpace(*req.Code) != "" {
			newCode := model.NormalizeSlugCode(*req.Code)
			if newCode != "" && !strings.EqualFold(newCode, t.Code) {
				if existing, _ := s.repo.GetTenantByCode(newCode); existing != nil && existing.ID != t.ID {
					return nil, fmt.Errorf("tenant code '%s' already exists", newCode)
				}
				t.Code = newCode
			}
		}
		if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
			t.Name = strings.TrimSpace(*req.Name)
		}
		if req.Description != nil {
			t.Description = strings.TrimSpace(*req.Description)
		}
		if req.Status != nil {
			st := strings.ToUpper(strings.TrimSpace(*req.Status))
			if id == 1 && st == model.TenantStatusSuspended {
				return nil, errors.New("cannot suspend primary default tenant")
			}
			if st == model.TenantStatusActive || st == model.TenantStatusSuspended {
				t.Status = st
			}
		}
		if req.MaxSites != nil && *req.MaxSites >= 0 {
			t.MaxSites = *req.MaxSites
		}
		if req.MaxDevices != nil && *req.MaxDevices >= 0 {
			t.MaxDevices = *req.MaxDevices
		}
		if req.ContactName != nil {
			t.ContactName = strings.TrimSpace(*req.ContactName)
		}
		if req.ContactEmail != nil {
			t.ContactEmail = strings.TrimSpace(*req.ContactEmail)
		}
		if req.ContactPhone != nil {
			t.ContactPhone = strings.TrimSpace(*req.ContactPhone)
		}
		if req.CustomerName != nil {
			t.CustomerName = strings.TrimSpace(*req.CustomerName)
		}
		if req.CustomerSubtitle != nil {
			t.CustomerSubtitle = strings.TrimSpace(*req.CustomerSubtitle)
		}
		if req.ClearCustomerLogo {
			t.CustomerLogo = ""
			t.CustomerHasLogo = false
		} else if req.CustomerLogo != nil && strings.TrimSpace(*req.CustomerLogo) != "" {
			t.CustomerLogo = *req.CustomerLogo
			t.CustomerHasLogo = true
		}
	}

	if err := s.repo.UpdateTenant(t); err != nil {
		return nil, err
	}
	s.invalidateKeyCache()

	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &t.ID,
			Username:  scope.Username,
			Action:    "TENANT_UPDATE",
			Resource:  "TENANT",
			Details:   fmt.Sprintf("Updated tenant '%s' (%s)", t.Name, t.Code),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}
	return s.repo.GetTenantByID(t.ID)
}

func (s *TenantService) DeleteTenant(id uint, actorUsername, ip string) error {
	if id == 1 {
		return errors.New("cannot delete primary default tenant (#1)")
	}
	t, err := s.repo.GetTenantByID(id)
	if err != nil || t == nil {
		return errors.New("tenant not found")
	}
	devCount, _ := s.repo.CountDevicesByTenant(id)
	if devCount > 0 {
		return fmt.Errorf("cannot delete tenant '%s' while %d active device(s) are still registered under it", t.Name, devCount)
	}
	if err := s.repo.DeleteTenant(id); err != nil {
		return err
	}
	s.invalidateKeyCache()
	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			Username:  actorUsername,
			Action:    "TENANT_DELETE",
			Resource:  "TENANT",
			Details:   fmt.Sprintf("Deleted tenant '%s' (ID=%d)", t.Name, id),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}
	return nil
}

// CheckDeviceQuota verifies that adding newDevices to tenantID does not exceed tenant.MaxDevices (0 = unlimited).
func (s *TenantService) CheckDeviceQuota(tenantID uint, additional int) error {
	if tenantID == 0 || s == nil || s.repo == nil {
		return nil
	}
	t, err := s.repo.GetTenantByID(tenantID)
	if err != nil || t == nil || t.MaxDevices <= 0 {
		return nil
	}
	curr, _ := s.repo.CountDevicesByTenant(tenantID)
	if int(curr)+additional > t.MaxDevices {
		return fmt.Errorf("kuota perangkat tenant '%s' telah penuh (%d/%d perangkat)", t.Name, curr, t.MaxDevices)
	}
	return nil
}

// --- Site / Station Operations ---

func (s *TenantService) ListSites(scope tenantctx.Scope, explicitTenantID uint) ([]model.Site, error) {
	tid := scope.TenantID
	if scope.IsSuperAdmin && explicitTenantID > 0 {
		tid = explicitTenantID
	} else if !scope.IsSuperAdmin && tid == 0 {
		tid = tenantctx.DefaultTenantID
	}
	return s.repo.ListSites(tid, scope.AllowedSiteIDs)
}

func (s *TenantService) GetSiteByID(id uint, scope tenantctx.Scope) (*model.Site, error) {
	st, err := s.repo.GetSiteByID(id)
	if err != nil || st == nil {
		return nil, errors.New("site not found")
	}
	if !scope.CanAccessRecord(st.TenantID, st.ID) {
		return nil, errors.New("access denied: site belongs to another tenant")
	}
	return st, nil
}

// ResolveEffectiveSite returns the active Site for a Scope (falling back to the Tenant's first Site if SiteID == 0).
func (s *TenantService) ResolveEffectiveSite(scope tenantctx.Scope) (*model.Site, error) {
	if scope.SiteID > 0 {
		st, err := s.repo.GetSiteByID(scope.SiteID)
		if err == nil && st != nil && scope.CanAccessRecord(st.TenantID, st.ID) {
			return st, nil
		}
	}
	tid := scope.TenantID
	if tid == 0 {
		tid = tenantctx.DefaultTenantID
	}
	if len(scope.AllowedSiteIDs) > 0 {
		if st, err := s.repo.GetSiteByID(scope.AllowedSiteIDs[0]); err == nil && st != nil && st.TenantID == tid {
			return st, nil
		}
	}
	return s.repo.GetDefaultSiteForTenant(tid)
}

func (s *TenantService) CreateSite(req CreateSiteRequest, scope tenantctx.Scope, ip string) (*model.Site, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("site name is required")
	}
	tenantID := scope.WriteTenantID(req.TenantID)
	t, err := s.repo.GetTenantByID(tenantID)
	if err != nil || t == nil {
		return nil, errors.New("target tenant not found")
	}
	if t.MaxSites > 0 {
		curr, _ := s.repo.CountSitesByTenant(tenantID)
		if int(curr) >= t.MaxSites {
			return nil, fmt.Errorf("kuota stasiun tenant '%s' telah mencapai batas maksimum (%d/%d)", t.Name, curr, t.MaxSites)
		}
	}

	code := model.NormalizeSlugCode(req.Code)
	if code == "" {
		code = model.NormalizeSlugCode(name)
	}
	if code == "" {
		code = fmt.Sprintf("station-%d", time.Now().Unix())
	}
	if existing, _ := s.repo.GetSiteByCode(tenantID, code); existing != nil {
		return nil, fmt.Errorf("site code '%s' already exists in tenant '%s'", code, t.Name)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	locSrc := strings.ToUpper(strings.TrimSpace(req.LocationSource))
	if locSrc == "" {
		if req.Latitude != nil && req.Longitude != nil {
			locSrc = model.LocationSourceManual
		} else {
			locSrc = model.LocationSourceUnconfigured
		}
	}

	st := &model.Site{
		TenantID:         tenantID,
		Code:             code,
		Name:             name,
		Description:      strings.TrimSpace(req.Description),
		MonitoringDomain: req.MonitoringDomain,
		StationTypeMode:  req.StationTypeMode,
		IngestFormat:     model.NormalizeIngestFormat(req.IngestFormat),
		LegacyStationID:  strings.TrimSpace(req.LegacyStationID),
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		LocationSource:   locSrc,
		Address:          strings.TrimSpace(req.Address),
		IsConfigured:     req.Latitude != nil && req.Longitude != nil,
		Timezone:         req.Timezone,
		Enabled:          enabled,
		UpdatedBy:        scope.Username,
	}
	if err := s.repo.CreateSite(st); err != nil {
		return nil, err
	}

	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &tenantID,
			Username:  scope.Username,
			Action:    "SITE_CREATE",
			Resource:  "SITE",
			Details:   fmt.Sprintf("Created station '%s' (%s, domain=%s)", st.Name, st.Code, st.MonitoringDomain),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}
	return s.repo.GetSiteByID(st.ID)
}

func (s *TenantService) UpdateSite(id uint, req UpdateSiteRequest, scope tenantctx.Scope, ip string) (*model.Site, error) {
	st, err := s.repo.GetSiteByID(id)
	if err != nil || st == nil {
		return nil, errors.New("site not found")
	}
	if !scope.CanAccessRecord(st.TenantID, st.ID) {
		return nil, errors.New("access denied: site belongs to another tenant")
	}

	if req.Code != nil && strings.TrimSpace(*req.Code) != "" {
		newCode := model.NormalizeSlugCode(*req.Code)
		if newCode != "" && !strings.EqualFold(newCode, st.Code) {
			if existing, _ := s.repo.GetSiteByCode(st.TenantID, newCode); existing != nil && existing.ID != st.ID {
				return nil, fmt.Errorf("site code '%s' already exists in this tenant", newCode)
			}
			st.Code = newCode
		}
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		st.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		st.Description = strings.TrimSpace(*req.Description)
	}
	if req.MonitoringDomain != nil && strings.TrimSpace(*req.MonitoringDomain) != "" {
		st.MonitoringDomain = strings.ToUpper(strings.TrimSpace(*req.MonitoringDomain))
	}
	if req.StationTypeMode != nil && strings.TrimSpace(*req.StationTypeMode) != "" {
		st.StationTypeMode = strings.ToUpper(strings.TrimSpace(*req.StationTypeMode))
	}
	if req.IngestFormat != nil {
		st.IngestFormat = model.NormalizeIngestFormat(*req.IngestFormat)
	}
	if req.LegacyStationID != nil {
		st.LegacyStationID = strings.TrimSpace(*req.LegacyStationID)
	}
	if req.Latitude != nil {
		st.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		st.Longitude = req.Longitude
	}
	if req.LocationSource != nil && strings.TrimSpace(*req.LocationSource) != "" {
		st.LocationSource = strings.ToUpper(strings.TrimSpace(*req.LocationSource))
	}
	if req.Address != nil {
		st.Address = strings.TrimSpace(*req.Address)
	}
	if req.Timezone != nil && strings.TrimSpace(*req.Timezone) != "" {
		st.Timezone = strings.TrimSpace(*req.Timezone)
	}
	if req.Enabled != nil {
		st.Enabled = *req.Enabled
	}
	st.IsConfigured = st.Latitude != nil && st.Longitude != nil
	st.UpdatedBy = scope.Username

	if err := s.repo.UpdateSite(st); err != nil {
		return nil, err
	}
	s.invalidateKeyCache()

	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &st.TenantID,
			Username:  scope.Username,
			Action:    "SITE_UPDATE",
			Resource:  "SITE",
			Details:   fmt.Sprintf("Updated station '%s' (%s, domain=%s)", st.Name, st.Code, st.MonitoringDomain),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}
	return s.repo.GetSiteByID(st.ID)
}

func (s *TenantService) DeleteSite(id uint, scope tenantctx.Scope, ip string) error {
	st, err := s.repo.GetSiteByID(id)
	if err != nil || st == nil {
		return errors.New("site not found")
	}
	if !scope.CanAccessRecord(st.TenantID, st.ID) {
		return errors.New("access denied: site belongs to another tenant")
	}
	devCount, _ := s.repo.CountDevicesBySite(id)
	if devCount > 0 {
		return fmt.Errorf("tidak dapat menghapus stasiun '%s' karena masih memiliki %d perangkat aktif", st.Name, devCount)
	}
	siteCount, _ := s.repo.CountSitesByTenant(st.TenantID)
	if siteCount <= 1 {
		return errors.New("setiap tenant wajib memiliki minimal 1 stasiun aktif")
	}
	if err := s.repo.DeleteSite(id); err != nil {
		return err
	}
	s.invalidateKeyCache()
	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &st.TenantID,
			Username:  scope.Username,
			Action:    "SITE_DELETE",
			Resource:  "SITE",
			Details:   fmt.Sprintf("Deleted station '%s' (ID=%d)", st.Name, id),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}
	return nil
}

// --- Edge Ingest API Keys ---

func hashIngestKey(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func cleanRawCustomToken(raw string) string {
	val := strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToLower(val), "authorization:") {
		val = strings.TrimSpace(val[len("authorization:"):])
	}
	lower := strings.ToLower(val)
	switch {
	case strings.HasPrefix(lower, "bearer "):
		val = strings.TrimSpace(val[7:])
	case strings.HasPrefix(lower, "token "):
		rest := strings.TrimSpace(val[6:])
		if strings.HasPrefix(strings.ToLower(rest), "token=") {
			rest = strings.TrimSpace(rest[6:])
		}
		val = rest
	}
	val = strings.Trim(val, `"'`)
	return strings.TrimSpace(val)
}

func formatKeyPrefix(rawKey string) string {
	if len(rawKey) > 14 {
		return rawKey[:10] + "..."
	}
	if len(rawKey) > 4 {
		return rawKey[:4] + "..."
	}
	return rawKey + "..."
}

func (s *TenantService) invalidateKeyCache() {
	s.keyMu.Lock()
	s.keyCache = make(map[string]cachedIngestKey)
	s.keyMu.Unlock()
}

func (s *TenantService) ListAPIKeys(scope tenantctx.Scope, explicitTenantID uint) ([]model.TenantAPIKey, error) {
	tid := scope.TenantID
	if scope.IsSuperAdmin && explicitTenantID > 0 {
		tid = explicitTenantID
	} else if !scope.IsSuperAdmin && tid == 0 {
		tid = tenantctx.DefaultTenantID
	}
	return s.repo.ListAPIKeysScoped(tid, scope.IsSuperAdmin)
}

func (s *TenantService) CreateAPIKey(req CreateAPIKeyRequest, scope tenantctx.Scope, ip string) (*model.TenantAPIKey, string, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, "", errors.New("API key name is required")
	}

	isGlobal := req.IsGlobal
	var tenantID uint
	var siteID *uint
	if isGlobal {
		if !scope.IsSuperAdmin {
			return nil, "", errors.New("access denied: hanya Superadmin yang dapat membuat Token Global Lintas-Tenant")
		}
		tenantID = tenantctx.DefaultTenantID
		if _, err := s.repo.GetTenantByID(tenantID); err != nil {
			tenants, _ := s.repo.ListTenants(true)
			if len(tenants) > 0 {
				tenantID = tenants[0].ID
			}
		}
	} else {
		tenantID = scope.WriteTenantID(req.TenantID)
		if _, err := s.repo.GetTenantByID(tenantID); err != nil {
			return nil, "", errors.New("target tenant not found")
		}
		if req.SiteID != nil && *req.SiteID > 0 {
			st, err := s.repo.GetSiteByID(*req.SiteID)
			if err != nil || st == nil || st.TenantID != tenantID {
				return nil, "", errors.New("selected site does not belong to target tenant")
			}
			sid := st.ID
			siteID = &sid
		}
	}

	rawKey := cleanRawCustomToken(req.CustomToken)
	var prefix string
	if rawKey != "" {
		if len(rawKey) < 4 {
			return nil, "", errors.New("token datalogger minimal 4 karakter")
		}
		if existing, _ := s.repo.GetAPIKeyByHash(hashIngestKey(rawKey)); existing != nil {
			return nil, "", errors.New("token tersebut sudah terdaftar pada kunci lain (gunakan mode Global Lintas-Tenant jika 1 token lama dipakai banyak tenant)")
		}
		prefix = formatKeyPrefix(rawKey)
	} else {
		rawBytes := make([]byte, 20)
		if _, err := rand.Read(rawBytes); err != nil {
			return nil, "", errors.New("failed to generate random API key")
		}
		rawKey = "dlg_live_" + hex.EncodeToString(rawBytes)
		prefix = rawKey[:17] + "..."
	}
	// Strict governance: devices and parameters must be registered manually by users (never auto-created).
	autoProv := false
	ingestFmt := model.NormalizeIngestFormat(req.IngestFormat)

	k := &model.TenantAPIKey{
		TenantID:      tenantID,
		SiteID:        siteID,
		IsGlobal:      isGlobal,
		Name:          name,
		KeyPrefix:     prefix,
		KeyHash:       hashIngestKey(rawKey),
		IngestFormat:  ingestFmt,
		AutoProvision: autoProv,
		Enabled:       true,
		CreatedBy:     scope.Username,
	}
	if err := s.repo.CreateAPIKey(k); err != nil {
		return nil, "", err
	}
	s.invalidateKeyCache()

	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &tenantID,
			Username:  scope.Username,
			Action:    "INGEST_KEY_CREATE",
			Resource:  "INGEST_KEY",
			Details:   fmt.Sprintf("Generated Edge Ingest API key '%s' (%s, format=%s, global=%v)", k.Name, k.KeyPrefix, k.IngestFormat, k.IsGlobal),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}

	loaded, err := s.repo.GetAPIKeyByID(k.ID)
	if err == nil && loaded != nil {
		return loaded, rawKey, nil
	}
	return k, rawKey, nil
}

func (s *TenantService) UpdateAPIKey(id uint, req UpdateAPIKeyRequest, scope tenantctx.Scope, ip string) (*model.TenantAPIKey, error) {
	k, err := s.repo.GetAPIKeyByID(id)
	if err != nil || k == nil {
		return nil, errors.New("API key not found")
	}
	if k.IsGlobal && !scope.IsSuperAdmin {
		return nil, errors.New("access denied: global cross-tenant API key requires superadmin")
	}
	if !k.IsGlobal && !scope.CanAccessTenant(k.TenantID) {
		return nil, errors.New("access denied: API key belongs to another tenant")
	}
	if req.Enabled != nil {
		k.Enabled = *req.Enabled
	}
	if req.IngestFormat != nil {
		k.IngestFormat = model.NormalizeIngestFormat(*req.IngestFormat)
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		k.Name = strings.TrimSpace(*req.Name)
	}
	if req.IsGlobal != nil {
		if !scope.IsSuperAdmin {
			return nil, errors.New("access denied: hanya Superadmin yang dapat mengatur mode Global Lintas-Tenant")
		}
		k.IsGlobal = *req.IsGlobal
		if k.IsGlobal {
			k.SiteID = nil
		}
	}
	if req.CustomToken != nil {
		cleaned := cleanRawCustomToken(*req.CustomToken)
		if cleaned == "" {
			return nil, errors.New("token pengganti tidak boleh kosong")
		}
		if len(cleaned) < 4 {
			return nil, errors.New("token datalogger minimal 4 karakter")
		}
		newHash := hashIngestKey(cleaned)
		if existing, _ := s.repo.GetAPIKeyByHash(newHash); existing != nil && existing.ID != k.ID {
			return nil, errors.New("token tersebut sudah terdaftar pada kunci lain")
		}
		k.KeyHash = newHash
		k.KeyPrefix = formatKeyPrefix(cleaned)
	}
	if !k.IsGlobal {
		if req.ClearSiteID {
			k.SiteID = nil
		} else if req.SiteID != nil {
			if *req.SiteID == 0 {
				k.SiteID = nil
			} else {
				st, sErr := s.repo.GetSiteByID(*req.SiteID)
				if sErr != nil || st == nil || st.TenantID != k.TenantID {
					return nil, errors.New("selected site does not belong to target tenant")
				}
				sid := st.ID
				k.SiteID = &sid
			}
		}
	}
	k.AutoProvision = false
	if err := s.repo.UpdateAPIKey(k); err != nil {
		return nil, err
	}
	s.invalidateKeyCache()
	if loaded, lErr := s.repo.GetAPIKeyByID(k.ID); lErr == nil && loaded != nil {
		return loaded, nil
	}
	return k, nil
}

func (s *TenantService) ToggleAPIKey(id uint, enabled bool, autoProvision *bool, scope tenantctx.Scope, ip string) (*model.TenantAPIKey, error) {
	return s.UpdateAPIKey(id, UpdateAPIKeyRequest{
		Enabled:       &enabled,
		AutoProvision: autoProvision,
	}, scope, ip)
}

func (s *TenantService) DeleteAPIKey(id uint, scope tenantctx.Scope, ip string) error {
	k, err := s.repo.GetAPIKeyByID(id)
	if err != nil || k == nil {
		return errors.New("API key not found")
	}
	if k.IsGlobal && !scope.IsSuperAdmin {
		return errors.New("access denied: global cross-tenant API key requires superadmin")
	}
	if !k.IsGlobal && !scope.CanAccessTenant(k.TenantID) {
		return errors.New("access denied: API key belongs to another tenant")
	}
	if err := s.repo.DeleteAPIKey(id); err != nil {
		return err
	}
	s.invalidateKeyCache()
	if s.sysRepo != nil {
		_ = s.sysRepo.AddAuditTrail(&model.AuditTrail{
			TenantID:  &k.TenantID,
			Username:  scope.Username,
			Action:    "INGEST_KEY_REVOKE",
			Resource:  "INGEST_KEY",
			Details:   fmt.Sprintf("Revoked Edge Ingest API key '%s' (%s)", k.Name, k.KeyPrefix),
			IPAddress: ip,
			CreatedAt: time.Now(),
		})
	}
	return nil
}

func (s *TenantService) ValidateIngestKey(rawKey string) (*model.TenantAPIKey, error) {
	rawKey = cleanRawCustomToken(rawKey)
	if rawKey == "" {
		return nil, errors.New("missing Edge Ingest API key")
	}
	hash := hashIngestKey(rawKey)
	now := time.Now()

	s.keyMu.RLock()
	if entry, ok := s.keyCache[hash]; ok && now.Before(entry.expiresAt) {
		k := entry.key
		s.keyMu.RUnlock()
		if k == nil || !k.Enabled {
			return nil, errors.New("invalid or disabled Edge Ingest API key")
		}
		if !k.IsGlobal && k.Tenant != nil && strings.EqualFold(k.Tenant.Status, model.TenantStatusSuspended) {
			return nil, errors.New("tenant account is suspended")
		}
		return k, nil
	}
	s.keyMu.RUnlock()

	k, err := s.repo.GetAPIKeyByHash(hash)
	if err != nil || k == nil || !k.Enabled {
		return nil, errors.New("invalid or disabled Edge Ingest API key")
	}
	if !k.IsGlobal && k.Tenant != nil && strings.EqualFold(k.Tenant.Status, model.TenantStatusSuspended) {
		return nil, errors.New("tenant account is suspended")
	}

	s.keyMu.Lock()
	s.keyCache[hash] = cachedIngestKey{
		key:       k,
		expiresAt: now.Add(45 * time.Second),
	}
	s.keyMu.Unlock()

	return k, nil
}


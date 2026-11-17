package repository

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

const siteIngestCacheTTL = 20 * time.Second

type cachedSiteIngestEntry struct {
	site      *model.Site
	expiresAt time.Time
}

type TenantRepository struct {
	db                *gorm.DB
	siteCacheMu       sync.RWMutex
	siteByIDCache     map[uint]cachedSiteIngestEntry
	siteByLookupCache map[string]cachedSiteIngestEntry
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{
		db:                db,
		siteByIDCache:     make(map[uint]cachedSiteIngestEntry),
		siteByLookupCache: make(map[string]cachedSiteIngestEntry),
	}
}

func cloneSiteForIngest(src *model.Site) *model.Site {
	if src == nil {
		return nil
	}
	cp := *src
	if src.Tenant != nil {
		tCp := *src.Tenant
		cp.Tenant = &tCp
	}
	return &cp
}

// InvalidateSiteIngestCache clears cached site & tenant metadata used on the high-throughput ingest path.
func (r *TenantRepository) InvalidateSiteIngestCache() {
	r.siteCacheMu.Lock()
	r.siteByIDCache = make(map[uint]cachedSiteIngestEntry)
	r.siteByLookupCache = make(map[string]cachedSiteIngestEntry)
	r.siteCacheMu.Unlock()
}

func (r *TenantRepository) cacheSiteEntry(lookupKey string, s *model.Site) *model.Site {
	if s == nil {
		return nil
	}
	exp := time.Now().Add(siteIngestCacheTTL)
	entry := cachedSiteIngestEntry{
		site:      cloneSiteForIngest(s),
		expiresAt: exp,
	}
	r.siteCacheMu.Lock()
	if r.siteByIDCache == nil {
		r.siteByIDCache = make(map[uint]cachedSiteIngestEntry)
	}
	if r.siteByLookupCache == nil {
		r.siteByLookupCache = make(map[string]cachedSiteIngestEntry)
	}
	if s.ID > 0 {
		r.siteByIDCache[s.ID] = entry
	}
	if lookupKey != "" {
		r.siteByLookupCache[lookupKey] = entry
	}
	r.siteCacheMu.Unlock()
	return cloneSiteForIngest(s)
}

func (r *TenantRepository) getCachedSiteByLookup(lookupKey string) (*model.Site, bool) {
	now := time.Now()
	r.siteCacheMu.RLock()
	entry, ok := r.siteByLookupCache[lookupKey]
	if ok && entry.site != nil && now.Before(entry.expiresAt) {
		cp := cloneSiteForIngest(entry.site)
		r.siteCacheMu.RUnlock()
		return cp, true
	}
	r.siteCacheMu.RUnlock()
	return nil, false
}

// GetCachedSiteByIDForIngest fetches lightweight site + tenant status for high-frequency HTTP/MQTT ingestion
// without running global GROUP BY device count queries on every request.
func (r *TenantRepository) GetCachedSiteByIDForIngest(id uint) (*model.Site, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	now := time.Now()
	r.siteCacheMu.RLock()
	if entry, ok := r.siteByIDCache[id]; ok && entry.site != nil && now.Before(entry.expiresAt) {
		cp := cloneSiteForIngest(entry.site)
		r.siteCacheMu.RUnlock()
		return cp, nil
	}
	r.siteCacheMu.RUnlock()

	var s model.Site
	if err := r.db.Preload("Tenant").First(&s, id).Error; err != nil {
		return nil, err
	}
	return r.cacheSiteEntry("", &s), nil
}

// GetCachedSiteByCodeForIngest resolves a tenant site by normalized/raw code using short-TTL RAM cache.
func (r *TenantRepository) GetCachedSiteByCodeForIngest(tenantID uint, code string) (*model.Site, error) {
	clean := model.NormalizeSlugCode(code)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	key := fmt.Sprintf("code:%d:%s", tenantID, strings.ToLower(clean))
	if cached, ok := r.getCachedSiteByLookup(key); ok {
		return cached, nil
	}
	var s model.Site
	if err := r.db.Preload("Tenant").Where("tenant_id = ? AND LOWER(code) = LOWER(?)", tenantID, clean).First(&s).Error; err != nil {
		return nil, err
	}
	return r.cacheSiteEntry(key, &s), nil
}

// GetCachedSiteByCodeGlobalForIngest resolves an enabled site across tenants by code using short-TTL RAM cache.
func (r *TenantRepository) GetCachedSiteByCodeGlobalForIngest(code string) (*model.Site, error) {
	clean := model.NormalizeSlugCode(code)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	key := "gcode:" + strings.ToLower(clean)
	if cached, ok := r.getCachedSiteByLookup(key); ok {
		return cached, nil
	}
	var s model.Site
	if err := r.db.Preload("Tenant").Where("enabled = ? AND LOWER(code) = LOWER(?)", true, clean).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return r.cacheSiteEntry(key, &s), nil
}

// GetCachedSiteByLegacyStationIDGlobalForIngest resolves an enabled site across tenants by legacy station_id using short-TTL RAM cache.
func (r *TenantRepository) GetCachedSiteByLegacyStationIDGlobalForIngest(legacyStationID string) (*model.Site, error) {
	clean := strings.TrimSpace(legacyStationID)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	key := "glegacy:" + strings.ToLower(clean)
	if cached, ok := r.getCachedSiteByLookup(key); ok {
		return cached, nil
	}
	var s model.Site
	if err := r.db.Preload("Tenant").Where("enabled = ? AND LOWER(legacy_station_id) = LOWER(?)", true, clean).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return r.cacheSiteEntry(key, &s), nil
}

// GetCachedSiteByLegacyStationIDForIngest resolves a site within a tenant by legacy station_id using short-TTL RAM cache.
func (r *TenantRepository) GetCachedSiteByLegacyStationIDForIngest(tenantID uint, legacyStationID string) (*model.Site, error) {
	clean := strings.TrimSpace(legacyStationID)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	key := fmt.Sprintf("legacy:%d:%s", tenantID, strings.ToLower(clean))
	if cached, ok := r.getCachedSiteByLookup(key); ok {
		return cached, nil
	}
	var s model.Site
	if err := r.db.Preload("Tenant").Where("tenant_id = ? AND LOWER(legacy_station_id) = LOWER(?)", tenantID, clean).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return r.cacheSiteEntry(key, &s), nil
}

// GetCachedDefaultSiteForTenantForIngest resolves a tenant's default site using short-TTL RAM cache.
func (r *TenantRepository) GetCachedDefaultSiteForTenantForIngest(tenantID uint) (*model.Site, error) {
	key := fmt.Sprintf("default:%d", tenantID)
	if cached, ok := r.getCachedSiteByLookup(key); ok {
		return cached, nil
	}
	var s model.Site
	if err := r.db.Preload("Tenant").Where("tenant_id = ?", tenantID).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return r.cacheSiteEntry(key, &s), nil
}

func (r *TenantRepository) GetDB() *gorm.DB {
	return r.db
}

// --- Tenants ---

func (r *TenantRepository) ListTenants(includeSites bool) ([]model.Tenant, error) {
	var tenants []model.Tenant
	q := r.db.Model(&model.Tenant{}).Order("id ASC")
	if includeSites {
		q = q.Preload("Sites", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		})
	}
	if err := q.Find(&tenants).Error; err != nil {
		return nil, err
	}
	r.populateTenantCounts(tenants)
	if includeSites {
		for i := range tenants {
			r.populateSiteCounts(tenants[i].Sites)
		}
	}
	return tenants, nil
}

func (r *TenantRepository) GetTenantByID(id uint) (*model.Tenant, error) {
	var t model.Tenant
	if err := r.db.Preload("Sites", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).First(&t, id).Error; err != nil {
		return nil, err
	}
	slice := []model.Tenant{t}
	r.populateTenantCounts(slice)
	r.populateSiteCounts(slice[0].Sites)
	return &slice[0], nil
}

func (r *TenantRepository) GetTenantByCode(code string) (*model.Tenant, error) {
	var t model.Tenant
	clean := model.NormalizeSlugCode(code)
	if err := r.db.Where("LOWER(code) = LOWER(?)", clean).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) CreateTenant(t *model.Tenant) error {
	err := r.db.Create(t).Error
	r.InvalidateSiteIngestCache()
	return err
}

func (r *TenantRepository) UpdateTenant(t *model.Tenant) error {
	err := r.db.Save(t).Error
	r.InvalidateSiteIngestCache()
	return err
}

func (r *TenantRepository) DeleteTenant(id uint) error {
	err := r.db.Delete(&model.Tenant{}, id).Error
	r.InvalidateSiteIngestCache()
	return err
}

func (r *TenantRepository) CountSitesByTenant(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Site{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

func (r *TenantRepository) CountDevicesByTenant(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Device{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

func (r *TenantRepository) populateTenantCounts(tenants []model.Tenant) {
	if len(tenants) == 0 || r.db == nil {
		return
	}
	type countRow struct {
		TenantID uint
		Total    int64
	}
	var siteCounts, devCounts, userCounts []countRow
	_ = r.db.Model(&model.Site{}).Select("tenant_id, COUNT(*) as total").Group("tenant_id").Scan(&siteCounts).Error
	_ = r.db.Model(&model.Device{}).Select("tenant_id, COUNT(*) as total").Group("tenant_id").Scan(&devCounts).Error
	_ = r.db.Model(&model.User{}).Where("tenant_id IS NOT NULL").Select("tenant_id, COUNT(*) as total").Group("tenant_id").Scan(&userCounts).Error

	sMap := make(map[uint]int64, len(siteCounts))
	for _, c := range siteCounts {
		sMap[c.TenantID] = c.Total
	}
	dMap := make(map[uint]int64, len(devCounts))
	for _, c := range devCounts {
		dMap[c.TenantID] = c.Total
	}
	uMap := make(map[uint]int64, len(userCounts))
	for _, c := range userCounts {
		uMap[c.TenantID] = c.Total
	}
	for i := range tenants {
		id := tenants[i].ID
		tenants[i].SiteCount = sMap[id]
		tenants[i].DeviceCount = dMap[id]
		tenants[i].UserCount = uMap[id]
		tenants[i].CustomerHasLogo = strings.TrimSpace(tenants[i].CustomerLogo) != ""
	}
}

// --- Sites ---

func (r *TenantRepository) ListSites(tenantID uint, allowedSiteIDs []uint) ([]model.Site, error) {
	var sites []model.Site
	q := r.db.Model(&model.Site{}).Preload("Tenant").Order("tenant_id ASC, id ASC")
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if len(allowedSiteIDs) > 0 {
		q = q.Where("id IN ?", allowedSiteIDs)
	}
	if err := q.Find(&sites).Error; err != nil {
		return nil, err
	}
	r.populateSiteCounts(sites)
	return sites, nil
}

func (r *TenantRepository) GetSiteByID(id uint) (*model.Site, error) {
	var s model.Site
	if err := r.db.Preload("Tenant").First(&s, id).Error; err != nil {
		return nil, err
	}
	slice := []model.Site{s}
	r.populateSiteCounts(slice)
	return &slice[0], nil
}

func (r *TenantRepository) GetSiteByCode(tenantID uint, code string) (*model.Site, error) {
	var s model.Site
	clean := model.NormalizeSlugCode(code)
	if err := r.db.Where("tenant_id = ? AND LOWER(code) = LOWER(?)", tenantID, clean).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *TenantRepository) GetSiteByLegacyStationID(tenantID uint, legacyStationID string) (*model.Site, error) {
	clean := strings.TrimSpace(legacyStationID)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var s model.Site
	if err := r.db.Where("tenant_id = ? AND LOWER(legacy_station_id) = LOWER(?)", tenantID, clean).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *TenantRepository) GetSiteByLegacyStationIDGlobal(legacyStationID string) (*model.Site, error) {
	clean := strings.TrimSpace(legacyStationID)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var s model.Site
	if err := r.db.Preload("Tenant").Where("enabled = ? AND LOWER(legacy_station_id) = LOWER(?)", true, clean).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *TenantRepository) GetSiteByCodeGlobal(code string) (*model.Site, error) {
	clean := model.NormalizeSlugCode(code)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var s model.Site
	if err := r.db.Preload("Tenant").Where("enabled = ? AND LOWER(code) = LOWER(?)", true, clean).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *TenantRepository) GetDefaultSiteForTenant(tenantID uint) (*model.Site, error) {
	var s model.Site
	if err := r.db.Where("tenant_id = ?", tenantID).Order("id ASC").First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *TenantRepository) CreateSite(s *model.Site) error {
	err := r.db.Create(s).Error
	r.InvalidateSiteIngestCache()
	return err
}

func (r *TenantRepository) UpdateSite(s *model.Site) error {
	err := r.db.Save(s).Error
	r.InvalidateSiteIngestCache()
	return err
}

func (r *TenantRepository) DeleteSite(id uint) error {
	err := r.db.Delete(&model.Site{}, id).Error
	r.InvalidateSiteIngestCache()
	return err
}

func (r *TenantRepository) CountDevicesBySite(siteID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Device{}).Where("site_id = ?", siteID).Count(&count).Error
	return count, err
}

func (r *TenantRepository) populateSiteCounts(sites []model.Site) {
	if len(sites) == 0 || r.db == nil {
		return
	}
	type siteStat struct {
		SiteID uint
		Total  int64
		Online int64
	}
	var rows []siteStat
	_ = r.db.Model(&model.Device{}).
		Select("site_id, COUNT(*) as total, SUM(CASE WHEN connection_status = 'ONLINE' THEN 1 ELSE 0 END) as online").
		Group("site_id").
		Scan(&rows).Error
	statMap := make(map[uint]siteStat, len(rows))
	for _, row := range rows {
		statMap[row.SiteID] = row
	}
	for i := range sites {
		if st, ok := statMap[sites[i].ID]; ok {
			sites[i].DeviceCount = st.Total
			sites[i].OnlineCount = st.Online
		}
	}
}

// --- Tenant Ingest API Keys ---

func (r *TenantRepository) ListAPIKeys(tenantID uint) ([]model.TenantAPIKey, error) {
	return r.ListAPIKeysScoped(tenantID, tenantID == 0)
}

func (r *TenantRepository) ListAPIKeysScoped(tenantID uint, isSuperAdmin bool) ([]model.TenantAPIKey, error) {
	var keys []model.TenantAPIKey
	q := r.db.Model(&model.TenantAPIKey{}).Preload("Tenant").Preload("Site").Order("is_global DESC, id DESC")
	if isSuperAdmin {
		if tenantID > 0 {
			q = q.Where("tenant_id = ? OR is_global = ?", tenantID, true)
		}
	} else {
		q = q.Where("tenant_id = ? AND (is_global = ? OR is_global IS NULL)", tenantID, false)
	}
	err := q.Find(&keys).Error
	return keys, err
}

func (r *TenantRepository) GetAPIKeyByID(id uint) (*model.TenantAPIKey, error) {
	var k model.TenantAPIKey
	if err := r.db.Preload("Tenant").Preload("Site").First(&k, id).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *TenantRepository) GetAPIKeyByHash(keyHash string) (*model.TenantAPIKey, error) {
	var k model.TenantAPIKey
	if err := r.db.Preload("Tenant").Preload("Site").Where("key_hash = ?", keyHash).First(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *TenantRepository) CreateAPIKey(k *model.TenantAPIKey) error {
	return r.db.Create(k).Error
}

func (r *TenantRepository) UpdateAPIKey(k *model.TenantAPIKey) error {
	return r.db.Save(k).Error
}

func (r *TenantRepository) DeleteAPIKey(id uint) error {
	return r.db.Delete(&model.TenantAPIKey{}, id).Error
}

func (r *TenantRepository) RecordAPIKeyUsage(id uint, ip string, ingestedCount int64) {
	now := time.Now().UTC()
	_ = r.db.Model(&model.TenantAPIKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_used_at":   &now,
		"last_used_ip":   ip,
		"total_ingested": gorm.Expr("total_ingested + ?", ingestedCount),
	}).Error
}


package repository

import (
	"strings"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
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
	return r.db.Create(t).Error
}

func (r *TenantRepository) UpdateTenant(t *model.Tenant) error {
	return r.db.Save(t).Error
}

func (r *TenantRepository) DeleteTenant(id uint) error {
	return r.db.Delete(&model.Tenant{}, id).Error
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
	return r.db.Create(s).Error
}

func (r *TenantRepository) UpdateSite(s *model.Site) error {
	return r.db.Save(s).Error
}

func (r *TenantRepository) DeleteSite(id uint) error {
	return r.db.Delete(&model.Site{}, id).Error
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


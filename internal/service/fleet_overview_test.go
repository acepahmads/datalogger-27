package service_test

import (
	"fmt"
	"testing"
	"time"

	"datalogger/internal/database"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestGetFleetOverviewMultiTenantAggregation(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_fleet_overview_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Clean seeded default rows to verify deterministic counts
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.Alarm{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.Parameter{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.DeviceConnection{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.Device{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.Site{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.Tenant{})

	tenantRepo := repository.NewTenantRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	tenantService := service.NewTenantService(tenantRepo, sysRepo)

	t1 := model.Tenant{Code: "klh-dki", Name: "DLH DKI Jakarta", Status: model.TenantStatusActive}
	t2 := model.Tenant{Code: "pt-industri", Name: "PT Industri Hijau", Status: model.TenantStatusActive}
	if err := db.Create(&t1).Error; err != nil {
		t.Fatalf("create t1: %v", err)
	}
	if err := db.Create(&t2).Error; err != nil {
		t.Fatalf("create t2: %v", err)
	}

	now := time.Now().UTC()

	// Site 1: Online AQMS with valid Site GPS (-6.1944, 106.8229)
	lat1, lng1 := -6.1944, 106.8229
	s1 := model.Site{
		TenantID:         t1.ID,
		Code:             "bundaran-hi",
		Name:             "Stasiun AQMS Bundaran HI",
		MonitoringDomain: "AQMS",
		Latitude:         &lat1,
		Longitude:        &lng1,
		LocationSource:   model.LocationSourceManual,
		Enabled:          true,
	}
	// Site 2: Delayed WQMS with Null-Island (0, 0) on Site, falling back to Device GPS (-6.9175, 107.6191)
	zeroLat, zeroLng := 0.0, 0.0
	s2 := model.Site{
		TenantID:         t1.ID,
		Code:             "sungai-citarum",
		Name:             "Pos WQMS Citarum",
		MonitoringDomain: "WQMS",
		Latitude:         &zeroLat,
		Longitude:        &zeroLng,
		LocationSource:   model.LocationSourceUnconfigured,
		Enabled:          true,
	}
	// Site 3: Offline SPARING with Critical Alarm & NO GPS anywhere
	s3 := model.Site{
		TenantID:         t2.ID,
		Code:             "sparing-outlet-01",
		Name:             "SPARING Outlet IPAL 01",
		MonitoringDomain: "SPARING",
		Latitude:         &zeroLat,
		Longitude:        &zeroLng,
		Enabled:          true,
	}
	// Site 4: Administratively Disabled Site
	s4 := model.Site{
		TenantID:         t2.ID,
		Code:             "cadangan",
		Name:             "Stasiun Cadangan",
		MonitoringDomain: "AQMS",
		Enabled:          false,
	}
	for _, s := range []*model.Site{&s1, &s2, &s3, &s4} {
		if err := db.Create(s).Error; err != nil {
			t.Fatalf("create site %s: %v", s.Code, err)
		}
	}
	// GORM skips false bool when default:true unless explicitly updated
	db.Model(&s4).Update("enabled", false)

	// Device 1 on Site 1: Fresh data 30s ago -> ONLINE
	tFresh := now.Add(-30 * time.Second)
	d1 := model.Device{
		TenantID:         t1.ID,
		SiteID:           s1.ID,
		DeviceCode:       "AQMS-HI-01",
		DeviceName:       "Logger Bundaran HI",
		DeviceType:       string(model.ProtocolEdgePush),
		Enabled:          true,
		Status:           model.DeviceStatusActive,
		ConnectionStatus: model.DeviceConnOnline,
		LastDataAt:       &tFresh,
		LastSeenAt:       &tFresh,
	}
	db.Create(&d1)
	db.Create(&model.DeviceConnection{DeviceID: d1.ID, Protocol: model.ProtocolEdgePush, Timeout: 300, Enabled: true})

	// Device 2 on Site 2: Delayed data 10m (600s) ago with timeout=300s + valid Device GPS -> DELAYED + DEVICE_FALLBACK
	tDelayed := now.Add(-10 * time.Minute)
	devLat, devLng := -6.9175, 107.6191
	d2 := model.Device{
		TenantID:         t1.ID,
		SiteID:           s2.ID,
		DeviceCode:       "WQMS-CTR-01",
		DeviceName:       "Logger Citarum",
		DeviceType:       string(model.ProtocolEdgePush),
		Latitude:         &devLat,
		Longitude:        &devLng,
		Enabled:          true,
		Status:           model.DeviceStatusActive,
		ConnectionStatus: model.DeviceConnOnline,
		LastDataAt:       &tDelayed,
		LastSeenAt:       &tDelayed,
	}
	db.Create(&d2)
	db.Create(&model.DeviceConnection{DeviceID: d2.ID, Protocol: model.ProtocolEdgePush, Timeout: 300, Enabled: true})

	// Device 3 on Site 3: Offline data 2h ago + Critical Alarm -> OFFLINE + CRITICAL + No-GPS
	tOld := now.Add(-2 * time.Hour)
	d3 := model.Device{
		TenantID:         t2.ID,
		SiteID:           s3.ID,
		DeviceCode:       "SPR-OUT-01",
		DeviceName:       "Logger IPAL 01",
		DeviceType:       string(model.ProtocolMQTT),
		Enabled:          true,
		Status:           model.DeviceStatusActive,
		ConnectionStatus: model.DeviceConnOffline,
		LastDataAt:       &tOld,
		LastSeenAt:       &tOld,
	}
	db.Create(&d3)

	p3 := model.Parameter{
		DeviceID:      d3.ID,
		ParameterCode: "cod",
		ParameterName: "COD",
		Enabled:       true,
	}
	db.Create(&p3)

	alCrit := model.Alarm{
		TenantID:    t2.ID,
		SiteID:      s3.ID,
		DeviceID:    &d3.ID,
		ParameterID: &p3.ID,
		AlarmCode:   "HIGH_CRITICAL",
		Severity:    model.SeverityCritical,
		Message:     "COD melebihi ambang kritis",
		Status:      model.AlarmActive,
		TriggeredAt: now.Add(-5 * time.Minute),
	}
	db.Create(&alCrit)

	overview, err := tenantService.GetFleetOverview()
	if err != nil {
		t.Fatalf("GetFleetOverview failed: %v", err)
	}

	if overview.Summary.TotalTenants != 2 {
		t.Errorf("expected 2 tenants, got %d", overview.Summary.TotalTenants)
	}
	if overview.Summary.TotalSites != 4 {
		t.Errorf("expected 4 sites, got %d", overview.Summary.TotalSites)
	}
	if overview.Summary.MappedSites != 2 || overview.Summary.NoGPSSites != 2 {
		t.Errorf("expected MappedSites=2 & NoGPSSites=2, got mapped=%d nogps=%d", overview.Summary.MappedSites, overview.Summary.NoGPSSites)
	}
	if overview.Summary.OnlineSites != 1 || overview.Summary.DelayedSites != 1 || overview.Summary.OfflineSites != 1 || overview.Summary.InactiveSites != 1 {
		t.Errorf("unexpected site delivery breakdown: %+v", overview.Summary)
	}
	if overview.Summary.CriticalAlarms != 1 || overview.Summary.SitesWithAlarms != 1 {
		t.Errorf("expected 1 critical alarm on 1 site, got %+v", overview.Summary)
	}

	// Verify Urgency Sort Order: Site 3 (CRITICAL) -> Site 2 (WARNING/DELAYED) -> Site 1 (NORMAL/ONLINE) -> Site 4 (INACTIVE)
	if len(overview.Sites) != 4 {
		t.Fatalf("expected 4 site items, got %d", len(overview.Sites))
	}
	if overview.Sites[0].SiteCode != "sparing-outlet-01" || overview.Sites[0].SeverityState != "CRITICAL" || overview.Sites[0].HasCoordinates {
		t.Errorf("expected sparing-outlet-01 first as CRITICAL No-GPS, got %+v", overview.Sites[0])
	}
	if overview.Sites[1].SiteCode != "sungai-citarum" || overview.Sites[1].DeliveryStatus != "DELAYED" || !overview.Sites[1].HasCoordinates || overview.Sites[1].LocationSource != "DEVICE_FALLBACK" {
		t.Errorf("expected sungai-citarum second as DELAYED with DEVICE_FALLBACK GPS, got %+v", overview.Sites[1])
	}
	if overview.Sites[2].SiteCode != "bundaran-hi" || overview.Sites[2].DeliveryStatus != "ONLINE" || overview.Sites[2].SeverityState != "NORMAL" {
		t.Errorf("expected bundaran-hi third as ONLINE/NORMAL, got %+v", overview.Sites[2])
	}
	if overview.Sites[3].SiteCode != "cadangan" || overview.Sites[3].DeliveryStatus != "INACTIVE" {
		t.Errorf("expected cadangan fourth as INACTIVE, got %+v", overview.Sites[3])
	}
}


package database_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/model"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func openFreshSQLite(t *testing.T, label string) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:mem_seed_cfg_%s_%d?mode=memory&cache=shared", label, time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite (%s): %v", label, err)
	}
	return db
}

func resetGlobalConfigCleanup(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		def := config.DefaultConfig()
		g := config.Get()
		if g != nil {
			*g = *def
		}
		_ = os.Unsetenv("CLOUD_MODE")
		_ = os.Unsetenv("SEED_USERS")
		_ = os.Unsetenv("SEED_SAMPLE_DATA")
		_ = os.Unsetenv("SUPERADMIN_USER")
		_ = os.Unsetenv("SUPERADMIN_PASSWORD")
	})
}

func TestConfigLoadSeedDefaultsForCloudAndEdge(t *testing.T) {
	resetGlobalConfigCleanup(t)
	tmpDir := t.TempDir()

	// 1. Edge AQMS config.json defaults seed_users=true, seed_sample_data=true
	edgeCfgPath := filepath.Join(tmpDir, "config-edge.json")
	if err := os.WriteFile(edgeCfgPath, []byte(`{
		"monitoring_domain": "AQMS",
		"data_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "data"))+`",
		"log_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "logs"))+`"
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfgEdge, err := config.Load(edgeCfgPath)
	if err != nil {
		t.Fatalf("Load edge config failed: %v", err)
	}
	if !cfgEdge.SeedUsers || !cfgEdge.SeedSampleData {
		t.Fatalf("expected Edge default seed_users=true & seed_sample_data=true, got %t / %t", cfgEdge.SeedUsers, cfgEdge.SeedSampleData)
	}

	// 2. Edge config.json with explicit seed_users=false, seed_sample_data=false
	edgeOffPath := filepath.Join(tmpDir, "config-edge-off.json")
	if err := os.WriteFile(edgeOffPath, []byte(`{
		"monitoring_domain": "AQMS",
		"seed_users": false,
		"seed_sample_data": false,
		"data_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "data"))+`",
		"log_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "logs"))+`"
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfgEdgeOff, err := config.Load(edgeOffPath)
	if err != nil {
		t.Fatalf("Load edge-off config failed: %v", err)
	}
	if cfgEdgeOff.SeedUsers || cfgEdgeOff.SeedSampleData {
		t.Fatalf("expected explicit Edge seed_users=false & seed_sample_data=false, got %t / %t", cfgEdgeOff.SeedUsers, cfgEdgeOff.SeedSampleData)
	}

	// 3. Cloud config.json without explicit seed flags -> defaults seed_users=false & seed_sample_data=false
	cloudCfgPath := filepath.Join(tmpDir, "config-cloud.json")
	if err := os.WriteFile(cloudCfgPath, []byte(`{
		"monitoring_domain": "CLOUD",
		"data_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "data"))+`",
		"log_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "logs"))+`"
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfgCloud, err := config.Load(cloudCfgPath)
	if err != nil {
		t.Fatalf("Load cloud config failed: %v", err)
	}
	if !cfgCloud.IsCloudMode() || cfgCloud.SeedUsers || cfgCloud.SeedSampleData {
		t.Fatalf("expected CLOUD default seed_users=false & seed_sample_data=false, got cloud=%t seed_users=%t seed_sample=%t",
			cfgCloud.IsCloudMode(), cfgCloud.SeedUsers, cfgCloud.SeedSampleData)
	}

	// 4. Edge template edited only to monitoring_domain="CLOUD" (while cloud_mode:false remained) -> still defaults seed=false
	cloudFromTemplatePath := filepath.Join(tmpDir, "config-cloud-from-tpl.json")
	if err := os.WriteFile(cloudFromTemplatePath, []byte(`{
		"monitoring_domain": "CLOUD",
		"cloud_mode": false,
		"seed_users": true,
		"seed_sample_data": true,
		"data_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "data"))+`",
		"log_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "logs"))+`"
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfgCloudTpl, err := config.Load(cloudFromTemplatePath)
	if err != nil {
		t.Fatalf("Load cloud-from-tpl config failed: %v", err)
	}
	if cfgCloudTpl.SeedUsers || cfgCloudTpl.SeedSampleData {
		t.Fatalf("expected template switched to CLOUD to default seed_users=false & seed_sample_data=false, got %t / %t",
			cfgCloudTpl.SeedUsers, cfgCloudTpl.SeedSampleData)
	}

	// 5. Explicit Cloud override (cloud_mode: true + seed_users: true) is honored
	cloudForceOnPath := filepath.Join(tmpDir, "config-cloud-on.json")
	if err := os.WriteFile(cloudForceOnPath, []byte(`{
		"monitoring_domain": "CLOUD",
		"cloud_mode": true,
		"seed_users": true,
		"seed_sample_data": true,
		"data_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "data"))+`",
		"log_dir": "`+filepath.ToSlash(filepath.Join(tmpDir, "logs"))+`"
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfgCloudOn, err := config.Load(cloudForceOnPath)
	if err != nil {
		t.Fatalf("Load cloud-on config failed: %v", err)
	}
	if !cfgCloudOn.SeedUsers || !cfgCloudOn.SeedSampleData {
		t.Fatalf("expected explicit Cloud seed_users=true & seed_sample_data=true to be honored, got %t / %t",
			cfgCloudOn.SeedUsers, cfgCloudOn.SeedSampleData)
	}
}

func TestCloudSeedOffSkipsDemoUsersAndSampleDevicesWhilePreservingPasswords(t *testing.T) {
	resetGlobalConfigCleanup(t)

	cfg := config.Get()
	cfg.MonitoringDomain = "CLOUD"
	cfg.CloudMode = true
	cfg.SeedUsers = false
	cfg.SeedSampleData = false
	cfg.SuperadminUsername = "admin"
	cfg.SuperadminPassword = "Password123!"

	db := openFreshSQLite(t, "cloud_no_seed")
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// 1. Only 1 initial bootstrap SUPERADMIN ("admin") must exist; NO demo accounts!
	var users []model.User
	if err := db.Find(&users).Error; err != nil {
		t.Fatalf("query users failed: %v", err)
	}
	if len(users) != 1 || users[0].Username != "admin" {
		t.Fatalf("expected strictly 1 bootstrap superadmin 'admin', got %d users: %+v", len(users), users)
	}
	for _, demo := range []string{"admin_user", "engineer_user", "operator_user"} {
		var cnt int64
		db.Model(&model.User{}).Where("username = ?", demo).Count(&cnt)
		if cnt != 0 {
			t.Fatalf("expected demo user %q NOT to be seeded when seed_users=false", demo)
		}
	}

	// 2. Zero sample devices (PM-01, TH-01, FM-01) must exist!
	var devCount int64
	db.Model(&model.Device{}).Count(&devCount)
	if devCount != 0 {
		t.Fatalf("expected 0 sample devices when seed_sample_data=false, got %d", devCount)
	}

	// 3. Change 'admin' password and re-run RunMigrations -> custom password MUST NOT be overwritten!
	customHash, _ := bcrypt.GenerateFromPassword([]byte("MyCustomCloudSecret#2026"), bcrypt.DefaultCost)
	if err := db.Model(&model.User{}).Where("username = ?", "admin").Update("password", string(customHash)).Error; err != nil {
		t.Fatalf("update admin password failed: %v", err)
	}
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("second RunMigrations failed: %v", err)
	}
	var updatedAdmin model.User
	if err := db.Where("username = ?", "admin").First(&updatedAdmin).Error; err != nil {
		t.Fatalf("reload admin failed: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedAdmin.Password), []byte("MyCustomCloudSecret#2026")) != nil {
		t.Fatalf("CRITICAL: RunMigrations overwrote modified admin password!")
	}

	// 4. Create custom superadmin "cbi.superadmin", delete "admin", and re-run RunMigrations -> "admin" MUST NOT reappear!
	db.Create(&model.User{
		Username: "cbi.superadmin",
		Email:    "superadmin@cbi.co.id",
		FullName: "CBI Cloud Root",
		Password: string(customHash),
		RoleID:   updatedAdmin.RoleID,
		IsActive: true,
	})
	db.Where("username = ?", "admin").Delete(&model.User{})

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("third RunMigrations failed: %v", err)
	}
	var adminCount int64
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&adminCount)
	if adminCount != 0 {
		t.Fatalf("expected deleted 'admin' NOT to be recreated when seed_users=false and another user exists")
	}

	// Cleanup any config.json generated during one-time migration check
	_ = os.Remove("config.json")
}


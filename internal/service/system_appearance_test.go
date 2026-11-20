package service_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestSaveSystemConfig_GlobalUIAppearancePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	t.Setenv("DATALOGGER_CONFIG", cfgPath)
	cfg := config.DefaultConfig()
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("failed to seed temp config: %v", err)
	}
	if _, err := config.Load(cfgPath); err != nil {
		t.Fatalf("failed to load temp config: %v", err)
	}

	dsn := fmt.Sprintf("file:mem_sys_appearance_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	sysRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	sysSvc := service.NewSystemService(sysRepo, phaseRepo, config.Get())

	preset := "eco-forest"
	layout := "split"
	density := "compact"
	lighting := "dark"

	res, err := sysSvc.SaveSystemConfig(service.SaveSystemConfigRequest{
		UIThemePreset:  &preset,
		UILayout:       &layout,
		UIDensity:      &density,
		UILightingMode: &lighting,
	}, 1, "admin", "127.0.0.1", "go-test")
	if err != nil {
		t.Fatalf("SaveSystemConfig failed: %v", err)
	}
	if !res.Saved || res.RestartRequired {
		t.Fatalf("expected Saved=true RestartRequired=false, got %+v", res)
	}

	gotCfg := sysSvc.GetSystemConfig()
	if gotCfg.UIThemePreset != "eco-forest" || gotCfg.UILayout != "split" || gotCfg.UIDensity != "compact" || gotCfg.UILightingMode != "dark" {
		t.Fatalf("unexpected GetSystemConfig global appearance: %+v", gotCfg)
	}

	// Reload from disk to confirm config.json persisted globally
	reloaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("reload config failed: %v", err)
	}
	if reloaded.UIThemePreset != "eco-forest" || reloaded.UILayout != "split" || reloaded.UIDensity != "compact" || reloaded.UILightingMode != "dark" {
		t.Fatalf("persisted config.json mismatch: %+v", reloaded)
	}
}

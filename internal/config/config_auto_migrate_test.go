package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAutoMigrateConfig_DefaultFalseAndOverrides(t *testing.T) {
	// 1. DefaultConfig must have AutoMigrate == false
	def := DefaultConfig()
	if def.AutoMigrate {
		t.Fatal("expected DefaultConfig().AutoMigrate to be false (default no)")
	}

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	// 2. Omitted in config.json -> stays false
	if err := os.WriteFile(cfgPath, []byte(`{"app_name":"test"}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("AUTO_MIGRATE", "")
	t.Setenv("DATALOGGER_AUTO_MIGRATE", "")
	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.AutoMigrate {
		t.Fatal("expected AutoMigrate=false when omitted in config.json")
	}

	// 3. Explicit true in config.json -> true
	if err := os.WriteFile(cfgPath, []byte(`{"auto_migrate":true}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	loaded, err = Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !loaded.AutoMigrate {
		t.Fatal("expected AutoMigrate=true when explicitly set in config.json")
	}

	// 4. Env AUTO_MIGRATE=no overrides file true -> false
	t.Setenv("AUTO_MIGRATE", "no")
	loaded, err = Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.AutoMigrate {
		t.Fatal("expected AUTO_MIGRATE=no to disable AutoMigrate")
	}

	// 5. Env DATALOGGER_AUTO_MIGRATE=yes enables AutoMigrate even when file is false
	if err := os.WriteFile(cfgPath, []byte(`{"auto_migrate":false}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("AUTO_MIGRATE", "")
	t.Setenv("DATALOGGER_AUTO_MIGRATE", "yes")
	loaded, err = Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !loaded.AutoMigrate {
		t.Fatal("expected DATALOGGER_AUTO_MIGRATE=yes to enable AutoMigrate")
	}
}

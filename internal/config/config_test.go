package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultGlobalMenuVisibilityHidden(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ShowOperatorConsole {
		t.Fatalf("expected ShowOperatorConsole default to be false (hidden), got true")
	}
	if cfg.ShowDevelopmentMenu {
		t.Fatalf("expected ShowDevelopmentMenu default to be false (hidden), got true")
	}

	// Legacy config.json without show_operator_console / show_development_menu must also default to false
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{"monitoring_domain":"AQMS","port":"8080"}`), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load temp config: %v", err)
	}
	if loaded.ShowOperatorConsole || loaded.ShowDevelopmentMenu {
		t.Fatalf("expected legacy config to default both visibility flags to false, got operator=%t dev=%t",
			loaded.ShowOperatorConsole, loaded.ShowDevelopmentMenu)
	}

	// Saving enabled flags persists and reloads cleanly
	loaded.ShowOperatorConsole = true
	loaded.ShowDevelopmentMenu = true
	if err := Save(cfgPath, loaded); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	reloaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if !reloaded.ShowOperatorConsole || !reloaded.ShowDevelopmentMenu {
		t.Fatalf("expected reloaded flags to be true, got operator=%t dev=%t",
			reloaded.ShowOperatorConsole, reloaded.ShowDevelopmentMenu)
	}
}


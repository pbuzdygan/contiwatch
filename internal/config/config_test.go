package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStoreRestrictsExistingConfigPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"scan_interval_sec":60,"global_policy":"notify_only"}`), 0o644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	if _, err := NewStore(path); err != nil {
		t.Fatalf("load config: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected permissions 0600, got %04o", got)
	}
}

func TestStoreUpdateKeepsPrivatePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	if _, err := store.Update(func(cfg *Config) {
		cfg.ScanIntervalSec = 120
	}); err != nil {
		t.Fatalf("update store: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected permissions 0600, got %04o", got)
	}
}

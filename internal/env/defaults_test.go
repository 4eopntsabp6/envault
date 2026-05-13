package env

import (
	"testing"

	"github.com/envault/envault/internal/store"
)

func newDefaultsVault(t *testing.T) *store.Vault {
	t.Helper()
	v, err := store.NewVault(t.TempDir()+"/test.vault", "password")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}
	return v
}

func TestApplyDefaultsSetsNewKeys(t *testing.T) {
	v := newDefaultsVault(t)
	entries := []DefaultEntry{
		{Key: "APP_ENV", DefaultValue: "development"},
		{Key: "LOG_LEVEL", DefaultValue: "info"},
	}
	applied, err := ApplyDefaults(v, entries)
	if err != nil {
		t.Fatalf("ApplyDefaults: %v", err)
	}
	if len(applied) != 2 {
		t.Fatalf("expected 2 applied, got %d", len(applied))
	}
	if val, _ := v.Get("APP_ENV"); val != "development" {
		t.Errorf("APP_ENV = %q, want %q", val, "development")
	}
}

func TestApplyDefaultsSkipsExistingKeys(t *testing.T) {
	v := newDefaultsVault(t)
	v.Set("APP_ENV", "production")
	entries := []DefaultEntry{
		{Key: "APP_ENV", DefaultValue: "development"},
	}
	applied, err := ApplyDefaults(v, entries)
	if err != nil {
		t.Fatalf("ApplyDefaults: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("expected 0 applied, got %d", len(applied))
	}
	if val, _ := v.Get("APP_ENV"); val != "production" {
		t.Errorf("APP_ENV should remain %q, got %q", "production", val)
	}
}

func TestApplyDefaultsInvalidKey(t *testing.T) {
	v := newDefaultsVault(t)
	entries := []DefaultEntry{
		{Key: "bad-key", DefaultValue: "value"},
	}
	_, err := ApplyDefaults(v, entries)
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}
}

func TestLoadDefaultsBasic(t *testing.T) {
	lines := []string{
		"# comment line",
		"",
		"APP_ENV=development # the app environment",
		"LOG_LEVEL=info",
	}
	entries, err := LoadDefaults(lines)
	if err != nil {
		t.Fatalf("LoadDefaults: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "APP_ENV" || entries[0].DefaultValue != "development" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[0].Description != "the app environment" {
		t.Errorf("unexpected description: %q", entries[0].Description)
	}
	if entries[1].Description != "" {
		t.Errorf("expected empty description, got %q", entries[1].Description)
	}
}

func TestLoadDefaultsMalformed(t *testing.T) {
	lines := []string{"NOEQUALS"}
	_, err := LoadDefaults(lines)
	if err == nil {
		t.Fatal("expected error for malformed line, got nil")
	}
}

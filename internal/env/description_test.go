package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newDescriptionVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	v := store.NewVault(filepath.Join(dir, "test.vault"), "password")
	if err := v.Set("API_KEY", "abc123"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := v.Set("DB_PASS", "secret"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	return v
}

func TestDescriptionPath(t *testing.T) {
	p := DescriptionPath("/home/user/.envault/prod.vault")
	want := "/home/user/.envault/prod.vault.descriptions.json"
	if p != want {
		t.Errorf("got %q, want %q", p, want)
	}
}

func TestLoadDescriptionsMissing(t *testing.T) {
	v := newDescriptionVault(t)
	m, err := LoadDescriptions(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Descriptions) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Descriptions)
	}
}

func TestSetAndGetDescription(t *testing.T) {
	v := newDescriptionVault(t)
	if err := SetDescription(v, "API_KEY", "Primary API key for external service"); err != nil {
		t.Fatalf("SetDescription: %v", err)
	}
	desc, err := GetDescription(v, "API_KEY")
	if err != nil {
		t.Fatalf("GetDescription: %v", err)
	}
	if desc != "Primary API key for external service" {
		t.Errorf("got %q", desc)
	}
}

func TestGetDescriptionMissingKey(t *testing.T) {
	v := newDescriptionVault(t)
	desc, err := GetDescription(v, "NONEXISTENT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "" {
		t.Errorf("expected empty string, got %q", desc)
	}
}

func TestSetDescriptionKeyNotInVault(t *testing.T) {
	v := newDescriptionVault(t)
	err := SetDescription(v, "MISSING_KEY", "should fail")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestDeleteDescription(t *testing.T) {
	v := newDescriptionVault(t)
	_ = SetDescription(v, "API_KEY", "to be deleted")
	if err := DeleteDescription(v, "API_KEY"); err != nil {
		t.Fatalf("DeleteDescription: %v", err)
	}
	desc, _ := GetDescription(v, "API_KEY")
	if desc != "" {
		t.Errorf("expected empty after delete, got %q", desc)
	}
}

func TestListDescriptions(t *testing.T) {
	v := newDescriptionVault(t)
	_ = SetDescription(v, "API_KEY", "api key desc")
	_ = SetDescription(v, "DB_PASS", "database password")
	all, err := ListDescriptions(v)
	if err != nil {
		t.Fatalf("ListDescriptions: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 descriptions, got %d", len(all))
	}
	if all["API_KEY"] != "api key desc" {
		t.Errorf("wrong description for API_KEY: %q", all["API_KEY"])
	}
}

func TestDescriptionFilePersists(t *testing.T) {
	v := newDescriptionVault(t)
	_ = SetDescription(v, "DB_PASS", "persistent description")
	p := DescriptionPath(v.Path())
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("description file not created: %v", err)
	}
}

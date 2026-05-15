package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func newCategoryVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path, "password")
	if err := v.Set("DB_HOST", "localhost"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("DB_PASS", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("API_KEY", "abc123"); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestCategoryPath(t *testing.T) {
	path := "/some/dir/my.vault"
	got := CategoryPath(path)
	want := "/some/dir/" + categoryManifestFile
	if got != want {
		t.Errorf("CategoryPath = %q, want %q", got, want)
	}
}

func TestLoadCategoryManifestMissing(t *testing.T) {
	v := newCategoryVault(t)
	m, err := LoadCategoryManifest(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Categories) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Categories)
	}
}

func TestSetAndGetCategory(t *testing.T) {
	v := newCategoryVault(t)
	if err := SetCategory(v, "DB_HOST", "database"); err != nil {
		t.Fatalf("SetCategory: %v", err)
	}
	cat, err := GetCategory(v, "DB_HOST")
	if err != nil {
		t.Fatalf("GetCategory: %v", err)
	}
	if cat != "database" {
		t.Errorf("GetCategory = %q, want %q", cat, "database")
	}
}

func TestGetCategoryMissingKey(t *testing.T) {
	v := newCategoryVault(t)
	cat, err := GetCategory(v, "NONEXISTENT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat != "" {
		t.Errorf("expected empty category, got %q", cat)
	}
}

func TestSetCategoryMissingVaultKey(t *testing.T) {
	v := newCategoryVault(t)
	err := SetCategory(v, "DOES_NOT_EXIST", "infra")
	if err == nil {
		t.Error("expected error for missing vault key, got nil")
	}
}

func TestKeysByCategory(t *testing.T) {
	v := newCategoryVault(t)
	_ = SetCategory(v, "DB_HOST", "database")
	_ = SetCategory(v, "DB_PASS", "database")
	_ = SetCategory(v, "API_KEY", "api")

	keys, err := KeysByCategory(v, "database")
	if err != nil {
		t.Fatalf("KeysByCategory: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "DB_HOST" || keys[1] != "DB_PASS" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestDeleteCategory(t *testing.T) {
	v := newCategoryVault(t)
	_ = SetCategory(v, "DB_HOST", "database")
	if err := DeleteCategory(v, "DB_HOST"); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
	cat, _ := GetCategory(v, "DB_HOST")
	if cat != "" {
		t.Errorf("expected empty category after delete, got %q", cat)
	}
}

func TestCategoryManifestPersists(t *testing.T) {
	v := newCategoryVault(t)
	_ = SetCategory(v, "API_KEY", "api")

	// Re-load manifest from disk
	m, err := LoadCategoryManifest(v.Path())
	if err != nil {
		t.Fatalf("LoadCategoryManifest: %v", err)
	}
	if m.Categories["API_KEY"] != "api" {
		t.Errorf("expected persisted category 'api', got %q", m.Categories["API_KEY"])
	}
	_ = os.Remove(CategoryPath(v.Path()))
}

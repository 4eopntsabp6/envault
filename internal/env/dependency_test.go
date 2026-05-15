package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newDepVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path, "password")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v
}

func TestDependencyPath(t *testing.T) {
	path := "/home/user/.envault/prod.vault"
	got := DependencyPath(path)
	want := "/home/user/.envault/.prod.vault.deps.json"
	if got != want {
		t.Errorf("DependencyPath = %q, want %q", got, want)
	}
}

func TestLoadDependenciesMissing(t *testing.T) {
	v := newDepVault(t)
	m, err := LoadDependencies(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty manifest, got %v", m)
	}
}

func TestSetAndGetDependencies(t *testing.T) {
	v := newDepVault(t)
	v.Set("DB_URL", "postgres://localhost/db")
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PORT", "5432")
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := SetDependencies(v, "DB_URL", []string{"DB_HOST", "DB_PORT"}); err != nil {
		t.Fatalf("SetDependencies: %v", err)
	}

	deps, err := GetDependencies(v, "DB_URL")
	if err != nil {
		t.Fatalf("GetDependencies: %v", err)
	}
	if len(deps) != 2 || deps[0] != "DB_HOST" || deps[1] != "DB_PORT" {
		t.Errorf("unexpected deps: %v", deps)
	}
}

func TestSetDependenciesMissingKey(t *testing.T) {
	v := newDepVault(t)
	err := SetDependencies(v, "NONEXISTENT", []string{})
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestSetDependenciesMissingDepKey(t *testing.T) {
	v := newDepVault(t)
	v.Set("API_KEY", "secret")
	v.Save()
	err := SetDependencies(v, "API_KEY", []string{"MISSING_DEP"})
	if err == nil {
		t.Error("expected error for missing dependency key")
	}
}

func TestCheckMissing(t *testing.T) {
	v := newDepVault(t)
	v.Set("APP_URL", "http://example.com")
	v.Set("BASE_URL", "http://example.com")
	v.Save()

	SetDependencies(v, "APP_URL", []string{"BASE_URL"})

	// Remove BASE_URL to simulate missing dep
	v.Delete("BASE_URL")
	v.Save()

	missing, err := CheckMissing(v, "APP_URL")
	if err != nil {
		t.Fatalf("CheckMissing: %v", err)
	}
	if len(missing) != 1 || missing[0] != "BASE_URL" {
		t.Errorf("expected [BASE_URL], got %v", missing)
	}
}

func TestCheckMissingNoDeps(t *testing.T) {
	v := newDepVault(t)
	v.Set("STANDALONE", "value")
	v.Save()

	missing, err := CheckMissing(v, "STANDALONE")
	if err != nil {
		t.Fatalf("CheckMissing: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing deps, got %v", missing)
	}
}

func TestSaveAndLoadDependencies(t *testing.T) {
	v := newDepVault(t)
	v.Set("X", "1")
	v.Set("Y", "2")
	v.Save()

	SetDependencies(v, "X", []string{"Y"})

	// Verify file exists
	if _, err := os.Stat(DependencyPath(v.Path())); err != nil {
		t.Fatalf("manifest file not created: %v", err)
	}

	m, err := LoadDependencies(v.Path())
	if err != nil {
		t.Fatalf("LoadDependencies: %v", err)
	}
	if len(m["X"]) != 1 || m["X"][0] != "Y" {
		t.Errorf("unexpected manifest: %v", m)
	}
}

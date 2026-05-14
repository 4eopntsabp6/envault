package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newScopeVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path, "password")
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PORT", "5432")
	v.Set("API_KEY", "secret123")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v, path
}

func TestScopePath(t *testing.T) {
	path := "/home/user/.envault/myproject.vault"
	got := ScopePath(path)
	want := "/home/user/.envault/myproject.scopes.json"
	if got != want {
		t.Errorf("ScopePath = %q, want %q", got, want)
	}
}

func TestLoadScopeManifestMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.vault")
	m, err := LoadScopeManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Scopes) != 0 {
		t.Errorf("expected empty scopes, got %d", len(m.Scopes))
	}
}

func TestSetAndApplyScope(t *testing.T) {
	v, path := newScopeVault(t)

	if err := SetScope(v, path, "dev", []string{"DB_HOST", "DB_PORT"}); err != nil {
		t.Fatalf("SetScope: %v", err)
	}

	// Modify values in vault
	v.Set("DB_HOST", "prod-host")
	v.Set("DB_PORT", "5433")

	applied, err := ApplyScope(v, path, "dev", true)
	if err != nil {
		t.Fatalf("ApplyScope: %v", err)
	}
	if len(applied) != 2 {
		t.Errorf("expected 2 applied keys, got %d", len(applied))
	}

	val, _ := v.Get("DB_HOST")
	if val != "localhost" {
		t.Errorf("DB_HOST = %q, want %q", val, "localhost")
	}
	val, _ = v.Get("DB_PORT")
	if val != "5432" {
		t.Errorf("DB_PORT = %q, want %q", val, "5432")
	}
}

func TestApplyScopeSkipsExistingWithoutOverwrite(t *testing.T) {
	v, path := newScopeVault(t)

	if err := SetScope(v, path, "staging", []string{"DB_HOST"}); err != nil {
		t.Fatalf("SetScope: %v", err)
	}
	v.Set("DB_HOST", "staging-host")

	applied, err := ApplyScope(v, path, "staging", false)
	if err != nil {
		t.Fatalf("ApplyScope: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("expected 0 applied keys (no overwrite), got %d", len(applied))
	}
	val, _ := v.Get("DB_HOST")
	if val != "staging-host" {
		t.Errorf("DB_HOST should remain %q, got %q", "staging-host", val)
	}
}

func TestApplyScopeMissing(t *testing.T) {
	v, path := newScopeVault(t)
	_, err := ApplyScope(v, path, "nonexistent", false)
	if err == nil {
		t.Error("expected error for missing scope")
	}
}

func TestListScopes(t *testing.T) {
	v, path := newScopeVault(t)

	_ = SetScope(v, path, "prod", []string{"API_KEY"})
	_ = SetScope(v, path, "dev", []string{"DB_HOST"})
	_ = SetScope(v, path, "staging", []string{"DB_PORT"})

	names, err := ListScopes(path)
	if err != nil {
		t.Fatalf("ListScopes: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(names))
	}
	if names[0] != "dev" || names[1] != "prod" || names[2] != "staging" {
		t.Errorf("unexpected order: %v", names)
	}
}

func TestSetScopeMissingKey(t *testing.T) {
	v, path := newScopeVault(t)
	err := SetScope(v, path, "bad", []string{"NONEXISTENT_KEY"})
	if err == nil {
		t.Error("expected error for missing key in vault")
	}
	_ = os.Remove(ScopePath(path))
}

package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newVisibilityVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")
	v := store.NewVault(vaultPath, "testpassword")
	if err := v.Set("PUBLIC_KEY", "hello"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("PRIVATE_KEY", "world"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("SECRET_KEY", "topsecret"); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestVisibilityPath(t *testing.T) {
	v := newVisibilityVault(t)
	p := VisibilityPath(v)
	if filepath.Base(p) != ".envault_visibility.json" {
		t.Errorf("unexpected filename: %s", filepath.Base(p))
	}
}

func TestLoadVisibilityManifestMissing(t *testing.T) {
	v := newVisibilityVault(t)
	m, err := LoadVisibilityManifest(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(m.Entries) != 0 {
		t.Errorf("expected empty manifest, got %d entries", len(m.Entries))
	}
}

func TestSetAndGetVisibility(t *testing.T) {
	v := newVisibilityVault(t)
	if err := SetVisibility(v, "PUBLIC_KEY", VisibilityPublic); err != nil {
		t.Fatalf("SetVisibility: %v", err)
	}
	level, err := GetVisibility(v, "PUBLIC_KEY")
	if err != nil {
		t.Fatalf("GetVisibility: %v", err)
	}
	if level != VisibilityPublic {
		t.Errorf("expected %q, got %q", VisibilityPublic, level)
	}
}

func TestGetVisibilityDefaultsToPrivate(t *testing.T) {
	v := newVisibilityVault(t)
	level, err := GetVisibility(v, "PRIVATE_KEY")
	if err != nil {
		t.Fatalf("GetVisibility: %v", err)
	}
	if level != VisibilityPrivate {
		t.Errorf("expected default %q, got %q", VisibilityPrivate, level)
	}
}

func TestSetVisibilityInvalidLevel(t *testing.T) {
	v := newVisibilityVault(t)
	err := SetVisibility(v, "PUBLIC_KEY", "invisible")
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestSetVisibilityMissingKey(t *testing.T) {
	v := newVisibilityVault(t)
	err := SetVisibility(v, "NONEXISTENT", VisibilityPublic)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestFilterByVisibility(t *testing.T) {
	v := newVisibilityVault(t)
	_ = SetVisibility(v, "PUBLIC_KEY", VisibilityPublic)
	_ = SetVisibility(v, "SECRET_KEY", VisibilitySecret)
	// PRIVATE_KEY defaults to private

	public, err := FilterByVisibility(v, VisibilityPublic)
	if err != nil {
		t.Fatalf("FilterByVisibility: %v", err)
	}
	if len(public) != 1 || public[0] != "PUBLIC_KEY" {
		t.Errorf("expected [PUBLIC_KEY], got %v", public)
	}

	private, err := FilterByVisibility(v, VisibilityPrivate)
	if err != nil {
		t.Fatalf("FilterByVisibility: %v", err)
	}
	if len(private) != 1 || private[0] != "PRIVATE_KEY" {
		t.Errorf("expected [PRIVATE_KEY], got %v", private)
	}
}

func TestVisibilityManifestPersists(t *testing.T) {
	v := newVisibilityVault(t)
	_ = SetVisibility(v, "SECRET_KEY", VisibilitySecret)

	if _, err := os.Stat(VisibilityPath(v)); err != nil {
		t.Fatalf("manifest file not created: %v", err)
	}

	m, err := LoadVisibilityManifest(v)
	if err != nil {
		t.Fatalf("LoadVisibilityManifest: %v", err)
	}
	if m.Entries["SECRET_KEY"] != VisibilitySecret {
		t.Errorf("expected %q, got %q", VisibilitySecret, m.Entries["SECRET_KEY"])
	}
}

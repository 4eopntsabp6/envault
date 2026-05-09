package expiry_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/envault/internal/expiry"
	"github.com/user/envault/internal/store"
)

func newTestVault(t *testing.T, password string) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path)
	if err := v.Save(password); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v, path
}

func TestManifestPath(t *testing.T) {
	got := expiry.ManifestPath("/tmp/proj.vault")
	want := "/tmp/proj.vault.expiry.json"
	if got != want {
		t.Errorf("ManifestPath = %q, want %q", got, want)
	}
}

func TestLoadManifestMissing(t *testing.T) {
	dir := t.TempDir()
	m, err := expiry.LoadManifest(filepath.Join(dir, "missing.vault"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Entries) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Entries)
	}
}

func TestSetAndLoadExpiry(t *testing.T) {
	_, path := newTestVault(t, "pass")
	if err := expiry.SetExpiry(path, "TOKEN", 5*time.Minute); err != nil {
		t.Fatalf("SetExpiry: %v", err)
	}
	m, err := expiry.LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	entry, ok := m.Entries["TOKEN"]
	if !ok {
		t.Fatal("expected TOKEN entry in manifest")
	}
	if time.Until(entry.ExpiresAt) <= 0 {
		t.Error("expected future expiry time")
	}
}

func TestPurgeExpiredRemovesKeys(t *testing.T) {
	v, path := newTestVault(t, "pass")
	v.Set("STALE", "old-value")
	v.Set("FRESH", "new-value")
	if err := v.Save("pass"); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Set STALE to already-expired time via manifest directly.
	m := &expiry.Manifest{
		Entries: map[string]expiry.Entry{
			"STALE": {Key: "STALE", ExpiresAt: time.Now().Add(-1 * time.Second)},
			"FRESH": {Key: "FRESH", ExpiresAt: time.Now().Add(10 * time.Minute)},
		},
	}
	if err := expiry.SaveManifest(path, m); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}

	purged, err := expiry.PurgeExpired(v, path, "pass")
	if err != nil {
		t.Fatalf("PurgeExpired: %v", err)
	}
	if len(purged) != 1 || purged[0] != "STALE" {
		t.Errorf("purged = %v, want [STALE]", purged)
	}
	if _, ok := v.Get("STALE"); ok {
		t.Error("STALE should have been removed from vault")
	}
	if _, ok := v.Get("FRESH"); !ok {
		t.Error("FRESH should still be in vault")
	}

	// Verify manifest on disk no longer contains STALE.
	m2, _ := expiry.LoadManifest(path)
	if _, found := m2.Entries["STALE"]; found {
		t.Error("STALE should be removed from manifest")
	}
}

func TestPurgeExpiredNothingToDo(t *testing.T) {
	v, path := newTestVault(t, "pass")
	purged, err := expiry.PurgeExpired(v, path, "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(purged) != 0 {
		t.Errorf("expected nothing purged, got %v", purged)
	}
	// Manifest file should not be created when nothing to purge.
	if _, err := os.Stat(expiry.ManifestPath(path)); !os.IsNotExist(err) {
		t.Error("manifest file should not exist when nothing was purged")
	}
}

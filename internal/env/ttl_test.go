package env

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/envault/internal/store"
)

func newTTLVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault("password")
	if err := v.Save(path); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v, path
}

func TestTTLPath(t *testing.T) {
	p := TTLPath("/some/dir/my.vault")
	want := "/some/dir/.my.vault.ttl.json"
	if p != want {
		t.Errorf("got %q, want %q", p, want)
	}
}

func TestLoadTTLManifestMissing(t *testing.T) {
	v, path := newTTLVault(t)
	_ = v
	m, err := LoadTTLManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Entries) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Entries)
	}
}

func TestSetTTLMissingKey(t *testing.T) {
	v, path := newTTLVault(t)
	err := SetTTL(v, path, "MISSING", time.Minute)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestSetAndLoadTTL(t *testing.T) {
	v, path := newTTLVault(t)
	v.Set("MY_KEY", "value")

	if err := SetTTL(v, path, "MY_KEY", 5*time.Minute); err != nil {
		t.Fatalf("SetTTL: %v", err)
	}

	m, err := LoadTTLManifest(path)
	if err != nil {
		t.Fatalf("LoadTTLManifest: %v", err)
	}
	exp, ok := m.Entries["MY_KEY"]
	if !ok {
		t.Fatal("expected MY_KEY in manifest")
	}
	if time.Until(exp) < 4*time.Minute {
		t.Errorf("expiry too soon: %v", exp)
	}
}

func TestPurgeTTLExpiredRemovesKeys(t *testing.T) {
	v, path := newTTLVault(t)
	v.Set("EXPIRED_KEY", "old")
	v.Set("LIVE_KEY", "alive")

	// Set EXPIRED_KEY with a negative TTL so it is already expired.
	if err := SetTTL(v, path, "EXPIRED_KEY", -time.Second); err != nil {
		t.Fatalf("SetTTL expired: %v", err)
	}
	if err := SetTTL(v, path, "LIVE_KEY", 10*time.Minute); err != nil {
		t.Fatalf("SetTTL live: %v", err)
	}

	purged, err := PurgeTTLExpired(v, path)
	if err != nil {
		t.Fatalf("PurgeTTLExpired: %v", err)
	}
	if len(purged) != 1 || purged[0] != "EXPIRED_KEY" {
		t.Errorf("expected [EXPIRED_KEY], got %v", purged)
	}
	if _, ok := v.Get("EXPIRED_KEY"); ok {
		t.Error("EXPIRED_KEY should have been deleted")
	}
	if _, ok := v.Get("LIVE_KEY"); !ok {
		t.Error("LIVE_KEY should still exist")
	}
}

func TestPurgeTTLNothingExpired(t *testing.T) {
	v, path := newTTLVault(t)
	v.Set("KEY", "val")
	if err := SetTTL(v, path, "KEY", time.Hour); err != nil {
		t.Fatalf("SetTTL: %v", err)
	}
	purged, err := PurgeTTLExpired(v, path)
	if err != nil {
		t.Fatalf("PurgeTTLExpired: %v", err)
	}
	if len(purged) != 0 {
		t.Errorf("expected no purged keys, got %v", purged)
	}
}

func TestTTLManifestFilePermissions(t *testing.T) {
	v, path := newTTLVault(t)
	v.Set("K", "v")
	if err := SetTTL(v, path, "K", time.Minute); err != nil {
		t.Fatalf("SetTTL: %v", err)
	}
	info, err := os.Stat(TTLPath(path))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

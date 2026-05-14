package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func newLockVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	v := store.NewVault(filepath.Join(dir, "test.vault"))
	if err := v.Set("API_KEY", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("DB_PASS", "hunter2"); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestLockPath(t *testing.T) {
	path := LockPath("/home/user/.envault/myproject.vault")
	want := "/home/user/.envault/.myproject.vault.lock.json"
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
}

func TestLoadLockManifestMissing(t *testing.T) {
	v := newLockVault(t)
	m, err := LoadLockManifest(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Locked) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Locked)
	}
}

func TestLockKey(t *testing.T) {
	v := newLockVault(t)
	if err := LockKey(v, "API_KEY", "do not rotate"); err != nil {
		t.Fatalf("LockKey failed: %v", err)
	}
	locked, err := IsLocked(v, "API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if !locked {
		t.Error("expected API_KEY to be locked")
	}
}

func TestLockKeyMissing(t *testing.T) {
	v := newLockVault(t)
	err := LockKey(v, "NONEXISTENT", "")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestUnlockKey(t *testing.T) {
	v := newLockVault(t)
	if err := LockKey(v, "DB_PASS", ""); err != nil {
		t.Fatal(err)
	}
	if err := UnlockKey(v, "DB_PASS"); err != nil {
		t.Fatalf("UnlockKey failed: %v", err)
	}
	locked, err := IsLocked(v, "DB_PASS")
	if err != nil {
		t.Fatal(err)
	}
	if locked {
		t.Error("expected DB_PASS to be unlocked")
	}
}

func TestUnlockKeyNotLocked(t *testing.T) {
	v := newLockVault(t)
	err := UnlockKey(v, "API_KEY")
	if err == nil {
		t.Error("expected error when unlocking a non-locked key")
	}
}

func TestListLocked(t *testing.T) {
	v := newLockVault(t)
	if err := LockKey(v, "API_KEY", "reason1"); err != nil {
		t.Fatal(err)
	}
	if err := LockKey(v, "DB_PASS", "reason2"); err != nil {
		t.Fatal(err)
	}
	locked, err := ListLocked(v)
	if err != nil {
		t.Fatal(err)
	}
	if len(locked) != 2 {
		t.Errorf("expected 2 locked keys, got %d", len(locked))
	}
	if locked["API_KEY"].Reason != "reason1" {
		t.Errorf("unexpected reason: %q", locked["API_KEY"].Reason)
	}
}

func TestLockManifestPersists(t *testing.T) {
	v := newLockVault(t)
	if err := LockKey(v, "API_KEY", "persistent"); err != nil {
		t.Fatal(err)
	}
	// reload from disk
	m, err := LoadLockManifest(v.Path())
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := m.Locked["API_KEY"]
	if !ok {
		t.Fatal("API_KEY not found in reloaded manifest")
	}
	if entry.Reason != "persistent" {
		t.Errorf("expected reason %q, got %q", "persistent", entry.Reason)
	}
	if entry.LockedAt.IsZero() {
		t.Error("expected non-zero LockedAt timestamp")
	}
}

func TestLockFilePermissions(t *testing.T) {
	v := newLockVault(t)
	if err := LockKey(v, "API_KEY", ""); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(LockPath(v.Path()))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %v", info.Mode().Perm())
	}
}

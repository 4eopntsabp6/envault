package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/store"
)

func newWatchVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path, "password")
	return v, path
}

func TestWatchPath(t *testing.T) {
	p := WatchPath("/tmp/myproject.vault")
	if !strings.Contains(p, ".myproject.vault.watch") {
		t.Errorf("unexpected watch path: %s", p)
	}
}

func TestFingerprint(t *testing.T) {
	v, path := newWatchVault(t)
	_ = path
	v.Set("KEY_A", "alpha")
	v.Set("KEY_B", "beta")

	fp1, err := Fingerprint(v)
	if err != nil {
		t.Fatalf("Fingerprint error: %v", err)
	}
	if fp1 == "" {
		t.Fatal("expected non-empty fingerprint")
	}

	v.Set("KEY_C", "gamma")
	fp2, err := Fingerprint(v)
	if err != nil {
		t.Fatalf("Fingerprint error: %v", err)
	}
	if fp1 == fp2 {
		t.Error("fingerprint should change after adding a key")
	}
}

func TestSaveAndLoadWatchState(t *testing.T) {
	v, path := newWatchVault(t)
	v.Set("FOO", "bar")

	if err := SaveWatchState(path, v); err != nil {
		t.Fatalf("SaveWatchState: %v", err)
	}

	ws, err := LoadWatchState(path)
	if err != nil {
		t.Fatalf("LoadWatchState: %v", err)
	}
	if ws == nil {
		t.Fatal("expected non-nil watch state")
	}
	if ws.Keys["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got %s", ws.Keys["FOO"])
	}
	if ws.Fingerprint == "" {
		t.Error("expected non-empty fingerprint in saved state")
	}
}

func TestLoadWatchStateMissingFile(t *testing.T) {
	ws, err := LoadWatchState("/tmp/nonexistent_envault_watch_xyz.vault")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if ws != nil {
		t.Error("expected nil watch state for missing file")
	}
}

func TestHasChangedNoState(t *testing.T) {
	v, path := newWatchVault(t)
	v.Set("X", "1")

	changed, err := HasChanged(path, v)
	if err != nil {
		t.Fatalf("HasChanged: %v", err)
	}
	if !changed {
		t.Error("expected HasChanged=true when no watch state exists")
	}
}

func TestHasChangedDetectsModification(t *testing.T) {
	v, path := newWatchVault(t)
	v.Set("X", "1")

	if err := SaveWatchState(path, v); err != nil {
		t.Fatalf("SaveWatchState: %v", err)
	}

	changed, err := HasChanged(path, v)
	if err != nil {
		t.Fatalf("HasChanged: %v", err)
	}
	if changed {
		t.Error("expected HasChanged=false when vault is unchanged")
	}

	v.Set("Y", "2")
	changed, err = HasChanged(path, v)
	if err != nil {
		t.Fatalf("HasChanged: %v", err)
	}
	if !changed {
		t.Error("expected HasChanged=true after adding a key")
	}
}

func TestWatchStateFilePermissions(t *testing.T) {
	v, path := newWatchVault(t)
	v.Set("SECRET", "value")

	if err := SaveWatchState(path, v); err != nil {
		t.Fatalf("SaveWatchState: %v", err)
	}

	info, err := os.Stat(WatchPath(path))
	if err != nil {
		t.Fatalf("stat watch file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %v", info.Mode().Perm())
	}
}

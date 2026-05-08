package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/snapshot"
	"github.com/user/envault/internal/store"
)

func newTestVault(t *testing.T) *store.Vault {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.env")
	v := store.NewVault(path, "testpass")
	v.Set("FOO", "bar")
	v.Set("BAZ", "qux")
	return v
}

func TestTakeSnapshot(t *testing.T) {
	v := newTestVault(t)
	snap, err := snapshot.Take(v, "testpass")
	if err != nil {
		t.Fatalf("Take: %v", err)
	}
	if len(snap.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(snap.Secrets))
	}
	if snap.Secrets["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got %s", snap.Secrets["FOO"])
	}
}

func TestSaveAndLoad(t *testing.T) {
	v := newTestVault(t)
	snap, _ := snapshot.Take(v, "testpass")
	dir := t.TempDir()
	path, err := snapshot.Save(snap, dir)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := snapshot.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Secrets["BAZ"] != "qux" {
		t.Errorf("expected BAZ=qux, got %s", loaded.Secrets["BAZ"])
	}
	if loaded.VaultPath != snap.VaultPath {
		t.Errorf("vault path mismatch")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error loading missing file")
	}
}

func TestDiff(t *testing.T) {
	before := &snapshot.Snapshot{
		Secrets: map[string]string{"FOO": "bar", "OLD": "val"},
	}
	after := &snapshot.Snapshot{
		Secrets: map[string]string{"FOO": "newbar", "NEW": "val2"},
	}
	added, removed, changed := snapshot.Diff(before, after)
	if len(added) != 1 || added[0] != "NEW" {
		t.Errorf("expected NEW added, got %v", added)
	}
	if len(removed) != 1 || removed[0] != "OLD" {
		t.Errorf("expected OLD removed, got %v", removed)
	}
	if len(changed) != 1 || changed[0] != "FOO" {
		t.Errorf("expected FOO changed, got %v", changed)
	}
}

func TestDiffNoChanges(t *testing.T) {
	secrets := map[string]string{"A": "1", "B": "2"}
	before := &snapshot.Snapshot{Secrets: secrets}
	after := &snapshot.Snapshot{Secrets: map[string]string{"A": "1", "B": "2"}}
	added, removed, changed := snapshot.Diff(before, after)
	if len(added)+len(removed)+len(changed) != 0 {
		t.Errorf("expected no diff, got added=%v removed=%v changed=%v", added, removed, changed)
	}
}

func TestSaveCreatesDir(t *testing.T) {
	v := newTestVault(t)
	snap, _ := snapshot.Take(v, "testpass")
	dir := filepath.Join(t.TempDir(), "deep", "snapshots")
	path, err := snapshot.Save(snap, dir)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("snapshot file not created: %v", err)
	}
}

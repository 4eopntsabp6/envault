package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func newFreezeVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	v := store.NewVault(filepath.Join(dir, "test.vault"))
	if err := v.Unlock("password"); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	return v
}

func TestFreezePath(t *testing.T) {
	v := newFreezeVault(t)
	p := FreezePath(v.Path())
	if filepath.Ext(p) != ".json" {
		t.Errorf("expected .json extension, got %s", p)
	}
}

func TestLoadFreezeManifestMissing(t *testing.T) {
	v := newFreezeVault(t)
	m, err := LoadFreezeManifest(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Frozen) != 0 {
		t.Errorf("expected empty manifest, got %d entries", len(m.Frozen))
	}
}

func TestFreezeKey(t *testing.T) {
	v := newFreezeVault(t)
	v.Set("API_KEY", "secret")

	if err := FreezeKey(v, "API_KEY", "do not change"); err != nil {
		t.Fatalf("FreezeKey: %v", err)
	}

	ok, err := IsFrozen(v, "API_KEY")
	if err != nil {
		t.Fatalf("IsFrozen: %v", err)
	}
	if !ok {
		t.Error("expected API_KEY to be frozen")
	}
}

func TestFreezeKeyMissing(t *testing.T) {
	v := newFreezeVault(t)
	err := FreezeKey(v, "MISSING", "")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestUnfreezeKey(t *testing.T) {
	v := newFreezeVault(t)
	v.Set("DB_PASS", "hunter2")
	if err := FreezeKey(v, "DB_PASS", ""); err != nil {
		t.Fatalf("FreezeKey: %v", err)
	}
	if err := UnfreezeKey(v, "DB_PASS"); err != nil {
		t.Fatalf("UnfreezeKey: %v", err)
	}
	ok, err := IsFrozen(v, "DB_PASS")
	if err != nil {
		t.Fatalf("IsFrozen: %v", err)
	}
	if ok {
		t.Error("expected DB_PASS to be unfrozen")
	}
}

func TestUnfreezeKeyNotFrozen(t *testing.T) {
	v := newFreezeVault(t)
	v.Set("TOKEN", "abc")
	err := UnfreezeKey(v, "TOKEN")
	if err == nil {
		t.Error("expected error when unfreezing a non-frozen key")
	}
}

func TestFrozenKeys(t *testing.T) {
	v := newFreezeVault(t)
	v.Set("A", "1")
	v.Set("B", "2")
	v.Set("C", "3")
	FreezeKey(v, "A", "reason a")
	FreezeKey(v, "C", "reason c")

	_, keys, err := FrozenKeys(v)
	if err != nil {
		t.Fatalf("FrozenKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 frozen keys, got %d", len(keys))
	}
}

func TestFreezeManifestPersists(t *testing.T) {
	v := newFreezeVault(t)
	v.Set("PERSIST", "val")
	FreezeKey(v, "PERSIST", "keep")

	// Reload from disk
	m, err := LoadFreezeManifest(v.Path())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := m.Frozen["PERSIST"]
	if !ok {
		t.Fatal("PERSIST not found in reloaded manifest")
	}
	if e.Reason != "keep" {
		t.Errorf("expected reason 'keep', got %q", e.Reason)
	}
	if e.FrozenAt.IsZero() {
		t.Error("FrozenAt should not be zero")
	}
	_ = os.Remove(FreezePath(v.Path()))
}

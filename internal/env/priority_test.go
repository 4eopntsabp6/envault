package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newPriorityVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	v := store.NewVault(filepath.Join(dir, "test.vault"), "pass")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v
}

func TestPriorityPath(t *testing.T) {
	p := PriorityPath("/home/user/.envault/prod.vault")
	if p != "/home/user/.envault/.prod.vault.priority.json" {
		t.Errorf("unexpected path: %s", p)
	}
}

func TestLoadPriorityManifestMissing(t *testing.T) {
	v := newPriorityVault(t)
	m, err := LoadPriorityManifest(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty manifest, got %v", m)
	}
}

func TestSetAndGetPriority(t *testing.T) {
	v := newPriorityVault(t)
	v.Set("API_KEY", "secret")
	v.Set("DEBUG", "true")
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := SetPriority(v, PriorityHigh)("API_KEY"); err != nil {
		t.Fatalf("set priority: %v", err)
	}
	if err := SetPriority(v, PriorityLow)("DEBUG"); err != nil {
		t.Fatalf("set priority: %v", err)
	}

	lvl, err := GetPriority(v, "API_KEY")
	if err != nil {
		t.Fatalf("get priority: %v", err)
	}
	if lvl != PriorityHigh {
		t.Errorf("expected high, got %d", lvl)
	}

	lvl, err = GetPriority(v, "DEBUG")
	if err != nil {
		t.Fatalf("get priority: %v", err)
	}
	if lvl != PriorityLow {
		t.Errorf("expected low, got %d", lvl)
	}
}

func TestGetPriorityDefaultsToNormal(t *testing.T) {
	v := newPriorityVault(t)
	v.Set("FOO", "bar")
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	lvl, err := GetPriority(v, "FOO")
	if err != nil {
		t.Fatalf("get priority: %v", err)
	}
	if lvl != PriorityNormal {
		t.Errorf("expected normal, got %d", lvl)
	}
}

func TestSetPriorityMissingKey(t *testing.T) {
	v := newPriorityVault(t)
	err := SetPriority(v, PriorityHigh)("MISSING")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestKeysByPriority(t *testing.T) {
	v := newPriorityVault(t)
	v.Set("LOW_KEY", "a")
	v.Set("HIGH_KEY", "b")
	v.Set("NORMAL_KEY", "c")
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	_ = SetPriority(v, PriorityHigh)("HIGH_KEY")
	_ = SetPriority(v, PriorityLow)("LOW_KEY")

	keys, err := KeysByPriority(v)
	if err != nil {
		t.Fatalf("keys by priority: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "HIGH_KEY" {
		t.Errorf("first key should be HIGH_KEY, got %s", keys[0])
	}
	if keys[len(keys)-1] != "LOW_KEY" {
		t.Errorf("last key should be LOW_KEY, got %s", keys[len(keys)-1])
	}
}

func TestPriorityManifestPersists(t *testing.T) {
	v := newPriorityVault(t)
	v.Set("K", "v")
	_ = v.Save()
	_ = SetPriority(v, PriorityHigh)("K")

	m, err := LoadPriorityManifest(v.Path())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if m["K"] != PriorityHigh {
		t.Errorf("expected high, got %d", m["K"])
	}
	_ = os.Remove(PriorityPath(v.Path()))
}

package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newReadonlyVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	v := store.NewVault(filepath.Join(dir, "test.vault"), "password")
	if err := v.Set("KEY_A", "value_a"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := v.Set("KEY_B", "value_b"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	return v
}

func TestReadonlyPath(t *testing.T) {
	p := ReadonlyPath("/tmp/foo/my.vault")
	if p != "/tmp/foo/.my.vault.readonly.json" {
		t.Errorf("unexpected path: %s", p)
	}
}

func TestLoadReadonlyManifestMissing(t *testing.T) {
	v := newReadonlyVault(t)
	m, err := LoadReadonlyManifest(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Keys) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Keys)
	}
}

func TestSetAndIsReadonly(t *testing.T) {
	v := newReadonlyVault(t)
	if err := SetReadonly(v, "KEY_A", true); err != nil {
		t.Fatalf("SetReadonly: %v", err)
	}
	ro, err := IsReadonly(v, "KEY_A")
	if err != nil {
		t.Fatalf("IsReadonly: %v", err)
	}
	if !ro {
		t.Error("expected KEY_A to be read-only")
	}
	ro2, _ := IsReadonly(v, "KEY_B")
	if ro2 {
		t.Error("KEY_B should not be read-only")
	}
}

func TestUnsetReadonly(t *testing.T) {
	v := newReadonlyVault(t)
	_ = SetReadonly(v, "KEY_A", true)
	_ = SetReadonly(v, "KEY_A", false)
	ro, err := IsReadonly(v, "KEY_A")
	if err != nil {
		t.Fatalf("IsReadonly: %v", err)
	}
	if ro {
		t.Error("expected KEY_A to no longer be read-only")
	}
}

func TestReadonlyKeys(t *testing.T) {
	v := newReadonlyVault(t)
	_ = SetReadonly(v, "KEY_A", true)
	_ = SetReadonly(v, "KEY_B", true)
	keys, err := ReadonlyKeys(v)
	if err != nil {
		t.Fatalf("ReadonlyKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 readonly keys, got %d", len(keys))
	}
}

func TestGuardReadonly(t *testing.T) {
	v := newReadonlyVault(t)
	_ = SetReadonly(v, "KEY_A", true)
	if err := GuardReadonly(v, "KEY_A"); err == nil {
		t.Error("expected error for read-only key")
	}
	if err := GuardReadonly(v, "KEY_B"); err != nil {
		t.Errorf("expected no error for writable key: %v", err)
	}
}

func TestSetReadonlyMissingKey(t *testing.T) {
	v := newReadonlyVault(t)
	err := SetReadonly(v, "MISSING", true)
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestReadonlyManifestPersists(t *testing.T) {
	v := newReadonlyVault(t)
	_ = SetReadonly(v, "KEY_A", true)
	_, err := os.Stat(ReadonlyPath(v.Path()))
	if err != nil {
		t.Errorf("manifest file not created: %v", err)
	}
}

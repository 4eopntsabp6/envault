package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewVault(t *testing.T) {
	v := NewVault()
	if v.Version != 1 {
		t.Fatalf("expected version 1, got %d", v.Version)
	}
	if v.Entries == nil {
		t.Fatal("expected non-nil entries map")
	}
}

func TestSetAndGet(t *testing.T) {
	v := NewVault()
	v.Set("API_KEY", "encrypted-value")

	val, ok := v.Get("API_KEY")
	if !ok {
		t.Fatal("expected key to be present")
	}
	if val != "encrypted-value" {
		t.Fatalf("unexpected value: %s", val)
	}
}

func TestDelete(t *testing.T) {
	v := NewVault()
	v.Set("TO_DELETE", "some-value")
	v.Delete("TO_DELETE")

	_, ok := v.Get("TO_DELETE")
	if ok {
		t.Fatal("expected key to be absent after deletion")
	}
}

func TestKeys(t *testing.T) {
	v := NewVault()
	v.Set("A", "1")
	v.Set("B", "2")

	keys := v.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	v := NewVault()
	v.Set("SECRET", "ciphertext-abc")

	if err := Save(dir, v); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists with restricted permissions.
	info, err := os.Stat(filepath.Join(dir, vaultFileName))
	if err != nil {
		t.Fatalf("vault file not found: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected file mode 0600, got %v", info.Mode().Perm())
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	val, ok := loaded.Get("SECRET")
	if !ok {
		t.Fatal("expected SECRET key after reload")
	}
	if val != "ciphertext-abc" {
		t.Fatalf("unexpected value after reload: %s", val)
	}
}

func TestLoadVaultNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err != ErrVaultNotFound {
		t.Fatalf("expected ErrVaultNotFound, got %v", err)
	}
}

package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func newPinVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path, "password")
	return v, path
}

func TestPinPath(t *testing.T) {
	path := "/home/user/.envault/myproject.vault"
	got := PinPath(path)
	want := "/home/user/.envault/myproject.pins.json"
	if got != want {
		t.Errorf("PinPath = %q, want %q", got, want)
	}
}

func TestLoadPinManifestMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no.vault")
	m, err := LoadPinManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Pinned) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Pinned)
	}
}

func TestPinKey(t *testing.T) {
	v, path := newPinVault(t)
	v.Set("DB_PASS", "supersecret")

	if err := PinKey(v, path, "DB_PASS"); err != nil {
		t.Fatalf("PinKey: %v", err)
	}

	m, err := LoadPinManifest(path)
	if err != nil {
		t.Fatalf("LoadPinManifest: %v", err)
	}
	if m.Pinned["DB_PASS"] != "supersecret" {
		t.Errorf("expected pinned value %q, got %q", "supersecret", m.Pinned["DB_PASS"])
	}
}

func TestPinKeyMissing(t *testing.T) {
	v, path := newPinVault(t)
	err := PinKey(v, path, "MISSING")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestUnpinKey(t *testing.T) {
	v, path := newPinVault(t)
	v.Set("API_KEY", "abc123")

	if err := PinKey(v, path, "API_KEY"); err != nil {
		t.Fatalf("PinKey: %v", err)
	}
	if err := UnpinKey(path, "API_KEY"); err != nil {
		t.Fatalf("UnpinKey: %v", err)
	}

	pinned, err := IsPinned(path, "API_KEY")
	if err != nil {
		t.Fatalf("IsPinned: %v", err)
	}
	if pinned {
		t.Error("expected key to be unpinned")
	}
}

func TestUnpinKeyNotPinned(t *testing.T) {
	_, path := newPinVault(t)
	err := UnpinKey(path, "GHOST")
	if err == nil {
		t.Fatal("expected error for unpinning non-pinned key")
	}
}

func TestIsPinned(t *testing.T) {
	v, path := newPinVault(t)
	v.Set("TOKEN", "xyz")

	before, _ := IsPinned(path, "TOKEN")
	if before {
		t.Error("key should not be pinned before PinKey")
	}

	_ = PinKey(v, path, "TOKEN")

	after, err := IsPinned(path, "TOKEN")
	if err != nil {
		t.Fatalf("IsPinned: %v", err)
	}
	if !after {
		t.Error("key should be pinned after PinKey")
	}
}

func TestPinManifestPersistence(t *testing.T) {
	v, path := newPinVault(t)
	v.Set("SECRET", "val")
	_ = PinKey(v, path, "SECRET")

	pinFile := PinPath(path)
	if _, err := os.Stat(pinFile); err != nil {
		t.Fatalf("pin file not created: %v", err)
	}

	m, err := LoadPinManifest(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if m.Pinned["SECRET"] != "val" {
		t.Errorf("persisted value mismatch: got %q", m.Pinned["SECRET"])
	}
}

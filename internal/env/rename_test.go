package env

import (
	"testing"

	"github.com/yourusername/envault/internal/store"
)

func newRenameVault(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault("password")
	v.Set("OLD_KEY", "hello")
	v.Set("EXISTING_KEY", "world")
	return v
}

func TestRenameKeySuccess(t *testing.T) {
	v := newRenameVault(t)
	res, err := RenameKey(v, "OLD_KEY", "NEW_KEY", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.OldKey != "OLD_KEY" || res.NewKey != "NEW_KEY" {
		t.Errorf("unexpected result keys: %+v", res)
	}
	if res.Overwrote {
		t.Error("expected Overwrote=false")
	}
	if _, ok := v.Get("OLD_KEY"); ok {
		t.Error("old key should be deleted")
	}
	if val, ok := v.Get("NEW_KEY"); !ok || val != "hello" {
		t.Errorf("new key should hold original value, got %q", val)
	}
}

func TestRenameKeyMissingSource(t *testing.T) {
	v := newRenameVault(t)
	_, err := RenameKey(v, "MISSING", "NEW_KEY", false)
	if err == nil {
		t.Fatal("expected error for missing source key")
	}
}

func TestRenameKeyExistsNoOverwrite(t *testing.T) {
	v := newRenameVault(t)
	_, err := RenameKey(v, "OLD_KEY", "EXISTING_KEY", false)
	if err == nil {
		t.Fatal("expected error when destination exists and overwrite=false")
	}
}

func TestRenameKeyExistsWithOverwrite(t *testing.T) {
	v := newRenameVault(t)
	res, err := RenameKey(v, "OLD_KEY", "EXISTING_KEY", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Overwrote {
		t.Error("expected Overwrote=true")
	}
	if val, ok := v.Get("EXISTING_KEY"); !ok || val != "hello" {
		t.Errorf("expected overwritten value 'hello', got %q", val)
	}
}

func TestRenameKeyInvalidNewKey(t *testing.T) {
	v := newRenameVault(t)
	_, err := RenameKey(v, "OLD_KEY", "bad-key!", false)
	if err == nil {
		t.Fatal("expected error for invalid new key")
	}
}

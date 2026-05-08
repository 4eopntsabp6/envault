package copy_test

import (
	"testing"

	"github.com/user/envault/internal/copy"
	"github.com/user/envault/internal/store"
)

func newVault(t *testing.T, pairs map[string]string) *store.Vault {
	t.Helper()
	v := store.NewVault("password")
	for k, val := range pairs {
		v.Set(k, val)
	}
	return v
}

func TestCopyAllKeys(t *testing.T) {
	src := newVault(t, map[string]string{"A": "1", "B": "2"})
	dst := store.NewVault("password2")

	res, err := copy.Copy(src, dst, copy.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Copied) != 2 {
		t.Errorf("expected 2 copied, got %d", len(res.Copied))
	}
	if v, ok := dst.Get("A"); !ok || v != "1" {
		t.Errorf("expected A=1 in dst")
	}
}

func TestCopySkipsExistingWithoutOverwrite(t *testing.T) {
	src := newVault(t, map[string]string{"KEY": "new"})
	dst := newVault(t, map[string]string{"KEY": "old"})

	res, err := copy.Copy(src, dst, copy.Options{Overwrite: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(res.Skipped))
	}
	if v, _ := dst.Get("KEY"); v != "old" {
		t.Errorf("expected KEY to remain 'old'")
	}
}

func TestCopyOverwrite(t *testing.T) {
	src := newVault(t, map[string]string{"KEY": "new"})
	dst := newVault(t, map[string]string{"KEY": "old"})

	_, err := copy.Copy(src, dst, copy.Options{Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, _ := dst.Get("KEY"); v != "new" {
		t.Errorf("expected KEY to be 'new', got %q", v)
	}
}

func TestCopySelectedKeys(t *testing.T) {
	src := newVault(t, map[string]string{"A": "1", "B": "2", "C": "3"})
	dst := store.NewVault("password2")

	res, err := copy.Copy(src, dst, copy.Options{Keys: []string{"A", "C"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Copied) != 2 {
		t.Errorf("expected 2 copied, got %d", len(res.Copied))
	}
	if _, ok := dst.Get("B"); ok {
		t.Errorf("B should not have been copied")
	}
}

func TestCopyNilVaultReturnsError(t *testing.T) {
	_, err := copy.Copy(nil, store.NewVault("pw"), copy.Options{})
	if err == nil {
		t.Error("expected error for nil src")
	}
}

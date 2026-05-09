package env_test

import (
	"testing"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

func newMergeVault(t *testing.T, pairs map[string]string) *store.Vault {
	t.Helper()
	v := store.NewVault(t.TempDir()+"/vault.enc", "password")
	for k, val := range pairs {
		if err := v.Set(k, val); err != nil {
			t.Fatalf("setup Set(%q): %v", k, err)
		}
	}
	return v
}

func TestMergeAddsNewKeys(t *testing.T) {
	dst := newMergeVault(t, map[string]string{"A": "1"})
	src := newMergeVault(t, map[string]string{"B": "2", "C": "3"})

	res, err := env.MergeVaults(dst, src, env.MergeSkip)
	if err != nil {
		t.Fatalf("MergeVaults: %v", err)
	}
	if len(res.Added) != 2 {
		t.Errorf("expected 2 added, got %d", len(res.Added))
	}
	if len(res.Skipped) != 0 || len(res.Overwritten) != 0 {
		t.Errorf("unexpected skipped/overwritten: %v / %v", res.Skipped, res.Overwritten)
	}
}

func TestMergeSkipsExistingKeys(t *testing.T) {
	dst := newMergeVault(t, map[string]string{"A": "original"})
	src := newMergeVault(t, map[string]string{"A": "new", "B": "2"})

	res, err := env.MergeVaults(dst, src, env.MergeSkip)
	if err != nil {
		t.Fatalf("MergeVaults: %v", err)
	}
	if len(res.Skipped) != 1 || res.Skipped[0] != "A" {
		t.Errorf("expected A to be skipped, got %v", res.Skipped)
	}
	val, _ := dst.Get("A")
	if val != "original" {
		t.Errorf("expected original value, got %q", val)
	}
}

func TestMergeOverwritesExistingKeys(t *testing.T) {
	dst := newMergeVault(t, map[string]string{"A": "original"})
	src := newMergeVault(t, map[string]string{"A": "new"})

	res, err := env.MergeVaults(dst, src, env.MergeOverwrite)
	if err != nil {
		t.Fatalf("MergeVaults: %v", err)
	}
	if len(res.Overwritten) != 1 || res.Overwritten[0] != "A" {
		t.Errorf("expected A to be overwritten, got %v", res.Overwritten)
	}
	val, _ := dst.Get("A")
	if val != "new" {
		t.Errorf("expected new value, got %q", val)
	}
}

func TestMergeEmptySource(t *testing.T) {
	dst := newMergeVault(t, map[string]string{"A": "1"})
	src := newMergeVault(t, map[string]string{})

	res, err := env.MergeVaults(dst, src, env.MergeSkip)
	if err != nil {
		t.Fatalf("MergeVaults: %v", err)
	}
	if len(res.Added)+len(res.Skipped)+len(res.Overwritten) != 0 {
		t.Errorf("expected empty result, got %+v", res)
	}
}

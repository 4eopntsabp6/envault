package env

import (
	"testing"

	"github.com/envault/envault/internal/store"
)

func newPromoteVault(t *testing.T, pairs map[string]string) *store.Vault {
	t.Helper()
	v := store.NewVault("password")
	for k, val := range pairs {
		v.Set(k, val)
	}
	return v
}

func TestPromoteAllKeys(t *testing.T) {
	src := newPromoteVault(t, map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432"})
	dst := store.NewVault("password")

	res, err := PromoteVault(src, dst, PromoteOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Promoted) != 2 {
		t.Errorf("expected 2 promoted, got %d", len(res.Promoted))
	}
	if v, ok := dst.Get("DB_HOST"); !ok || v != "localhost" {
		t.Errorf("expected DB_HOST=localhost in dst")
	}
}

func TestPromoteSkipsExistingWithoutOverwrite(t *testing.T) {
	src := newPromoteVault(t, map[string]string{"API_KEY": "new-value"})
	dst := newPromoteVault(t, map[string]string{"API_KEY": "old-value"})

	res, err := PromoteVault(src, dst, PromoteOptions{Overwrite: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Skipped) != 1 || res.Skipped[0] != "API_KEY" {
		t.Errorf("expected API_KEY to be skipped")
	}
	if v, _ := dst.Get("API_KEY"); v != "old-value" {
		t.Errorf("expected old-value to be preserved")
	}
}

func TestPromoteOverwrite(t *testing.T) {
	src := newPromoteVault(t, map[string]string{"API_KEY": "new-value"})
	dst := newPromoteVault(t, map[string]string{"API_KEY": "old-value"})

	_, err := PromoteVault(src, dst, PromoteOptions{Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, _ := dst.Get("API_KEY"); v != "new-value" {
		t.Errorf("expected new-value after overwrite")
	}
}

func TestPromoteSelectedKeys(t *testing.T) {
	src := newPromoteVault(t, map[string]string{"A": "1", "B": "2", "C": "3"})
	dst := store.NewVault("password")

	res, err := PromoteVault(src, dst, PromoteOptions{Keys: []string{"A", "C"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Promoted) != 2 {
		t.Errorf("expected 2 promoted, got %d", len(res.Promoted))
	}
	if _, ok := dst.Get("B"); ok {
		t.Errorf("B should not have been promoted")
	}
}

func TestPromoteDryRun(t *testing.T) {
	src := newPromoteVault(t, map[string]string{"SECRET": "val"})
	dst := store.NewVault("password")

	res, err := PromoteVault(src, dst, PromoteOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Promoted) != 1 {
		t.Errorf("expected 1 in promoted list even for dry run")
	}
	if _, ok := dst.Get("SECRET"); ok {
		t.Errorf("dry run should not write to dst vault")
	}
}

func TestPromoteMissingSourceKey(t *testing.T) {
	src := store.NewVault("password")
	dst := store.NewVault("password")

	_, err := PromoteVault(src, dst, PromoteOptions{Keys: []string{"MISSING"}})
	if err == nil {
		t.Error("expected error for missing source key")
	}
}

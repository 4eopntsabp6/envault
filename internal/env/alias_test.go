package env

import (
	"testing"

	"github.com/envault/envault/internal/store"
)

func newAliasVault(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault(t.TempDir()+"/test.vault", "password")
	if err := v.Set("DB_HOST", "localhost"); err != nil {
		t.Fatalf("setup Set: %v", err)
	}
	if err := v.Set("DB_PORT", "5432"); err != nil {
		t.Fatalf("setup Set: %v", err)
	}
	return v
}

func TestSetAlias(t *testing.T) {
	v := newAliasVault(t)
	if err := SetAlias(v, "DATABASE_HOST", "DB_HOST"); err != nil {
		t.Fatalf("SetAlias: %v", err)
	}
	aliases := ListAliases(v)
	if target, ok := aliases["DATABASE_HOST"]; !ok || target != "DB_HOST" {
		t.Errorf("expected alias DATABASE_HOST->DB_HOST, got %v", aliases)
	}
}

func TestSetAliasMissingTarget(t *testing.T) {
	v := newAliasVault(t)
	err := SetAlias(v, "GHOST", "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for missing target key")
	}
}

func TestSetAliasInvalidName(t *testing.T) {
	v := newAliasVault(t)
	err := SetAlias(v, "bad-alias", "DB_HOST")
	if err == nil {
		t.Fatal("expected error for invalid alias name")
	}
}

func TestResolveAlias(t *testing.T) {
	v := newAliasVault(t)
	_ = SetAlias(v, "HOST", "DB_HOST")
	val, err := ResolveAlias(v, "HOST")
	if err != nil {
		t.Fatalf("ResolveAlias: %v", err)
	}
	if val != "localhost" {
		t.Errorf("expected localhost, got %q", val)
	}
}

func TestResolveAliasFallback(t *testing.T) {
	v := newAliasVault(t)
	// No alias registered — should fall back to direct key lookup
	val, err := ResolveAlias(v, "DB_PORT")
	if err != nil {
		t.Fatalf("ResolveAlias fallback: %v", err)
	}
	if val != "5432" {
		t.Errorf("expected 5432, got %q", val)
	}
}

func TestDeleteAlias(t *testing.T) {
	v := newAliasVault(t)
	_ = SetAlias(v, "HOST", "DB_HOST")
	if err := DeleteAlias(v, "HOST"); err != nil {
		t.Fatalf("DeleteAlias: %v", err)
	}
	aliases := ListAliases(v)
	if _, ok := aliases["HOST"]; ok {
		t.Error("alias should have been deleted")
	}
}

func TestDeleteAliasMissing(t *testing.T) {
	v := newAliasVault(t)
	err := DeleteAlias(v, "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error deleting non-existent alias")
	}
}

func TestListAliasesEmpty(t *testing.T) {
	v := newAliasVault(t)
	aliases := ListAliases(v)
	if len(aliases) != 0 {
		t.Errorf("expected no aliases, got %v", aliases)
	}
}

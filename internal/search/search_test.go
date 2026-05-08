package search_test

import (
	"testing"

	"github.com/user/envault/internal/search"
	"github.com/user/envault/internal/store"
)

func newPopulatedVault(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault("test-password")
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PORT", "5432")
	v.Set("DB_PASSWORD", "secret")
	v.Set("API_KEY", "abc123")
	v.Set("API_SECRET", "xyz789")
	v.Set("APP_ENV", "production")
	return v
}

func TestByKeyPrefixCaseInsensitive(t *testing.T) {
	v := newPopulatedVault(t)
	results := search.ByKeyPrefix(v, "db_", search.Options{})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestByKeyPrefixCaseSensitive(t *testing.T) {
	v := newPopulatedVault(t)
	results := search.ByKeyPrefix(v, "db_", search.Options{CaseSensitive: true})
	if len(results) != 0 {
		t.Fatalf("expected 0 results for lowercase prefix with CaseSensitive=true, got %d", len(results))
	}
	results = search.ByKeyPrefix(v, "DB_", search.Options{CaseSensitive: true})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestByKeyContains(t *testing.T) {
	v := newPopulatedVault(t)
	results := search.ByKeyContains(v, "SECRET", search.Options{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestShowValues(t *testing.T) {
	v := newPopulatedVault(t)
	results := search.ByKeyPrefix(v, "API_KEY", search.Options{ShowValues: true})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != "abc123" {
		t.Errorf("expected value abc123, got %q", results[0].Value)
	}
}

func TestNoValuesWithoutShowValues(t *testing.T) {
	v := newPopulatedVault(t)
	results := search.ByKeyPrefix(v, "API_KEY", search.Options{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != "" {
		t.Errorf("expected empty value when ShowValues=false, got %q", results[0].Value)
	}
}

func TestNoMatch(t *testing.T) {
	v := newPopulatedVault(t)
	results := search.ByKeyContains(v, "NOTFOUND", search.Options{})
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

package env

import (
	"path/filepath"
	"testing"

	"github.com/envault/envault/internal/store"
)

func newGroupVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.vault")
	v := store.NewVault(path)
	return v, "password"
}

func TestGroupByPrefix(t *testing.T) {
	v, pw := newGroupVault(t)
	keys := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"AWS_KEY":     "abc",
		"AWS_SECRET":  "xyz",
		"NOUNDERSCORE": "val",
	}
	for k, val := range keys {
		if err := v.Set(pw, k, val); err != nil {
			t.Fatalf("Set %s: %v", k, err)
		}
	}

	groups, err := GroupByPrefix(v, pw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(groups["DB"]) != 2 {
		t.Errorf("expected 2 DB keys, got %d", len(groups["DB"]))
	}
	if len(groups["AWS"]) != 2 {
		t.Errorf("expected 2 AWS keys, got %d", len(groups["AWS"]))
	}
	if len(groups["_"]) != 1 {
		t.Errorf("expected 1 ungrouped key, got %d", len(groups["_"]))
	}
}

func TestGroupByPrefixEmpty(t *testing.T) {
	v, pw := newGroupVault(t)
	groups, err := GroupByPrefix(v, pw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected empty groups, got %d", len(groups))
	}
}

func TestGroupNamesSorted(t *testing.T) {
	groups := map[string][]string{
		"ZEBRA": {"ZEBRA_ONE"},
		"ALPHA": {"ALPHA_ONE"},
		"BETA":  {"BETA_ONE"},
	}
	names := GroupNames(groups)
	expected := []string{"ALPHA", "BETA", "ZEBRA"}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("position %d: want %s, got %s", i, expected[i], n)
		}
	}
}

func TestPrefixOf(t *testing.T) {
	cases := []struct {
		key    string
		want   string
	}{
		{"DB_HOST", "DB"},
		{"AWS_REGION_EXTRA", "AWS"},
		{"NOUNDERSCORE", "_"},
		{"_LEADING", "_"},
	}
	for _, tc := range cases {
		got := prefixOf(tc.key)
		if got != tc.want {
			t.Errorf("prefixOf(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

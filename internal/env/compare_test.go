package env

import (
	"path/filepath"
	"testing"

	"github.com/envault/envault/internal/store"
)

func newCompareVault(t *testing.T, pass string, pairs map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.env")
	v := store.NewVault(path, pass)
	for k, val := range pairs {
		v.Set(k, val)
	}
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return path
}

func TestCompareVaultsMatch(t *testing.T) {
	left := newCompareVault(t, "pass", map[string]string{"FOO": "bar", "BAZ": "qux"})
	right := newCompareVault(t, "pass", map[string]string{"FOO": "bar", "BAZ": "qux"})

	results, err := CompareVaults(left, "pass", right, "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Status != "match" {
			t.Errorf("key %s expected match, got %s", r.Key, r.Status)
		}
	}
}

func TestCompareVaultsMismatch(t *testing.T) {
	left := newCompareVault(t, "pass", map[string]string{"FOO": "bar"})
	right := newCompareVault(t, "pass", map[string]string{"FOO": "different"})

	results, err := CompareVaults(left, "pass", right, "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "mismatch" {
		t.Errorf("expected mismatch, got %+v", results)
	}
}

func TestCompareVaultsLeftOnly(t *testing.T) {
	left := newCompareVault(t, "pass", map[string]string{"ONLY_LEFT": "val"})
	right := newCompareVault(t, "pass", map[string]string{})

	results, err := CompareVaults(left, "pass", right, "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "left_only" {
		t.Errorf("expected left_only, got %+v", results)
	}
}

func TestCompareVaultsRightOnly(t *testing.T) {
	left := newCompareVault(t, "pass", map[string]string{})
	right := newCompareVault(t, "pass", map[string]string{"ONLY_RIGHT": "val"})

	results, err := CompareVaults(left, "pass", right, "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "right_only" {
		t.Errorf("expected right_only, got %+v", results)
	}
}

func TestFormatCompareHidesValues(t *testing.T) {
	results := []CompareResult{
		{Key: "SECRET", LeftVal: "abc", RightVal: "xyz", Status: "mismatch"},
	}
	out := FormatCompare(results, false)
	if contains(out, "abc") || contains(out, "xyz") {
		t.Errorf("expected values to be hidden, got: %s", out)
	}
	if !contains(out, "values differ") {
		t.Errorf("expected 'values differ' in output, got: %s", out)
	}
}

func TestFormatCompareShowsValues(t *testing.T) {
	results := []CompareResult{
		{Key: "KEY", LeftVal: "foo", RightVal: "bar", Status: "mismatch"},
	}
	out := FormatCompare(results, true)
	if !contains(out, "foo") || !contains(out, "bar") {
		t.Errorf("expected values in output, got: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

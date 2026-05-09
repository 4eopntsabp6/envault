package lint_test

import (
	"testing"

	"github.com/user/envault/internal/lint"
	"github.com/user/envault/internal/store"
)

func newTestVault(t *testing.T, pairs map[string]string) *store.Vault {
	t.Helper()
	v := store.NewVault("password")
	for k, val := range pairs {
		v.Set(k, val)
	}
	return v
}

func TestRuleEmptyValue(t *testing.T) {
	v := newTestVault(t, map[string]string{"EMPTY_KEY": "", "GOOD_KEY": "value"})
	issues := lint.Run(v, []lint.Rule{lint.RuleEmptyValue})
	if len(issues) != 1 || issues[0].Key != "EMPTY_KEY" {
		t.Fatalf("expected 1 issue for EMPTY_KEY, got %v", issues)
	}
}

func TestRuleKeyUpperCase(t *testing.T) {
	v := newTestVault(t, map[string]string{"lower_key": "val", "UPPER_KEY": "val"})
	issues := lint.Run(v, []lint.Rule{lint.RuleKeyUpperCase})
	if len(issues) != 1 || issues[0].Key != "lower_key" {
		t.Fatalf("expected 1 issue for lower_key, got %v", issues)
	}
}

func TestRuleNoSpacesInKey(t *testing.T) {
	v := newTestVault(t, map[string]string{"KEY WITH SPACE": "val", "GOOD": "val"})
	issues := lint.Run(v, []lint.Rule{lint.RuleNoSpacesInKey})
	if len(issues) != 1 || issues[0].Key != "KEY WITH SPACE" {
		t.Fatalf("expected 1 issue for spaced key, got %v", issues)
	}
}

func TestRuleWeakSecret(t *testing.T) {
	v := newTestVault(t, map[string]string{"DB_PASS": "changeme", "API_KEY": "abc123xyz"})
	issues := lint.Run(v, []lint.Rule{lint.RuleWeakSecret})
	if len(issues) != 1 || issues[0].Key != "DB_PASS" {
		t.Fatalf("expected 1 weak-secret issue, got %v", issues)
	}
}

func TestRunAllDefaultRules(t *testing.T) {
	v := newTestVault(t, map[string]string{
		"GOOD_KEY":  "strong-random-value-xyz",
		"bad_key":   "also fine value",
		"WEAK":      "password",
		"EMPTY_ONE": "",
	})
	issues := lint.Run(v, lint.DefaultRules)
	// bad_key -> upper-case, WEAK -> weak secret, EMPTY_ONE -> empty value
	if len(issues) < 3 {
		t.Fatalf("expected at least 3 issues, got %d: %v", len(issues), issues)
	}
}

func TestRunCleanVault(t *testing.T) {
	v := newTestVault(t, map[string]string{
		"DATABASE_URL": "postgres://user:xK92mQp@host/db",
		"API_SECRET":   "s3cur3-r4nd0m-t0k3n",
	})
	issues := lint.Run(v, lint.DefaultRules)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for clean vault, got %v", issues)
	}
}

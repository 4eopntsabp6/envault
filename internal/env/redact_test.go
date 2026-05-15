package env

import (
	"testing"

	"github.com/envault/envault/internal/store"
)

func newRedactVault(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault(t.TempDir()+"/test.vault", "password")
	return v
}

func TestSetAndIsRedacted(t *testing.T) {
	v := newRedactVault(t)
	v.Set("API_KEY", "supersecret")

	if err := SetRedacted(v, "API_KEY", true); err != nil {
		t.Fatalf("SetRedacted: %v", err)
	}
	if !IsRedacted(v, "API_KEY") {
		t.Error("expected API_KEY to be redacted")
	}
}

func TestUnsetRedacted(t *testing.T) {
	v := newRedactVault(t)
	v.Set("TOKEN", "abc123")
	_ = SetRedacted(v, "TOKEN", true)
	_ = SetRedacted(v, "TOKEN", false)

	if IsRedacted(v, "TOKEN") {
		t.Error("expected TOKEN to not be redacted after unset")
	}
}

func TestMaybeRedact(t *testing.T) {
	v := newRedactVault(t)
	v.Set("SECRET", "mysecret")
	_ = SetRedacted(v, "SECRET", true)

	got := MaybeRedact(v, "SECRET", "mysecret")
	if got != RedactedPlaceholder {
		t.Errorf("expected %q, got %q", RedactedPlaceholder, got)
	}

	v.Set("PLAIN", "visible")
	got2 := MaybeRedact(v, "PLAIN", "visible")
	if got2 != "visible" {
		t.Errorf("expected %q, got %q", "visible", got2)
	}
}

func TestRedactedKeys(t *testing.T) {
	v := newRedactVault(t)
	v.Set("A", "1")
	v.Set("B", "2")
	_ = SetRedacted(v, "A", true)
	_ = SetRedacted(v, "B", true)

	keys := RedactedKeys(v)
	if len(keys) != 2 {
		t.Errorf("expected 2 redacted keys, got %d", len(keys))
	}
}

func TestAutoDetectSensitive(t *testing.T) {
	cases := []struct {
		key      string
		expected bool
	}{
		{"API_KEY", true},
		{"DB_PASSWORD", true},
		{"AUTH_TOKEN", true},
		{"DATABASE_URL", false},
		{"APP_NAME", false},
		{"PRIVATE_KEY", true},
	}
	for _, tc := range cases {
		got := AutoDetectSensitive(tc.key)
		if got != tc.expected {
			t.Errorf("AutoDetectSensitive(%q) = %v, want %v", tc.key, got, tc.expected)
		}
	}
}

func TestRedactedKeysEmpty(t *testing.T) {
	v := newRedactVault(t)
	keys := RedactedKeys(v)
	if len(keys) != 0 {
		t.Errorf("expected 0 redacted keys, got %d", len(keys))
	}
}

package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envault/envault/internal/store"
)

func newTransformVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path)
	if err := v.Set("pass", "MY_KEY", "hello world"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := v.Save("pass"); err != nil {
		t.Fatalf("save: %v", err)
	}
	return v, path
}

func TestTransformUpper(t *testing.T) {
	v, _ := newTransformVault(t)
	result, err := ApplyTransform(v, "pass", "MY_KEY", "upper")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "HELLO WORLD" {
		t.Errorf("expected HELLO WORLD, got %q", result)
	}
	val, _ := v.Get("pass", "MY_KEY")
	if val != "HELLO WORLD" {
		t.Errorf("vault not updated: got %q", val)
	}
}

func TestTransformLower(t *testing.T) {
	v, _ := newTransformVault(t)
	_ = v.Set("pass", "MY_KEY", "HELLO")
	result, err := ApplyTransform(v, "pass", "MY_KEY", "lower")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected hello, got %q", result)
	}
}

func TestTransformTrim(t *testing.T) {
	v, _ := newTransformVault(t)
	_ = v.Set("pass", "MY_KEY", "  spaced  ")
	result, err := ApplyTransform(v, "pass", "MY_KEY", "trim")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "spaced" {
		t.Errorf("expected 'spaced', got %q", result)
	}
}

func TestTransformReverse(t *testing.T) {
	v, _ := newTransformVault(t)
	_ = v.Set("pass", "MY_KEY", "abcde")
	result, err := ApplyTransform(v, "pass", "MY_KEY", "reverse")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "edcba" {
		t.Errorf("expected edcba, got %q", result)
	}
}

func TestTransformUnknown(t *testing.T) {
	v, _ := newTransformVault(t)
	_, err := ApplyTransform(v, "pass", "MY_KEY", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown transform")
	}
}

func TestTransformMissingKey(t *testing.T) {
	v, _ := newTransformVault(t)
	_, err := ApplyTransform(v, "pass", "NO_SUCH_KEY", "upper")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestTransformBase64NonEmpty(t *testing.T) {
	v, _ := newTransformVault(t)
	_ = v.Set("pass", "MY_KEY", "abc")
	result, err := ApplyTransform(v, "pass", "MY_KEY", "base64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" || result == "abc" {
		t.Errorf("expected base64 encoded value, got %q", result)
	}
}

var _ = os.DevNull // suppress unused import

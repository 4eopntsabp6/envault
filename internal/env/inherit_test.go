package env_test

import (
	"os"
	"testing"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

func newInheritVault(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault("password")
	return v
}

func TestInheritFromOSNoPrefix(t *testing.T) {
	t.Setenv("ENVAULT_TEST_FOO", "bar")
	t.Setenv("ENVAULT_TEST_BAZ", "qux")

	v := newInheritVault(t)
	n, err := env.InheritFromOS(v, env.InheritOptions{Prefix: "ENVAULT_TEST_"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n < 2 {
		t.Fatalf("expected at least 2 keys imported, got %d", n)
	}
	val, ok := v.Get("ENVAULT_TEST_FOO")
	if !ok || val != "bar" {
		t.Errorf("expected ENVAULT_TEST_FOO=bar, got %q ok=%v", val, ok)
	}
}

func TestInheritFromOSSkipsExisting(t *testing.T) {
	t.Setenv("ENVAULT_INHERIT_KEY", "from_os")

	v := newInheritVault(t)
	_ = v.Set("ENVAULT_INHERIT_KEY", "original")

	_, err := env.InheritFromOS(v, env.InheritOptions{Prefix: "ENVAULT_INHERIT_"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := v.Get("ENVAULT_INHERIT_KEY")
	if val != "original" {
		t.Errorf("expected original value to be preserved, got %q", val)
	}
}

func TestInheritFromOSOverwrite(t *testing.T) {
	t.Setenv("ENVAULT_OW_KEY", "new_value")

	v := newInheritVault(t)
	_ = v.Set("ENVAULT_OW_KEY", "old_value")

	_, err := env.InheritFromOS(v, env.InheritOptions{Prefix: "ENVAULT_OW_", Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := v.Get("ENVAULT_OW_KEY")
	if val != "new_value" {
		t.Errorf("expected new_value, got %q", val)
	}
}

func TestExportToOS(t *testing.T) {
	v := newInheritVault(t)
	_ = v.Set("ENVAULT_EXPORT_X", "hello")
	_ = v.Set("ENVAULT_EXPORT_Y", "world")

	os.Unsetenv("ENVAULT_EXPORT_X")
	os.Unsetenv("ENVAULT_EXPORT_Y")

	n, err := env.ExportToOS(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 keys exported, got %d", n)
	}
	if got := os.Getenv("ENVAULT_EXPORT_X"); got != "hello" {
		t.Errorf("expected ENVAULT_EXPORT_X=hello, got %q", got)
	}
	if got := os.Getenv("ENVAULT_EXPORT_Y"); got != "world" {
		t.Errorf("expected ENVAULT_EXPORT_Y=world, got %q", got)
	}
}

func TestInheritFromOSEmptyVault(t *testing.T) {
	v := newInheritVault(t)
	n, err := env.InheritFromOS(v, env.InheritOptions{Prefix: "__ENVAULT_NONEXISTENT_PREFIX_XYZ__"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 keys imported, got %d", n)
	}
}

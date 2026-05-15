package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/store"
)

func setupReadonlyVault(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	password := "testpass"
	v := store.NewVault(path, password)
	_ = v.Set("API_KEY", "secret")
	_ = v.Set("DB_URL", "postgres://localhost")
	_ = v.Save()
	return path, password
}

func TestRunReadonlySet(t *testing.T) {
	path, pass := setupReadonlyVault(t)
	var buf bytes.Buffer
	err := RunReadonly([]string{"set", "API_KEY"}, pass, path, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "read-only") {
		t.Errorf("expected confirmation message, got: %s", buf.String())
	}
}

func TestRunReadonlyUnset(t *testing.T) {
	path, pass := setupReadonlyVault(t)
	_ = RunReadonly([]string{"set", "API_KEY"}, pass, path, &bytes.Buffer{})
	var buf bytes.Buffer
	err := RunReadonly([]string{"unset", "API_KEY"}, pass, path, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no longer read-only") {
		t.Errorf("expected unset message, got: %s", buf.String())
	}
}

func TestRunReadonlyList(t *testing.T) {
	path, pass := setupReadonlyVault(t)
	_ = RunReadonly([]string{"set", "API_KEY"}, pass, path, &bytes.Buffer{})
	_ = RunReadonly([]string{"set", "DB_URL"}, pass, path, &bytes.Buffer{})
	var buf bytes.Buffer
	err := RunReadonly([]string{"list"}, pass, path, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "API_KEY") || !strings.Contains(out, "DB_URL") {
		t.Errorf("expected both keys in list, got: %s", out)
	}
}

func TestRunReadonlyListEmpty(t *testing.T) {
	path, pass := setupReadonlyVault(t)
	var buf bytes.Buffer
	err := RunReadonly([]string{"list"}, pass, path, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No read-only keys") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestRunReadonlySetMissingKey(t *testing.T) {
	path, pass := setupReadonlyVault(t)
	err := RunReadonly([]string{"set", "NONEXISTENT"}, pass, path, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestRunReadonlyUnknownSubcommand(t *testing.T) {
	path, pass := setupReadonlyVault(t)
	err := RunReadonly([]string{"freeze", "API_KEY"}, pass, path, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}

func TestRunReadonlyBadPassword(t *testing.T) {
	path, _ := setupReadonlyVault(t)
	err := RunReadonly([]string{"list"}, "wrongpass", path, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for bad password")
	}
}

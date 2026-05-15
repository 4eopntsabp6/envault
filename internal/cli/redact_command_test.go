package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/envault/envault/internal/store"
)

func setupRedactVault(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	password := "redactpass"
	v := store.NewVault(path, password)
	v.Set("API_KEY", "secret123")
	v.Set("DB_PASSWORD", "hunter2")
	v.Set("APP_NAME", "envault")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return path, password
}

func TestRunRedactSet(t *testing.T) {
	path, password := setupRedactVault(t)
	var out bytes.Buffer
	err := RunRedact([]string{"set", path, "API_KEY"}, password, &out)
	if err != nil {
		t.Fatalf("RunRedact set: %v", err)
	}
	if !strings.Contains(out.String(), "redacted: API_KEY") {
		t.Errorf("expected redacted confirmation, got: %s", out.String())
	}
}

func TestRunRedactList(t *testing.T) {
	path, password := setupRedactVault(t)
	_ = RunRedact([]string{"set", path, "API_KEY"}, password, &bytes.Buffer{})

	var out bytes.Buffer
	err := RunRedact([]string{"list", path}, password, &out)
	if err != nil {
		t.Fatalf("RunRedact list: %v", err)
	}
	if !strings.Contains(out.String(), "API_KEY") {
		t.Errorf("expected API_KEY in list, got: %s", out.String())
	}
}

func TestRunRedactUnset(t *testing.T) {
	path, password := setupRedactVault(t)
	_ = RunRedact([]string{"set", path, "API_KEY"}, password, &bytes.Buffer{})

	var out bytes.Buffer
	err := RunRedact([]string{"unset", path, "API_KEY"}, password, &out)
	if err != nil {
		t.Fatalf("RunRedact unset: %v", err)
	}
	if !strings.Contains(out.String(), "unredacted: API_KEY") {
		t.Errorf("expected unredacted confirmation, got: %s", out.String())
	}
}

func TestRunRedactAutoDetect(t *testing.T) {
	path, password := setupRedactVault(t)
	var out bytes.Buffer
	err := RunRedact([]string{"set", path, "--auto"}, password, &out)
	if err != nil {
		t.Fatalf("RunRedact auto: %v", err)
	}
	outStr := out.String()
	if !strings.Contains(outStr, "API_KEY") || !strings.Contains(outStr, "DB_PASSWORD") {
		t.Errorf("expected sensitive keys auto-detected, got: %s", outStr)
	}
	if strings.Contains(outStr, "APP_NAME") {
		t.Errorf("APP_NAME should not be auto-detected as sensitive")
	}
}

func TestRunRedactListEmpty(t *testing.T) {
	path, password := setupRedactVault(t)
	var out bytes.Buffer
	err := RunRedact([]string{"list", path}, password, &out)
	if err != nil {
		t.Fatalf("RunRedact list empty: %v", err)
	}
	if !strings.Contains(out.String(), "no redacted keys") {
		t.Errorf("expected empty list message, got: %s", out.String())
	}
}

func TestRunRedactUnknownSubcommand(t *testing.T) {
	path, password := setupRedactVault(t)
	err := RunRedact([]string{"bogus", path}, password, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}

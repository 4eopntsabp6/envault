package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/envault/envault/internal/store"
)

func setupDefaultsVault(t *testing.T) (vaultPath, password string) {
	t.Helper()
	dir := t.TempDir()
	vaultPath = filepath.Join(dir, "test.vault")
	password = "testpass"
	v, err := store.NewVault(vaultPath, password)
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}
	v.Set("EXISTING_KEY", "already-set")
	if err := v.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return
}

func writeDefaultsFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "defaults-*.env")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	return f.Name()
}

func TestRunDefaultsAppliesNewKeys(t *testing.T) {
	vaultPath, password := setupDefaultsVault(t)
	defFile := writeDefaultsFile(t, "APP_ENV=development\nLOG_LEVEL=info\n")
	var buf bytes.Buffer
	if err := RunDefaults(vaultPath, password, defFile, &buf); err != nil {
		t.Fatalf("RunDefaults: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "APP_ENV") || !strings.Contains(out, "LOG_LEVEL") {
		t.Errorf("expected applied keys in output, got: %s", out)
	}
	v, _ := store.Load(vaultPath, password)
	if val, ok := v.Get("APP_ENV"); !ok || val != "development" {
		t.Errorf("APP_ENV = %q, want %q", val, "development")
	}
}

func TestRunDefaultsSkipsExistingKeys(t *testing.T) {
	vaultPath, password := setupDefaultsVault(t)
	defFile := writeDefaultsFile(t, "EXISTING_KEY=new-value\n")
	var buf bytes.Buffer
	if err := RunDefaults(vaultPath, password, defFile, &buf); err != nil {
		t.Fatalf("RunDefaults: %v", err)
	}
	v, _ := store.Load(vaultPath, password)
	if val, _ := v.Get("EXISTING_KEY"); val != "already-set" {
		t.Errorf("EXISTING_KEY should be unchanged, got %q", val)
	}
	if !strings.Contains(buf.String(), "no new defaults") {
		t.Errorf("expected 'no new defaults' message, got: %s", buf.String())
	}
}

func TestRunDefaultsMissingFile(t *testing.T) {
	vaultPath, password := setupDefaultsVault(t)
	err := RunDefaults(vaultPath, password, "/nonexistent/file.env", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing defaults file, got nil")
	}
}

func TestRunDefaultsListPrintsEntries(t *testing.T) {
	defFile := writeDefaultsFile(t, "APP_ENV=development # env setting\nDEBUG=false\n")
	var buf bytes.Buffer
	if err := RunDefaultsList(defFile, &buf); err != nil {
		t.Fatalf("RunDefaultsList: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "APP_ENV=development") {
		t.Errorf("expected APP_ENV in output, got: %s", out)
	}
	if !strings.Contains(out, "env setting") {
		t.Errorf("expected description in output, got: %s", out)
	}
	if !strings.Contains(out, "DEBUG=false") {
		t.Errorf("expected DEBUG in output, got: %s", out)
	}
}

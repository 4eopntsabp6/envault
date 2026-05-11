package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newTestVault(t *testing.T) *store.Vault {
	t.Helper()
	v, err := store.NewVault("testproject", "testpassword")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}
	return v
}

func TestImportFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")

	content := "FOO=bar\nBAZ=qux\n"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	v := newTestVault(t)
	n, err := ImportFile(envFile, v)
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 imported entries, got %d", n)
	}

	val, ok := v.Get("FOO")
	if !ok || val != "bar" {
		t.Errorf("FOO: got %q, ok=%v", val, ok)
	}
	val, ok = v.Get("BAZ")
	if !ok || val != "qux" {
		t.Errorf("BAZ: got %q, ok=%v", val, ok)
	}
}

func TestImportFileMissing(t *testing.T) {
	v := newTestVault(t)
	_, err := ImportFile("/nonexistent/.env", v)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestImportFileEmpty(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")

	if err := os.WriteFile(envFile, []byte(""), 0600); err != nil {
		t.Fatalf("write empty env file: %v", err)
	}

	v := newTestVault(t)
	n, err := ImportFile(envFile, v)
	if err != nil {
		t.Fatalf("ImportFile on empty file: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 imported entries for empty file, got %d", n)
	}
}

func TestExportFile(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "exported.env")

	v := newTestVault(t)
	v.Set("KEY1", "value1")
	v.Set("KEY2", "value2")

	n, err := ExportFile(outFile, v)
	if err != nil {
		t.Fatalf("ExportFile: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 exported entries, got %d", n)
	}

	// Re-import and verify round-trip
	v2 := newTestVault(t)
	n2, err := ImportFile(outFile, v2)
	if err != nil {
		t.Fatalf("re-ImportFile: %v", err)
	}
	if n2 != 2 {
		t.Errorf("expected 2 re-imported entries, got %d", n2)
	}
	val, ok := v2.Get("KEY1")
	if !ok || val != "value1" {
		t.Errorf("KEY1 round-trip: got %q, ok=%v", val, ok)
	}
}

func TestExportFileEmptyVault(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "empty.env")

	v := newTestVault(t)
	n, err := ExportFile(outFile, v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 entries, got %d", n)
	}
	// File should not have been created
	if _, statErr := os.Stat(outFile); !os.IsNotExist(statErr) {
		t.Error("expected output file to not exist for empty vault")
	}
}

package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/cli"
)

func tempVaultPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "vault.json")
}

func TestRunSetAndGet(t *testing.T) {
	vp := tempVaultPath(t)
	if err := cli.RunSet(vp, "DB_URL", "postgres://localhost/db"); err != nil {
		t.Fatalf("RunSet: %v", err)
	}
	// Capture stdout via redirect is complex; just verify no error on Get.
	if err := cli.RunGet(vp, "DB_URL"); err != nil {
		t.Fatalf("RunGet: %v", err)
	}
}

func TestRunGetMissingKey(t *testing.T) {
	vp := tempVaultPath(t)
	err := cli.RunGet(vp, "MISSING")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestRunDelete(t *testing.T) {
	vp := tempVaultPath(t)
	if err := cli.RunSet(vp, "TOKEN", "abc123"); err != nil {
		t.Fatalf("RunSet: %v", err)
	}
	if err := cli.RunDelete(vp, "TOKEN"); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}
	if err := cli.RunGet(vp, "TOKEN"); err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestRunList(t *testing.T) {
	vp := tempVaultPath(t)
	if err := cli.RunList(vp); err != nil {
		t.Fatalf("RunList on empty vault: %v", err)
	}
	cli.RunSet(vp, "A", "1")
	cli.RunSet(vp, "B", "2")
	if err := cli.RunList(vp); err != nil {
		t.Fatalf("RunList: %v", err)
	}
}

func TestRunExport(t *testing.T) {
	vp := tempVaultPath(t)
	cli.RunSet(vp, "PORT", "8080")
	if err := cli.RunExport(vp, "bash"); err != nil {
		t.Fatalf("RunExport bash: %v", err)
	}
	if err := cli.RunExport(vp, "fish"); err != nil {
		t.Fatalf("RunExport fish: %v", err)
	}
}

func TestRunImport(t *testing.T) {
	vp := tempVaultPath(t)
	envFile := filepath.Join(t.TempDir(), ".env")
	content := "HELLO=world\nFOO=bar\n"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := cli.RunImport(vp, envFile); err != nil {
		t.Fatalf("RunImport: %v", err)
	}
	if err := cli.RunGet(vp, "HELLO"); err != nil {
		t.Fatalf("RunGet after import: %v", err)
	}
}

func TestRunImportMissingFile(t *testing.T) {
	vp := tempVaultPath(t)
	err := cli.RunImport(vp, "/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for missing import file")
	}
}

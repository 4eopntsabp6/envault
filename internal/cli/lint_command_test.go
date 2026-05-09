package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/cli"
	"github.com/user/envault/internal/store"
)

func setupLintVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	v := store.NewVault("lintpass")
	v.Set("GOOD_KEY", "strong-value-abc")
	if err := store.Save(v, path, "lintpass"); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return path
}

func TestRunLintClean(t *testing.T) {
	path := setupLintVault(t)
	var buf bytes.Buffer
	err := cli.RunLint(path, "lintpass", &buf)
	if err != nil {
		t.Fatalf("expected no error for clean vault, got: %v", err)
	}
	if !strings.Contains(buf.String(), "No issues") {
		t.Errorf("expected success message, got: %s", buf.String())
	}
}

func TestRunLintDetectsWeakSecret(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	v := store.NewVault("lintpass")
	v.Set("DB_PASSWORD", "changeme")
	store.Save(v, path, "lintpass")

	var buf bytes.Buffer
	err := cli.RunLint(path, "lintpass", &buf)
	if err == nil {
		t.Fatal("expected error for weak secret, got nil")
	}
	if !strings.Contains(buf.String(), "DB_PASSWORD") {
		t.Errorf("expected DB_PASSWORD in output, got: %s", buf.String())
	}
}

func TestRunLintDetectsEmptyValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	v := store.NewVault("lintpass")
	v.Set("MISSING_VAL", "")
	store.Save(v, path, "lintpass")

	var buf bytes.Buffer
	err := cli.RunLint(path, "lintpass", &buf)
	if err == nil {
		t.Fatal("expected error for empty value")
	}
	if !strings.Contains(buf.String(), "MISSING_VAL") {
		t.Errorf("expected MISSING_VAL in output, got: %s", buf.String())
	}
}

func TestRunLintBadPassword(t *testing.T) {
	path := setupLintVault(t)
	var buf bytes.Buffer
	err := cli.RunLint(path, "wrongpass", &buf)
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
}

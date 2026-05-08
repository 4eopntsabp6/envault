package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/cli"
	"github.com/user/envault/internal/store"
)

func setupSearchVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "search.vault")
	v := store.NewVault("searchpass")
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PORT", "5432")
	v.Set("API_KEY", "abc123")
	v.Set("API_SECRET", "xyz")
	if err := v.Save(path); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return path
}

func TestRunSearchPrefix(t *testing.T) {
	path := setupSearchVault(t)
	var buf bytes.Buffer
	if err := cli.RunSearch(path, "searchpass", "DB_", "prefix", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DB_HOST") || !strings.Contains(out, "DB_PORT") {
		t.Errorf("expected DB keys in output, got: %s", out)
	}
	if strings.Contains(out, "API_KEY") {
		t.Errorf("unexpected API_KEY in output: %s", out)
	}
}

func TestRunSearchContains(t *testing.T) {
	path := setupSearchVault(t)
	var buf bytes.Buffer
	if err := cli.RunSearch(path, "searchpass", "KEY", "contains", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("expected API_KEY in output, got: %s", out)
	}
}

func TestRunSearchShowValues(t *testing.T) {
	path := setupSearchVault(t)
	var buf bytes.Buffer
	if err := cli.RunSearch(path, "searchpass", "API_KEY", "prefix", true, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "abc123") {
		t.Errorf("expected value abc123 in output, got: %s", out)
	}
}

func TestRunSearchNoMatch(t *testing.T) {
	path := setupSearchVault(t)
	var buf bytes.Buffer
	if err := cli.RunSearch(path, "searchpass", "NOTFOUND", "contains", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no matches found") {
		t.Errorf("expected no-match message, got: %s", buf.String())
	}
}

func TestRunSearchInvalidMode(t *testing.T) {
	path := setupSearchVault(t)
	var buf bytes.Buffer
	err := cli.RunSearch(path, "searchpass", "DB", "fuzzy", false, &buf)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "unknown search mode") {
		t.Errorf("unexpected error message: %v", err)
	}
}

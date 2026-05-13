package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/envault/internal/store"
)

func setupRenameVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault("pass")
	v.Set("OLD_KEY", "value1")
	v.Set("KEEP_KEY", "value2")
	if err := v.Save(path); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return path
}

func TestRunRenameSuccess(t *testing.T) {
	path := setupRenameVault(t)
	passwordReader = strings.NewReader("pass\n")
	var buf bytes.Buffer
	if err := RunRename([]string{path, "OLD_KEY", "NEW_KEY"}, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "renamed") {
		t.Errorf("expected 'renamed' in output, got: %s", buf.String())
	}
	v := store.NewVault("pass")
	v.Load(path)
	if _, ok := v.Get("OLD_KEY"); ok {
		t.Error("OLD_KEY should be gone")
	}
	if val, ok := v.Get("NEW_KEY"); !ok || val != "value1" {
		t.Errorf("NEW_KEY should be 'value1', got %q", val)
	}
}

func TestRunRenameMissingKey(t *testing.T) {
	path := setupRenameVault(t)
	passwordReader = strings.NewReader("pass\n")
	var buf bytes.Buffer
	err := RunRename([]string{path, "MISSING", "NEW_KEY"}, &buf)
	if err == nil {
		t.Fatal("expected error for missing source key")
	}
}

func TestRunRenameExistsNoOverwrite(t *testing.T) {
	path := setupRenameVault(t)
	passwordReader = strings.NewReader("pass\n")
	var buf bytes.Buffer
	err := RunRename([]string{path, "OLD_KEY", "KEEP_KEY"}, &buf)
	if err == nil {
		t.Fatal("expected error when destination exists without --overwrite")
	}
}

func TestRunRenameExistsWithOverwrite(t *testing.T) {
	path := setupRenameVault(t)
	passwordReader = strings.NewReader("pass\n")
	var buf bytes.Buffer
	if err := RunRename([]string{path, "OLD_KEY", "KEEP_KEY", "--overwrite"}, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "overwrote") {
		t.Errorf("expected overwrite notice in output: %s", buf.String())
	}
}

func TestRunRenameBadArgs(t *testing.T) {
	var buf bytes.Buffer
	err := RunRename([]string{"only-one"}, &buf)
	if err == nil {
		t.Fatal("expected error for insufficient args")
	}
}

func TestRunRenameBadPassword(t *testing.T) {
	path := setupRenameVault(t)
	passwordReader = strings.NewReader("wrong\n")
	var buf bytes.Buffer
	err := RunRename([]string{path, "OLD_KEY", "NEW_KEY"}, &buf)
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	_ = os.Getenv("")
}

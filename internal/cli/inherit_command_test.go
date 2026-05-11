package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/envault/envault/internal/cli"
	"github.com/envault/envault/internal/store"
)

func setupInheritVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault("pass")
	if err := store.Save(v, path, "pass"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

func TestRunInheritImportsPrefix(t *testing.T) {
	t.Setenv("ENVAULT_CLI_TEST_A", "alpha")
	t.Setenv("ENVAULT_CLI_TEST_B", "beta")

	path := setupInheritVault(t)
	var buf bytes.Buffer

	err := cli.RunInherit(path, "pass", "ENVAULT_CLI_TEST_", false, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Imported") {
		t.Errorf("expected Imported message, got %q", buf.String())
	}

	v, err := store.Load(path, "pass")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	val, ok := v.Get("ENVAULT_CLI_TEST_A")
	if !ok || val != "alpha" {
		t.Errorf("expected ENVAULT_CLI_TEST_A=alpha, got %q ok=%v", val, ok)
	}
}

func TestRunInheritSkipsExistingWithoutOverwrite(t *testing.T) {
	t.Setenv("ENVAULT_CLI_SKIP", "from_os")

	path := setupInheritVault(t)
	v, _ := store.Load(path, "pass")
	_ = v.Set("ENVAULT_CLI_SKIP", "original")
	_ = store.Save(v, path, "pass")

	var buf bytes.Buffer
	_ = cli.RunInherit(path, "pass", "ENVAULT_CLI_SKIP", false, &buf)

	v2, _ := store.Load(path, "pass")
	val, _ := v2.Get("ENVAULT_CLI_SKIP")
	if val != "original" {
		t.Errorf("expected original preserved, got %q", val)
	}
}

func TestRunInheritOverwrite(t *testing.T) {
	t.Setenv("ENVAULT_CLI_OW", "updated")

	path := setupInheritVault(t)
	v, _ := store.Load(path, "pass")
	_ = v.Set("ENVAULT_CLI_OW", "stale")
	_ = store.Save(v, path, "pass")

	var buf bytes.Buffer
	_ = cli.RunInherit(path, "pass", "ENVAULT_CLI_OW", true, &buf)

	v2, _ := store.Load(path, "pass")
	val, _ := v2.Get("ENVAULT_CLI_OW")
	if val != "updated" {
		t.Errorf("expected updated, got %q", val)
	}
}

func TestRunInheritBadPassword(t *testing.T) {
	path := setupInheritVault(t)
	var buf bytes.Buffer
	err := cli.RunInherit(path, "wrong", "", false, &buf)
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestRunExportToOS(t *testing.T) {
	path := setupInheritVault(t)
	v, _ := store.Load(path, "pass")
	_ = v.Set("ENVAULT_EXP_OUT", "exported")
	_ = store.Save(v, path, "pass")

	os.Unsetenv("ENVAULT_EXP_OUT")

	var buf bytes.Buffer
	err := cli.RunExportToOS(path, "pass", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("ENVAULT_EXP_OUT"); got != "exported" {
		t.Errorf("expected ENVAULT_EXP_OUT=exported, got %q", got)
	}
}

package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/store"
)

func setupLabelVault(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")
	password := "labelpass"
	v := store.NewVault(password)
	_ = v.Set("API_KEY", "abc")
	_ = v.Set("DB_PASS", "secret")
	_ = v.Set("PORT", "9090")
	if err := store.Save(v, vaultPath, password); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return vaultPath, password
}

func TestRunLabelSet(t *testing.T) {
	vaultPath, password := setupLabelVault(t)
	var buf bytes.Buffer
	err := RunLabel([]string{"set", "API_KEY", "sensitive,external"}, password, vaultPath, &buf)
	if err != nil {
		t.Fatalf("RunLabel set: %v", err)
	}
	if !strings.Contains(buf.String(), "sensitive") {
		t.Errorf("expected output to contain label name, got: %s", buf.String())
	}
}

func TestRunLabelGet(t *testing.T) {
	vaultPath, password := setupLabelVault(t)
	_ = RunLabel([]string{"set", "API_KEY", "sensitive"}, password, vaultPath, &bytes.Buffer{})

	var buf bytes.Buffer
	err := RunLabel([]string{"get", "API_KEY"}, password, vaultPath, &buf)
	if err != nil {
		t.Fatalf("RunLabel get: %v", err)
	}
	if !strings.Contains(buf.String(), "sensitive") {
		t.Errorf("expected 'sensitive' in output, got: %s", buf.String())
	}
}

func TestRunLabelGetNoLabels(t *testing.T) {
	vaultPath, password := setupLabelVault(t)
	var buf bytes.Buffer
	err := RunLabel([]string{"get", "PORT"}, password, vaultPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no labels") {
		t.Errorf("expected 'no labels' message, got: %s", buf.String())
	}
}

func TestRunLabelFilter(t *testing.T) {
	vaultPath, password := setupLabelVault(t)
	_ = RunLabel([]string{"set", "API_KEY", "sensitive"}, password, vaultPath, &bytes.Buffer{})
	_ = RunLabel([]string{"set", "DB_PASS", "sensitive,db"}, password, vaultPath, &bytes.Buffer{})

	var buf bytes.Buffer
	err := RunLabel([]string{"filter", "sensitive"}, password, vaultPath, &buf)
	if err != nil {
		t.Fatalf("RunLabel filter: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "API_KEY") || !strings.Contains(out, "DB_PASS") {
		t.Errorf("expected both keys in filter output, got: %s", out)
	}
}

func TestRunLabelFilterNoMatch(t *testing.T) {
	vaultPath, password := setupLabelVault(t)
	var buf bytes.Buffer
	err := RunLabel([]string{"filter", "nonexistent"}, password, vaultPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no keys") {
		t.Errorf("expected 'no keys' message, got: %s", buf.String())
	}
}

func TestRunLabelDelete(t *testing.T) {
	vaultPath, password := setupLabelVault(t)
	_ = RunLabel([]string{"set", "API_KEY", "sensitive"}, password, vaultPath, &bytes.Buffer{})

	var buf bytes.Buffer
	err := RunLabel([]string{"delete", "API_KEY"}, password, vaultPath, &buf)
	if err != nil {
		t.Fatalf("RunLabel delete: %v", err)
	}
	if !strings.Contains(buf.String(), "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", buf.String())
	}
}

func TestRunLabelUnknownSubcommand(t *testing.T) {
	vaultPath, password := setupLabelVault(t)
	err := RunLabel([]string{"bogus"}, password, vaultPath, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unknown sub-command")
	}
}

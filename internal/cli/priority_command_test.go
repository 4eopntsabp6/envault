package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/store"
)

func setupPriorityVault(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "test.vault")
	v := store.NewVault(p, "pass")
	v.Set("API_KEY", "secret")
	v.Set("DEBUG", "true")
	v.Set("DB_URL", "postgres://localhost/db")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return p, "pass"
}

func TestRunPrioritySet(t *testing.T) {
	p, pass := setupPriorityVault(t)
	var buf bytes.Buffer
	err := RunPriority([]string{"set", "API_KEY", "high"}, p, pass, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "API_KEY") {
		t.Errorf("expected output to mention API_KEY, got: %s", buf.String())
	}
}

func TestRunPriorityGet(t *testing.T) {
	p, pass := setupPriorityVault(t)
	var buf bytes.Buffer
	_ = RunPriority([]string{"set", "API_KEY", "high"}, p, pass, &buf)
	buf.Reset()
	err := RunPriority([]string{"get", "API_KEY"}, p, pass, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "high") {
		t.Errorf("expected 'high' in output, got: %s", out)
	}
}

func TestRunPriorityList(t *testing.T) {
	p, pass := setupPriorityVault(t)
	var buf bytes.Buffer
	_ = RunPriority([]string{"set", "API_KEY", "high"}, p, pass, &buf)
	_ = RunPriority([]string{"set", "DEBUG", "low"}, p, pass, &buf)
	buf.Reset()
	err := RunPriority([]string{"list"}, p, pass, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("expected API_KEY in list output")
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if lines[0] != "" && !strings.HasPrefix(lines[0], "API_KEY") {
		t.Errorf("expected API_KEY first (highest priority), got: %s", lines[0])
	}
}

func TestRunPriorityNumericLevel(t *testing.T) {
	p, pass := setupPriorityVault(t)
	var buf bytes.Buffer
	err := RunPriority([]string{"set", "DB_URL", "7"}, p, pass, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	buf.Reset()
	_ = RunPriority([]string{"get", "DB_URL"}, p, pass, &buf)
	if !strings.Contains(buf.String(), "7") {
		t.Errorf("expected level 7 in output, got: %s", buf.String())
	}
}

func TestRunPriorityInvalidLevel(t *testing.T) {
	p, pass := setupPriorityVault(t)
	var buf bytes.Buffer
	err := RunPriority([]string{"set", "API_KEY", "ultra"}, p, pass, &buf)
	if err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestRunPriorityMissingKey(t *testing.T) {
	p, pass := setupPriorityVault(t)
	var buf bytes.Buffer
	err := RunPriority([]string{"set", "NONEXISTENT", "high"}, p, pass, &buf)
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestRunPriorityUnknownSubcommand(t *testing.T) {
	p, pass := setupPriorityVault(t)
	var buf bytes.Buffer
	err := RunPriority([]string{"frobnicate"}, p, pass, &buf)
	if err == nil {
		t.Error("expected error for unknown sub-command")
	}
}

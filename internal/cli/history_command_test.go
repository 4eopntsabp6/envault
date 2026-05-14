package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envault/internal/env"
	"github.com/nicholasgasior/envault/internal/store"
)

func setupHistoryVault(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "hist.vault")
	const pass = "histpass"
	v := store.NewVault(p, pass)
	v.Set("DB_URL", "postgres://localhost")
	v.Set("API_KEY", "abc123")
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := env.RecordHistory(v, "set", "DB_URL", "", "postgres://localhost"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := env.RecordHistory(v, "set", "API_KEY", "", "abc123"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := env.RecordHistory(v, "delete", "API_KEY", "abc123", ""); err != nil {
		t.Fatalf("record: %v", err)
	}
	return p, pass
}

func TestRunHistoryAll(t *testing.T) {
	p, pass := setupHistoryVault(t)
	var buf bytes.Buffer
	if err := RunHistory(p, pass, "", &buf); err != nil {
		t.Fatalf("RunHistory: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DB_URL") {
		t.Errorf("expected DB_URL in output, got:\n%s", out)
	}
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("expected API_KEY in output, got:\n%s", out)
	}
	if !strings.Contains(out, "delete") {
		t.Errorf("expected delete action in output, got:\n%s", out)
	}
}

func TestRunHistoryFilterKey(t *testing.T) {
	p, pass := setupHistoryVault(t)
	var buf bytes.Buffer
	if err := RunHistory(p, pass, "API_KEY", &buf); err != nil {
		t.Fatalf("RunHistory: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "DB_URL") {
		t.Errorf("DB_URL should be filtered out, got:\n%s", out)
	}
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("expected API_KEY in filtered output, got:\n%s", out)
	}
}

func TestRunHistoryNoMatch(t *testing.T) {
	p, pass := setupHistoryVault(t)
	var buf bytes.Buffer
	if err := RunHistory(p, pass, "NONEXISTENT", &buf); err != nil {
		t.Fatalf("RunHistory: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "no history for key") {
		t.Errorf("expected no-match message, got:\n%s", out)
	}
}

func TestRunHistoryBadPassword(t *testing.T) {
	p, _ := setupHistoryVault(t)
	var buf bytes.Buffer
	err := RunHistory(p, "wrongpass", "", &buf)
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

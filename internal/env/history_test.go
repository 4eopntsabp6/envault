package env

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nicholasgasior/envault/internal/store"
)

func newHistoryVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "test.vault")
	v := store.NewVault(p, "pass")
	v.Set("KEY1", "val1")
	v.Set("KEY2", "val2")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v
}

func TestHistoryPath(t *testing.T) {
	p := "/tmp/myproject.vault"
	got := HistoryPath(p)
	want := "/tmp/myproject.history.json"
	if got != want {
		t.Errorf("HistoryPath = %q, want %q", got, want)
	}
}

func TestLoadHistoryMissing(t *testing.T) {
	v := newHistoryVault(t)
	m, err := LoadHistory(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Entries) != 0 {
		t.Errorf("expected empty history, got %d entries", len(m.Entries))
	}
}

func TestRecordAndLoadHistory(t *testing.T) {
	v := newHistoryVault(t)

	if err := RecordHistory(v, "set", "KEY1", "", "val1"); err != nil {
		t.Fatalf("RecordHistory: %v", err)
	}
	if err := RecordHistory(v, "set", "KEY2", "", "val2"); err != nil {
		t.Fatalf("RecordHistory: %v", err)
	}

	m, err := LoadHistory(v.Path())
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}
	if m.Entries[0].Key != "KEY1" || m.Entries[0].Action != "set" {
		t.Errorf("unexpected first entry: %+v", m.Entries[0])
	}
	if m.Entries[1].NewValue != "val2" {
		t.Errorf("unexpected new value: %q", m.Entries[1].NewValue)
	}
}

func TestRecordHistoryTimestamp(t *testing.T) {
	v := newHistoryVault(t)
	before := time.Now().UTC()
	if err := RecordHistory(v, "delete", "KEY1", "val1", ""); err != nil {
		t.Fatalf("RecordHistory: %v", err)
	}
	after := time.Now().UTC()

	m, _ := LoadHistory(v.Path())
	ts := m.Entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in range [%v, %v]", ts, before, after)
	}
}

func TestHistoryFilePermissions(t *testing.T) {
	v := newHistoryVault(t)
	if err := RecordHistory(v, "set", "KEY1", "", "secret"); err != nil {
		t.Fatalf("RecordHistory: %v", err)
	}
	info, err := os.Stat(HistoryPath(v.Path()))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %v", info.Mode().Perm())
	}
}

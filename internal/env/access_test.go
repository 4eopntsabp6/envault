package env

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/envault/internal/store"
)

func newAccessVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	v := store.NewVault(filepath.Join(dir, "test.vault"), "password")
	return v
}

func TestAccessPath(t *testing.T) {
	v := newAccessVault(t)
	p := AccessPath(v)
	if p == "" {
		t.Fatal("expected non-empty access path")
	}
	if filepath.Base(p) != ".envault_access.json" {
		t.Fatalf("unexpected access file name: %s", filepath.Base(p))
	}
}

func TestLoadAccessManifestMissing(t *testing.T) {
	v := newAccessVault(t)
	m, err := LoadAccessManifest(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(m.Entries))
	}
}

func TestRecordAccess(t *testing.T) {
	v := newAccessVault(t)

	if err := RecordAccess(v, "API_KEY"); err != nil {
		t.Fatalf("RecordAccess: %v", err)
	}
	if err := RecordAccess(v, "DB_PASS"); err != nil {
		t.Fatalf("RecordAccess: %v", err)
	}
	if err := RecordAccess(v, "API_KEY"); err != nil {
		t.Fatalf("RecordAccess: %v", err)
	}

	m, err := LoadAccessManifest(v)
	if err != nil {
		t.Fatalf("LoadAccessManifest: %v", err)
	}
	if len(m.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(m.Entries))
	}
}

func TestRecordAccessTimestamp(t *testing.T) {
	v := newAccessVault(t)
	before := time.Now().UTC().Add(-time.Second)

	if err := RecordAccess(v, "SECRET"); err != nil {
		t.Fatalf("RecordAccess: %v", err)
	}

	m, err := LoadAccessManifest(v)
	if err != nil {
		t.Fatalf("LoadAccessManifest: %v", err)
	}
	if len(m.Entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	entry := m.Entries[0]
	if entry.AccessedAt.Before(before) {
		t.Errorf("timestamp %v is before test start %v", entry.AccessedAt, before)
	}
}

func TestAccessesForKey(t *testing.T) {
	v := newAccessVault(t)

	_ = RecordAccess(v, "API_KEY")
	_ = RecordAccess(v, "DB_PASS")
	_ = RecordAccess(v, "API_KEY")

	entries, err := AccessesForKey(v, "API_KEY")
	if err != nil {
		t.Fatalf("AccessesForKey: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for API_KEY, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Key != "API_KEY" {
			t.Errorf("unexpected key in result: %s", e.Key)
		}
	}
}

func TestAccessesForKeyNoMatch(t *testing.T) {
	v := newAccessVault(t)
	_ = RecordAccess(v, "DB_PASS")

	entries, err := AccessesForKey(v, "API_KEY")
	if err != nil {
		t.Fatalf("AccessesForKey: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestAccessManifestPersistence(t *testing.T) {
	v := newAccessVault(t)
	_ = RecordAccess(v, "TOKEN")

	p := AccessPath(v)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Fatal("access manifest file was not created")
	}
}

package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func newChecksumVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "test.vault")
	v := store.NewVault(p, "password")
	return v
}

func TestChecksumPath(t *testing.T) {
	p := "/tmp/myproject.vault"
	got := ChecksumPath(p)
	want := "/tmp/.myproject.vault.checksums.json"
	if got != want {
		t.Errorf("ChecksumPath = %q, want %q", got, want)
	}
}

func TestLoadChecksumManifestMissing(t *testing.T) {
	v := newChecksumVault(t)
	m, err := LoadChecksumManifest(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Checksums) != 0 {
		t.Errorf("expected empty checksums, got %v", m.Checksums)
	}
}

func TestRecordAndVerifyChecksums(t *testing.T) {
	v := newChecksumVault(t)
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PASS", "s3cr3t")

	if err := RecordChecksums(v, "password"); err != nil {
		t.Fatalf("RecordChecksums error: %v", err)
	}

	results, err := VerifyChecksums(v, "password")
	if err != nil {
		t.Fatalf("VerifyChecksums error: %v", err)
	}
	for k, status := range results {
		if status != "ok" {
			t.Errorf("key %q: expected ok, got %q", k, status)
		}
	}
}

func TestVerifyChecksumDetectsMismatch(t *testing.T) {
	v := newChecksumVault(t)
	v.Set("API_KEY", "original")

	if err := RecordChecksums(v, "password"); err != nil {
		t.Fatalf("RecordChecksums error: %v", err)
	}

	// Mutate the value after recording
	v.Set("API_KEY", "changed")

	results, err := VerifyChecksums(v, "password")
	if err != nil {
		t.Fatalf("VerifyChecksums error: %v", err)
	}
	if results["API_KEY"] != "mismatch" {
		t.Errorf("expected mismatch for API_KEY, got %q", results["API_KEY"])
	}
}

func TestVerifyChecksumDetectsMissing(t *testing.T) {
	v := newChecksumVault(t)
	v.Set("SECRET", "value")

	if err := RecordChecksums(v, "password"); err != nil {
		t.Fatalf("RecordChecksums error: %v", err)
	}

	// Delete the key from vault
	v.Delete("SECRET")

	results, err := VerifyChecksums(v, "password")
	if err != nil {
		t.Fatalf("VerifyChecksums error: %v", err)
	}
	if results["SECRET"] != "missing" {
		t.Errorf("expected missing for SECRET, got %q", results["SECRET"])
	}
}

func TestChecksumManifestPersistence(t *testing.T) {
	v := newChecksumVault(t)
	v.Set("FOO", "bar")

	if err := RecordChecksums(v, "password"); err != nil {
		t.Fatalf("RecordChecksums error: %v", err)
	}

	// Ensure file exists
	if _, err := os.Stat(ChecksumPath(v.Path())); err != nil {
		t.Errorf("checksum file not created: %v", err)
	}

	// Reload and verify
	m, err := LoadChecksumManifest(v.Path())
	if err != nil {
		t.Fatalf("LoadChecksumManifest error: %v", err)
	}
	if _, ok := m.Checksums["FOO"]; !ok {
		t.Error("expected FOO in loaded manifest")
	}
}

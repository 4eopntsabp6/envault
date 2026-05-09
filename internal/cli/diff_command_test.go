package cli_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/envault/internal/cli"
	"github.com/yourorg/envault/internal/snapshot"
	"github.com/yourorg/envault/internal/store"
)

func setupDiffVault(t *testing.T) (vaultPath, password string) {
	t.Helper()
	dir := t.TempDir()
	vaultPath = filepath.Join(dir, "test.vault")
	password = "diffpass"
	v := store.NewVault(vaultPath, password)
	v.Set("API_KEY", "abc123")
	v.Set("DB_PASS", "secret")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return
}

func writeSnapshot(t *testing.T, vaultPath, name string, secrets map[string]string) {
	t.Helper()
	snapshotDir := filepath.Join(filepath.Dir(vaultPath), ".snapshots")
	if err := os.MkdirAll(snapshotDir, 0700); err != nil {
		t.Fatalf("mkdir snapshots: %v", err)
	}
	snap := &snapshot.Snapshot{Secrets: secrets}
	snapshotFile := fmt.Sprintf("%s/%s.json", snapshotDir, name)
	if err := snapshot.Save(snap, snapshotFile); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}
}

func TestRunDiffNoChanges(t *testing.T) {
	vaultPath, password := setupDiffVault(t)
	writeSnapshot(t, vaultPath, "baseline", map[string]string{
		"API_KEY": "abc123",
		"DB_PASS": "secret",
	})
	var buf bytes.Buffer
	if err := cli.RunDiff(vaultPath, password, "baseline", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "No differences found.\n" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestRunDiffDetectsModified(t *testing.T) {
	vaultPath, password := setupDiffVault(t)
	writeSnapshot(t, vaultPath, "old", map[string]string{
		"API_KEY": "old_value",
		"DB_PASS": "secret",
	})
	var buf bytes.Buffer
	if err := cli.RunDiff(vaultPath, password, "old", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("~ API_KEY")) {
		t.Errorf("expected modified API_KEY in output: %q", buf.String())
	}
}

func TestRunDiffDetectsAdded(t *testing.T) {
	vaultPath, password := setupDiffVault(t)
	writeSnapshot(t, vaultPath, "before", map[string]string{
		"API_KEY": "abc123",
	})
	var buf bytes.Buffer
	if err := cli.RunDiff(vaultPath, password, "before", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("+ DB_PASS")) {
		t.Errorf("expected added DB_PASS in output: %q", buf.String())
	}
}

func TestRunDiffMissingSnapshot(t *testing.T) {
	vaultPath, password := setupDiffVault(t)
	var buf bytes.Buffer
	err := cli.RunDiff(vaultPath, password, "nonexistent", false, &buf)
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

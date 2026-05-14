package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newRollbackVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")
	v := store.NewVault(vaultPath, "password")
	return v, vaultPath
}

func TestRollbackPath(t *testing.T) {
	p := RollbackPath("/home/user/.envault/project.vault")
	expected := "/home/user/.envault/.project.vault.rollback.json"
	if p != expected {
		t.Errorf("expected %q, got %q", expected, p)
	}
}

func TestLoadRollbackMissing(t *testing.T) {
	_, vaultPath := newRollbackVault(t)
	entries, err := LoadRollback(vaultPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(entries))
	}
}

func TestCheckpointAndLoad(t *testing.T) {
	v, vaultPath := newRollbackVault(t)
	v.Set("KEY1", "value1")
	v.Set("KEY2", "value2")

	if err := Checkpoint(v, vaultPath, "before-change"); err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	entries, err := LoadRollback(vaultPath)
	if err != nil {
		t.Fatalf("LoadRollback failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Label != "before-change" {
		t.Errorf("expected label 'before-change', got %q", entries[0].Label)
	}
	if entries[0].Snapshot["KEY1"] != "value1" {
		t.Errorf("snapshot missing KEY1")
	}
}

func TestRollbackRestoresState(t *testing.T) {
	v, vaultPath := newRollbackVault(t)
	v.Set("KEY1", "original")
	v.Set("KEY2", "original2")

	if err := Checkpoint(v, vaultPath, "v1"); err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	v.Set("KEY1", "modified")
	v.Set("KEY3", "new")
	v.Delete("KEY2")

	if err := Rollback(v, vaultPath, "v1"); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	if val, _ := v.Get("KEY1"); val != "original" {
		t.Errorf("expected 'original', got %q", val)
	}
	if val, _ := v.Get("KEY2"); val != "original2" {
		t.Errorf("expected 'original2', got %q", val)
	}
	if _, ok := v.Get("KEY3"); ok {
		t.Error("KEY3 should not exist after rollback")
	}
}

func TestRollbackLatestWhenNoLabel(t *testing.T) {
	v, vaultPath := newRollbackVault(t)
	v.Set("K", "first")
	_ = Checkpoint(v, vaultPath, "cp1")
	v.Set("K", "second")
	_ = Checkpoint(v, vaultPath, "cp2")
	v.Set("K", "third")

	if err := Rollback(v, vaultPath, ""); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if val, _ := v.Get("K"); val != "second" {
		t.Errorf("expected 'second' (latest checkpoint), got %q", val)
	}
}

func TestRollbackNoCheckpoints(t *testing.T) {
	v, vaultPath := newRollbackVault(t)
	err := Rollback(v, vaultPath, "")
	if err == nil {
		t.Error("expected error when no checkpoints exist")
	}
}

func TestRollbackLabelNotFound(t *testing.T) {
	v, vaultPath := newRollbackVault(t)
	v.Set("X", "y")
	_ = Checkpoint(v, vaultPath, "existing")
	err := Rollback(v, vaultPath, "nonexistent")
	if err == nil {
		t.Error("expected error for missing label")
	}
}

func TestRollbackFilePermissions(t *testing.T) {
	v, vaultPath := newRollbackVault(t)
	v.Set("A", "b")
	_ = Checkpoint(v, vaultPath, "perm-test")

	info, err := os.Stat(RollbackPath(vaultPath))
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %v", info.Mode().Perm())
	}
}

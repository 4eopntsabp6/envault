package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/cli"
	"github.com/user/envault/internal/store"
)

func setupSnapshotVault(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.env")
	v := store.NewVault(path, "pass")
	v.Set("KEY1", "val1")
	v.Set("KEY2", "val2")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return path
}

func TestRunSnapshotCreatesFile(t *testing.T) {
	path := setupSnapshotVault(t)
	var buf bytes.Buffer
	if err := cli.RunSnapshot(path, "pass", &buf); err != nil {
		t.Fatalf("RunSnapshot: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Snapshot saved") {
		t.Errorf("expected 'Snapshot saved' in output, got: %s", out)
	}
	if !strings.Contains(out, "2 secrets") {
		t.Errorf("expected '2 secrets' in output, got: %s", out)
	}
}

func TestRunSnapshotDiffNoChanges(t *testing.T) {
	path := setupSnapshotVault(t)
	var buf bytes.Buffer
	_ = cli.RunSnapshot(path, "pass", &buf)
	buf.Reset()
	if err := cli.RunSnapshotDiff(path, "pass", &buf); err != nil {
		t.Fatalf("RunSnapshotDiff: %v", err)
	}
	if !strings.Contains(buf.String(), "No changes") {
		t.Errorf("expected 'No changes', got: %s", buf.String())
	}
}

func TestRunSnapshotDiffDetectsChanges(t *testing.T) {
	path := setupSnapshotVault(t)
	var buf bytes.Buffer
	_ = cli.RunSnapshot(path, "pass", &buf)

	v := store.NewVault(path, "pass")
	_ = v.Load()
	v.Set("KEY1", "changed")
	v.Set("KEY3", "new")
	v.Delete("KEY2")
	_ = v.Save()

	buf.Reset()
	if err := cli.RunSnapshotDiff(path, "pass", &buf); err != nil {
		t.Fatalf("RunSnapshotDiff: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "+ KEY3") {
		t.Errorf("expected added KEY3, got: %s", out)
	}
	if !strings.Contains(out, "- KEY2") {
		t.Errorf("expected removed KEY2, got: %s", out)
	}
	if !strings.Contains(out, "~ KEY1") {
		t.Errorf("expected changed KEY1, got: %s", out)
	}
}

func TestRunSnapshotDiffNoSnapshots(t *testing.T) {
	path := setupSnapshotVault(t)
	var buf bytes.Buffer
	err := cli.RunSnapshotDiff(path, "pass", &buf)
	if err == nil && !strings.Contains(buf.String(), "No snapshots") {
		// either an error or 'No snapshots' message is acceptable
		if !strings.Contains(buf.String(), "No snapshots") {
			t.Errorf("expected 'No snapshots' message, got: %s", buf.String())
		}
	}
}

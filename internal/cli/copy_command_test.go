package cli_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/cli"
	"github.com/user/envault/internal/store"
)

func setupCopyVault(t *testing.T, pairs map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "vault.env")
	v := store.NewVault("pass")
	for k, val := range pairs {
		v.Set(k, val)
	}
	if err := v.Save(p, "pass"); err != nil {
		t.Fatalf("save: %v", err)
	}
	return p
}

func TestRunCopyAllKeys(t *testing.T) {
	src := setupCopyVault(t, map[string]string{"FOO": "bar", "BAZ": "qux"})
	dst := setupCopyVault(t, map[string]string{})

	var buf bytes.Buffer
	if err := cli.RunCopy(&buf, src, "pass", dst, "pass", "", false); err != nil {
		t.Fatalf("RunCopy: %v", err)
	}
	out := buf.String()
	if !contains(out, "Copied 2") {
		t.Errorf("expected 'Copied 2' in output, got: %s", out)
	}
}

func TestRunCopySkipsExisting(t *testing.T) {
	src := setupCopyVault(t, map[string]string{"KEY": "new"})
	dst := setupCopyVault(t, map[string]string{"KEY": "old"})

	var buf bytes.Buffer
	if err := cli.RunCopy(&buf, src, "pass", dst, "pass", "", false); err != nil {
		t.Fatalf("RunCopy: %v", err)
	}
	out := buf.String()
	if !contains(out, "skipped 1") {
		t.Errorf("expected 'skipped 1' in output, got: %s", out)
	}
}

func TestRunCopyOverwrite(t *testing.T) {
	src := setupCopyVault(t, map[string]string{"KEY": "new"})
	dst := setupCopyVault(t, map[string]string{"KEY": "old"})

	var buf bytes.Buffer
	if err := cli.RunCopy(&buf, src, "pass", dst, "pass", "", true); err != nil {
		t.Fatalf("RunCopy: %v", err)
	}
	out := buf.String()
	if !contains(out, "Copied 1") {
		t.Errorf("expected 'Copied 1' in output, got: %s", out)
	}
}

func TestRunCopySelectedKeys(t *testing.T) {
	src := setupCopyVault(t, map[string]string{"A": "1", "B": "2", "C": "3"})
	dst := setupCopyVault(t, map[string]string{})

	var buf bytes.Buffer
	if err := cli.RunCopy(&buf, src, "pass", dst, "pass", "A,C", false); err != nil {
		t.Fatalf("RunCopy: %v", err)
	}
	out := buf.String()
	if !contains(out, "Copied 2") {
		t.Errorf("expected 'Copied 2' in output, got: %s", out)
	}
	if contains(out, "+ B") {
		t.Errorf("B should not appear in output")
	}
}

func TestRunCopyBadSourcePassword(t *testing.T) {
	src := setupCopyVault(t, map[string]string{"K": "v"})
	dst := setupCopyVault(t, map[string]string{})

	var buf bytes.Buffer
	err := cli.RunCopy(&buf, src, "wrong", dst, "pass", "", false)
	if err == nil {
		t.Error("expected error with wrong source password")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}

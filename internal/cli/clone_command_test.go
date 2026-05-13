package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func setupCloneVault(t *testing.T) (srcPath, dstPath string) {
	t.Helper()
	dir := t.TempDir()
	srcPath = filepath.Join(dir, "src.vault")
	dstPath = filepath.Join(dir, "dst.vault")

	v := store.NewVault(srcPath, "pass")
	_ = v.Set("KEY_A", "alpha")
	_ = v.Set("KEY_B", "beta")
	_ = v.Set("KEY_C", "gamma")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	return srcPath, dstPath
}

func TestRunCloneAllKeys(t *testing.T) {
	src, dst := setupCloneVault(t)
	var buf bytes.Buffer
	if err := RunClone([]string{src, dst}, "pass", nil, false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "Copied:  3") {
		t.Errorf("expected 3 copied, got: %s", out)
	}
}

func TestRunCloneSelectedKeys(t *testing.T) {
	src, dst := setupCloneVault(t)
	var buf bytes.Buffer
	if err := RunClone([]string{src, dst}, "pass", []string{"KEY_A", "KEY_C"}, false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "Copied:  2") {
		t.Errorf("expected 2 copied, got: %s", out)
	}
}

func TestRunCloneSkipsExisting(t *testing.T) {
	src, dst := setupCloneVault(t)

	// Pre-populate dst with KEY_A
	v := store.NewVault(dst, "pass")
	_ = v.Set("KEY_A", "existing")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunClone([]string{src, dst}, "pass", nil, false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "Copied:  2") {
		t.Errorf("expected 2 copied, got: %s", out)
	}
	if !contains(out, "Skipped: 1") {
		t.Errorf("expected 1 skipped, got: %s", out)
	}
}

func TestRunCloneOverwrite(t *testing.T) {
	src, dst := setupCloneVault(t)

	v := store.NewVault(dst, "pass")
	_ = v.Set("KEY_A", "old")
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunClone([]string{src, dst}, "pass", nil, true, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "Copied:  3") {
		t.Errorf("expected 3 copied with overwrite, got: %s", out)
	}
}

func TestRunCloneMissingArgs(t *testing.T) {
	var buf bytes.Buffer
	err := RunClone([]string{"only-one"}, "pass", nil, false, &buf)
	if err == nil {
		t.Fatal("expected error for missing destination")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

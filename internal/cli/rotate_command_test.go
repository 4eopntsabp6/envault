package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/envault/internal/cli"
	"github.com/user/envault/internal/store"
)

func setupRotateVault(t *testing.T, password string, secrets map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rotate.vault")
	v := store.NewVault(path, password)
	for k, val := range secrets {
		v.Set(k, val)
	}
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return path
}

func TestRunRotateSuccess(t *testing.T) {
	path := setupRotateVault(t, "oldpass", map[string]string{
		"DB_URL": "postgres://localhost",
		"API_KEY": "abc123",
	})

	var out bytes.Buffer
	err := cli.RunRotate(path, "oldpass", "newpass", &out)
	if err != nil {
		t.Fatalf("RunRotate: %v", err)
	}
	if !strings.Contains(out.String(), "Rotated 2 secret(s)") {
		t.Errorf("unexpected output: %q", out.String())
	}

	v, err := store.Load(path, "newpass")
	if err != nil {
		t.Fatalf("load with new password: %v", err)
	}
	val, ok := v.Get("DB_URL")
	if !ok || val != "postgres://localhost" {
		t.Errorf("DB_URL not preserved after rotation")
	}
}

func TestRunRotateSamePassword(t *testing.T) {
	path := setupRotateVault(t, "pass", map[string]string{"X": "y"})
	var out bytes.Buffer
	err := cli.RunRotate(path, "pass", "pass", &out)
	if err == nil {
		t.Error("expected error for same password")
	}
}

func TestRunRotateWrongOldPassword(t *testing.T) {
	path := setupRotateVault(t, "correct", map[string]string{"K": "v"})
	var out bytes.Buffer
	err := cli.RunRotate(path, "wrong", "newpass", &out)
	if err == nil {
		t.Error("expected error for wrong old password")
	}
}

func TestRunRotateOutputContainsPath(t *testing.T) {
	path := setupRotateVault(t, "old", map[string]string{"FOO": "bar"})
	var out bytes.Buffer
	_ = cli.RunRotate(path, "old", "new", &out)
	if !strings.Contains(out.String(), path) {
		t.Errorf("expected vault path in output, got: %q", out.String())
	}
}

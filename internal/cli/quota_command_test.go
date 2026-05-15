package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func setupQuotaVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "quota.vault")
	v := store.NewVault(path, "secret")
	v.Set("FOO", "bar")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v, path
}

func TestRunQuotaSet(t *testing.T) {
	_, path := setupQuotaVault(t)
	var buf bytes.Buffer
	err := RunQuota([]string{path, "set", "20", "512"}, "secret", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "max_keys=20") {
		t.Errorf("expected max_keys=20 in output, got: %s", out)
	}
	if !strings.Contains(out, "max_value_bytes=512") {
		t.Errorf("expected max_value_bytes=512 in output, got: %s", out)
	}
}

func TestRunQuotaShow(t *testing.T) {
	_, path := setupQuotaVault(t)
	var buf bytes.Buffer
	// First set a quota.
	if err := RunQuota([]string{path, "set", "5", "128"}, "secret", &buf); err != nil {
		t.Fatalf("set: %v", err)
	}
	buf.Reset()
	if err := RunQuota([]string{path, "show"}, "secret", &buf); err != nil {
		t.Fatalf("show: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "max_keys: 5") {
		t.Errorf("expected max_keys: 5, got: %s", out)
	}
	if !strings.Contains(out, "max_value_bytes: 128") {
		t.Errorf("expected max_value_bytes: 128, got: %s", out)
	}
}

func TestRunQuotaShowNoQuota(t *testing.T) {
	_, path := setupQuotaVault(t)
	var buf bytes.Buffer
	if err := RunQuota([]string{path, "show"}, "secret", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no quota configured") {
		t.Errorf("expected no quota message, got: %s", buf.String())
	}
}

func TestRunQuotaUnknownSubcommand(t *testing.T) {
	_, path := setupQuotaVault(t)
	var buf bytes.Buffer
	err := RunQuota([]string{path, "purge"}, "secret", &buf)
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown quota subcommand") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunQuotaMissingArgs(t *testing.T) {
	var buf bytes.Buffer
	err := RunQuota([]string{}, "secret", &buf)
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestRunQuotaInvalidMaxKeys(t *testing.T) {
	_, path := setupQuotaVault(t)
	var buf bytes.Buffer
	err := RunQuota([]string{path, "set", "notanumber", "100"}, "secret", &buf)
	if err == nil {
		t.Fatal("expected error for invalid max-keys")
	}
	if !strings.Contains(err.Error(), "invalid max-keys") {
		t.Errorf("unexpected error: %v", err)
	}
}

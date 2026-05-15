package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func newQuotaVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path, "pass")
	return v, path
}

func TestQuotaPath(t *testing.T) {
	p := QuotaPath("/tmp/myproject.vault")
	if p != "/tmp/myproject.quota.json" {
		t.Fatalf("unexpected quota path: %s", p)
	}
}

func TestLoadQuotaMissing(t *testing.T) {
	_, path := newQuotaVault(t)
	m, err := LoadQuota(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.MaxKeys != 0 || m.MaxValueSize != 0 {
		t.Fatalf("expected zero-value manifest, got %+v", m)
	}
}

func TestSetAndLoadQuota(t *testing.T) {
	_, path := newQuotaVault(t)
	if err := SetQuota(path, 10, 256); err != nil {
		t.Fatalf("SetQuota error: %v", err)
	}
	m, err := LoadQuota(path)
	if err != nil {
		t.Fatalf("LoadQuota error: %v", err)
	}
	if m.MaxKeys != 10 {
		t.Errorf("expected MaxKeys=10, got %d", m.MaxKeys)
	}
	if m.MaxValueSize != 256 {
		t.Errorf("expected MaxValueSize=256, got %d", m.MaxValueSize)
	}
}

func TestCheckQuotaKeyLimit(t *testing.T) {
	v, path := newQuotaVault(t)
	if err := SetQuota(path, 2, 0); err != nil {
		t.Fatalf("SetQuota: %v", err)
	}
	v.Set("KEY1", "val1")
	v.Set("KEY2", "val2")
	err := CheckQuota(v, path, "KEY3", "val3")
	if err == nil {
		t.Fatal("expected quota error, got nil")
	}
	if !strings.Contains(err.Error(), "at most 2 keys") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCheckQuotaAllowsExistingKey(t *testing.T) {
	v, path := newQuotaVault(t)
	if err := SetQuota(path, 1, 0); err != nil {
		t.Fatalf("SetQuota: %v", err)
	}
	v.Set("KEY1", "old")
	if err := CheckQuota(v, path, "KEY1", "new"); err != nil {
		t.Errorf("updating existing key should not trigger quota: %v", err)
	}
}

func TestCheckQuotaValueSize(t *testing.T) {
	v, path := newQuotaVault(t)
	if err := SetQuota(path, 0, 5); err != nil {
		t.Fatalf("SetQuota: %v", err)
	}
	err := CheckQuota(v, path, "KEY", "toolongvalue")
	if err == nil {
		t.Fatal("expected value size error, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds limit") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCheckQuotaNoLimits(t *testing.T) {
	v, path := newQuotaVault(t)
	for i := 0; i < 50; i++ {
		v.Set("KEY", "val")
	}
	if err := CheckQuota(v, path, "NEWKEY", strings.Repeat("x", 1000)); err != nil {
		t.Errorf("no quota set, should not error: %v", err)
	}
}

func TestQuotaFilePermissions(t *testing.T) {
	_, path := newQuotaVault(t)
	if err := SetQuota(path, 5, 128); err != nil {
		t.Fatalf("SetQuota: %v", err)
	}
	info, err := os.Stat(QuotaPath(path))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

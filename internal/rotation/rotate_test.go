package rotation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/rotation"
	"github.com/user/envault/internal/store"
)

func tempVaultPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test.vault")
}

func TestRotateSuccess(t *testing.T) {
	path := tempVaultPath(t)
	v := store.NewVault(path, "oldpass")
	v.Set("KEY1", "value1")
	v.Set("KEY2", "value2")
	if err := v.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	result, err := rotation.Rotate(path, "oldpass", "newpass")
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if result.KeysCount != 2 {
		t.Errorf("expected 2 keys rotated, got %d", result.KeysCount)
	}

	newVault, err := store.Load(path, "newpass")
	if err != nil {
		t.Fatalf("load with new password: %v", err)
	}
	val, ok := newVault.Get("KEY1")
	if !ok || val != "value1" {
		t.Errorf("expected KEY1=value1, got %q ok=%v", val, ok)
	}
}

func TestRotateSamePassword(t *testing.T) {
	path := tempVaultPath(t)
	_, err := rotation.Rotate(path, "same", "same")
	if err == nil {
		t.Error("expected error for same password")
	}
}

func TestRotateWrongOldPassword(t *testing.T) {
	path := tempVaultPath(t)
	v := store.NewVault(path, "correct")
	v.Set("X", "y")
	_ = v.Save()

	_, err := rotation.Rotate(path, "wrong", "newpass")
	if err == nil {
		t.Error("expected error for wrong old password")
	}
}

func TestRotateEmptyVault(t *testing.T) {
	path := tempVaultPath(t)
	v := store.NewVault(path, "pass")
	_ = v.Save()

	_, err := rotation.Rotate(path, "pass", "newpass")
	if err == nil {
		t.Error("expected error for empty vault")
	}
}

func TestRotateMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.vault")
	_, err := rotation.Rotate(path, "old", "new")
	if err == nil {
		t.Error("expected error for missing vault file")
	}
}

func TestRotateOldPasswordNoLongerWorks(t *testing.T) {
	path := tempVaultPath(t)
	v := store.NewVault(path, "oldpass")
	v.Set("SECRET", "mysecret")
	_ = v.Save()

	_, err := rotation.Rotate(path, "oldpass", "newpass")
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}

	_, err = store.Load(path, "oldpass")
	if err == nil {
		t.Error("expected old password to fail after rotation")
	}
	_ = os.Remove(path)
}

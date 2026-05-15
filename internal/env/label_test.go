package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/store"
)

func newLabelVault(t *testing.T) (*store.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")
	v := store.NewVault("password")
	if err := v.Set("API_KEY", "abc123"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("DB_PASS", "secret"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("PORT", "8080"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	return v, vaultPath
}

func TestLabelPath(t *testing.T) {
	p := LabelPath("/tmp/project.vault")
	if p != "/tmp/project.vault.labels.json" {
		t.Errorf("unexpected path: %s", p)
	}
}

func TestLoadLabelManifestMissing(t *testing.T) {
	dir := t.TempDir()
	m, err := LoadLabelManifest(filepath.Join(dir, "missing.vault"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Labels) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Labels)
	}
}

func TestSetAndGetLabels(t *testing.T) {
	v, vaultPath := newLabelVault(t)
	if err := SetLabels(vaultPath, v, "API_KEY", []string{"sensitive", "external"}); err != nil {
		t.Fatalf("SetLabels: %v", err)
	}
	labels, err := GetLabels(vaultPath, "API_KEY")
	if err != nil {
		t.Fatalf("GetLabels: %v", err)
	}
	if len(labels) != 2 || labels[0] != "sensitive" || labels[1] != "external" {
		t.Errorf("unexpected labels: %v", labels)
	}
}

func TestGetLabelsMissingKey(t *testing.T) {
	_, vaultPath := newLabelVault(t)
	labels, err := GetLabels(vaultPath, "NONEXISTENT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if labels != nil {
		t.Errorf("expected nil labels, got %v", labels)
	}
}

func TestSetLabelsKeyNotInVault(t *testing.T) {
	v, vaultPath := newLabelVault(t)
	err := SetLabels(vaultPath, v, "GHOST", []string{"x"})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestFilterByLabel(t *testing.T) {
	v, vaultPath := newLabelVault(t)
	_ = SetLabels(vaultPath, v, "API_KEY", []string{"sensitive"})
	_ = SetLabels(vaultPath, v, "DB_PASS", []string{"sensitive", "db"})
	_ = SetLabels(vaultPath, v, "PORT", []string{"config"})

	matched, err := FilterByLabel(vaultPath, "sensitive")
	if err != nil {
		t.Fatalf("FilterByLabel: %v", err)
	}
	if len(matched) != 2 {
		t.Errorf("expected 2 matches, got %d: %v", len(matched), matched)
	}
}

func TestDeleteLabels(t *testing.T) {
	v, vaultPath := newLabelVault(t)
	_ = SetLabels(vaultPath, v, "API_KEY", []string{"sensitive"})
	if err := DeleteLabels(vaultPath, "API_KEY"); err != nil {
		t.Fatalf("DeleteLabels: %v", err)
	}
	labels, _ := GetLabels(vaultPath, "API_KEY")
	if labels != nil {
		t.Errorf("expected nil after delete, got %v", labels)
	}
}

func TestLabelManifestPersistence(t *testing.T) {
	v, vaultPath := newLabelVault(t)
	_ = SetLabels(vaultPath, v, "PORT", []string{"infra"})

	// Reload from disk
	m, err := LoadLabelManifest(vaultPath)
	if err != nil {
		t.Fatalf("LoadLabelManifest: %v", err)
	}
	if lbls := m.Labels["PORT"]; len(lbls) != 1 || lbls[0] != "infra" {
		t.Errorf("expected [infra], got %v", lbls)
	}
	_ = os.Remove(LabelPath(vaultPath))
}

package tags_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/tags"
)

func tempVaultPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test.vault")
}

func TestManifestPath(t *testing.T) {
	vp := "/home/user/.envault/myproject.vault"
	got := tags.ManifestPath(vp)
	want := "/home/user/.envault/.myproject.vault.tags.json"
	if got != want {
		t.Errorf("ManifestPath = %q, want %q", got, want)
	}
}

func TestLoadManifestMissing(t *testing.T) {
	vp := tempVaultPath(t)
	m, err := tags.LoadManifest(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty manifest, got %v", m)
	}
}

func TestSetAndGetTags(t *testing.T) {
	m := tags.Manifest{}
	tags.SetTags(m, "DB_PASSWORD", []string{"prod", "database"})
	got := tags.GetTags(m, "DB_PASSWORD")
	if len(got) != 2 || got[0] != "database" || got[1] != "prod" {
		t.Errorf("GetTags = %v, want [database prod]", got)
	}
}

func TestFilterByTag(t *testing.T) {
	m := tags.Manifest{}
	tags.SetTags(m, "DB_PASSWORD", []string{"prod", "database"})
	tags.SetTags(m, "API_KEY", []string{"prod"})
	tags.SetTags(m, "DEV_TOKEN", []string{"dev"})

	prod := tags.FilterByTag(m, "prod")
	if len(prod) != 2 {
		t.Fatalf("expected 2 prod keys, got %d: %v", len(prod), prod)
	}
	dev := tags.FilterByTag(m, "dev")
	if len(dev) != 1 || dev[0] != "DEV_TOKEN" {
		t.Errorf("dev keys = %v, want [DEV_TOKEN]", dev)
	}
}

func TestRemoveKey(t *testing.T) {
	m := tags.Manifest{}
	tags.SetTags(m, "SECRET", []string{"prod"})
	tags.RemoveKey(m, "SECRET")
	if _, ok := m["SECRET"]; ok {
		t.Error("expected key to be removed")
	}
}

func TestSaveAndLoad(t *testing.T) {
	vp := tempVaultPath(t)
	m := tags.Manifest{}
	tags.SetTags(m, "DB_URL", []string{"infra", "prod"})
	tags.SetTags(m, "REDIS_URL", []string{"infra"})

	if err := tags.SaveManifest(vp, m); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}
	if _, err := os.Stat(tags.ManifestPath(vp)); err != nil {
		t.Fatalf("manifest file missing: %v", err)
	}

	loaded, err := tags.LoadManifest(vp)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 keys, got %d", len(loaded))
	}
	got := tags.GetTags(loaded, "DB_URL")
	if len(got) != 2 || got[0] != "infra" || got[1] != "prod" {
		t.Errorf("DB_URL tags = %v", got)
	}
}

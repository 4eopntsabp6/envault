package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/store"
)

func newBookmarkVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	v := store.NewVault(filepath.Join(dir, "test.vault"))
	if err := v.Save("password"); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return v
}

func TestBookmarkPath(t *testing.T) {
	v := newBookmarkVault(t)
	p := BookmarkPath(v.Path())
	dir := filepath.Dir(v.Path())
	expected := filepath.Join(dir, ".envault_bookmarks.json")
	if p != expected {
		t.Errorf("got %q, want %q", p, expected)
	}
}

func TestLoadBookmarkManifestMissing(t *testing.T) {
	v := newBookmarkVault(t)
	m, err := LoadBookmarkManifest(v.Path())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Bookmarks) != 0 {
		t.Errorf("expected empty manifest, got %d entries", len(m.Bookmarks))
	}
}

func TestAddAndListBookmarks(t *testing.T) {
	v := newBookmarkVault(t)

	if err := AddBookmark(v, "prod", "/vaults/prod.vault", "production vault"); err != nil {
		t.Fatalf("AddBookmark: %v", err)
	}
	if err := AddBookmark(v, "staging", "/vaults/staging.vault", ""); err != nil {
		t.Fatalf("AddBookmark: %v", err)
	}

	names, m, err := ListBookmarks(v)
	if err != nil {
		t.Fatalf("ListBookmarks: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 bookmarks, got %d", len(names))
	}
	if names[0] != "prod" || names[1] != "staging" {
		t.Errorf("unexpected order: %v", names)
	}
	if m.Bookmarks["prod"].VaultPath != "/vaults/prod.vault" {
		t.Errorf("wrong vault path for prod")
	}
	if m.Bookmarks["prod"].Note != "production vault" {
		t.Errorf("wrong note for prod")
	}
	if m.Bookmarks["prod"].CreatedAt.IsZero() {
		t.Errorf("expected non-zero CreatedAt")
	}
}

func TestAddBookmarkEmptyName(t *testing.T) {
	v := newBookmarkVault(t)
	err := AddBookmark(v, "", "/vaults/prod.vault", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemoveBookmark(t *testing.T) {
	v := newBookmarkVault(t)

	_ = AddBookmark(v, "dev", "/vaults/dev.vault", "")
	if err := RemoveBookmark(v, "dev"); err != nil {
		t.Fatalf("RemoveBookmark: %v", err)
	}

	names, _, err := ListBookmarks(v)
	if err != nil {
		t.Fatalf("ListBookmarks: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 bookmarks after removal, got %d", len(names))
	}
}

func TestRemoveBookmarkNotFound(t *testing.T) {
	v := newBookmarkVault(t)
	err := RemoveBookmark(v, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing bookmark")
	}
}

func TestBookmarkManifestPersists(t *testing.T) {
	v := newBookmarkVault(t)
	_ = AddBookmark(v, "alpha", "/alpha.vault", "alpha note")

	// reload from disk
	m, err := LoadBookmarkManifest(v.Path())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	entry, ok := m.Bookmarks["alpha"]
	if !ok {
		t.Fatal("bookmark 'alpha' not found after reload")
	}
	if entry.VaultPath != "/alpha.vault" {
		t.Errorf("wrong path: %q", entry.VaultPath)
	}

	// ensure file exists with restricted permissions
	info, err := os.Stat(BookmarkPath(v.Path()))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %v", info.Mode().Perm())
	}
}

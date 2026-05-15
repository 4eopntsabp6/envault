package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/nicholasgasior/envault/internal/store"
)

// BookmarkManifest maps bookmark names to vault paths with metadata.
type BookmarkManifest struct {
	Bookmarks map[string]BookmarkEntry `json:"bookmarks"`
}

// BookmarkEntry holds the vault path and creation time for a bookmark.
type BookmarkEntry struct {
	VaultPath string    `json:"vault_path"`
	CreatedAt time.Time `json:"created_at"`
	Note      string    `json:"note,omitempty"`
}

// BookmarkPath returns the path to the bookmark manifest for a given vault.
func BookmarkPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	return filepath.Join(dir, ".envault_bookmarks.json")
}

// LoadBookmarkManifest loads the bookmark manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadBookmarkManifest(vaultPath string) (*BookmarkManifest, error) {
	p := BookmarkPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &BookmarkManifest{Bookmarks: map[string]BookmarkEntry{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("bookmark: read manifest: %w", err)
	}
	var m BookmarkManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("bookmark: parse manifest: %w", err)
	}
	if m.Bookmarks == nil {
		m.Bookmarks = map[string]BookmarkEntry{}
	}
	return &m, nil
}

// SaveBookmarkManifest writes the bookmark manifest to disk.
func SaveBookmarkManifest(vaultPath string, m *BookmarkManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("bookmark: marshal manifest: %w", err)
	}
	return os.WriteFile(BookmarkPath(vaultPath), data, 0600)
}

// AddBookmark creates or updates a bookmark with the given name pointing to targetVaultPath.
func AddBookmark(v *store.Vault, name, targetVaultPath, note string) error {
	if name == "" {
		return fmt.Errorf("bookmark: name must not be empty")
	}
	m, err := LoadBookmarkManifest(v.Path())
	if err != nil {
		return err
	}
	m.Bookmarks[name] = BookmarkEntry{
		VaultPath: targetVaultPath,
		CreatedAt: time.Now().UTC(),
		Note:      note,
	}
	return SaveBookmarkManifest(v.Path(), m)
}

// RemoveBookmark deletes a bookmark by name.
func RemoveBookmark(v *store.Vault, name string) error {
	m, err := LoadBookmarkManifest(v.Path())
	if err != nil {
		return err
	}
	if _, ok := m.Bookmarks[name]; !ok {
		return fmt.Errorf("bookmark: %q not found", name)
	}
	delete(m.Bookmarks, name)
	return SaveBookmarkManifest(v.Path(), m)
}

// ListBookmarks returns bookmark names in sorted order.
func ListBookmarks(v *store.Vault) ([]string, *BookmarkManifest, error) {
	m, err := LoadBookmarkManifest(v.Path())
	if err != nil {
		return nil, nil, err
	}
	names := make([]string, 0, len(m.Bookmarks))
	for k := range m.Bookmarks {
		names = append(names, k)
	}
	sort.Strings(names)
	return names, m, nil
}

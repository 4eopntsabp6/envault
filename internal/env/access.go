package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/envault/internal/store"
)

// AccessEntry records a single read access to a key.
type AccessEntry struct {
	Key       string    `json:"key"`
	AccessedAt time.Time `json:"accessed_at"`
}

// AccessManifest holds access records keyed by vault path.
type AccessManifest struct {
	Entries []AccessEntry `json:"entries"`
}

// AccessPath returns the path to the access log for a given vault.
func AccessPath(v *store.Vault) string {
	return filepath.Join(filepath.Dir(v.Path()), ".envault_access.json")
}

// LoadAccessManifest loads the access manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadAccessManifest(v *store.Vault) (*AccessManifest, error) {
	p := AccessPath(v)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &AccessManifest{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("access: read manifest: %w", err)
	}
	var m AccessManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("access: parse manifest: %w", err)
	}
	return &m, nil
}

// SaveAccessManifest writes the access manifest to disk.
func SaveAccessManifest(v *store.Vault, m *AccessManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("access: marshal manifest: %w", err)
	}
	return os.WriteFile(AccessPath(v), data, 0600)
}

// RecordAccess appends an access entry for the given key.
func RecordAccess(v *store.Vault, key string) error {
	m, err := LoadAccessManifest(v)
	if err != nil {
		return err
	}
	m.Entries = append(m.Entries, AccessEntry{
		Key:        key,
		AccessedAt: time.Now().UTC(),
	})
	return SaveAccessManifest(v, m)
}

// AccessesForKey returns all access entries for a specific key.
func AccessesForKey(v *store.Vault, key string) ([]AccessEntry, error) {
	m, err := LoadAccessManifest(v)
	if err != nil {
		return nil, err
	}
	var result []AccessEntry
	for _, e := range m.Entries {
		if e.Key == key {
			result = append(result, e)
		}
	}
	return result, nil
}

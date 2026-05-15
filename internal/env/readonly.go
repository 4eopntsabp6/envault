package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/envault/internal/store"
)

// ReadonlyManifest holds the set of keys marked as read-only.
type ReadonlyManifest struct {
	Keys map[string]bool `json:"keys"`
}

// ReadonlyPath returns the path to the readonly manifest for a vault.
func ReadonlyPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".readonly.json")
}

// LoadReadonlyManifest loads the readonly manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadReadonlyManifest(vaultPath string) (*ReadonlyManifest, error) {
	p := ReadonlyPath(vaultPath)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &ReadonlyManifest{Keys: map[string]bool{}}, nil
		}
		return nil, err
	}
	var m ReadonlyManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Keys == nil {
		m.Keys = map[string]bool{}
	}
	return &m, nil
}

// SaveReadonlyManifest writes the readonly manifest to disk.
func SaveReadonlyManifest(vaultPath string, m *ReadonlyManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ReadonlyPath(vaultPath), data, 0600)
}

// SetReadonly marks or unmarks a key as read-only.
func SetReadonly(v *store.Vault, key string, readonly bool) error {
	if _, err := v.Get(key); err != nil {
		return fmt.Errorf("key %q not found", key)
	}
	m, err := LoadReadonlyManifest(v.Path())
	if err != nil {
		return err
	}
	if readonly {
		m.Keys[key] = true
	} else {
		delete(m.Keys, key)
	}
	return SaveReadonlyManifest(v.Path(), m)
}

// IsReadonly reports whether the given key is marked read-only.
func IsReadonly(v *store.Vault, key string) (bool, error) {
	m, err := LoadReadonlyManifest(v.Path())
	if err != nil {
		return false, err
	}
	return m.Keys[key], nil
}

// ReadonlyKeys returns all keys currently marked as read-only.
func ReadonlyKeys(v *store.Vault) ([]string, error) {
	m, err := LoadReadonlyManifest(v.Path())
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(m.Keys))
	for k := range m.Keys {
		keys = append(keys, k)
	}
	return keys, nil
}

// GuardReadonly returns an error if the key is marked read-only.
func GuardReadonly(v *store.Vault, key string) error {
	ro, err := IsReadonly(v, key)
	if err != nil {
		return err
	}
	if ro {
		return fmt.Errorf("key %q is read-only and cannot be modified", key)
	}
	return nil
}

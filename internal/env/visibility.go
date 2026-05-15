package env

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/user/envault/internal/store"
)

// Visibility levels for vault keys.
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
	VisibilitySecret  = "secret"
)

var validVisibilities = map[string]bool{
	VisibilityPublic:  true,
	VisibilityPrivate: true,
	VisibilitySecret:  true,
}

// VisibilityManifest maps key names to their visibility level.
type VisibilityManifest struct {
	Entries map[string]string `json:"entries"`
}

// VisibilityPath returns the path to the visibility manifest for a vault.
func VisibilityPath(v *store.Vault) string {
	return filepath.Join(filepath.Dir(v.Path()), ".envault_visibility.json")
}

// LoadVisibilityManifest loads the manifest from disk, returning an empty one if missing.
func LoadVisibilityManifest(v *store.Vault) (*VisibilityManifest, error) {
	path := VisibilityPath(v)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &VisibilityManifest{Entries: map[string]string{}}, nil
		}
		return nil, err
	}
	var m VisibilityManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Entries == nil {
		m.Entries = map[string]string{}
	}
	return &m, nil
}

// SaveVisibilityManifest persists the manifest to disk.
func SaveVisibilityManifest(v *store.Vault, m *VisibilityManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(VisibilityPath(v), data, 0600)
}

// SetVisibility assigns a visibility level to a key.
func SetVisibility(v *store.Vault, key, level string) error {
	if !validVisibilities[level] {
		return errors.New("invalid visibility level: must be public, private, or secret")
	}
	if _, err := v.Get(key); err != nil {
		return errors.New("key not found: " + key)
	}
	m, err := LoadVisibilityManifest(v)
	if err != nil {
		return err
	}
	m.Entries[key] = level
	return SaveVisibilityManifest(v, m)
}

// GetVisibility returns the visibility level for a key, defaulting to "private".
func GetVisibility(v *store.Vault, key string) (string, error) {
	m, err := LoadVisibilityManifest(v)
	if err != nil {
		return "", err
	}
	level, ok := m.Entries[key]
	if !ok {
		return VisibilityPrivate, nil
	}
	return level, nil
}

// FilterByVisibility returns keys whose visibility matches the given level.
func FilterByVisibility(v *store.Vault, level string) ([]string, error) {
	m, err := LoadVisibilityManifest(v)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, k := range v.Keys() {
		kLevel, ok := m.Entries[k]
		if !ok {
			kLevel = VisibilityPrivate
		}
		if kLevel == level {
			result = append(result, k)
		}
	}
	return result, nil
}

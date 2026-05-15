package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/envault/internal/store"
)

// DescriptionManifest maps key names to human-readable descriptions.
type DescriptionManifest struct {
	Descriptions map[string]string `json:"descriptions"`
}

// DescriptionPath returns the path to the description manifest for a vault.
func DescriptionPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, base+".descriptions.json")
}

// LoadDescriptions loads the description manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadDescriptions(vaultPath string) (*DescriptionManifest, error) {
	p := DescriptionPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &DescriptionManifest{Descriptions: map[string]string{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load descriptions: %w", err)
	}
	var m DescriptionManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse descriptions: %w", err)
	}
	if m.Descriptions == nil {
		m.Descriptions = map[string]string{}
	}
	return &m, nil
}

// SaveDescriptions writes the description manifest to disk.
func SaveDescriptions(vaultPath string, m *DescriptionManifest) error {
	p := DescriptionPath(vaultPath)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal descriptions: %w", err)
	}
	return os.WriteFile(p, data, 0600)
}

// SetDescription attaches a description to a key in the vault.
// Returns an error if the key does not exist in the vault.
func SetDescription(v *store.Vault, key, description string) error {
	if _, err := v.Get(key); err != nil {
		return fmt.Errorf("key %q not found", key)
	}
	m, err := LoadDescriptions(v.Path())
	if err != nil {
		return err
	}
	m.Descriptions[key] = description
	return SaveDescriptions(v.Path(), m)
}

// GetDescription returns the description for a key, or an empty string if none is set.
func GetDescription(v *store.Vault, key string) (string, error) {
	m, err := LoadDescriptions(v.Path())
	if err != nil {
		return "", err
	}
	return m.Descriptions[key], nil
}

// DeleteDescription removes the description for a key.
func DeleteDescription(v *store.Vault, key string) error {
	m, err := LoadDescriptions(v.Path())
	if err != nil {
		return err
	}
	delete(m.Descriptions, key)
	return SaveDescriptions(v.Path(), m)
}

// ListDescriptions returns all key→description pairs present in the manifest.
func ListDescriptions(v *store.Vault) (map[string]string, error) {
	m, err := LoadDescriptions(v.Path())
	if err != nil {
		return nil, err
	}
	return m.Descriptions, nil
}

package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/envault/internal/store"
)

// LabelManifest maps key names to a set of string labels.
type LabelManifest struct {
	Labels map[string][]string `json:"labels"`
}

// LabelPath returns the path to the label manifest for the given vault.
func LabelPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, base+".labels.json")
}

// LoadLabelManifest reads the label manifest from disk, returning an empty one if missing.
func LoadLabelManifest(vaultPath string) (*LabelManifest, error) {
	p := LabelPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &LabelManifest{Labels: make(map[string][]string)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("label: read manifest: %w", err)
	}
	var m LabelManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("label: unmarshal manifest: %w", err)
	}
	if m.Labels == nil {
		m.Labels = make(map[string][]string)
	}
	return &m, nil
}

// SaveLabelManifest writes the label manifest to disk.
func SaveLabelManifest(vaultPath string, m *LabelManifest) error {
	p := LabelPath(vaultPath)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("label: marshal manifest: %w", err)
	}
	return os.WriteFile(p, data, 0600)
}

// SetLabels assigns labels to a key, replacing any existing labels.
func SetLabels(vaultPath string, v *store.Vault, key string, labels []string) error {
	if _, err := v.Get(key); err != nil {
		return fmt.Errorf("label: key %q not found", key)
	}
	m, err := LoadLabelManifest(vaultPath)
	if err != nil {
		return err
	}
	m.Labels[key] = labels
	return SaveLabelManifest(vaultPath, m)
}

// GetLabels returns the labels for a key, or nil if none are set.
func GetLabels(vaultPath string, key string) ([]string, error) {
	m, err := LoadLabelManifest(vaultPath)
	if err != nil {
		return nil, err
	}
	return m.Labels[key], nil
}

// FilterByLabel returns all keys that have the given label.
func FilterByLabel(vaultPath string, label string) ([]string, error) {
	m, err := LoadLabelManifest(vaultPath)
	if err != nil {
		return nil, err
	}
	var matched []string
	for key, labels := range m.Labels {
		for _, l := range labels {
			if l == label {
				matched = append(matched, key)
				break
			}
		}
	}
	return matched, nil
}

// DeleteLabels removes all labels for a key.
func DeleteLabels(vaultPath string, key string) error {
	m, err := LoadLabelManifest(vaultPath)
	if err != nil {
		return err
	}
	delete(m.Labels, key)
	return SaveLabelManifest(vaultPath, m)
}

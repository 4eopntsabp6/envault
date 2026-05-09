// Package tags provides key tagging functionality for envault vaults.
// Tags allow grouping and filtering of secrets by arbitrary labels.
package tags

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// Manifest maps each key to its set of tags.
type Manifest map[string][]string

// ManifestPath returns the path to the tags manifest for a given vault path.
func ManifestPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".tags.json")
}

// LoadManifest reads the tags manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadManifest(vaultPath string) (Manifest, error) {
	p := ManifestPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return Manifest{}, nil
	}
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// SaveManifest writes the tags manifest to disk.
func SaveManifest(vaultPath string, m Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ManifestPath(vaultPath), data, 0600)
}

// SetTags replaces the tags for a key.
func SetTags(m Manifest, key string, tags []string) {
	sorted := make([]string, len(tags))
	copy(sorted, tags)
	sort.Strings(sorted)
	m[key] = sorted
}

// GetTags returns the tags for a key.
func GetTags(m Manifest, key string) []string {
	return m[key]
}

// FilterByTag returns all keys that have the given tag.
func FilterByTag(m Manifest, tag string) []string {
	var result []string
	for key, tags := range m {
		for _, t := range tags {
			if t == tag {
				result = append(result, key)
				break
			}
		}
	}
	sort.Strings(result)
	return result
}

// RemoveKey deletes all tag entries for a key.
func RemoveKey(m Manifest, key string) {
	delete(m, key)
}

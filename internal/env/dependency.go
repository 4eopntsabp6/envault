package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/envault/internal/store"
)

// DependencyManifest maps a key to the list of keys it depends on.
type DependencyManifest map[string][]string

// DependencyPath returns the path to the dependency manifest file for the given vault.
func DependencyPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".deps.json")
}

// LoadDependencies loads the dependency manifest for the given vault.
// Returns an empty manifest if the file does not exist.
func LoadDependencies(vaultPath string) (DependencyManifest, error) {
	path := DependencyPath(vaultPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DependencyManifest{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("dependency: read manifest: %w", err)
	}
	var m DependencyManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("dependency: parse manifest: %w", err)
	}
	return m, nil
}

// SaveDependencies writes the dependency manifest to disk.
func SaveDependencies(vaultPath string, m DependencyManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("dependency: marshal manifest: %w", err)
	}
	if err := os.WriteFile(DependencyPath(vaultPath), data, 0600); err != nil {
		return fmt.Errorf("dependency: write manifest: %w", err)
	}
	return nil
}

// SetDependencies records that key depends on the given list of keys.
// All dependency keys must exist in the vault.
func SetDependencies(v *store.Vault, key string, deps []string) error {
	if _, err := v.Get(key); err != nil {
		return fmt.Errorf("dependency: key %q not found in vault", key)
	}
	for _, dep := range deps {
		if _, err := v.Get(dep); err != nil {
			return fmt.Errorf("dependency: dependency key %q not found in vault", dep)
		}
	}
	m, err := LoadDependencies(v.Path())
	if err != nil {
		return err
	}
	m[key] = deps
	return SaveDependencies(v.Path(), m)
}

// GetDependencies returns the list of keys that key depends on.
func GetDependencies(v *store.Vault, key string) ([]string, error) {
	m, err := LoadDependencies(v.Path())
	if err != nil {
		return nil, err
	}
	return m[key], nil
}

// CheckMissing returns any dependency keys that are not present in the vault.
func CheckMissing(v *store.Vault, key string) ([]string, error) {
	deps, err := GetDependencies(v, key)
	if err != nil {
		return nil, err
	}
	var missing []string
	for _, dep := range deps {
		if _, err := v.Get(dep); err != nil {
			missing = append(missing, dep)
		}
	}
	return missing, nil
}

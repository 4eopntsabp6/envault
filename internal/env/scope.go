package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/user/envault/internal/store"
)

// Scope represents a named environment context (e.g. "dev", "staging", "prod").
type Scope struct {
	Name string            `json:"name"`
	Keys map[string]string `json:"keys"`
}

// ScopeManifest holds all scopes for a vault.
type ScopeManifest struct {
	Scopes map[string]Scope `json:"scopes"`
}

// ScopePath returns the path to the scope manifest for a given vault path.
func ScopePath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return filepath.Join(dir, name+".scopes.json")
}

// LoadScopeManifest loads the scope manifest from disk, returning an empty one if missing.
func LoadScopeManifest(vaultPath string) (*ScopeManifest, error) {
	p := ScopePath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &ScopeManifest{Scopes: make(map[string]Scope)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scope: read manifest: %w", err)
	}
	var m ScopeManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("scope: parse manifest: %w", err)
	}
	if m.Scopes == nil {
		m.Scopes = make(map[string]Scope)
	}
	return &m, nil
}

// SaveScopeManifest writes the scope manifest to disk.
func SaveScopeManifest(vaultPath string, m *ScopeManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("scope: marshal manifest: %w", err)
	}
	return os.WriteFile(ScopePath(vaultPath), data, 0600)
}

// SetScope saves a snapshot of the given keys from the vault under a named scope.
func SetScope(v *store.Vault, vaultPath, scopeName string, keys []string) error {
	m, err := LoadScopeManifest(vaultPath)
	if err != nil {
		return err
	}
	captured := make(map[string]string)
	for _, k := range keys {
		val, err := v.Get(k)
		if err != nil {
			return fmt.Errorf("scope: key %q not found", k)
		}
		captured[k] = val
	}
	m.Scopes[scopeName] = Scope{Name: scopeName, Keys: captured}
	return SaveScopeManifest(vaultPath, m)
}

// ApplyScope restores keys from a named scope into the vault.
func ApplyScope(v *store.Vault, vaultPath, scopeName string, overwrite bool) ([]string, error) {
	m, err := LoadScopeManifest(vaultPath)
	if err != nil {
		return nil, err
	}
	sc, ok := m.Scopes[scopeName]
	if !ok {
		return nil, fmt.Errorf("scope: %q not found", scopeName)
	}
	var applied []string
	for k, val := range sc.Keys {
		if !overwrite {
			if _, err := v.Get(k); err == nil {
				continue
			}
		}
		v.Set(k, val)
		applied = append(applied, k)
	}
	sort.Strings(applied)
	return applied, nil
}

// ListScopes returns all scope names in alphabetical order.
func ListScopes(vaultPath string) ([]string, error) {
	m, err := LoadScopeManifest(vaultPath)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(m.Scopes))
	for name := range m.Scopes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

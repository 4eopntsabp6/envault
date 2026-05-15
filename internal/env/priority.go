package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/user/envault/internal/store"
)

// PriorityLevel represents the numeric priority of a key (higher = more important).
type PriorityLevel int

const (
	PriorityLow    PriorityLevel = 1
	PriorityNormal PriorityLevel = 5
	PriorityHigh   PriorityLevel = 10
)

// PriorityManifest maps key names to their priority levels.
type PriorityManifest map[string]PriorityLevel

// PriorityPath returns the path to the priority manifest file for a vault.
func PriorityPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".priority.json")
}

// LoadPriorityManifest loads the priority manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadPriorityManifest(vaultPath string) (PriorityManifest, error) {
	p := PriorityPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return PriorityManifest{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("priority: read manifest: %w", err)
	}
	var m PriorityManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("priority: parse manifest: %w", err)
	}
	return m, nil
}

// SavePriorityManifest writes the priority manifest to disk.
func SavePriorityManifest(vaultPath string, m PriorityManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("priority: marshal manifest: %w", err)
	}
	return os.WriteFile(PriorityPath(vaultPath), data, 0600)
}

// SetPriority assigns a priority level to a key in the vault.
func SetPriority(v *store.Vault, level PriorityLevel) func(key string) error {
	return func(key string) error {
		if _, ok := v.Get(key); !ok {
			return fmt.Errorf("priority: key %q not found", key)
		}
		m, err := LoadPriorityManifest(v.Path())
		if err != nil {
			return err
		}
		m[key] = level
		return SavePriorityManifest(v.Path(), m)
	}
}

// GetPriority returns the priority level for a key (defaults to PriorityNormal).
func GetPriority(v *store.Vault, key string) (PriorityLevel, error) {
	m, err := LoadPriorityManifest(v.Path())
	if err != nil {
		return 0, err
	}
	if lvl, ok := m[key]; ok {
		return lvl, nil
	}
	return PriorityNormal, nil
}

// KeysByPriority returns all vault keys sorted by priority descending.
func KeysByPriority(v *store.Vault) ([]string, error) {
	m, err := LoadPriorityManifest(v.Path())
	if err != nil {
		return nil, err
	}
	keys := v.Keys()
	sort.Slice(keys, func(i, j int) bool {
		pi := m[keys[i]]
		if pi == 0 {
			pi = PriorityNormal
		}
		pj := m[keys[j]]
		if pj == 0 {
			pj = PriorityNormal
		}
		if pi != pj {
			return pi > pj
		}
		return keys[i] < keys[j]
	})
	return keys, nil
}

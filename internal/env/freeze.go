package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nicholasgasior/envault/internal/store"
)

// FreezeManifest holds freeze state for vault keys.
type FreezeManifest struct {
	Frozen map[string]FreezeEntry `json:"frozen"`
}

// FreezeEntry records when a key was frozen and by whom.
type FreezeEntry struct {
	FrozenAt time.Time `json:"frozen_at"`
	Reason   string    `json:"reason,omitempty"`
}

// FreezePath returns the path to the freeze manifest for a vault.
func FreezePath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return filepath.Join(dir, name+".freeze.json")
}

// LoadFreezeManifest loads the freeze manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadFreezeManifest(vaultPath string) (*FreezeManifest, error) {
	p := FreezePath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &FreezeManifest{Frozen: make(map[string]FreezeEntry)}, nil
	}
	if err != nil {
		return nil, err
	}
	var m FreezeManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Frozen == nil {
		m.Frozen = make(map[string]FreezeEntry)
	}
	return &m, nil
}

// SaveFreezeManifest writes the freeze manifest to disk.
func SaveFreezeManifest(vaultPath string, m *FreezeManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(FreezePath(vaultPath), data, 0600)
}

// FreezeKey marks a key as frozen, preventing modification.
func FreezeKey(v *store.Vault, key, reason string) error {
	if _, err := v.Get(key); err != nil {
		return fmt.Errorf("key %q not found", key)
	}
	m, err := LoadFreezeManifest(v.Path())
	if err != nil {
		return err
	}
	m.Frozen[key] = FreezeEntry{FrozenAt: time.Now().UTC(), Reason: reason}
	return SaveFreezeManifest(v.Path(), m)
}

// UnfreezeKey removes the freeze on a key.
func UnfreezeKey(v *store.Vault, key string) error {
	m, err := LoadFreezeManifest(v.Path())
	if err != nil {
		return err
	}
	if _, ok := m.Frozen[key]; !ok {
		return fmt.Errorf("key %q is not frozen", key)
	}
	delete(m.Frozen, key)
	return SaveFreezeManifest(v.Path(), m)
}

// IsFrozen reports whether a key is currently frozen.
func IsFrozen(v *store.Vault, key string) (bool, error) {
	m, err := LoadFreezeManifest(v.Path())
	if err != nil {
		return false, err
	}
	_, ok := m.Frozen[key]
	return ok, nil
}

// FrozenKeys returns all currently frozen keys.
func FrozenKeys(v *store.Vault) ([]FreezeEntry, []string, error) {
	m, err := LoadFreezeManifest(v.Path())
	if err != nil {
		return nil, nil, err
	}
	keys := make([]string, 0, len(m.Frozen))
	entries := make([]FreezeEntry, 0, len(m.Frozen))
	for k, e := range m.Frozen {
		keys = append(keys, k)
		entries = append(entries, e)
	}
	return entries, keys, nil
}

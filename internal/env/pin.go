// Package env provides utilities for managing environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicholasgasior/envault/internal/store"
)

// PinManifest maps key names to pinned (locked) values.
type PinManifest struct {
	Pinned map[string]string `json:"pinned"`
}

// PinPath returns the path to the pin manifest for a given vault.
func PinPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return filepath.Join(dir, name+".pins.json")
}

// LoadPinManifest loads the pin manifest from disk, returning an empty one if missing.
func LoadPinManifest(vaultPath string) (*PinManifest, error) {
	path := PinPath(vaultPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &PinManifest{Pinned: make(map[string]string)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read pin manifest: %w", err)
	}
	var m PinManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse pin manifest: %w", err)
	}
	if m.Pinned == nil {
		m.Pinned = make(map[string]string)
	}
	return &m, nil
}

// SavePinManifest writes the pin manifest to disk.
func SavePinManifest(vaultPath string, m *PinManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pin manifest: %w", err)
	}
	return os.WriteFile(PinPath(vaultPath), data, 0600)
}

// PinKey pins the current value of key in the vault, preventing future overwrites.
func PinKey(v *store.Vault, vaultPath, key string) error {
	val, ok := v.Get(key)
	if !ok {
		return fmt.Errorf("key %q not found", key)
	}
	m, err := LoadPinManifest(vaultPath)
	if err != nil {
		return err
	}
	m.Pinned[key] = val
	return SavePinManifest(vaultPath, m)
}

// UnpinKey removes the pin for key.
func UnpinKey(vaultPath, key string) error {
	m, err := LoadPinManifest(vaultPath)
	if err != nil {
		return err
	}
	if _, ok := m.Pinned[key]; !ok {
		return fmt.Errorf("key %q is not pinned", key)
	}
	delete(m.Pinned, key)
	return SavePinManifest(vaultPath, m)
}

// IsPinned returns true if key is currently pinned.
func IsPinned(vaultPath, key string) (bool, error) {
	m, err := LoadPinManifest(vaultPath)
	if err != nil {
		return false, err
	}
	_, ok := m.Pinned[key]
	return ok, nil
}

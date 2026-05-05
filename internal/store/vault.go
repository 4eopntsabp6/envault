package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const vaultFileName = ".envault"

// Vault holds encrypted secret entries for a project.
type Vault struct {
	Version int               `json:"version"`
	Entries map[string]string `json:"entries"` // key -> base64-encoded ciphertext
}

// NewVault creates an empty vault.
func NewVault() *Vault {
	return &Vault{
		Version: 1,
		Entries: make(map[string]string),
	}
}

// VaultPath returns the path to the vault file in the given directory.
func VaultPath(dir string) string {
	return filepath.Join(dir, vaultFileName)
}

// Load reads and deserialises a vault from the given directory.
func Load(dir string) (*Vault, error) {
	path := VaultPath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	var v Vault
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Save serialises the vault and writes it to the given directory.
func Save(dir string, v *Vault) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	path := VaultPath(dir)
	return os.WriteFile(path, data, 0600)
}

// Set stores a ciphertext entry under the given key.
func (v *Vault) Set(key, ciphertext string) {
	v.Entries[key] = ciphertext
}

// Get retrieves the ciphertext for the given key.
func (v *Vault) Get(key string) (string, bool) {
	val, ok := v.Entries[key]
	return val, ok
}

// Delete removes an entry from the vault.
func (v *Vault) Delete(key string) {
	delete(v.Entries, key)
}

// Keys returns all secret keys stored in the vault.
func (v *Vault) Keys() []string {
	keys := make([]string, 0, len(v.Entries))
	for k := range v.Entries {
		keys = append(keys, k)
	}
	return keys
}

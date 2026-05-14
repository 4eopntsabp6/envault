package env

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/nicholasgasior/envault/internal/store"
)

// LockManifest holds lock state for vault keys.
type LockManifest struct {
	Locked map[string]LockEntry `json:"locked"`
}

// LockEntry records when a key was locked and by whom.
type LockEntry struct {
	LockedAt time.Time `json:"locked_at"`
	Reason   string    `json:"reason,omitempty"`
}

// LockPath returns the path to the lock manifest for a given vault.
func LockPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".lock.json")
}

// LoadLockManifest reads the lock manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadLockManifest(vaultPath string) (LockManifest, error) {
	path := LockPath(vaultPath)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return LockManifest{Locked: map[string]LockEntry{}}, nil
	}
	if err != nil {
		return LockManifest{}, err
	}
	var m LockManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return LockManifest{}, err
	}
	if m.Locked == nil {
		m.Locked = map[string]LockEntry{}
	}
	return m, nil
}

// SaveLockManifest writes the lock manifest to disk.
func SaveLockManifest(vaultPath string, m LockManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(LockPath(vaultPath), data, 0600)
}

// LockKey marks a key as locked, preventing modification.
func LockKey(v *store.Vault, key, reason string) error {
	if _, err := v.Get(key); err != nil {
		return errors.New("key not found: " + key)
	}
	m, err := LoadLockManifest(v.Path())
	if err != nil {
		return err
	}
	m.Locked[key] = LockEntry{LockedAt: time.Now().UTC(), Reason: reason}
	return SaveLockManifest(v.Path(), m)
}

// UnlockKey removes the lock from a key.
func UnlockKey(v *store.Vault, key string) error {
	m, err := LoadLockManifest(v.Path())
	if err != nil {
		return err
	}
	if _, ok := m.Locked[key]; !ok {
		return errors.New("key is not locked: " + key)
	}
	delete(m.Locked, key)
	return SaveLockManifest(v.Path(), m)
}

// IsLocked reports whether a key is currently locked.
func IsLocked(v *store.Vault, key string) (bool, error) {
	m, err := LoadLockManifest(v.Path())
	if err != nil {
		return false, err
	}
	_, ok := m.Locked[key]
	return ok, nil
}

// ListLocked returns all locked keys and their entries.
func ListLocked(v *store.Vault) (map[string]LockEntry, error) {
	m, err := LoadLockManifest(v.Path())
	if err != nil {
		return nil, err
	}
	return m.Locked, nil
}

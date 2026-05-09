// Package expiry provides time-based expiration for vault secrets.
package expiry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/envault/internal/store"
)

// Entry holds expiration metadata for a single key.
type Entry struct {
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Manifest maps keys to their expiry entries.
type Manifest struct {
	Entries map[string]Entry `json:"entries"`
}

// ManifestPath returns the path to the expiry manifest for a vault.
func ManifestPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, base+".expiry.json")
}

// LoadManifest reads the expiry manifest from disk, returning an empty one if missing.
func LoadManifest(vaultPath string) (*Manifest, error) {
	path := ManifestPath(vaultPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Manifest{Entries: make(map[string]Entry)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("expiry: read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("expiry: parse manifest: %w", err)
	}
	if m.Entries == nil {
		m.Entries = make(map[string]Entry)
	}
	return &m, nil
}

// SaveManifest writes the expiry manifest to disk.
func SaveManifest(vaultPath string, m *Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("expiry: marshal manifest: %w", err)
	}
	return os.WriteFile(ManifestPath(vaultPath), data, 0600)
}

// SetExpiry assigns an expiration duration to a key.
func SetExpiry(vaultPath, key string, ttl time.Duration) error {
	m, err := LoadManifest(vaultPath)
	if err != nil {
		return err
	}
	m.Entries[key] = Entry{Key: key, ExpiresAt: time.Now().Add(ttl)}
	return SaveManifest(vaultPath, m)
}

// PurgeExpired removes all expired keys from the vault and the manifest.
func PurgeExpired(v *store.Vault, vaultPath, password string) ([]string, error) {
	m, err := LoadManifest(vaultPath)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var purged []string
	for key, entry := range m.Entries {
		if now.After(entry.ExpiresAt) {
			v.Delete(key)
			delete(m.Entries, key)
			purged = append(purged, key)
		}
	}
	if len(purged) > 0 {
		if err := v.Save(password); err != nil {
			return nil, fmt.Errorf("expiry: save vault: %w", err)
		}
		if err := SaveManifest(vaultPath, m); err != nil {
			return nil, err
		}
	}
	return purged, nil
}

package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/envault/internal/store"
)

// TTLManifest maps key names to their expiration times.
type TTLManifest struct {
	Entries map[string]time.Time `json:"entries"`
}

// TTLPath returns the path to the TTL manifest for the given vault.
func TTLPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".ttl.json")
}

// LoadTTLManifest loads the TTL manifest from disk, returning an empty one if missing.
func LoadTTLManifest(vaultPath string) (*TTLManifest, error) {
	p := TTLPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &TTLManifest{Entries: make(map[string]time.Time)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ttl: read manifest: %w", err)
	}
	var m TTLManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("ttl: parse manifest: %w", err)
	}
	if m.Entries == nil {
		m.Entries = make(map[string]time.Time)
	}
	return &m, nil
}

// SaveTTLManifest writes the TTL manifest to disk.
func SaveTTLManifest(vaultPath string, m *TTLManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("ttl: marshal manifest: %w", err)
	}
	return os.WriteFile(TTLPath(vaultPath), data, 0600)
}

// SetTTL assigns a time-to-live duration to a key. The key must exist in the vault.
func SetTTL(v *store.Vault, vaultPath string, key string, ttl time.Duration) error {
	if _, ok := v.Get(key); !ok {
		return fmt.Errorf("ttl: key %q not found", key)
	}
	m, err := LoadTTLManifest(vaultPath)
	if err != nil {
		return err
	}
	m.Entries[key] = time.Now().Add(ttl)
	return SaveTTLManifest(vaultPath, m)
}

// PurgeTTLExpired removes all keys from the vault whose TTL has elapsed.
// Returns the list of purged keys.
func PurgeTTLExpired(v *store.Vault, vaultPath string) ([]string, error) {
	m, err := LoadTTLManifest(vaultPath)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var purged []string
	for key, exp := range m.Entries {
		if now.After(exp) {
			v.Delete(key)
			delete(m.Entries, key)
			purged = append(purged, key)
		}
	}
	if len(purged) > 0 {
		if err := SaveTTLManifest(vaultPath, m); err != nil {
			return purged, err
		}
	}
	return purged, nil
}

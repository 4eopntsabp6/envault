// Package snapshot provides point-in-time snapshots of vault secrets,
// enabling diffing and rollback of secret values.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/envault/internal/store"
)

// Snapshot represents a saved state of all secrets in a vault.
type Snapshot struct {
	CreatedAt time.Time         `json:"created_at"`
	VaultPath string            `json:"vault_path"`
	Secrets   map[string]string `json:"secrets"`
}

// Take captures the current state of all secrets in the vault.
func Take(v *store.Vault, password string) (*Snapshot, error) {
	keys := v.Keys()
	secrets := make(map[string]string, len(keys))
	for _, k := range keys {
		val, ok := v.Get(k)
		if !ok {
			continue
		}
		secrets[k] = val
	}
	return &Snapshot{
		CreatedAt: time.Now().UTC(),
		VaultPath: v.Path(),
		Secrets:   secrets,
	}, nil
}

// Save writes the snapshot to a JSON file inside snapshotDir.
func Save(snap *Snapshot, snapshotDir string) (string, error) {
	if err := os.MkdirAll(snapshotDir, 0700); err != nil {
		return "", fmt.Errorf("create snapshot dir: %w", err)
	}
	filename := fmt.Sprintf("%d.json", snap.CreatedAt.UnixNano())
	path := filepath.Join(snapshotDir, filename)
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal snapshot: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", fmt.Errorf("write snapshot: %w", err)
	}
	return path, nil
}

// Load reads a snapshot from the given file path.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return &snap, nil
}

// Diff compares two snapshots and returns added, removed, and changed keys.
func Diff(before, after *Snapshot) (added, removed, changed []string) {
	for k, v := range after.Secrets {
		if _, ok := before.Secrets[k]; !ok {
			added = append(added, k)
		} else if before.Secrets[k] != v {
			changed = append(changed, k)
		}
	}
	for k := range before.Secrets {
		if _, ok := after.Secrets[k]; !ok {
			removed = append(removed, k)
		}
	}
	return
}

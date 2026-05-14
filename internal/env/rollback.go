package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/envault/internal/store"
)

// RollbackEntry represents a single rollback checkpoint.
type RollbackEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Label     string            `json:"label"`
	Snapshot  map[string]string `json:"snapshot"`
}

// RollbackPath returns the path to the rollback journal for a vault.
func RollbackPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".rollback.json")
}

// LoadRollback reads all rollback entries for the vault.
func LoadRollback(vaultPath string) ([]RollbackEntry, error) {
	p := RollbackPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []RollbackEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// SaveRollback persists rollback entries to disk.
func SaveRollback(vaultPath string, entries []RollbackEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(RollbackPath(vaultPath), data, 0600)
}

// Checkpoint saves the current vault state as a named rollback point.
func Checkpoint(v *store.Vault, vaultPath, label string) error {
	keys := v.Keys()
	snap := make(map[string]string, len(keys))
	for _, k := range keys {
		val, _ := v.Get(k)
		snap[k] = val
	}
	entries, err := LoadRollback(vaultPath)
	if err != nil {
		return err
	}
	entries = append(entries, RollbackEntry{
		Timestamp: time.Now().UTC(),
		Label:     label,
		Snapshot:  snap,
	})
	return SaveRollback(vaultPath, entries)
}

// Rollback restores the vault to the most recent checkpoint matching label.
// If label is empty, the latest checkpoint is used.
func Rollback(v *store.Vault, vaultPath, label string) error {
	entries, err := LoadRollback(vaultPath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no rollback checkpoints found")
	}
	var target *RollbackEntry
	for i := len(entries) - 1; i >= 0; i-- {
		if label == "" || entries[i].Label == label {
			target = &entries[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("no checkpoint found with label %q", label)
	}
	for _, k := range v.Keys() {
		v.Delete(k)
	}
	for k, val := range target.Snapshot {
		v.Set(k, val)
	}
	return nil
}

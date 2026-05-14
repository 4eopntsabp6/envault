package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nicholasgasior/envault/internal/store"
)

// HistoryEntry records a single change to a key.
type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // set, delete, rename
	Key       string    `json:"key"`
	OldValue  string    `json:"old_value,omitempty"`
	NewValue  string    `json:"new_value,omitempty"`
}

// HistoryManifest holds all history entries for a vault.
type HistoryManifest struct {
	Entries []HistoryEntry `json:"entries"`
}

// HistoryPath returns the path to the history file for a given vault.
func HistoryPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return filepath.Join(dir, name+".history.json")
}

// LoadHistory reads the history manifest from disk, returning an empty one if missing.
func LoadHistory(vaultPath string) (*HistoryManifest, error) {
	p := HistoryPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &HistoryManifest{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("history: read: %w", err)
	}
	var m HistoryManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("history: unmarshal: %w", err)
	}
	return &m, nil
}

// SaveHistory writes the history manifest to disk.
func SaveHistory(vaultPath string, m *HistoryManifest) error {
	p := HistoryPath(vaultPath)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}
	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("history: write: %w", err)
	}
	return nil
}

// RecordHistory appends an entry to the history manifest for the given vault.
func RecordHistory(v *store.Vault, action, key, oldValue, newValue string) error {
	m, err := LoadHistory(v.Path())
	if err != nil {
		return err
	}
	m.Entries = append(m.Entries, HistoryEntry{
		Timestamp: time.Now().UTC(),
		Action:    action,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  newValue,
	})
	return SaveHistory(v.Path(), m)
}

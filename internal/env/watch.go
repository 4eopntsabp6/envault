package env

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/envault/internal/store"
)

// WatchState records a fingerprint of vault contents at a point in time.
type WatchState struct {
	Fingerprint string            `json:"fingerprint"`
	Keys        map[string]string `json:"keys"`
	RecordedAt  time.Time         `json:"recorded_at"`
}

// WatchPath returns the path to the watch-state file for a vault.
func WatchPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".watch")
}

// Fingerprint computes a deterministic hash of all key-value pairs in the vault.
func Fingerprint(v *store.Vault) (string, error) {
	keys := v.Keys()
	h := sha256.New()
	for _, k := range keys {
		val, ok := v.Get(k)
		if !ok {
			continue
		}
		fmt.Fprintf(h, "%s=%s\n", k, val)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// SaveWatchState writes the current vault state to a watch file.
func SaveWatchState(vaultPath string, v *store.Vault) error {
	fp, err := Fingerprint(v)
	if err != nil {
		return err
	}
	keys := v.Keys()
	kv := make(map[string]string, len(keys))
	for _, k := range keys {
		val, _ := v.Get(k)
		kv[k] = val
	}
	ws := WatchState{
		Fingerprint: fp,
		Keys:        kv,
		RecordedAt:  time.Now().UTC(),
	}
	data, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(WatchPath(vaultPath), data, 0600)
}

// LoadWatchState reads the last saved watch state for a vault.
func LoadWatchState(vaultPath string) (*WatchState, error) {
	data, err := os.ReadFile(WatchPath(vaultPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ws WatchState
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

// HasChanged returns true if the vault contents differ from the last saved watch state.
func HasChanged(vaultPath string, v *store.Vault) (bool, error) {
	ws, err := LoadWatchState(vaultPath)
	if err != nil {
		return false, err
	}
	if ws == nil {
		return true, nil
	}
	current, err := Fingerprint(v)
	if err != nil {
		return false, err
	}
	return current != ws.Fingerprint, nil
}

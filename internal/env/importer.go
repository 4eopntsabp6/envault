package env

import (
	"fmt"
	"os"

	"github.com/user/envault/internal/store"
)

// ImportFile reads a .env file from path and stores every key-value pair into
// the provided Vault, overwriting existing keys.
func ImportFile(path string, v *store.Vault) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	entries, err := Parse(f)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}

	for _, e := range entries {
		v.Set(e.Key, e.Value)
	}

	return len(entries), nil
}

// ExportFile writes all secrets from the Vault to a .env file at path.
// If the file already exists it is overwritten.
func ExportFile(path string, v *store.Vault) (int, error) {
	keys := v.Keys()
	if len(keys) == 0 {
		return 0, nil
	}

	entries := make([]Entry, 0, len(keys))
	for _, k := range keys {
		val, ok := v.Get(k)
		if !ok {
			continue
		}
		entries = append(entries, Entry{Key: k, Value: val})
	}

	contents := Format(entries)

	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		return 0, fmt.Errorf("write %s: %w", path, err)
	}

	return len(entries), nil
}

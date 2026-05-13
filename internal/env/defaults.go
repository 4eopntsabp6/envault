package env

import (
	"fmt"
	"strings"

	"github.com/envault/envault/internal/store"
)

// DefaultEntry holds a key, its default value, and an optional description.
type DefaultEntry struct {
	Key          string
	DefaultValue string
	Description  string
}

// ApplyDefaults sets keys in the vault only if they are not already present.
// It returns the list of keys that were actually set.
func ApplyDefaults(v *store.Vault, entries []DefaultEntry) ([]string, error) {
	var applied []string
	for _, e := range entries {
		if err := ValidateKey(e.Key); err != nil {
			return applied, fmt.Errorf("invalid key %q: %w", e.Key, err)
		}
		if _, exists := v.Get(e.Key); exists {
			continue
		}
		v.Set(e.Key, e.DefaultValue)
		applied = append(applied, e.Key)
	}
	return applied, nil
}

// LoadDefaults parses a slice of "KEY=VALUE" or "KEY=VALUE # description" lines
// into DefaultEntry values. Lines starting with '#' and blank lines are skipped.
func LoadDefaults(lines []string) ([]DefaultEntry, error) {
	var entries []DefaultEntry
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		desc := ""
		if idx := strings.Index(line, " # "); idx != -1 {
			desc = strings.TrimSpace(line[idx+3:])
			line = line[:idx]
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed default entry: %q", raw)
		}
		entries = append(entries, DefaultEntry{
			Key:          strings.TrimSpace(parts[0]),
			DefaultValue: strings.TrimSpace(parts[1]),
			Description:  desc,
		})
	}
	return entries, nil
}

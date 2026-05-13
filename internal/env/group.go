package env

import (
	"fmt"
	"sort"
	"strings"

	"github.com/envault/envault/internal/store"
)

// GroupByPrefix groups vault keys by their prefix (the part before the first underscore).
// Keys without an underscore are placed under the "_" group.
func GroupByPrefix(v *store.Vault, password string) (map[string][]string, error) {
	keys, err := v.Keys(password)
	if err != nil {
		return nil, fmt.Errorf("group: list keys: %w", err)
	}

	groups := make(map[string][]string)
	for _, k := range keys {
		prefix := prefixOf(k)
		groups[prefix] = append(groups[prefix], k)
	}

	// Sort keys within each group for deterministic output.
	for prefix := range groups {
		sort.Strings(groups[prefix])
	}

	return groups, nil
}

// GroupNames returns a sorted list of group names from a grouped map.
func GroupNames(groups map[string][]string) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// prefixOf returns the portion of key before the first underscore,
// upper-cased. If there is no underscore the sentinel "_" is returned.
func prefixOf(key string) string {
	if idx := strings.Index(key, "_"); idx > 0 {
		return strings.ToUpper(key[:idx])
	}
	return "_"
}

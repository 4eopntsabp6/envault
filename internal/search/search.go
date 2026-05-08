// Package search provides functionality to search and filter vault secrets by key patterns.
package search

import (
	"strings"

	"github.com/user/envault/internal/store"
)

// Result holds a single search match.
type Result struct {
	Key   string
	Value string
}

// Options controls search behaviour.
type Options struct {
	// CaseSensitive determines whether matching is case-sensitive.
	CaseSensitive bool
	// ShowValues includes secret values in results (use with caution).
	ShowValues bool
}

// ByKeyPrefix returns all secrets whose key starts with the given prefix.
func ByKeyPrefix(v *store.Vault, prefix string, opts Options) []Result {
	return filter(v, func(k string) bool {
		if !opts.CaseSensitive {
			return strings.HasPrefix(strings.ToLower(k), strings.ToLower(prefix))
		}
		return strings.HasPrefix(k, prefix)
	}, opts)
}

// ByKeyContains returns all secrets whose key contains the given substring.
func ByKeyContains(v *store.Vault, substr string, opts Options) []Result {
	return filter(v, func(k string) bool {
		if !opts.CaseSensitive {
			return strings.Contains(strings.ToLower(k), strings.ToLower(substr))
		}
		return strings.Contains(k, substr)
	}, opts)
}

// filter applies a predicate over all vault keys and returns matching Results.
func filter(v *store.Vault, predicate func(string) bool, opts Options) []Result {
	keys := v.Keys()
	results := make([]Result, 0, len(keys))
	for _, k := range keys {
		if predicate(k) {
			r := Result{Key: k}
			if opts.ShowValues {
				val, _ := v.Get(k)
				r.Value = val
			}
			results = append(results, r)
		}
	}
	return results
}

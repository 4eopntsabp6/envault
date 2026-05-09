package env

import (
	"fmt"

	"github.com/user/envault/internal/store"
)

// MergeStrategy controls how conflicts are handled during a merge.
type MergeStrategy int

const (
	// MergeSkip keeps the existing value when a key already exists.
	MergeSkip MergeStrategy = iota
	// MergeOverwrite replaces the existing value with the incoming value.
	MergeOverwrite
)

// MergeResult holds the outcome of a merge operation.
type MergeResult struct {
	Added    []string
	Skipped  []string
	Overwritten []string
}

// MergeVaults merges all keys from src into dst using the given strategy.
// Both vaults must already be unlocked (i.e. loaded with the correct password).
func MergeVaults(dst, src *store.Vault, strategy MergeStrategy) (MergeResult, error) {
	var result MergeResult

	srcKeys := src.Keys()
	for _, key := range srcKeys {
		val, err := src.Get(key)
		if err != nil {
			return result, fmt.Errorf("merge: reading key %q from source: %w", key, err)
		}

		existing, getErr := dst.Get(key)
		keyExists := getErr == nil && existing != ""

		switch {
		case !keyExists:
			if err := dst.Set(key, val); err != nil {
				return result, fmt.Errorf("merge: writing key %q to destination: %w", key, err)
			}
			result.Added = append(result.Added, key)
		case strategy == MergeOverwrite:
			if err := dst.Set(key, val); err != nil {
				return result, fmt.Errorf("merge: overwriting key %q in destination: %w", key, err)
			}
			result.Overwritten = append(result.Overwritten, key)
		default:
			result.Skipped = append(result.Skipped, key)
		}
	}

	return result, nil
}

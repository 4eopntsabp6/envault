// Package copy provides functionality to copy secrets between vaults.
package copy

import (
	"fmt"

	"github.com/user/envault/internal/store"
)

// Options controls the behaviour of a copy operation.
type Options struct {
	// Overwrite existing keys in the destination vault.
	Overwrite bool
	// Keys is an optional allow-list; when non-empty only these keys are copied.
	Keys []string
}

// Result holds a summary of what happened during a copy.
type Result struct {
	Copied  []string
	Skipped []string
}

// Copy copies secrets from src into dst according to opts.
// Both vaults must already be loaded; the caller is responsible for saving dst.
func Copy(src, dst *store.Vault, opts Options) (Result, error) {
	if src == nil || dst == nil {
		return Result{}, fmt.Errorf("copy: src and dst must not be nil")
	}

	keys := opts.Keys
	if len(keys) == 0 {
		keys = src.Keys()
	}

	var result Result
	for _, k := range keys {
		val, ok := src.Get(k)
		if !ok {
			continue
		}

		_, exists := dst.Get(k)
		if exists && !opts.Overwrite {
			result.Skipped = append(result.Skipped, k)
			continue
		}

		dst.Set(k, val)
		result.Copied = append(result.Copied, k)
	}

	return result, nil
}

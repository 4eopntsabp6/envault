// Package env provides utilities for parsing, formatting, and managing
// environment variable files and vault operations.
package env

import (
	"fmt"

	"github.com/yourusername/envault/internal/store"
)

// CloneOptions controls the behaviour of CloneVault.
type CloneOptions struct {
	// Keys restricts cloning to the specified keys. If empty, all keys are cloned.
	Keys []string
	// Overwrite replaces existing keys in the destination vault.
	Overwrite bool
	// Transform is an optional function applied to each value before writing.
	// If nil, values are copied verbatim.
	Transform func(key, value string) string
}

// CloneVault copies secrets from src into dst according to opts.
// It returns the number of keys written and any error encountered.
func CloneVault(src, dst *store.Vault, opts CloneOptions) (int, error) {
	if src == nil {
		return 0, fmt.Errorf("clone: source vault must not be nil")
	}
	if dst == nil {
		return 0, fmt.Errorf("clone: destination vault must not be nil")
	}

	keys := opts.Keys
	if len(keys) == 0 {
		keys = src.Keys()
	}

	written := 0
	for _, k := range keys {
		v, ok := src.Get(k)
		if !ok {
			return written, fmt.Errorf("clone: key %q not found in source vault", k)
		}

		if !opts.Overwrite {
			if _, exists := dst.Get(k); exists {
				continue
			}
		}

		if opts.Transform != nil {
			v = opts.Transform(k, v)
		}

		dst.Set(k, v)
		written++
	}

	return written, nil
}

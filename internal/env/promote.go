package env

import (
	"fmt"

	"github.com/envault/envault/internal/store"
)

// PromoteOptions controls how promotion behaves.
type PromoteOptions struct {
	Overwrite bool
	Keys      []string // if empty, promote all keys
	DryRun    bool
}

// PromoteResult records what happened during promotion.
type PromoteResult struct {
	Promoted []string
	Skipped  []string
}

// PromoteVault copies keys from src vault into dst vault, optionally filtered
// by a list of keys. It is similar to Copy but is scoped to environment
// promotion workflows (e.g. staging → production).
func PromoteVault(src, dst *store.Vault, opts PromoteOptions) (PromoteResult, error) {
	var result PromoteResult

	keys := opts.Keys
	if len(keys) == 0 {
		keys = src.Keys()
	}

	for _, k := range keys {
		val, ok := src.Get(k)
		if !ok {
			return result, fmt.Errorf("promote: key %q not found in source vault", k)
		}

		if err := ValidateKey(k); err != nil {
			return result, fmt.Errorf("promote: invalid key %q: %w", k, err)
		}

		_, exists := dst.Get(k)
		if exists && !opts.Overwrite {
			result.Skipped = append(result.Skipped, k)
			continue
		}

		if !opts.DryRun {
			dst.Set(k, val)
		}
		result.Promoted = append(result.Promoted, k)
	}

	return result, nil
}

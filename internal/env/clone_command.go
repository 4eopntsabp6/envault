package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunClone copies secrets from one vault to another, optionally filtering by keys.
// Usage: envault clone <src-vault> <dst-vault> [--keys k1,k2] [--overwrite]
func RunClone(args []string, password string, keys []string, overwrite bool, w io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("clone requires source and destination vault paths")
	}
	srcPath := args[0]
	dstPath := args[1]

	src, err := store.Load(srcPath, password)
	if err != nil {
		return fmt.Errorf("failed to load source vault %q: %w", srcPath, err)
	}

	var dst *store.Vault
	if _, statErr := os.Stat(dstPath); os.IsNotExist(statErr) {
		dst = store.NewVault(dstPath, password)
	} else {
		dst, err = store.Load(dstPath, password)
		if err != nil {
			return fmt.Errorf("failed to load destination vault %q: %w", dstPath, err)
		}
	}

	copied, skipped, err := env.CloneVault(src, dst, keys, overwrite)
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	if err := dst.Save(); err != nil {
		return fmt.Errorf("failed to save destination vault: %w", err)
	}

	fmt.Fprintf(w, "Cloned vault: %s → %s\n", srcPath, dstPath)
	fmt.Fprintf(w, "  Copied:  %d keys\n", copied)
	fmt.Fprintf(w, "  Skipped: %d keys\n", skipped)
	return nil
}

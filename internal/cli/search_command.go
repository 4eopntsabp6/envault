package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/envault/internal/search"
	"github.com/user/envault/internal/store"
)

// RunSearch searches vault keys by prefix or substring and prints matches.
// mode must be "prefix" or "contains".
func RunSearch(vaultPath, password, query, mode string, showValues bool, w io.Writer) error {
	v := store.NewVault(password)
	if err := v.Load(vaultPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vault not found at %s", vaultPath)
		}
		return fmt.Errorf("failed to load vault: %w", err)
	}

	opts := search.Options{
		CaseSensitive: false,
		ShowValues:    showValues,
	}

	var results []search.Result
	switch strings.ToLower(mode) {
	case "prefix":
		results = search.ByKeyPrefix(v, query, opts)
	case "contains":
		results = search.ByKeyContains(v, query, opts)
	default:
		return fmt.Errorf("unknown search mode %q: use 'prefix' or 'contains'", mode)
	}

	if len(results) == 0 {
		fmt.Fprintln(w, "no matches found")
		return nil
	}

	for _, r := range results {
		if showValues {
			fmt.Fprintf(w, "%s=%s\n", r.Key, r.Value)
		} else {
			fmt.Fprintln(w, r.Key)
		}
	}
	return nil
}

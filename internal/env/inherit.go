package env

import (
	"os"
	"strings"

	"github.com/envault/envault/internal/store"
)

// InheritOptions controls how environment variables are inherited.
type InheritOptions struct {
	// Overwrite existing vault keys with OS env values.
	Overwrite bool
	// Prefix filters which OS env vars to import (empty = all).
	Prefix string
}

// InheritFromOS reads variables from the current OS environment and stores
// them in the vault. Only variables passing the prefix filter are imported.
// Returns the number of keys written.
func InheritFromOS(v *store.Vault, opts InheritOptions) (int, error) {
	count := 0
	for _, entry := range os.Environ() {
		idx := strings.IndexByte(entry, '=')
		if idx < 0 {
			continue
		}
		key := entry[:idx]
		val := entry[idx+1:]

		if opts.Prefix != "" && !strings.HasPrefix(key, opts.Prefix) {
			continue
		}

		if err := ValidateKey(key); err != nil {
			continue
		}

		if !opts.Overwrite {
			if _, ok := v.Get(key); ok {
				continue
			}
		}

		if err := v.Set(key, val); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// ExportToOS writes all keys from the vault into the current process
// environment using os.Setenv. Returns the number of keys exported.
func ExportToOS(v *store.Vault) (int, error) {
	count := 0
	for _, key := range v.Keys() {
		val, ok := v.Get(key)
		if !ok {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

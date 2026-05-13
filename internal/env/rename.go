package env

import (
	"fmt"

	"github.com/yourusername/envault/internal/store"
)

// RenameResult holds the outcome of a rename operation.
type RenameResult struct {
	OldKey string
	NewKey string
	Overwrote bool
}

// RenameKey renames a key in the vault from oldKey to newKey.
// If newKey already exists and overwrite is false, an error is returned.
// The old key is removed after the value is copied to the new key.
func RenameKey(v *store.Vault, oldKey, newKey string, overwrite bool) (RenameResult, error) {
	if err := ValidateKey(oldKey); err != nil {
		return RenameResult{}, fmt.Errorf("invalid old key: %w", err)
	}
	if err := ValidateKey(newKey); err != nil {
		return RenameResult{}, fmt.Errorf("invalid new key: %w", err)
	}

	val, ok := v.Get(oldKey)
	if !ok {
		return RenameResult{}, fmt.Errorf("key %q not found", oldKey)
	}

	_, exists := v.Get(newKey)
	if exists && !overwrite {
		return RenameResult{}, fmt.Errorf("key %q already exists; use --overwrite to replace it", newKey)
	}

	v.Set(newKey, val)
	v.Delete(oldKey)

	return RenameResult{
		OldKey:    oldKey,
		NewKey:    newKey,
		Overwrote: exists,
	}, nil
}

// Package env provides utilities for managing environment variables within a vault.
package env

import (
	"fmt"

	"github.com/envault/envault/internal/store"
)

// AliasManifest maps alias names to their canonical key names.
type AliasManifest map[string]string

// SetAlias creates an alias for an existing key in the vault.
// The alias itself is stored as a key with a special prefix so it can be
// resolved transparently. Returns an error if the target key does not exist.
func SetAlias(v *store.Vault, alias, targetKey string) error {
	if err := ValidateKey(alias); err != nil {
		return fmt.Errorf("invalid alias name: %w", err)
	}
	if err := ValidateKey(targetKey); err != nil {
		return fmt.Errorf("invalid target key: %w", err)
	}
	if _, err := v.Get(targetKey); err != nil {
		return fmt.Errorf("target key %q does not exist", targetKey)
	}
	aliasKey := aliasStoreKey(alias)
	return v.Set(aliasKey, targetKey)
}

// ResolveAlias returns the value of the key that alias points to.
// If alias is not registered, it falls back to looking up alias directly as a key.
func ResolveAlias(v *store.Vault, alias string) (string, error) {
	aliasKey := aliasStoreKey(alias)
	target, err := v.Get(aliasKey)
	if err == nil {
		// alias exists — resolve to the canonical key
		val, err2 := v.Get(target)
		if err2 != nil {
			return "", fmt.Errorf("alias %q points to missing key %q", alias, target)
		}
		return val, nil
	}
	// fall back to direct lookup
	return v.Get(alias)
}

// DeleteAlias removes a previously registered alias. It does not affect the
// underlying key the alias pointed to.
func DeleteAlias(v *store.Vault, alias string) error {
	aliasKey := aliasStoreKey(alias)
	if _, err := v.Get(aliasKey); err != nil {
		return fmt.Errorf("alias %q not found", alias)
	}
	return v.Delete(aliasKey)
}

// ListAliases returns a map of alias → target key for all registered aliases.
func ListAliases(v *store.Vault) AliasManifest {
	const prefix = "__alias__"
	manifest := make(AliasManifest)
	for _, k := range v.Keys() {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			alias := k[len(prefix):]
			if target, err := v.Get(k); err == nil {
				manifest[alias] = target
			}
		}
	}
	return manifest
}

func aliasStoreKey(alias string) string {
	return "__alias__" + alias
}

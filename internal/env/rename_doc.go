// Package env provides utilities for parsing, formatting, importing,
// exporting, validating, merging, and renaming environment variable entries
// stored in envault vaults.
//
// # Rename
//
// RenameKey moves a secret from one key name to another within the same vault.
// The source key must exist and the destination key must be a valid environment
// variable name. By default, renaming onto an existing key is rejected; pass
// overwrite=true to allow replacement.
//
// Example:
//
//	res, err := env.RenameKey(vault, "DB_PASS", "DATABASE_PASSWORD", false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("renamed %s -> %s\n", res.OldKey, res.NewKey)
package env

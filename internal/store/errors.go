package store

import "errors"

// ErrVaultNotFound is returned when no vault file exists in the target directory.
var ErrVaultNotFound = errors.New("vault not found: run 'envault init' first")

// ErrKeyNotFound is returned when a requested secret key does not exist in the vault.
var ErrKeyNotFound = errors.New("key not found in vault")

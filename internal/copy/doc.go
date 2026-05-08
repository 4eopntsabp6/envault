// Package copy implements secret-copying between envault vaults.
//
// It supports copying all keys or a selected subset, with optional overwrite
// of existing keys in the destination vault. The source and destination vaults
// may use different passwords — the caller is responsible for loading and
// saving each vault with the appropriate credentials.
//
// Example usage:
//
//	res, err := copy.Copy(srcVault, dstVault, copy.Options{
//		Overwrite: false,
//		Keys:      []string{"DB_URL", "API_KEY"},
//	})
package copy

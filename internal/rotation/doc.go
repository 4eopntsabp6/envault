// Package rotation provides functionality to re-encrypt vault secrets
// using a new master password without loss of data.
//
// The rotation process:
//  1. Loads all secrets from the vault using the old password.
//  2. Creates a fresh vault instance keyed with the new password.
//  3. Re-encrypts every secret and writes the vault back to disk.
//
// After a successful rotation, the vault can no longer be opened
// with the old password. Callers should record the rotation event
// via the audit log for compliance tracking.
package rotation

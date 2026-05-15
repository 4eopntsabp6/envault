// Package env provides utilities for managing environment variable secrets
// stored in encrypted vaults.
//
// # Quota
//
// The quota subsystem lets operators cap the number of keys and the maximum
// byte length of any single value stored in a vault.  Limits are persisted
// in a sidecar JSON file alongside the vault (e.g. myproject.quota.json).
//
// Usage:
//
//	// Set a quota: at most 50 keys, values ≤ 1 KiB each.
//	_ = env.SetQuota("/path/to/project.vault", 50, 1024)
//
//	// Before writing a new key, verify the quota is not exceeded.
//	if err := env.CheckQuota(v, "/path/to/project.vault", key, value); err != nil {
//	    log.Fatal(err)
//	}
//
// A MaxKeys or MaxValueSize of 0 means unlimited.
package env

// Package env provides utilities for managing environment variables within
// envault vaults.
//
// # Priority
//
// The priority sub-feature allows each key in a vault to be assigned a
// numeric importance level:
//
//   - PriorityLow    (1)  — informational or rarely-needed keys
//   - PriorityNormal (5)  — default level assigned to all new keys
//   - PriorityHigh   (10) — critical secrets that should be reviewed first
//
// Priority metadata is stored in a sidecar JSON file alongside the vault:
//
//	<vaultname>.priority.json
//
// Usage:
//
//	env.SetPriority(vault, env.PriorityHigh)("DATABASE_URL")
//	keys, _ := env.KeysByPriority(vault) // sorted high → low
package env

// Package env provides utilities for managing environment variable secrets
// stored in encrypted vaults.
//
// # Labels
//
// The label feature allows associating one or more string labels with any key
// stored in a vault. Labels are persisted in a sidecar JSON file alongside the
// vault (e.g. "project.vault.labels.json") and are never encrypted, as they
// are considered organisational metadata rather than secret data.
//
// Typical usage:
//
//	// Tag a key as sensitive and belonging to the payments team
//	env.SetLabels(vaultPath, vault, "STRIPE_KEY", []string{"sensitive", "payments"})
//
//	// Retrieve labels for a key
//	labels, _ := env.GetLabels(vaultPath, "STRIPE_KEY")
//
//	// Find all keys tagged as sensitive
//	keys, _ := env.FilterByLabel(vaultPath, "sensitive")
//
//	// Remove all labels from a key
//	env.DeleteLabels(vaultPath, "STRIPE_KEY")
//
// Labels are free-form strings. Consumers are encouraged to adopt a consistent
// labelling convention (e.g. environment names, team names, sensitivity tiers)
// to make filtering meaningful across large vaults.
package env

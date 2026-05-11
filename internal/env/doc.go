// Package env provides utilities for working with .env-style files and
// environment variable management within envault.
//
// Sub-features:
//
//   - Parse / Format: read and write KEY=VALUE lines from .env files.
//   - ImportFile / ExportFile: bulk-load or dump vault contents to disk.
//   - MergeVaults: combine two vaults with configurable overwrite semantics.
//   - ValidateKey / ValidateValue: enforce naming rules before storing secrets.
//   - InheritFromOS / ExportToOS: bridge between the OS process environment
//     and an envault vault, enabling shell-less secret injection.
package env

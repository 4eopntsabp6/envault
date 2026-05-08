// Package snapshot provides point-in-time capture and comparison of vault secrets.
//
// A snapshot is a JSON file containing all key-value pairs from a vault at a
// specific moment in time. Snapshots are stored under
// .envault/<vault-name>/snapshots/ relative to the vault file.
//
// # Usage
//
//	snap, err := snapshot.Take(vault, password)
//	if err != nil { ... }
//
//	path, err := snapshot.Save(snap, "/path/to/snapshots")
//
//	loaded, err := snapshot.Load(path)
//
//	added, removed, changed := snapshot.Diff(before, after)
//
// Diffs report:
//   - added:   keys present in after but not before
//   - removed: keys present in before but not after
//   - changed: keys present in both but with different values
package snapshot

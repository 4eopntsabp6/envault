package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/yourorg/envault/internal/diff"
	"github.com/yourorg/envault/internal/snapshot"
	"github.com/yourorg/envault/internal/store"
)

// RunDiff compares the current vault state against a named snapshot
// and prints the diff to stdout.
//
// Usage: envault diff <vault-path> <password> <snapshot-name> [--show-unchanged]
func RunDiff(vaultPath, password, snapshotName string, showUnchanged bool, out io.Writer) error {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	snapshotDir := snapshotDirForVault(vaultPath)
	snapshotFile := fmt.Sprintf("%s/%s.json", snapshotDir, snapshotName)

	snap, err := snapshot.Load(snapshotFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("snapshot %q not found", snapshotName)
		}
		return fmt.Errorf("load snapshot: %w", err)
	}

	// Build current state map
	current := make(map[string]string)
	for _, k := range v.Keys() {
		val, _ := v.Get(k)
		current[k] = val
	}

	changes := diff.Compare(snap.Secrets, current)

	if len(changes) == 0 {
		fmt.Fprintln(out, "No differences found.")
		return nil
	}

	fmt.Fprintf(out, "Diff against snapshot %q:\n", snapshotName)
	fmt.Fprint(out, diff.Format(changes, showUnchanged))
	return nil
}

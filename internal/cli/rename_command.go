package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/yourusername/envault/internal/env"
	"github.com/yourusername/envault/internal/store"
)

// RunRename renames a key within a vault.
// Usage: envault rename <vault> <old-key> <new-key> [--overwrite]
func RunRename(args []string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: envault rename <vault> <old-key> <new-key> [--overwrite]")
	}

	vaultPath := args[0]
	oldKey := args[1]
	newKey := args[2]

	overwrite := false
	for _, a := range args[3:] {
		if a == "--overwrite" {
			overwrite = true
		}
	}

	password, err := readPassword("Vault password: ")
	if err != nil {
		return fmt.Errorf("reading password: %w", err)
	}

	v := store.NewVault(password)
	if err := v.Load(vaultPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vault not found: %s", vaultPath)
		}
		return fmt.Errorf("loading vault: %w", err)
	}

	res, err := env.RenameKey(v, oldKey, newKey, overwrite)
	if err != nil {
		return err
	}

	if err := v.Save(vaultPath); err != nil {
		return fmt.Errorf("saving vault: %w", err)
	}

	RecordAudit(vaultPath, fmt.Sprintf("rename %s -> %s", res.OldKey, res.NewKey))

	if res.Overwrote {
		fmt.Fprintf(out, "renamed %q to %q (overwrote existing)\n", res.OldKey, res.NewKey)
	} else {
		fmt.Fprintf(out, "renamed %q to %q\n", res.OldKey, res.NewKey)
	}
	return nil
}

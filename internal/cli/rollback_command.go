package cli

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunRollback dispatches rollback subcommands: checkpoint, restore, list.
func RunRollback(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: rollback <checkpoint|restore|list> [label]")
	}
	switch args[0] {
	case "checkpoint":
		return runRollbackCheckpoint(args[1:], vaultPath, password, out)
	case "restore":
		return runRollbackRestore(args[1:], vaultPath, password, out)
	case "list":
		return runRollbackList(vaultPath, out)
	default:
		return fmt.Errorf("unknown rollback subcommand %q", args[0])
	}
}

func runRollbackCheckpoint(args []string, vaultPath, password string, out io.Writer) error {
	label := ""
	if len(args) > 0 {
		label = args[0]
	}
	v, err := loadVaultForRollback(vaultPath, password)
	if err != nil {
		return err
	}
	if err := env.Checkpoint(v, vaultPath, label); err != nil {
		return fmt.Errorf("checkpoint failed: %w", err)
	}
	fmt.Fprintf(out, "checkpoint saved: %q\n", label)
	return nil
}

func runRollbackRestore(args []string, vaultPath, password string, out io.Writer) error {
	label := ""
	if len(args) > 0 {
		label = args[0]
	}
	v, err := loadVaultForRollback(vaultPath, password)
	if err != nil {
		return err
	}
	if err := env.Rollback(v, vaultPath, label); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	if err := v.Save(password); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}
	fmt.Fprintf(out, "vault restored to checkpoint %q\n", label)
	return nil
}

func runRollbackList(vaultPath string, out io.Writer) error {
	entries, err := env.LoadRollback(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to load rollback journal: %w", err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(out, "no checkpoints found")
		return nil
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tLABEL\tKEYS")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%d\n", e.Timestamp.Format("2006-01-02T15:04:05Z"), e.Label, len(e.Snapshot))
	}
	return w.Flush()
}

func loadVaultForRollback(vaultPath, password string) (*store.Vault, error) {
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("vault not found: %s", vaultPath)
	}
	v := store.NewVault(vaultPath, password)
	if err := v.Load(password); err != nil {
		return nil, fmt.Errorf("failed to unlock vault: %w", err)
	}
	return v, nil
}

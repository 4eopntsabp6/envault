package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/nicholasgasior/envault/internal/env"
	"github.com/nicholasgasior/envault/internal/store"
)

// RunQuota dispatches quota subcommands: set, show.
func RunQuota(args []string, password string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault quota <vault> <set|show> [max-keys] [max-value-bytes]")
	}
	vaultPath := args[0]
	subcmd := args[1]

	switch subcmd {
	case "set":
		return runQuotaSet(vaultPath, password, args[2:], out)
	case "show":
		return runQuotaShow(vaultPath, password, out)
	default:
		return fmt.Errorf("unknown quota subcommand: %s", subcmd)
	}
}

func runQuotaSet(vaultPath, password string, args []string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault quota <vault> set <max-keys> <max-value-bytes>")
	}
	maxKeys, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid max-keys %q: %w", args[0], err)
	}
	maxValue, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid max-value-bytes %q: %w", args[1], err)
	}
	// Validate vault is accessible.
	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil && !isNotExist(err) {
		return fmt.Errorf("loading vault: %w", err)
	}
	if err := env.SetQuota(vaultPath, maxKeys, maxValue); err != nil {
		return fmt.Errorf("saving quota: %w", err)
	}
	fmt.Fprintf(out, "quota set: max_keys=%d max_value_bytes=%d\n", maxKeys, maxValue)
	return nil
}

func runQuotaShow(vaultPath, password string, out io.Writer) error {
	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil && !isNotExist(err) {
		return fmt.Errorf("loading vault: %w", err)
	}
	m, err := env.LoadQuota(vaultPath)
	if err != nil {
		return fmt.Errorf("loading quota: %w", err)
	}
	if m.MaxKeys == 0 && m.MaxValueSize == 0 {
		fmt.Fprintln(out, "no quota configured")
		return nil
	}
	if m.MaxKeys > 0 {
		fmt.Fprintf(out, "max_keys: %d\n", m.MaxKeys)
	} else {
		fmt.Fprintln(out, "max_keys: unlimited")
	}
	if m.MaxValueSize > 0 {
		fmt.Fprintf(out, "max_value_bytes: %d\n", m.MaxValueSize)
	} else {
		fmt.Fprintln(out, "max_value_bytes: unlimited")
	}
	return nil
}

// isNotExist returns true for vault-not-found errors to allow first-time setup.
func isNotExist(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == store.ErrVaultNotFound.Error()
}

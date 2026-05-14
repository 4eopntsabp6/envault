package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/nicholasgasior/envault/internal/env"
	"github.com/nicholasgasior/envault/internal/store"
)

// RunLock handles the "lock" subcommand: lock, unlock, or list locked keys.
//
// Usage:
//
//	envault lock <vault> <password> set <key> [reason]
//	envault lock <vault> <password> unset <key>
//	envault lock <vault> <password> list
func RunLock(args []string, out io.Writer) error {
	if len(args) < 3 {
		return errors.New("usage: lock <vault> <password> <set|unset|list> [key] [reason]")
	}
	vaultPath := args[0]
	password := args[1]
	subcmd := args[2]

	v := store.NewVault(vaultPath)
	if err := v.Load(password); err != nil {
		return fmt.Errorf("failed to load vault: %w", err)
	}

	switch subcmd {
	case "set":
		return runLockSet(v, args[3:], out)
	case "unset":
		return runLockUnset(v, args[3:], out)
	case "list":
		return runLockList(v, out)
	default:
		return fmt.Errorf("unknown subcommand %q; expected set, unset, or list", subcmd)
	}
}

func runLockSet(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 1 {
		return errors.New("lock set requires a key name")
	}
	key := args[0]
	reason := ""
	if len(args) >= 2 {
		reason = args[1]
	}
	if err := env.LockKey(v, key, reason); err != nil {
		return err
	}
	fmt.Fprintf(out, "locked %s\n", key)
	return nil
}

func runLockUnset(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 1 {
		return errors.New("lock unset requires a key name")
	}
	key := args[0]
	if err := env.UnlockKey(v, key); err != nil {
		return err
	}
	fmt.Fprintf(out, "unlocked %s\n", key)
	return nil
}

func runLockList(v *store.Vault, out io.Writer) error {
	locked, err := env.ListLocked(v)
	if err != nil {
		return err
	}
	if len(locked) == 0 {
		fmt.Fprintln(out, "no locked keys")
		return nil
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tLOCKED AT\tREASON")
	for key, entry := range locked {
		fmt.Fprintf(w, "%s\t%s\t%s\n", key, entry.LockedAt.Format("2006-01-02 15:04:05"), entry.Reason)
	}
	return w.Flush()
}

// GuardLocked returns an error if the given key is locked, preventing writes.
func GuardLocked(v *store.Vault, key string) error {
	locked, err := env.IsLocked(v, key)
	if err != nil {
		return err
	}
	if locked {
		return fmt.Errorf("key %q is locked and cannot be modified; unlock it first", key)
	}
	return nil
}

// lockMain is the entry point wired into main.go.
func lockMain(args []string) {
	if err := RunLock(args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/nicholasgasior/envault/internal/env"
	"github.com/nicholasgasior/envault/internal/store"
)

// RunPin handles the 'pin' subcommand: pin, unpin, or list pinned keys.
//
// Usage:
//
//	envault pin set   <vault> <password> <key>
//	envault pin unset <vault> <password> <key>
//	envault pin list  <vault> <password>
func RunPin(args []string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: pin <set|unset|list> <vault> <password> [key]")
	}
	subcmd := args[0]
	vaultPath := args[1]
	password := args[2]

	v, err := loadPinVault(vaultPath, password)
	if err != nil {
		return err
	}

	switch subcmd {
	case "set":
		if len(args) < 4 {
			return fmt.Errorf("pin set requires a key argument")
		}
		key := args[3]
		if err := env.PinKey(v, vaultPath, key); err != nil {
			return fmt.Errorf("pin set: %w", err)
		}
		fmt.Fprintf(out, "pinned %q\n", key)

	case "unset":
		if len(args) < 4 {
			return fmt.Errorf("pin unset requires a key argument")
		}
		key := args[3]
		if err := env.UnpinKey(vaultPath, key); err != nil {
			return fmt.Errorf("pin unset: %w", err)
		}
		fmt.Fprintf(out, "unpinned %q\n", key)

	case "list":
		m, err := env.LoadPinManifest(vaultPath)
		if err != nil {
			return fmt.Errorf("pin list: %w", err)
		}
		if len(m.Pinned) == 0 {
			fmt.Fprintln(out, "no pinned keys")
			return nil
		}
		for k, val := range m.Pinned {
			fmt.Fprintf(out, "%s=%s\n", k, val)
		}

	default:
		return fmt.Errorf("unknown pin subcommand %q; use set, unset, or list", subcmd)
	}
	return nil
}

func loadPinVault(vaultPath, password string) (*store.Vault, error) {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vault not found: %s", vaultPath)
		}
		return nil, fmt.Errorf("load vault: %w", err)
	}
	return v, nil
}

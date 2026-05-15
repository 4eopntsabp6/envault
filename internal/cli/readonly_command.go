package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunReadonly dispatches readonly subcommands: set, unset, list.
func RunReadonly(args []string, password string, vaultPath string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault readonly <set|unset|list> [key]")
	}

	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	switch args[0] {
	case "set":
		return runReadonlySet(v, args[1:], out)
	case "unset":
		return runReadonlyUnset(v, args[1:], out)
	case "list":
		return runReadonlyList(v, out)
	default:
		return fmt.Errorf("unknown subcommand %q; expected set, unset, or list", args[0])
	}
}

func runReadonlySet(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault readonly set <key>")
	}
	key := args[0]
	if err := env.SetReadonly(v, key, true); err != nil {
		return err
	}
	fmt.Fprintf(out, "Key %q marked as read-only.\n", key)
	return nil
}

func runReadonlyUnset(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault readonly unset <key>")
	}
	key := args[0]
	if err := env.SetReadonly(v, key, false); err != nil {
		return err
	}
	fmt.Fprintf(out, "Key %q is no longer read-only.\n", key)
	return nil
}

func runReadonlyList(v *store.Vault, out io.Writer) error {
	keys, err := env.ReadonlyKeys(v)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		fmt.Fprintln(out, "No read-only keys.")
		return nil
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintln(out, k)
	}
	return nil
}

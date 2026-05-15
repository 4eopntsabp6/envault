package cli

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunDescription dispatches description sub-commands: set, get, delete, list.
func RunDescription(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: envault description <set|get|delete|list> [key] [description]")
	}

	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	switch args[0] {
	case "set":
		return runDescriptionSet(v, args[1:], out)
	case "get":
		return runDescriptionGet(v, args[1:], out)
	case "delete":
		return runDescriptionDelete(v, args[1:], out)
	case "list":
		return runDescriptionList(v, out)
	default:
		return fmt.Errorf("unknown sub-command %q; expected set, get, delete, or list", args[0])
	}
}

func runDescriptionSet(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault description set <key> <description>")
	}
	key, desc := args[0], args[1]
	if err := env.SetDescription(v, key, desc); err != nil {
		return err
	}
	fmt.Fprintf(out, "description set for %q\n", key)
	return nil
}

func runDescriptionGet(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault description get <key>")
	}
	desc, err := env.GetDescription(v, args[0])
	if err != nil {
		return err
	}
	if desc == "" {
		fmt.Fprintf(out, "(no description set for %q)\n", args[0])
	} else {
		fmt.Fprintln(out, desc)
	}
	return nil
}

func runDescriptionDelete(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault description delete <key>")
	}
	if err := env.DeleteDescription(v, args[0]); err != nil {
		return err
	}
	fmt.Fprintf(out, "description removed for %q\n", args[0])
	return nil
}

func runDescriptionList(v *store.Vault, out io.Writer) error {
	all, err := env.ListDescriptions(v)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Fprintln(out, "(no descriptions set)")
		return nil
	}
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(out, "%-20s  %s\n", k, all[k])
	}
	_ = os.Stdout // satisfy import if out is redirected
	return nil
}

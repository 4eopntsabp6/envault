package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

// RunAlias dispatches alias sub-commands: set, resolve, delete, list.
// Usage:
//
//	envault alias set   <alias> <target-key> <vault> <password>
//	envault alias get   <alias> <vault> <password>
//	envault alias delete <alias> <vault> <password>
//	envault alias list  <vault> <password>
func RunAlias(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: alias <set|get|delete|list> [args...]")
	}
	subcmd := args[0]
	rest := args[1:]

	switch subcmd {
	case "set":
		return runAliasSet(rest, out)
	case "get":
		return runAliasGet(rest, out)
	case "delete":
		return runAliasDelete(rest, out)
	case "list":
		return runAliasList(rest, out)
	default:
		return fmt.Errorf("unknown alias sub-command %q", subcmd)
	}
}

func runAliasSet(args []string, out io.Writer) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: alias set <alias> <target-key> <vault> <password>")
	}
	alias, target, vaultPath, password := args[0], args[1], args[2], args[3]
	v, err := loadVault(vaultPath, password)
	if err != nil {
		return err
	}
	if err := env.SetAlias(v, alias, target); err != nil {
		return err
	}
	if err := v.Save(password); err != nil {
		return err
	}
	fmt.Fprintf(out, "alias %q -> %q set\n", alias, target)
	return nil
}

func runAliasGet(args []string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: alias get <alias> <vault> <password>")
	}
	alias, vaultPath, password := args[0], args[1], args[2]
	v, err := loadVault(vaultPath, password)
	if err != nil {
		return err
	}
	val, err := env.ResolveAlias(v, alias)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, val)
	return nil
}

func runAliasDelete(args []string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: alias delete <alias> <vault> <password>")
	}
	alias, vaultPath, password := args[0], args[1], args[2]
	v, err := loadVault(vaultPath, password)
	if err != nil {
		return err
	}
	if err := env.DeleteAlias(v, alias); err != nil {
		return err
	}
	if err := v.Save(password); err != nil {
		return err
	}
	fmt.Fprintf(out, "alias %q deleted\n", alias)
	return nil
}

func runAliasList(args []string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: alias list <vault> <password>")
	}
	vaultPath, password := args[0], args[1]
	v, err := loadVault(vaultPath, password)
	if err != nil {
		return err
	}
	manifest := env.ListAliases(v)
	if len(manifest) == 0 {
		fmt.Fprintln(out, "(no aliases defined)")
		return nil
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	keys := make([]string, 0, len(manifest))
	for k := range manifest {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(tw, "%s\t->\t%s\n", k, manifest[k])
	}
	return tw.Flush()
}

func loadVault(vaultPath, password string) (*store.Vault, error) {
	v := store.NewVault(vaultPath, password)
	if _, err := os.Stat(vaultPath); err == nil {
		if err := v.Load(password); err != nil {
			return nil, fmt.Errorf("failed to load vault: %w", err)
		}
	}
	return v, nil
}

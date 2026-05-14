package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunScope dispatches scope subcommands: set, apply, list.
func RunScope(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("scope: subcommand required (set|apply|list)")
	}
	switch args[0] {
	case "set":
		return runScopeSet(args[1:], vaultPath, password, out)
	case "apply":
		return runScopeApply(args[1:], vaultPath, password, out)
	case "list":
		return runScopeList(vaultPath, out)
	default:
		return fmt.Errorf("scope: unknown subcommand %q", args[0])
	}
}

// runScopeSet captures specified keys (or all) from the vault into a named scope.
// Usage: scope set <name> [key1 key2 ...]
func runScopeSet(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("scope set: requires <name>")
	}
	scopeName := args[0]

	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("scope set: %w", err)
	}

	keys := args[1:]
	if len(keys) == 0 {
		keys = v.Keys()
	}
	if len(keys) == 0 {
		return fmt.Errorf("scope set: vault is empty")
	}

	if err := env.SetScope(v, vaultPath, scopeName, keys); err != nil {
		return fmt.Errorf("scope set: %w", err)
	}
	fmt.Fprintf(out, "scope %q saved with %d key(s)\n", scopeName, len(keys))
	return nil
}

// runScopeApply restores keys from a named scope into the vault.
// Usage: scope apply <name> [--overwrite]
func runScopeApply(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("scope apply: requires <name>")
	}
	scopeName := args[0]
	overwrite := false
	for _, a := range args[1:] {
		if a == "--overwrite" {
			overwrite = true
		}
	}

	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("scope apply: %w", err)
	}

	applied, err := env.ApplyScope(v, vaultPath, scopeName, overwrite)
	if err != nil {
		return fmt.Errorf("scope apply: %w", err)
	}

	if err := v.Save(); err != nil {
		return fmt.Errorf("scope apply: save: %w", err)
	}

	if len(applied) == 0 {
		fmt.Fprintln(out, "no keys applied (all exist; use --overwrite to force)")
	} else {
		fmt.Fprintf(out, "applied scope %q: %s\n", scopeName, strings.Join(applied, ", "))
	}
	return nil
}

// runScopeList prints all scope names for the vault.
func runScopeList(vaultPath string, out io.Writer) error {
	names, err := env.ListScopes(vaultPath)
	if err != nil {
		return fmt.Errorf("scope list: %w", err)
	}
	if len(names) == 0 {
		fmt.Fprintln(out, "no scopes defined")
		return nil
	}
	for _, n := range names {
		fmt.Fprintln(out, n)
	}
	return nil
}

func init() {
	_ = os.Stderr // ensure os import used
}

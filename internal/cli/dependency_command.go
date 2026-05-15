package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunDependency dispatches dependency sub-commands: set, get, check.
func RunDependency(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("dependency: sub-command required (set|get|check)")
	}
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("dependency: load vault: %w", err)
	}
	switch args[0] {
	case "set":
		return runDependencySet(args[1:], v, out)
	case "get":
		return runDependencyGet(args[1:], v, out)
	case "check":
		return runDependencyCheck(args[1:], v, out)
	default:
		return fmt.Errorf("dependency: unknown sub-command %q", args[0])
	}
}

// runDependencySet records dependencies for a key.
// Usage: set <key> <dep1> [dep2 ...]
func runDependencySet(args []string, v *store.Vault, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("dependency set: usage: set <key> <dep1> [dep2 ...]")
	}
	key := args[0]
	deps := args[1:]
	if err := env.SetDependencies(v, key, deps); err != nil {
		return err
	}
	fmt.Fprintf(out, "dependency: set %d dep(s) for %q\n", len(deps), key)
	return nil
}

// runDependencyGet prints the dependencies for a key.
// Usage: get <key>
func runDependencyGet(args []string, v *store.Vault, out io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("dependency get: usage: get <key>")
	}
	key := args[0]
	deps, err := env.GetDependencies(v, key)
	if err != nil {
		return err
	}
	if len(deps) == 0 {
		fmt.Fprintf(out, "%s has no dependencies\n", key)
		return nil
	}
	fmt.Fprintf(out, "%s depends on: %s\n", key, strings.Join(deps, ", "))
	return nil
}

// runDependencyCheck reports any missing dependency keys for a given key.
// Usage: check <key>
func runDependencyCheck(args []string, v *store.Vault, out io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("dependency check: usage: check <key>")
	}
	key := args[0]
	missing, err := env.CheckMissing(v, key)
	if err != nil {
		return err
	}
	if len(missing) == 0 {
		fmt.Fprintf(out, "all dependencies satisfied for %q\n", key)
		return nil
	}
	fmt.Fprintf(out, "missing dependencies for %q:\n", key)
	for _, m := range missing {
		fmt.Fprintf(out, "  - %s\n", m)
	}
	os.Exit(1)
	return nil
}

package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

// RunRedact handles the `envault redact` sub-commands.
// Usage:
//
//	envault redact set   <vault> <key>  [--auto]
//	envault redact unset <vault> <key>
//	envault redact list  <vault>
func RunRedact(args []string, password string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: redact <set|unset|list> <vault> [key]")
	}
	subcmd := args[0]
	vaultPath := args[1]

	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load vault: %w", err)
	}

	switch subcmd {
	case "set":
		return runRedactSet(args[2:], v, vaultPath, password, out)
	case "unset":
		return runRedactUnset(args[2:], v, vaultPath, password, out)
	case "list":
		return runRedactList(v, out)
	default:
		return fmt.Errorf("unknown redact subcommand: %s", subcmd)
	}
}

func runRedactSet(args []string, v *store.Vault, vaultPath, password string, out io.Writer) error {
	auto := len(args) == 1 && args[0] == "--auto"
	var keys []string

	if auto {
		for _, k := range v.Keys() {
			if env.AutoDetectSensitive(k) {
				keys = append(keys, k)
			}
		}
		if len(keys) == 0 {
			fmt.Fprintln(out, "no sensitive keys auto-detected")
			return nil
		}
	} else {
		if len(args) == 0 {
			return fmt.Errorf("usage: redact set <vault> <key|--auto>")
		}
		keys = args
	}

	for _, k := range keys {
		if err := env.SetRedacted(v, k, true); err != nil {
			return fmt.Errorf("set redacted %s: %w", k, err)
		}
		fmt.Fprintf(out, "redacted: %s\n", k)
	}
	return v.Save()
}

func runRedactUnset(args []string, v *store.Vault, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: redact unset <vault> <key>")
	}
	for _, k := range args {
		if err := env.SetRedacted(v, k, false); err != nil {
			return fmt.Errorf("unset redacted %s: %w", k, err)
		}
		fmt.Fprintf(out, "unredacted: %s\n", k)
	}
	return v.Save()
}

func runRedactList(v *store.Vault, out io.Writer) error {
	keys := env.RedactedKeys(v)
	if len(keys) == 0 {
		fmt.Fprintln(out, "no redacted keys")
		return nil
	}
	sort.Strings(keys)
	fmt.Fprintln(out, strings.Join(keys, "\n"))
	return nil
}

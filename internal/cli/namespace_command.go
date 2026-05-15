package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

// RunNamespace dispatches namespace subcommands: assign, get, list, remove.
func RunNamespace(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: envault namespace <assign|get|list|remove> [args...]")
	}
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("failed to load vault: %w", err)
	}
	switch args[0] {
	case "assign":
		return runNamespaceAssign(v, args[1:], out)
	case "get":
		return runNamespaceGet(v, args[1:], out)
	case "list":
		return runNamespaceList(v, out)
	case "remove":
		return runNamespaceRemove(v, args[1:], out)
	default:
		return fmt.Errorf("unknown namespace subcommand: %q", args[0])
	}
}

func runNamespaceAssign(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault namespace assign <namespace> <key> [key...]")
	}
	namespace := args[0]
	keys := args[1:]
	for _, key := range keys {
		if err := env.AssignNamespace(v, namespace, key); err != nil {
			return fmt.Errorf("assign %q to %q: %w", key, namespace, err)
		}
		fmt.Fprintf(out, "assigned %q to namespace %q\n", key, namespace)
	}
	return nil
}

func runNamespaceGet(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault namespace get <namespace>")
	}
	namespace := args[0]
	keys, err := env.GetNamespaceKeys(v, namespace)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		fmt.Fprintf(out, "namespace %q is empty\n", namespace)
		return nil
	}
	fmt.Fprintf(out, "namespace %q:\n", namespace)
	for _, k := range keys {
		fmt.Fprintf(out, "  %s\n", k)
	}
	return nil
}

func runNamespaceList(v *store.Vault, out io.Writer) error {
	names, err := env.ListNamespaces(v)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintln(out, "no namespaces defined")
		return nil
	}
	fmt.Fprintln(out, "namespaces:")
	for _, ns := range names {
		keys, _ := env.GetNamespaceKeys(v, ns)
		fmt.Fprintf(out, "  %-20s [%s]\n", ns, strings.Join(keys, ", "))
	}
	return nil
}

func runNamespaceRemove(v *store.Vault, args []string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault namespace remove <namespace> <key> [key...]")
	}
	namespace := args[0]
	keys := args[1:]
	for _, key := range keys {
		if err := env.RemoveFromNamespace(v, namespace, key); err != nil {
			return fmt.Errorf("remove %q from %q: %w", key, namespace, err)
		}
		fmt.Fprintf(out, "removed %q from namespace %q\n", key, namespace)
	}
	return nil
}

// init registers the namespace command in the CLI if a registry is used.
func init() {
	_ = os.Stderr // ensure os is used
}

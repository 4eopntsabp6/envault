package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunLabel dispatches label sub-commands: set, get, filter, delete.
func RunLabel(args []string, password string, vaultPath string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("label: expected sub-command: set|get|filter|delete")
	}
	switch args[0] {
	case "set":
		return runLabelSet(args[1:], password, vaultPath, out)
	case "get":
		return runLabelGet(args[1:], vaultPath, out)
	case "filter":
		return runLabelFilter(args[1:], vaultPath, out)
	case "delete":
		return runLabelDelete(args[1:], vaultPath, out)
	default:
		return fmt.Errorf("label: unknown sub-command %q", args[0])
	}
}

func runLabelSet(args []string, password, vaultPath string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("label set: usage: set <key> <label>[,<label>...]")
	}
	key := args[0]
	labels := strings.Split(args[1], ",")
	for i, l := range labels {
		labels[i] = strings.TrimSpace(l)
	}

	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("label set: %w", err)
	}
	if err := env.SetLabels(vaultPath, v, key, labels); err != nil {
		return err
	}
	fmt.Fprintf(out, "labels set for %q: %s\n", key, strings.Join(labels, ", "))
	return nil
}

func runLabelGet(args []string, vaultPath string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("label get: usage: get <key>")
	}
	labels, err := env.GetLabels(vaultPath, args[0])
	if err != nil {
		return err
	}
	if len(labels) == 0 {
		fmt.Fprintf(out, "no labels for %q\n", args[0])
		return nil
	}
	fmt.Fprintf(out, "%s\n", strings.Join(labels, ", "))
	return nil
}

func runLabelFilter(args []string, vaultPath string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("label filter: usage: filter <label>")
	}
	keys, err := env.FilterByLabel(vaultPath, args[0])
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		fmt.Fprintf(out, "no keys with label %q\n", args[0])
		return nil
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintln(out, k)
	}
	return nil
}

func runLabelDelete(args []string, vaultPath string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("label delete: usage: delete <key>")
	}
	if err := env.DeleteLabels(vaultPath, args[0]); err != nil {
		return err
	}
	fmt.Fprintf(out, "labels deleted for %q\n", args[0])
	_ = os.Stdout // suppress unused import
	return nil
}

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

// RunTransform applies a named transform to the value of a key in a vault.
// Usage: envault transform <vault> <key> <transform> [--password=<pw>]
//
// Supported transforms: upper, lower, trim, base64, reverse
func RunTransform(args []string, password string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: transform <vault-path> <key> <transform>")
	}
	vaultPath := args[0]
	key := args[1]
	transformName := strings.ToLower(args[2])

	v := store.NewVault(vaultPath)
	if err := v.Load(password); err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	origVal, err := v.Get(password, key)
	if err != nil {
		return fmt.Errorf("key %q not found in vault", key)
	}

	newVal, err := env.ApplyTransform(v, password, key, transformName)
	if err != nil {
		return fmt.Errorf("transform: %w", err)
	}

	if err := v.Save(password); err != nil {
		return fmt.Errorf("save vault: %w", err)
	}

	RecordAudit(vaultPath, fmt.Sprintf("transform key=%s transform=%s", key, transformName))

	fmt.Fprintf(out, "Transformed %q with %q\n", key, transformName)
	fmt.Fprintf(out, "  before: %s\n", origVal)
	fmt.Fprintf(out, "  after:  %s\n", newVal)
	return nil
}

// RunTransformList prints all available built-in transforms.
func RunTransformList(out io.Writer) error {
	fmt.Fprintln(out, "Available transforms:")
	for name := range env.BuiltinTransforms {
		fmt.Fprintf(out, "  - %s\n", name)
	}
	return nil
}

// transformMain is the entry point called from main.go dispatch.
func transformMain(args []string) {
	if len(args) == 0 || args[0] == "list" {
		if err := RunTransformList(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	password := os.Getenv("ENVAULT_PASSWORD")
	if password == "" {
		var err error
		password, err = readPassword("Password: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading password: %v\n", err)
			os.Exit(1)
		}
	}
	if err := RunTransform(args, password, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

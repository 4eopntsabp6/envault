package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/user/envault/internal/lint"
	"github.com/user/envault/internal/store"
)

// RunLint loads the vault at vaultPath and prints all lint issues to out.
// It returns a non-nil error if the vault cannot be loaded.
// When issues are found it writes them and returns a sentinel so callers can
// exit with a non-zero status code.
func RunLint(vaultPath, password string, out io.Writer) error {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	issues := lint.Run(v, lint.DefaultRules)
	if len(issues) == 0 {
		fmt.Fprintln(out, "✓ No issues found.")
		return nil
	}

	fmt.Fprintf(out, "Found %d issue(s):\n", len(issues))
	for _, iss := range issues {
		fmt.Fprintf(out, "  [WARN] %s\n", iss)
	}
	return fmt.Errorf("%d lint issue(s) detected", len(issues))
}

// RunLintCmd is the entry point wired up from main.
func RunLintCmd(args []string, vaultPath, password string) {
	err := RunLint(vaultPath, password, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

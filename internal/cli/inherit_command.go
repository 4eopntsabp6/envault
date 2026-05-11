package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

// RunInherit imports OS environment variables into the vault at vaultPath.
// prefix restricts which variables are imported; overwrite controls whether
// existing vault keys are replaced.
func RunInherit(vaultPath, password, prefix string, overwrite bool, out io.Writer) error {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	opts := env.InheritOptions{
		Prefix:    prefix,
		Overwrite: overwrite,
	}

	n, err := env.InheritFromOS(v, opts)
	if err != nil {
		return fmt.Errorf("inherit: %w", err)
	}

	if err := store.Save(v, vaultPath, password); err != nil {
		return fmt.Errorf("save vault: %w", err)
	}

	fmt.Fprintf(out, "Imported %d variable(s) from environment.\n", n)
	RecordAudit(vaultPath, fmt.Sprintf("inherit prefix=%q overwrite=%v count=%d", prefix, overwrite, n))
	return nil
}

// RunExportToOS loads the vault and sets all keys in the current process
// environment. Useful when envault is used as a library / subprocess launcher.
func RunExportToOS(vaultPath, password string, out io.Writer) error {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	n, err := env.ExportToOS(v)
	if err != nil {
		return fmt.Errorf("export to OS: %w", err)
	}

	fmt.Fprintf(out, "Exported %d variable(s) to process environment.\n", n)
	return nil
}

// inheritMain is a thin wrapper called from main.go for the "inherit" sub-command.
func inheritMain(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: envault inherit <vault> [--prefix PREFIX] [--overwrite]")
		os.Exit(1)
	}
	vaultPath := args[0]
	password := args[1]
	prefix := ""
	overwrite := false
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--overwrite":
			overwrite = true
		case "--prefix":
			i++
			if i < len(args) {
				prefix = args[i]
			}
		}
	}
	if err := RunInherit(vaultPath, password, prefix, overwrite, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

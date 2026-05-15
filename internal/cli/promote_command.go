package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

// RunPromote promotes keys from one vault (source) to another (destination).
// Usage: envault promote <src-vault> <dst-vault> [flags]
func RunPromote(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("promote", flag.ContinueOnError)
	overwrite := fs.Bool("overwrite", false, "overwrite existing keys in destination")
	dryRun := fs.Bool("dry-run", false, "show what would be promoted without making changes")
	keysFlag := fs.String("keys", "", "comma-separated list of keys to promote (default: all)")
	password := fs.String("password", "", "shared password for both vaults")

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) < 2 {
		return fmt.Errorf("usage: promote <src-vault> <dst-vault> [flags]")
	}

	srcPath := remaining[0]
	dstPath := remaining[1]

	pwd := *password
	if pwd == "" {
		pwd = os.Getenv("ENVAULT_PASSWORD")
	}
	if pwd == "" {
		return fmt.Errorf("password required: use --password or ENVAULT_PASSWORD")
	}

	srcVault, err := store.Load(srcPath, pwd)
	if err != nil {
		return fmt.Errorf("load source vault: %w", err)
	}

	dstVault, err := store.Load(dstPath, pwd)
	if err != nil {
		return fmt.Errorf("load destination vault: %w", err)
	}

	var selectedKeys []string
	if *keysFlag != "" {
		for _, k := range strings.Split(*keysFlag, ",") {
			if k = strings.TrimSpace(k); k != "" {
				selectedKeys = append(selectedKeys, k)
			}
		}
	}

	opts := env.PromoteOptions{
		Overwrite: *overwrite,
		DryRun:    *dryRun,
		Keys:      selectedKeys,
	}

	result, err := env.PromoteVault(srcVault, dstVault, opts)
	if err != nil {
		return fmt.Errorf("promote: %w", err)
	}

	if *dryRun {
		fmt.Fprintf(stdout, "[dry-run] would promote %d key(s):\n", len(result.Promoted))
		for _, k := range result.Promoted {
			fmt.Fprintf(stdout, "  + %s\n", k)
		}
		return nil
	}

	if err := store.Save(dstVault, dstPath, pwd); err != nil {
		return fmt.Errorf("save destination vault: %w", err)
	}

	fmt.Fprintf(stdout, "promoted %d key(s) to %s\n", len(result.Promoted), dstPath)
	for _, k := range result.Skipped {
		fmt.Fprintf(stdout, "  skipped (exists): %s\n", k)
	}
	return nil
}

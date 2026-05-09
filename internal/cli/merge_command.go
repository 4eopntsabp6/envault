package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunMerge merges secrets from srcPath into dstPath.
// Both vaults are opened with their respective passwords.
// strategy should be "skip" (default) or "overwrite".
func RunMerge(dstPath, srcPath, dstPassword, srcPassword, strategy string, w io.Writer) error {
	dst, err := store.Load(dstPath, dstPassword)
	if err != nil {
		return fmt.Errorf("open destination vault: %w", err)
	}

	src, err := store.Load(srcPath, srcPassword)
	if err != nil {
		return fmt.Errorf("open source vault: %w", err)
	}

	ms := env.MergeSkip
	if strategy == "overwrite" {
		ms = env.MergeOverwrite
	}

	result, err := env.MergeVaults(dst, src, ms)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}

	if err := store.Save(dst, dstPath, dstPassword); err != nil {
		return fmt.Errorf("save destination vault: %w", err)
	}

	for _, k := range result.Added {
		fmt.Fprintf(w, "added:       %s\n", k)
	}
	for _, k := range result.Overwritten {
		fmt.Fprintf(w, "overwritten: %s\n", k)
	}
	for _, k := range result.Skipped {
		fmt.Fprintf(w, "skipped:     %s\n", k)
	}

	fmt.Fprintf(w, "merge complete: %d added, %d overwritten, %d skipped\n",
		len(result.Added), len(result.Overwritten), len(result.Skipped))

	return nil
}

// RunMergeCmd is the CLI entry point wired from main.
func RunMergeCmd(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: envault merge <dst-vault> <src-vault> [--overwrite]")
		os.Exit(1)
	}
	dstPath := args[0]
	srcPath := args[1]
	strategy := "skip"
	for _, a := range args[2:] {
		if a == "--overwrite" {
			strategy = "overwrite"
		}
	}
	dstPass, _ := readPassword("Destination vault password: ")
	srcPass, _ := readPassword("Source vault password: ")
	if err := RunMerge(dstPath, srcPath, dstPass, srcPass, strategy, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

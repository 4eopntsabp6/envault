package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nicholasgasior/envault/internal/env"
	"github.com/nicholasgasior/envault/internal/store"
)

// RunHistory prints the change history for a vault, optionally filtered by key.
// Usage: envault history <vault> <password> [key]
func RunHistory(vaultPath, password, filterKey string, w io.Writer) error {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("history: load vault: %w", err)
	}
	_ = v

	m, err := env.LoadHistory(vaultPath)
	if err != nil {
		return fmt.Errorf("history: %w", err)
	}

	if len(m.Entries) == 0 {
		fmt.Fprintln(w, "no history recorded")
		return nil
	}

	count := 0
	for _, e := range m.Entries {
		if filterKey != "" && !strings.EqualFold(e.Key, filterKey) {
			continue
		}
		line := fmt.Sprintf("%s  %-8s  %s", e.Timestamp.Format("2006-01-02T15:04:05Z"), e.Action, e.Key)
		if e.OldValue != "" {
			line += fmt.Sprintf("  (old: %s)", e.OldValue)
		}
		if e.NewValue != "" {
			line += fmt.Sprintf("  (new: %s)", e.NewValue)
		}
		fmt.Fprintln(w, line)
		count++
	}

	if count == 0 {
		fmt.Fprintf(w, "no history for key %q\n", filterKey)
	}
	return nil
}

// RunHistoryCmd is the CLI entry point wired from main.
func RunHistoryCmd(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: envault history <vault> <password> [key]")
		os.Exit(1)
	}
	vaultPath := args[0]
	password := args[1]
	filterKey := ""
	if len(args) >= 3 {
		filterKey = args[2]
	}
	if err := RunHistory(vaultPath, password, filterKey, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

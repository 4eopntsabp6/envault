package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/envault/envault/internal/env"
	"github.com/envault/envault/internal/store"
)

// RunDefaults applies default values from a .env-style defaults file to the
// vault, skipping keys that are already set.
//
// Usage: envault defaults <vault> <defaults-file>
func RunDefaults(vaultPath, password, defaultsFile string, out io.Writer) error {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("load vault: %w", err)
	}

	f, err := os.Open(defaultsFile)
	if err != nil {
		return fmt.Errorf("open defaults file: %w", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read defaults file: %w", err)
	}

	entries, err := env.LoadDefaults(lines)
	if err != nil {
		return fmt.Errorf("parse defaults: %w", err)
	}

	applied, err := env.ApplyDefaults(v, entries)
	if err != nil {
		return fmt.Errorf("apply defaults: %w", err)
	}

	if err := v.Save(); err != nil {
		return fmt.Errorf("save vault: %w", err)
	}

	if len(applied) == 0 {
		fmt.Fprintln(out, "no new defaults applied (all keys already set)")
		return nil
	}
	fmt.Fprintf(out, "applied %d default(s):\n", len(applied))
	for _, k := range applied {
		fmt.Fprintf(out, "  + %s\n", k)
	}
	return nil
}

// RunDefaultsList prints the parsed entries from a defaults file without
// modifying the vault — useful for previewing what would be applied.
func RunDefaultsList(defaultsFile string, out io.Writer) error {
	f, err := os.Open(defaultsFile)
	if err != nil {
		return fmt.Errorf("open defaults file: %w", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read defaults file: %w", err)
	}

	entries, err := env.LoadDefaults(lines)
	if err != nil {
		return fmt.Errorf("parse defaults: %w", err)
	}

	for _, e := range entries {
		line := fmt.Sprintf("%s=%s", e.Key, e.DefaultValue)
		if e.Description != "" {
			line += fmt.Sprintf(" # %s", e.Description)
		}
		fmt.Fprintln(out, strings.TrimSpace(line))
	}
	return nil
}

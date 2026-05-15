package cli

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/nicholasgasior/envault/internal/env"
	"github.com/nicholasgasior/envault/internal/store"
)

// RunChecksum dispatches checksum subcommands: record, verify.
func RunChecksum(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("checksum: subcommand required (record|verify)")
	}
	switch args[0] {
	case "record":
		return runChecksumRecord(vaultPath, password, out)
	case "verify":
		return runChecksumVerify(vaultPath, password, out)
	default:
		return fmt.Errorf("checksum: unknown subcommand %q", args[0])
	}
}

func runChecksumRecord(vaultPath, password string, out io.Writer) error {
	v, err := loadChecksumVault(vaultPath, password)
	if err != nil {
		return err
	}
	if err := env.RecordChecksums(v, password); err != nil {
		return fmt.Errorf("checksum record: %w", err)
	}
	keys := v.Keys()
	fmt.Fprintf(out, "Recorded checksums for %d key(s) in %s\n", len(keys), vaultPath)
	return nil
}

func runChecksumVerify(vaultPath, password string, out io.Writer) error {
	v, err := loadChecksumVault(vaultPath, password)
	if err != nil {
		return err
	}
	results, err := env.VerifyChecksums(v, password)
	if err != nil {
		return fmt.Errorf("checksum verify: %w", err)
	}
	if len(results) == 0 {
		fmt.Fprintln(out, "No checksums recorded. Run 'checksum record' first.")
		return nil
	}

	keys := make([]string, 0, len(results))
	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	allOK := true
	for _, k := range keys {
		status := results[k]
		marker := "✓"
		if status != "ok" {
			marker = "✗"
			allOK = false
		}
		fmt.Fprintf(out, "  %s %-30s %s\n", marker, k, status)
	}

	if !allOK {
		return fmt.Errorf("checksum verify: one or more keys failed verification")
	}
	fmt.Fprintln(out, "All checksums verified successfully.")
	return nil
}

func loadChecksumVault(vaultPath, password string) (*store.Vault, error) {
	v, err := store.Load(vaultPath, password)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("vault not found: %s", vaultPath)
	}
	if err != nil {
		return nil, fmt.Errorf("checksum: open vault: %w", err)
	}
	return v, nil
}

package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunTTL dispatches ttl subcommands: set, purge.
func RunTTL(args []string, password string, w io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault ttl <vault> <set|purge> [args...]")
	}
	vaultPath := args[0]
	subcmd := args[1]
	rest := args[2:]

	switch subcmd {
	case "set":
		return runTTLSet(vaultPath, password, rest, w)
	case "purge":
		return runTTLPurge(vaultPath, password, w)
	default:
		return fmt.Errorf("ttl: unknown subcommand %q", subcmd)
	}
}

// runTTLSet sets a TTL on a key: ttl set <key> <duration>
func runTTLSet(vaultPath, password string, args []string, w io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault ttl <vault> set <key> <duration>")
	}
	key := args[0]
	rawDur := args[1]

	dur, err := time.ParseDuration(rawDur)
	if err != nil {
		return fmt.Errorf("ttl: invalid duration %q: %w", rawDur, err)
	}
	if dur <= 0 {
		return fmt.Errorf("ttl: duration must be positive")
	}

	v, err := loadTTLVault(vaultPath, password)
	if err != nil {
		return err
	}
	if err := env.SetTTL(v, vaultPath, key, dur); err != nil {
		return err
	}
	fmt.Fprintf(w, "ttl: set %s to expire in %s\n", key, dur)
	return nil
}

// runTTLPurge removes all expired keys from the vault.
func runTTLPurge(vaultPath, password string, w io.Writer) error {
	v, err := loadTTLVault(vaultPath, password)
	if err != nil {
		return err
	}
	purged, err := env.PurgeTTLExpired(v, vaultPath)
	if err != nil {
		return err
	}
	if len(purged) == 0 {
		fmt.Fprintln(w, "ttl: no expired keys found")
		return nil
	}
	if err := v.Save(vaultPath); err != nil {
		return fmt.Errorf("ttl: save vault: %w", err)
	}
	fmt.Fprintf(w, "ttl: purged %d key(s): %s\n", len(purged), strings.Join(purged, ", "))
	return nil
}

func loadTTLVault(vaultPath, password string) (*store.Vault, error) {
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return nil, fmt.Errorf("ttl: load vault: %w", err)
	}
	return v, nil
}

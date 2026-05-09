package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/user/envault/internal/expiry"
	"github.com/user/envault/internal/store"
)

// RunExpiry sets a TTL (in seconds) on a vault key, or purges expired keys.
//
// Usage:
//
//	envault expiry set   <vault> <password> <key> <ttl-seconds>
//	envault expiry purge <vault> <password>
func RunExpiry(args []string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: expiry <set|purge> <vault> <password> [key] [ttl]")
	}
	subcmd := args[0]
	vaultPath := args[1]
	password := args[2]

	switch subcmd {
	case "set":
		return runExpirySet(args[3:], vaultPath, password, out)
	case "purge":
		return runExpiryPurge(vaultPath, password, out)
	default:
		return fmt.Errorf("expiry: unknown subcommand %q (use set or purge)", subcmd)
	}
}

func runExpirySet(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("expiry set requires <key> and <ttl-seconds>")
	}
	key := args[0]
	ttlSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || ttlSec <= 0 {
		return fmt.Errorf("expiry set: ttl must be a positive integer (seconds), got %q", args[1])
	}
	ttl := time.Duration(ttlSec) * time.Second
	if err := expiry.SetExpiry(vaultPath, key, ttl); err != nil {
		return err
	}
	fmt.Fprintf(out, "expiry set: %s will expire in %s\n", key, ttl)
	return nil
}

func runExpiryPurge(vaultPath, password string, out io.Writer) error {
	v := store.NewVault(vaultPath)
	if _, err := os.Stat(vaultPath); err == nil {
		if err := v.Load(password); err != nil {
			return fmt.Errorf("expiry purge: load vault: %w", err)
		}
	}
	purged, err := expiry.PurgeExpired(v, vaultPath, password)
	if err != nil {
		return err
	}
	if len(purged) == 0 {
		fmt.Fprintln(out, "expiry purge: no expired keys found")
		return nil
	}
	for _, key := range purged {
		fmt.Fprintf(out, "expiry purge: removed %s\n", key)
	}
	return nil
}

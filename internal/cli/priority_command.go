package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunPriority is the entry point for the `envault priority` sub-command.
//
// Usage:
//
//	envault priority set <key> <low|normal|high|1-10> --vault=PATH --password=PASS
//	envault priority get <key>                        --vault=PATH --password=PASS
//	envault priority list                             --vault=PATH --password=PASS
func RunPriority(args []string, vaultPath, password string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("priority: sub-command required (set|get|list)")
	}
	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("priority: load vault: %w", err)
	}
	switch args[0] {
	case "set":
		return runPrioritySet(args[1:], v, out)
	case "get":
		return runPriorityGet(args[1:], v, out)
	case "list":
		return runPriorityList(v, out)
	default:
		return fmt.Errorf("priority: unknown sub-command %q", args[0])
	}
}

func runPrioritySet(args []string, v *store.Vault, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("priority set: usage: set <key> <low|normal|high|1-10>")
	}
	key := args[0]
	lvl, err := parsePriorityLevel(args[1])
	if err != nil {
		return err
	}
	if err := env.SetPriority(v, lvl)(key); err != nil {
		return err
	}
	fmt.Fprintf(out, "priority: set %s = %d\n", key, lvl)
	return nil
}

func runPriorityGet(args []string, v *store.Vault, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("priority get: usage: get <key>")
	}
	key := args[0]
	lvl, err := env.GetPriority(v, key)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s: %s (%d)\n", key, levelName(lvl), lvl)
	return nil
}

func runPriorityList(v *store.Vault, out io.Writer) error {
	keys, err := env.KeysByPriority(v)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		fmt.Fprintln(out, "(no keys)")
		return nil
	}
	m, _ := env.LoadPriorityManifest(v.Path())
	for _, k := range keys {
		lvl := m[k]
		if lvl == 0 {
			lvl = env.PriorityNormal
		}
		fmt.Fprintf(out, "%-30s %s (%d)\n", k, levelName(lvl), lvl)
	}
	return nil
}

func parsePriorityLevel(s string) (env.PriorityLevel, error) {
	switch strings.ToLower(s) {
	case "low":
		return env.PriorityLow, nil
	case "normal", "default":
		return env.PriorityNormal, nil
	case "high":
		return env.PriorityHigh, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 100 {
		return 0, fmt.Errorf("priority: invalid level %q (use low/normal/high or 1-100)", s)
	}
	return env.PriorityLevel(n), nil
}

func levelName(lvl env.PriorityLevel) string {
	switch {
	case lvl >= env.PriorityHigh:
		return "high"
	case lvl <= env.PriorityLow:
		return "low"
	default:
		return "normal"
	}
}

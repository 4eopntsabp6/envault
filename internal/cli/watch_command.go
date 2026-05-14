package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/store"
)

// RunWatch checks whether a vault has changed since the last watch checkpoint
// and optionally saves a new checkpoint.
//
// Usage:
//
//	envault watch [--save] <vault-path> <password>
func RunWatch(vaultPath, password string, save bool, out io.Writer) error {
	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("load vault: %w", err)
		}
	}

	changed, err := env.HasChanged(vaultPath, v)
	if err != nil {
		return fmt.Errorf("check watch state: %w", err)
	}

	ws, err := env.LoadWatchState(vaultPath)
	if err != nil {
		return fmt.Errorf("load watch state: %w", err)
	}

	if ws == nil {
		fmt.Fprintln(out, "status: no checkpoint recorded")
	} else if changed {
		fmt.Fprintln(out, "status: CHANGED since", ws.RecordedAt.Format(time.RFC3339))
		printWatchDiff(ws, v, out)
	} else {
		fmt.Fprintln(out, "status: unchanged since", ws.RecordedAt.Format(time.RFC3339))
	}

	if save {
		if err := env.SaveWatchState(vaultPath, v); err != nil {
			return fmt.Errorf("save watch state: %w", err)
		}
		fmt.Fprintln(out, "checkpoint saved")
	}
	return nil
}

// printWatchDiff shows keys added, removed, or modified compared to the saved state.
func printWatchDiff(ws *env.WatchState, v *store.Vault, out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	defer w.Flush()

	currentKeys := map[string]string{}
	for _, k := range v.Keys() {
		val, _ := v.Get(k)
		currentKeys[k] = val
	}

	all := map[string]struct{}{}
	for k := range ws.Keys {
		all[k] = struct{}{}
	}
	for k := range currentKeys {
		all[k] = struct{}{}
	}

	sorted := make([]string, 0, len(all))
	for k := range all {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, k := range sorted {
		old, hadOld := ws.Keys[k]
		cur, hasCur := currentKeys[k]
		switch {
		case hadOld && !hasCur:
			fmt.Fprintf(w, "  removed\t%s\n", k)
		case !hadOld && hasCur:
			fmt.Fprintf(w, "  added\t%s\n", k)
		case old != cur:
			fmt.Fprintf(w, "  modified\t%s\n", k)
		}
	}
}

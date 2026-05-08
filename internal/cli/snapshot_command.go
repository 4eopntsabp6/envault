package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/user/envault/internal/snapshot"
	"github.com/user/envault/internal/store"
)

// snapshotDirForVault returns the directory used to store snapshots for a vault.
func snapshotDirForVault(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := strings.TrimSuffix(filepath.Base(vaultPath), filepath.Ext(vaultPath))
	return filepath.Join(dir, ".envault", base, "snapshots")
}

// RunSnapshot takes a snapshot of the vault and saves it to disk.
func RunSnapshot(vaultPath, password string, out io.Writer) error {
	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load vault: %w", err)
	}
	snap, err := snapshot.Take(v, password)
	if err != nil {
		return fmt.Errorf("take snapshot: %w", err)
	}
	dir := snapshotDirForVault(vaultPath)
	path, err := snapshot.Save(snap, dir)
	if err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	fmt.Fprintf(out, "Snapshot saved: %s (%d secrets)\n", path, len(snap.Secrets))
	return nil
}

// RunSnapshotDiff compares the current vault state against the latest snapshot.
func RunSnapshotDiff(vaultPath, password string, out io.Writer) error {
	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load vault: %w", err)
	}
	dir := snapshotDirForVault(vaultPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read snapshots dir: %w", err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(out, "No snapshots found.")
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	latest := filepath.Join(dir, entries[len(entries)-1].Name())
	before, err := snapshot.Load(latest)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}
	current, err := snapshot.Take(v, password)
	if err != nil {
		return fmt.Errorf("take snapshot: %w", err)
	}
	added, removed, changed := snapshot.Diff(before, current)
	fmt.Fprintf(out, "Diff vs snapshot from %s:\n", before.CreatedAt.Format(time.RFC3339))
	for _, k := range added {
		fmt.Fprintf(out, "  + %s\n", k)
	}
	for _, k := range removed {
		fmt.Fprintf(out, "  - %s\n", k)
	}
	for _, k := range changed {
		fmt.Fprintf(out, "  ~ %s\n", k)
	}
	if len(added)+len(removed)+len(changed) == 0 {
		fmt.Fprintln(out, "  No changes.")
	}
	return nil
}

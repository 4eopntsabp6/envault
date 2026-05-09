// Package diff provides utilities for comparing two vault states
// and producing human-readable change summaries.
package diff

import (
	"fmt"
	"sort"
	"strings"
)

// ChangeKind describes the type of change detected.
type ChangeKind string

const (
	Added    ChangeKind = "added"
	Removed  ChangeKind = "removed"
	Modified ChangeKind = "modified"
	Unchanged ChangeKind = "unchanged"
)

// Change represents a single key-level difference.
type Change struct {
	Key  string
	Kind ChangeKind
}

// Compare takes two maps (old, new) of key→value and returns
// the list of changes between them, sorted by key.
func Compare(oldSecrets, newSecrets map[string]string) []Change {
	seen := make(map[string]bool)
	var changes []Change

	for key, oldVal := range oldSecrets {
		seen[key] = true
		newVal, exists := newSecrets[key]
		switch {
		case !exists:
			changes = append(changes, Change{Key: key, Kind: Removed})
		case oldVal != newVal:
			changes = append(changes, Change{Key: key, Kind: Modified})
		default:
			changes = append(changes, Change{Key: key, Kind: Unchanged})
		}
	}

	for key := range newSecrets {
		if !seen[key] {
			changes = append(changes, Change{Key: key, Kind: Added})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Key < changes[j].Key
	})
	return changes
}

// Format renders a slice of Changes into a human-readable string.
// Unchanged entries are omitted unless showUnchanged is true.
func Format(changes []Change, showUnchanged bool) string {
	var sb strings.Builder
	for _, c := range changes {
		if c.Kind == Unchanged && !showUnchanged {
			continue
		}
		var prefix string
		switch c.Kind {
		case Added:
			prefix = "+"
		case Removed:
			prefix = "-"
		case Modified:
			prefix = "~"
		default:
			prefix = " "
		}
		fmt.Fprintf(&sb, "%s %s\n", prefix, c.Key)
	}
	return sb.String()
}

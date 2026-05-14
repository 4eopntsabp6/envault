package env

import (
	"fmt"
	"sort"
	"strings"

	"github.com/envault/envault/internal/store"
)

// CompareResult holds the comparison outcome for a single key.
type CompareResult struct {
	Key      string
	LeftVal  string
	RightVal string
	Status   string // "match", "mismatch", "left_only", "right_only"
}

// CompareVaults compares two vaults using their respective passwords and
// returns a slice of CompareResult describing per-key differences.
func CompareVaults(leftPath, leftPass, rightPath, rightPass string) ([]CompareResult, error) {
	left, err := store.Load(leftPath, leftPass)
	if err != nil {
		return nil, fmt.Errorf("load left vault: %w", err)
	}

	right, err := store.Load(rightPath, rightPass)
	if err != nil {
		return nil, fmt.Errorf("load right vault: %w", err)
	}

	allKeys := mergeKeysets(left.Keys(), right.Keys())

	var results []CompareResult
	for _, key := range allKeys {
		lv, leftOK := left.Get(key)
		rv, rightOK := right.Get(key)

		var status string
		switch {
		case leftOK && rightOK && lv == rv:
			status = "match"
		case leftOK && rightOK:
			status = "mismatch"
		case leftOK:
			status = "left_only"
		default:
			status = "right_only"
		}

		results = append(results, CompareResult{
			Key:      key,
			LeftVal:  lv,
			RightVal: rv,
			Status:   status,
		})
	}
	return results, nil
}

// FormatCompare returns a human-readable summary of comparison results.
// If showValues is false, secret values are masked.
func FormatCompare(results []CompareResult, showValues bool) string {
	var sb strings.Builder
	for _, r := range results {
		switch r.Status {
		case "match":
			fmt.Fprintf(&sb, "  = %s\n", r.Key)
		case "mismatch":
			if showValues {
				fmt.Fprintf(&sb, "  ~ %s\t[%s] -> [%s]\n", r.Key, r.LeftVal, r.RightVal)
			} else {
				fmt.Fprintf(&sb, "  ~ %s\t(values differ)\n", r.Key)
			}
		case "left_only":
			fmt.Fprintf(&sb, "  < %s\t(left only)\n", r.Key)
		case "right_only":
			fmt.Fprintf(&sb, "  > %s\t(right only)\n", r.Key)
		}
	}
	return sb.String()
}

func mergeKeysets(a, b []string) []string {
	seen := make(map[string]struct{})
	for _, k := range a {
		seen[k] = struct{}{}
	}
	for _, k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

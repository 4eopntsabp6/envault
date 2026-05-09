package diff_test

import (
	"strings"
	"testing"

	"github.com/yourorg/envault/internal/diff"
)

func TestCompareAdded(t *testing.T) {
	old := map[string]string{"A": "1"}
	new_ := map[string]string{"A": "1", "B": "2"}
	changes := diff.Compare(old, new_)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes[0].Key != "A" || changes[0].Kind != diff.Unchanged {
		t.Errorf("unexpected change: %+v", changes[0])
	}
	if changes[1].Key != "B" || changes[1].Kind != diff.Added {
		t.Errorf("unexpected change: %+v", changes[1])
	}
}

func TestCompareRemoved(t *testing.T) {
	old := map[string]string{"A": "1", "B": "2"}
	new_ := map[string]string{"A": "1"}
	changes := diff.Compare(old, new_)
	found := false
	for _, c := range changes {
		if c.Key == "B" && c.Kind == diff.Removed {
			found = true
		}
	}
	if !found {
		t.Error("expected B to be marked as removed")
	}
}

func TestCompareModified(t *testing.T) {
	old := map[string]string{"SECRET": "old"}
	new_ := map[string]string{"SECRET": "new"}
	changes := diff.Compare(old, new_)
	if len(changes) != 1 || changes[0].Kind != diff.Modified {
		t.Errorf("expected Modified, got %+v", changes)
	}
}

func TestCompareEmpty(t *testing.T) {
	changes := diff.Compare(map[string]string{}, map[string]string{})
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestFormatHidesUnchanged(t *testing.T) {
	old := map[string]string{"A": "1", "B": "2"}
	new_ := map[string]string{"A": "1", "B": "changed"}
	changes := diff.Compare(old, new_)
	out := diff.Format(changes, false)
	if strings.Contains(out, " A") {
		t.Error("unchanged key A should be hidden")
	}
	if !strings.Contains(out, "~ B") {
		t.Error("modified key B should appear with ~ prefix")
	}
}

func TestFormatShowsUnchanged(t *testing.T) {
	old := map[string]string{"A": "1"}
	new_ := map[string]string{"A": "1"}
	changes := diff.Compare(old, new_)
	out := diff.Format(changes, true)
	if !strings.Contains(out, "  A") {
		t.Errorf("expected unchanged A to appear, got: %q", out)
	}
}

func TestFormatPrefixes(t *testing.T) {
	old := map[string]string{"OLD": "v"}
	new_ := map[string]string{"NEW": "v"}
	changes := diff.Compare(old, new_)
	out := diff.Format(changes, false)
	if !strings.Contains(out, "- OLD") {
		t.Error("expected - prefix for removed key")
	}
	if !strings.Contains(out, "+ NEW") {
		t.Error("expected + prefix for added key")
	}
}

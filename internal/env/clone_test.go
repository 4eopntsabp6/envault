package env

import (
	"strings"
	"testing"

	"github.com/yourusername/envault/internal/store"
)

func newCloneVault(t *testing.T, pairs map[string]string) *store.Vault {
	t.Helper()
	v := store.NewVault("password")
	for k, val := range pairs {
		v.Set(k, val)
	}
	return v
}

func TestCloneAllKeys(t *testing.T) {
	src := newCloneVault(t, map[string]string{"A": "1", "B": "2", "C": "3"})
	dst := store.NewVault("password")

	n, err := CloneVault(src, dst, CloneOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected 3 written, got %d", n)
	}
	for _, k := range []string{"A", "B", "C"} {
		if v, ok := dst.Get(k); !ok || v != src.MustGet(k) {
			t.Errorf("key %q: expected %q, got %q (ok=%v)", k, src.MustGet(k), v, ok)
		}
	}
}

func TestCloneSelectedKeys(t *testing.T) {
	src := newCloneVault(t, map[string]string{"A": "1", "B": "2", "C": "3"})
	dst := store.NewVault("password")

	n, err := CloneVault(src, dst, CloneOptions{Keys: []string{"A", "C"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 written, got %d", n)
	}
	if _, ok := dst.Get("B"); ok {
		t.Error("key B should not have been cloned")
	}
}

func TestCloneSkipsExistingWithoutOverwrite(t *testing.T) {
	src := newCloneVault(t, map[string]string{"X": "new"})
	dst := newCloneVault(t, map[string]string{"X": "old"})

	n, err := CloneVault(src, dst, CloneOptions{Overwrite: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 written, got %d", n)
	}
	if v, _ := dst.Get("X"); v != "old" {
		t.Errorf("expected 'old', got %q", v)
	}
}

func TestCloneOverwrite(t *testing.T) {
	src := newCloneVault(t, map[string]string{"X": "new"})
	dst := newCloneVault(t, map[string]string{"X": "old"})

	n, err := CloneVault(src, dst, CloneOptions{Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 written, got %d", n)
	}
	if v, _ := dst.Get("X"); v != "new" {
		t.Errorf("expected 'new', got %q", v)
	}
}

func TestCloneWithTransform(t *testing.T) {
	src := newCloneVault(t, map[string]string{"SECRET": "hello"})
	dst := store.NewVault("password")

	transform := func(_, v string) string { return strings.ToUpper(v) }
	_, err := CloneVault(src, dst, CloneOptions{Transform: transform})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, _ := dst.Get("SECRET"); v != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", v)
	}
}

func TestCloneMissingKeyReturnsError(t *testing.T) {
	src := store.NewVault("password")
	dst := store.NewVault("password")

	_, err := CloneVault(src, dst, CloneOptions{Keys: []string{"MISSING"}})
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

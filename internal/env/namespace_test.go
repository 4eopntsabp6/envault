package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envault/envault/internal/store"
)

func newNamespaceVault(t *testing.T) *store.Vault {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vault")
	v := store.NewVault(path, "password")
	if err := v.Set("DB_HOST", "localhost"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("DB_PORT", "5432"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("API_KEY", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestNamespacePath(t *testing.T) {
	path := "/tmp/myproject.vault"
	got := NamespacePath(path)
	want := "/tmp/myproject.namespaces.json"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoadNamespaceManifestMissing(t *testing.T) {
	v := newNamespaceVault(t)
	m, err := LoadNamespaceManifest(v.Path())
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Namespaces) != 0 {
		t.Errorf("expected empty manifest, got %v", m.Namespaces)
	}
}

func TestAssignAndGetNamespace(t *testing.T) {
	v := newNamespaceVault(t)
	if err := AssignNamespace(v, "database", "DB_HOST"); err != nil {
		t.Fatal(err)
	}
	if err := AssignNamespace(v, "database", "DB_PORT"); err != nil {
		t.Fatal(err)
	}
	keys, err := GetNamespaceKeys(v, "database")
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "DB_HOST" || keys[1] != "DB_PORT" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestAssignNamespaceMissingKey(t *testing.T) {
	v := newNamespaceVault(t)
	err := AssignNamespace(v, "database", "MISSING_KEY")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestAssignNamespaceDuplicate(t *testing.T) {
	v := newNamespaceVault(t)
	if err := AssignNamespace(v, "database", "DB_HOST"); err != nil {
		t.Fatal(err)
	}
	if err := AssignNamespace(v, "database", "DB_HOST"); err != nil {
		t.Fatal(err)
	}
	keys, _ := GetNamespaceKeys(v, "database")
	if len(keys) != 1 {
		t.Errorf("expected 1 key (no duplicates), got %d", len(keys))
	}
}

func TestListNamespaces(t *testing.T) {
	v := newNamespaceVault(t)
	_ = AssignNamespace(v, "api", "API_KEY")
	_ = AssignNamespace(v, "database", "DB_HOST")
	names, err := ListNamespaces(v)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "api" || names[1] != "database" {
		t.Errorf("unexpected namespaces: %v", names)
	}
}

func TestRemoveFromNamespace(t *testing.T) {
	v := newNamespaceVault(t)
	_ = AssignNamespace(v, "database", "DB_HOST")
	_ = AssignNamespace(v, "database", "DB_PORT")
	if err := RemoveFromNamespace(v, "database", "DB_HOST"); err != nil {
		t.Fatal(err)
	}
	keys, _ := GetNamespaceKeys(v, "database")
	if len(keys) != 1 || keys[0] != "DB_PORT" {
		t.Errorf("unexpected keys after remove: %v", keys)
	}
}

func TestRemoveLastKeyDeletesNamespace(t *testing.T) {
	v := newNamespaceVault(t)
	_ = AssignNamespace(v, "api", "API_KEY")
	_ = RemoveFromNamespace(v, "api", "API_KEY")
	names, _ := ListNamespaces(v)
	for _, n := range names {
		if n == "api" {
			t.Error("namespace 'api' should have been deleted")
		}
	}
}

func TestNamespaceManifestPersists(t *testing.T) {
	v := newNamespaceVault(t)
	_ = AssignNamespace(v, "database", "DB_HOST")
	m, err := LoadNamespaceManifest(v.Path())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Namespaces["database"]; !ok {
		t.Error("namespace manifest did not persist")
	}
	_ = os.Remove(NamespacePath(v.Path()))
}

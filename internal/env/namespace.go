package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/envault/envault/internal/store"
)

// NamespaceManifest maps keys to their assigned namespace.
type NamespaceManifest struct {
	Namespaces map[string][]string `json:"namespaces"` // namespace -> list of keys
}

// NamespacePath returns the path to the namespace manifest for a vault.
func NamespacePath(vaultPath string) string {
	dir := strings.TrimSuffix(vaultPath, filepath.Ext(vaultPath))
	return dir + ".namespaces.json"
}

// LoadNamespaceManifest loads the namespace manifest from disk.
func LoadNamespaceManifest(vaultPath string) (*NamespaceManifest, error) {
	path := NamespacePath(vaultPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &NamespaceManifest{Namespaces: make(map[string][]string)}, nil
	}
	if err != nil {
		return nil, err
	}
	var m NamespaceManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Namespaces == nil {
		m.Namespaces = make(map[string][]string)
	}
	return &m, nil
}

// SaveNamespaceManifest writes the namespace manifest to disk.
func SaveNamespaceManifest(vaultPath string, m *NamespaceManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(NamespacePath(vaultPath), data, 0600)
}

// AssignNamespace assigns a key to a namespace, creating the namespace if needed.
func AssignNamespace(v *store.Vault, namespace, key string) error {
	if _, err := v.Get(key); err != nil {
		return fmt.Errorf("key %q not found in vault", key)
	}
	m, err := LoadNamespaceManifest(v.Path())
	if err != nil {
		return err
	}
	keys := m.Namespaces[namespace]
	for _, k := range keys {
		if k == key {
			return nil // already assigned
		}
	}
	m.Namespaces[namespace] = append(keys, key)
	sort.Strings(m.Namespaces[namespace])
	return SaveNamespaceManifest(v.Path(), m)
}

// GetNamespaceKeys returns all keys assigned to the given namespace.
func GetNamespaceKeys(v *store.Vault, namespace string) ([]string, error) {
	m, err := LoadNamespaceManifest(v.Path())
	if err != nil {
		return nil, err
	}
	keys, ok := m.Namespaces[namespace]
	if !ok {
		return nil, fmt.Errorf("namespace %q not found", namespace)
	}
	return keys, nil
}

// ListNamespaces returns all namespace names in sorted order.
func ListNamespaces(v *store.Vault) ([]string, error) {
	m, err := LoadNamespaceManifest(v.Path())
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(m.Namespaces))
	for ns := range m.Namespaces {
		names = append(names, ns)
	}
	sort.Strings(names)
	return names, nil
}

// RemoveFromNamespace removes a key from a namespace.
func RemoveFromNamespace(v *store.Vault, namespace, key string) error {
	m, err := LoadNamespaceManifest(v.Path())
	if err != nil {
		return err
	}
	keys, ok := m.Namespaces[namespace]
	if !ok {
		return fmt.Errorf("namespace %q not found", namespace)
	}
	updated := keys[:0]
	for _, k := range keys {
		if k != key {
			updated = append(updated, k)
		}
	}
	if len(updated) == 0 {
		delete(m.Namespaces, namespace)
	} else {
		m.Namespaces[namespace] = updated
	}
	return SaveNamespaceManifest(v.Path(), m)
}

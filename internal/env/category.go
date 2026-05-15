package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/nicholasgasior/envault/internal/store"
)

const categoryManifestFile = ".envault_categories.json"

type CategoryManifest struct {
	Categories map[string]string `json:"categories"` // key -> category name
}

func CategoryPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	return filepath.Join(dir, categoryManifestFile)
}

func LoadCategoryManifest(vaultPath string) (*CategoryManifest, error) {
	p := CategoryPath(vaultPath)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &CategoryManifest{Categories: make(map[string]string)}, nil
		}
		return nil, err
	}
	var m CategoryManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Categories == nil {
		m.Categories = make(map[string]string)
	}
	return &m, nil
}

func SaveCategoryManifest(vaultPath string, m *CategoryManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(CategoryPath(vaultPath), data, 0600)
}

func SetCategory(v *store.Vault, key, category string) error {
	if _, err := v.Get(key); err != nil {
		return fmt.Errorf("key %q not found in vault", key)
	}
	m, err := LoadCategoryManifest(v.Path())
	if err != nil {
		return err
	}
	m.Categories[key] = category
	return SaveCategoryManifest(v.Path(), m)
}

func GetCategory(v *store.Vault, key string) (string, error) {
	m, err := LoadCategoryManifest(v.Path())
	if err != nil {
		return "", err
	}
	cat, ok := m.Categories[key]
	if !ok {
		return "", nil
	}
	return cat, nil
}

func KeysByCategory(v *store.Vault, category string) ([]string, error) {
	m, err := LoadCategoryManifest(v.Path())
	if err != nil {
		return nil, err
	}
	var result []string
	for k, c := range m.Categories {
		if c == category {
			result = append(result, k)
		}
	}
	sort.Strings(result)
	return result, nil
}

func DeleteCategory(v *store.Vault, key string) error {
	m, err := LoadCategoryManifest(v.Path())
	if err != nil {
		return err
	}
	delete(m.Categories, key)
	return SaveCategoryManifest(v.Path(), m)
}

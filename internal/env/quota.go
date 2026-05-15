package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicholasgasior/envault/internal/store"
)

// QuotaManifest holds per-vault key count and value size limits.
type QuotaManifest struct {
	MaxKeys      int `json:"max_keys,omitempty"`
	MaxValueSize int `json:"max_value_size,omitempty"` // bytes
}

// QuotaPath returns the path to the quota manifest for a vault.
func QuotaPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return filepath.Join(dir, name+".quota.json")
}

// LoadQuota loads the quota manifest; returns zero-value if missing.
func LoadQuota(vaultPath string) (QuotaManifest, error) {
	data, err := os.ReadFile(QuotaPath(vaultPath))
	if os.IsNotExist(err) {
		return QuotaManifest{}, nil
	}
	if err != nil {
		return QuotaManifest{}, err
	}
	var m QuotaManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return QuotaManifest{}, err
	}
	return m, nil
}

// SaveQuota persists the quota manifest to disk.
func SaveQuota(vaultPath string, m QuotaManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(QuotaPath(vaultPath), data, 0600)
}

// SetQuota updates the quota manifest for a vault.
func SetQuota(vaultPath string, maxKeys, maxValueSize int) error {
	m := QuotaManifest{
		MaxKeys:      maxKeys,
		MaxValueSize: maxValueSize,
	}
	return SaveQuota(vaultPath, m)
}

// CheckQuota validates that adding a key/value to the vault respects quotas.
// Pass value="" when only checking key count (e.g. before a Get).
func CheckQuota(v *store.Vault, vaultPath, key, value string) error {
	m, err := LoadQuota(vaultPath)
	if err != nil {
		return err
	}
	if m.MaxKeys > 0 {
		keys := v.Keys()
		_, exists := v.Get(key)
		if !exists && len(keys) >= m.MaxKeys {
			return fmt.Errorf("quota exceeded: vault allows at most %d keys", m.MaxKeys)
		}
	}
	if m.MaxValueSize > 0 && len(value) > m.MaxValueSize {
		return fmt.Errorf("quota exceeded: value size %d exceeds limit of %d bytes", len(value), m.MaxValueSize)
	}
	return nil
}

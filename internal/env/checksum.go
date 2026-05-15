package env

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/nicholasgasior/envault/internal/store"
)

// ChecksumManifest maps key names to their value checksums.
type ChecksumManifest struct {
	Checksums map[string]string `json:"checksums"`
}

// ChecksumPath returns the path to the checksum manifest for a given vault.
func ChecksumPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	base := filepath.Base(vaultPath)
	return filepath.Join(dir, "."+base+".checksums.json")
}

// hashValue returns a short SHA-256 hex digest of the given value.
func hashValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// LoadChecksumManifest loads the checksum manifest from disk.
// Returns an empty manifest if the file does not exist.
func LoadChecksumManifest(vaultPath string) (*ChecksumManifest, error) {
	p := ChecksumPath(vaultPath)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &ChecksumManifest{Checksums: map[string]string{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checksum: read manifest: %w", err)
	}
	var m ChecksumManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("checksum: parse manifest: %w", err)
	}
	if m.Checksums == nil {
		m.Checksums = map[string]string{}
	}
	return &m, nil
}

// SaveChecksumManifest persists the checksum manifest to disk.
func SaveChecksumManifest(vaultPath string, m *ChecksumManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("checksum: marshal manifest: %w", err)
	}
	return os.WriteFile(ChecksumPath(vaultPath), data, 0600)
}

// RecordChecksums computes and stores checksums for all keys in the vault.
func RecordChecksums(v *store.Vault, password string) error {
	keys := v.Keys()
	sort.Strings(keys)
	m := &ChecksumManifest{Checksums: map[string]string{}}
	for _, k := range keys {
		val, ok := v.Get(k)
		if !ok {
			continue
		}
		m.Checksums[k] = hashValue(val)
	}
	return SaveChecksumManifest(v.Path(), m)
}

// VerifyChecksums compares current vault values against stored checksums.
// Returns a map of key -> "ok" | "mismatch" | "missing".
func VerifyChecksums(v *store.Vault, password string) (map[string]string, error) {
	m, err := LoadChecksumManifest(v.Path())
	if err != nil {
		return nil, err
	}
	results := map[string]string{}
	for k, stored := range m.Checksums {
		val, ok := v.Get(k)
		if !ok {
			results[k] = "missing"
			continue
		}
		if hashValue(val) == stored {
			results[k] = "ok"
		} else {
			results[k] = "mismatch"
		}
	}
	return results, nil
}

package env

import (
	"strings"

	"github.com/envault/envault/internal/store"
)

// RedactManifestKey is the vault key used to store the redact manifest.
const RedactManifestKey = "__envault_redact_manifest__"

// RedactedPlaceholder is the string shown in place of a redacted value.
const RedactedPlaceholder = "[REDACTED]"

// sensitivePatterns are substrings that auto-suggest redaction.
var sensitivePatterns = []string{"secret", "password", "passwd", "token", "apikey", "api_key", "private"}

// SetRedacted marks a key as redacted in the manifest.
func SetRedacted(v *store.Vault, key string, redacted bool) error {
	keys := loadRedactManifest(v)
	if redacted {
		keys[key] = true
	} else {
		delete(keys, key)
	}
	return saveRedactManifest(v, keys)
}

// IsRedacted reports whether a key is marked as redacted.
func IsRedacted(v *store.Vault, key string) bool {
	keys := loadRedactManifest(v)
	return keys[key]
}

// RedactedKeys returns all keys currently marked as redacted.
func RedactedKeys(v *store.Vault) []string {
	keys := loadRedactManifest(v)
	out := make([]string, 0, len(keys))
	for k := range keys {
		out = append(out, k)
	}
	return out
}

// MaybeRedact returns the placeholder if the key is redacted, otherwise the value.
func MaybeRedact(v *store.Vault, key, value string) string {
	if IsRedacted(v, key) {
		return RedactedPlaceholder
	}
	return value
}

// AutoDetectSensitive returns true if the key name looks sensitive.
func AutoDetectSensitive(key string) bool {
	lower := strings.ToLower(key)
	for _, p := range sensitivePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func loadRedactManifest(v *store.Vault) map[string]bool {
	raw, ok := v.Get(RedactManifestKey)
	if !ok || raw == "" {
		return map[string]bool{}
	}
	m := map[string]bool{}
	for _, k := range strings.Split(raw, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			m[k] = true
		}
	}
	return m
}

func saveRedactManifest(v *store.Vault, keys map[string]bool) error {
	parts := make([]string, 0, len(keys))
	for k := range keys {
		parts = append(parts, k)
	}
	v.Set(RedactManifestKey, strings.Join(parts, ","))
	return nil
}

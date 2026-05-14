package env

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/envault/envault/internal/store"
)

// SupportedTypes lists all valid type hints.
var SupportedTypes = []string{"string", "int", "float", "bool"}

// TypeHintKey returns the vault key used to store the type hint for a given key.
func TypeHintKey(key string) string {
	return "__typehint__" + key
}

// SetTypeHint stores a type hint for the given key in the vault.
func SetTypeHint(v *store.Vault, key, typeName string) error {
	if !isSupportedType(typeName) {
		return fmt.Errorf("unsupported type %q: must be one of %s", typeName, strings.Join(SupportedTypes, ", "))
	}
	if _, ok := v.Get(key); !ok {
		return fmt.Errorf("key %q not found in vault", key)
	}
	v.Set(TypeHintKey(key), typeName)
	return nil
}

// GetTypeHint returns the type hint for the given key, or "string" if none is set.
func GetTypeHint(v *store.Vault, key string) string {
	if hint, ok := v.Get(TypeHintKey(key)); ok {
		return hint
	}
	return "string"
}

// CastValue casts the raw string value to the type hint associated with key.
// It returns the cast value as a string representation and any validation error.
func CastValue(v *store.Vault, key string) (string, error) {
	raw, ok := v.Get(key)
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}
	hint := GetTypeHint(v, key)
	switch hint {
	case "int":
		if _, err := strconv.ParseInt(raw, 10, 64); err != nil {
			return "", fmt.Errorf("value %q for key %q cannot be cast to int", raw, key)
		}
	case "float":
		if _, err := strconv.ParseFloat(raw, 64); err != nil {
			return "", fmt.Errorf("value %q for key %q cannot be cast to float", raw, key)
		}
	case "bool":
		norm := strings.ToLower(strings.TrimSpace(raw))
		if norm != "true" && norm != "false" && norm != "1" && norm != "0" {
			return "", fmt.Errorf("value %q for key %q cannot be cast to bool", raw, key)
		}
	}
	return raw, nil
}

// ValidateAllCasts checks every key with a type hint and returns a list of errors.
func ValidateAllCasts(v *store.Vault) []error {
	var errs []error
	for _, key := range v.Keys() {
		if strings.HasPrefix(key, "__typehint__") {
			continue
		}
		if _, err := CastValue(v, key); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func isSupportedType(t string) bool {
	for _, s := range SupportedTypes {
		if s == t {
			return true
		}
	}
	return false
}

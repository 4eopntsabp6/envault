package env

import (
	"fmt"
	"strings"
)

// ValidationError holds a key and the reason it failed validation.
type ValidationError struct {
	Key    string
	Reason string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("key %q: %s", e.Key, e.Reason)
}

// ValidateKey checks whether a key conforms to standard env var naming rules:
// must be non-empty, start with a letter or underscore, and contain only
// letters, digits, or underscores.
func ValidateKey(key string) error {
	if len(key) == 0 {
		return ValidationError{Key: key, Reason: "key must not be empty"}
	}
	first := rune(key[0])
	if !isLetter(first) && first != '_' {
		return ValidationError{Key: key, Reason: "key must start with a letter or underscore"}
	}
	for _, ch := range key[1:] {
		if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			return ValidationError{Key: key, Reason: fmt.Sprintf("key contains invalid character %q", ch)}
		}
	}
	return nil
}

// ValidateKeys validates a slice of keys and returns all errors found.
func ValidateKeys(keys []string) []error {
	var errs []error
	for _, k := range keys {
		if err := ValidateKey(k); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// ValidateValue checks that a value does not contain unescaped newlines.
func ValidateValue(key, value string) error {
	if strings.ContainsAny(value, "\n\r") {
		return ValidationError{Key: key, Reason: "value must not contain newline characters"}
	}
	return nil
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

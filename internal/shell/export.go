package shell

import (
	"fmt"
	"strings"

	"github.com/user/envault/internal/store"
)

// Format defines the shell export format to generate.
type Format int

const (
	FormatBash Format = iota
	FormatFish
	FormatDotenv
)

// ExportEnv generates shell-compatible export statements for all secrets
// stored in the given vault, using the specified shell format.
func ExportEnv(v *store.Vault, format Format) (string, error) {
	keys := v.Keys()
	if len(keys) == 0 {
		return "", nil
	}

	var sb strings.Builder

	for _, key := range keys {
		val, ok := v.Get(key)
		if !ok {
			continue
		}
		escaped := escapeValue(val)
		switch format {
		case FormatBash:
			fmt.Fprintf(&sb, "export %s=%q\n", key, escaped)
		case FormatFish:
			fmt.Fprintf(&sb, "set -x %s %q\n", key, escaped)
		case FormatDotenv:
			fmt.Fprintf(&sb, "%s=%q\n", key, escaped)
		default:
			return "", fmt.Errorf("unknown shell format: %d", format)
		}
	}

	return sb.String(), nil
}

// ParseFormat converts a string format name to a Format constant.
func ParseFormat(name string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "bash", "sh", "":
		return FormatBash, nil
	case "fish":
		return FormatFish, nil
	case "dotenv":
		return FormatDotenv, nil
	default:
		return FormatBash, fmt.Errorf("unsupported format %q: choose bash, fish, or dotenv", name)
	}
}

// escapeValue returns the value with backslashes and double-quotes escaped.
func escapeValue(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

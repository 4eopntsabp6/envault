// Package template provides functionality for rendering .env templates
// where secret values are substituted from the vault.
package template

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/yourusername/envault/internal/store"
)

// placeholder matches patterns like {{KEY}} or {{ KEY }} in template files.
var placeholder = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

// RenderResult holds the rendered output and metadata about substitutions.
type RenderResult struct {
	Output   string
	Resolved []string
	Missing  []string
}

// Render reads a template string and substitutes {{KEY}} placeholders
// with values from the vault. Missing keys are collected but do not
// cause an error unless strict is true.
func Render(tmpl string, v *store.Vault, strict bool) (*RenderResult, error) {
	result := &RenderResult{}
	resolved := map[string]bool{}
	missing := map[string]bool{}

	output := placeholder.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := strings.TrimSpace(placeholder.FindStringSubmatch(match)[1])
		val, ok := v.Get(key)
		if !ok {
			missing[key] = true
			return match
		}
		resolved[key] = true
		return val
	})

	for k := range resolved {
		result.Resolved = append(result.Resolved, k)
	}
	for k := range missing {
		result.Missing = append(result.Missing, k)
	}

	if strict && len(result.Missing) > 0 {
		return nil, fmt.Errorf("template: unresolved keys: %s", strings.Join(result.Missing, ", "))
	}

	result.Output = output
	return result, nil
}

// RenderFile reads a template from disk and renders it using the vault.
func RenderFile(path string, v *store.Vault, strict bool) (*RenderResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("template: read file %q: %w", path, err)
	}
	return Render(string(data), v, strict)
}

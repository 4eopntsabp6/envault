package env

import (
	"fmt"
	"strings"

	"github.com/envault/envault/internal/store"
)

// TransformFunc is a function that transforms a secret value.
type TransformFunc func(value string) (string, error)

// BuiltinTransforms maps transform names to their implementations.
var BuiltinTransforms = map[string]TransformFunc{
	"upper":   func(v string) (string, error) { return strings.ToUpper(v), nil },
	"lower":   func(v string) (string, error) { return strings.ToLower(v), nil },
	"trim":    func(v string) (string, error) { return strings.TrimSpace(v), nil },
	"base64":  transformBase64,
	"reverse": transformReverse,
}

func transformBase64(v string) (string, error) {
	import64 := func(s string) string {
		const enc = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
		var buf strings.Builder
		b := []byte(s)
		for i := 0; i < len(b); i += 3 {
			var chunk [3]byte
			pad := 0
			for j := 0; j < 3; j++ {
				if i+j < len(b) {
					chunk[j] = b[i+j]
				} else {
					pad++
				}
			}
			buf.WriteByte(enc[chunk[0]>>2])
			buf.WriteByte(enc[(chunk[0]&0x3)<<4|chunk[1]>>4])
			if pad < 2 {
				buf.WriteByte(enc[(chunk[1]&0xf)<<2|chunk[2]>>6])
			} else {
				buf.WriteByte('=')
			}
			if pad < 1 {
				buf.WriteByte(enc[chunk[2]&0x3f])
			} else {
				buf.WriteByte('=')
			}
		}
		return buf.String()
	}
	return import64(v), nil
}

func transformReverse(v string) (string, error) {
	r := []rune(v)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r), nil
}

// ApplyTransform applies a named transform to the value of a key in the vault.
// The transformed value is stored back under the same key.
func ApplyTransform(v *store.Vault, password, key, transformName string) (string, error) {
	fn, ok := BuiltinTransforms[transformName]
	if !ok {
		return "", fmt.Errorf("unknown transform %q: available: %s",
			transformName, strings.Join(availableTransforms(), ", "))
	}
	val, err := v.Get(password, key)
	if err != nil {
		return "", fmt.Errorf("get key %q: %w", key, err)
	}
	newVal, err := fn(val)
	if err != nil {
		return "", fmt.Errorf("transform %q: %w", transformName, err)
	}
	if err := v.Set(password, key, newVal); err != nil {
		return "", fmt.Errorf("set key %q: %w", key, err)
	}
	return newVal, nil
}

func availableTransforms() []string {
	names := make([]string, 0, len(BuiltinTransforms))
	for k := range BuiltinTransforms {
		names = append(names, k)
	}
	return names
}

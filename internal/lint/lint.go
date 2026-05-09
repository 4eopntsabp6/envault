// Package lint provides validation rules for vault secrets.
package lint

import (
	"fmt"
	"strings"

	"github.com/user/envault/internal/store"
)

// Issue represents a single lint finding for a key.
type Issue struct {
	Key     string
	Message string
}

func (i Issue) String() string {
	return fmt.Sprintf("%s: %s", i.Key, i.Message)
}

// Rule is a function that inspects a key/value pair and returns an issue or nil.
type Rule func(key, value string) *Issue

// DefaultRules is the set of rules applied by Run.
var DefaultRules = []Rule{
	RuleEmptyValue,
	RuleKeyUpperCase,
	RuleNoSpacesInKey,
	RuleWeakSecret,
}

// RuleEmptyValue flags keys whose value is empty.
func RuleEmptyValue(key, value string) *Issue {
	if strings.TrimSpace(value) == "" {
		return &Issue{Key: key, Message: "value is empty"}
	}
	return nil
}

// RuleKeyUpperCase flags keys that are not fully upper-case.
func RuleKeyUpperCase(key, value string) *Issue {
	if key != strings.ToUpper(key) {
		return &Issue{Key: key, Message: "key should be upper-case"}
	}
	return nil
}

// RuleNoSpacesInKey flags keys that contain spaces.
func RuleNoSpacesInKey(key, _ string) *Issue {
	if strings.Contains(key, " ") {
		return &Issue{Key: key, Message: "key contains spaces"}
	}
	return nil
}

// RuleWeakSecret flags values that look like placeholder secrets.
func RuleWeakSecret(key, value string) *Issue {
	weak := []string{"changeme", "secret", "password", "12345", "todo", "fixme"}
	lv := strings.ToLower(value)
	for _, w := range weak {
		if lv == w {
			return &Issue{Key: key, Message: fmt.Sprintf("value looks like a weak/placeholder secret (%q)", value)}
		}
	}
	return nil
}

// Run applies rules to every key in the vault and returns all issues found.
func Run(v *store.Vault, rules []Rule) []Issue {
	var issues []Issue
	for _, key := range v.Keys() {
		val, _ := v.Get(key)
		for _, rule := range rules {
			if issue := rule(key, val); issue != nil {
				issues = append(issues, *issue)
			}
		}
	}
	return issues
}

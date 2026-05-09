package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/envault/internal/store"
	"github.com/user/envault/internal/tags"
)

// RunTags dispatches tag subcommands: set, get, filter.
// args: [subcommand, ...]
func RunTags(vaultPath, password string, args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: envault tags <set|get|filter> [args...]")
	}
	switch args[0] {
	case "set":
		return runTagsSet(vaultPath, password, args[1:], out)
	case "get":
		return runTagsGet(vaultPath, args[1:], out)
	case "filter":
		return runTagsFilter(vaultPath, args[1:], out)
	default:
		return fmt.Errorf("unknown tags subcommand: %s", args[0])
	}
}

// runTagsSet assigns tags to a key: set <key> <tag1,tag2,...>
func runTagsSet(vaultPath, password string, args []string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: envault tags set <key> <tag1,tag2,...>")
	}
	key := args[0]
	tagList := strings.Split(args[1], ",")

	v, err := store.Load(vaultPath, password)
	if err != nil {
		return fmt.Errorf("load vault: %w", err)
	}
	if _, ok := v.Get(key); !ok {
		return fmt.Errorf("key %q not found in vault", key)
	}

	m, err := tags.LoadManifest(vaultPath)
	if err != nil {
		return fmt.Errorf("load tags: %w", err)
	}
	tags.SetTags(m, key, tagList)
	if err := tags.SaveManifest(vaultPath, m); err != nil {
		return fmt.Errorf("save tags: %w", err)
	}
	fmt.Fprintf(out, "tagged %s: %s\n", key, strings.Join(tagList, ", "))
	return nil
}

// runTagsGet prints tags for a key: get <key>
func runTagsGet(vaultPath string, args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault tags get <key>")
	}
	m, err := tags.LoadManifest(vaultPath)
	if err != nil {
		return fmt.Errorf("load tags: %w", err)
	}
	t := tags.GetTags(m, args[0])
	if len(t) == 0 {
		fmt.Fprintf(out, "no tags for %s\n", args[0])
		return nil
	}
	fmt.Fprintf(out, "%s: %s\n", args[0], strings.Join(t, ", "))
	return nil
}

// runTagsFilter lists keys with a given tag: filter <tag>
func runTagsFilter(vaultPath string, args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: envault tags filter <tag>")
	}
	m, err := tags.LoadManifest(vaultPath)
	if err != nil {
		return fmt.Errorf("load tags: %w", err)
	}
	keys := tags.FilterByTag(m, args[0])
	if len(keys) == 0 {
		fmt.Fprintf(out, "no keys with tag %q\n", args[0])
		return nil
	}
	for _, k := range keys {
		fmt.Fprintln(out, k)
	}
	return nil
}

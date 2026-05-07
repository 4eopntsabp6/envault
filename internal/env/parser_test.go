package env

import (
	"strings"
	"testing"
)

func TestParseBasic(t *testing.T) {
	input := `
# comment
FOO=bar
BAZ=hello world
`
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "FOO" || entries[0].Value != "bar" {
		t.Errorf("unexpected entry[0]: %+v", entries[0])
	}
	if entries[1].Key != "BAZ" || entries[1].Value != "hello world" {
		t.Errorf("unexpected entry[1]: %+v", entries[1])
	}
}

func TestParseQuotedValues(t *testing.T) {
	input := `KEY="quoted value"
KEY2='single quoted'`
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries[0].Value != "quoted value" {
		t.Errorf("expected 'quoted value', got %q", entries[0].Value)
	}
	if entries[1].Value != "single quoted" {
		t.Errorf("expected 'single quoted', got %q", entries[1].Value)
	}
}

func TestParseExportPrefix(t *testing.T) {
	input := `export MY_VAR=secret`
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Key != "MY_VAR" || entries[0].Value != "secret" {
		t.Errorf("unexpected entry: %+v", entries)
	}
}

func TestParseMissingEquals(t *testing.T) {
	input := `BADLINE`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for missing '='")
	}
}

func TestParseEmptyKey(t *testing.T) {
	input := `=value`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestFormat(t *testing.T) {
	entries := []Entry{
		{Key: "A", Value: "1"},
		{Key: "B", Value: "hello world"},
	}
	out := Format(entries)
	parsed, err := Parse(strings.NewReader(out))
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 entries after round-trip, got %d", len(parsed))
	}
	if parsed[1].Value != "hello world" {
		t.Errorf("value mismatch after round-trip: %q", parsed[1].Value)
	}
}

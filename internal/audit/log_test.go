package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, "myproject")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestRecord(t *testing.T) {
	dir := t.TempDir()
	l, _ := NewLogger(dir, "proj1")

	if err := l.Record("set", "DB_URL", true); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := l.Record("get", "API_KEY", false); err != nil {
		t.Fatalf("Record: %v", err)
	}

	entries, err := ReadAll(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Action != "set" || entries[0].Key != "DB_URL" || !entries[0].Success {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Action != "get" || entries[1].Key != "API_KEY" || entries[1].Success {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
	if entries[0].Project != "proj1" {
		t.Errorf("expected project proj1, got %s", entries[0].Project)
	}
}

func TestRecordTimestamp(t *testing.T) {
	dir := t.TempDir()
	l, _ := NewLogger(dir, "proj")
	before := time.Now().UTC()
	l.Record("delete", "SECRET", true)
	after := time.Now().UTC()

	entries, _ := ReadAll(filepath.Join(dir, "audit.jsonl"))
	if len(entries) == 0 {
		t.Fatal("no entries")
	}
	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v out of range [%v, %v]", ts, before, after)
	}
}

func TestReadAllMissingFile(t *testing.T) {
	entries, err := ReadAll("/nonexistent/audit.jsonl")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries, got %v", entries)
	}
}

func TestReadAllEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	os.WriteFile(path, []byte{}, 0600)

	entries, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll empty: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

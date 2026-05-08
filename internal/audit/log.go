package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Key       string    `json:"key,omitempty"`
	Project   string    `json:"project"`
	Success   bool      `json:"success"`
}

// Logger writes audit entries to a JSONL file.
type Logger struct {
	path    string
	project string
}

// NewLogger creates a Logger that appends to the given directory.
func NewLogger(dir, project string) (*Logger, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("audit: create log dir: %w", err)
	}
	return &Logger{
		path:    filepath.Join(dir, "audit.jsonl"),
		project: project,
	}, nil
}

// Record appends an entry to the audit log.
func (l *Logger) Record(action, key string, success bool) error {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("audit: open log: %w", err)
	}
	defer f.Close()

	entry := Entry{
		Timestamp: time.Now().UTC(),
		Action:    action,
		Key:       key,
		Project:   l.project,
		Success:   success,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// ReadAll reads all audit entries from the log file.
func ReadAll(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("audit: read log: %w", err)
	}

	var entries []Entry
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("audit: parse entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

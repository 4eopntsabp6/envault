package shell_test

import (
	"strings"
	"testing"

	"github.com/user/envault/internal/shell"
	"github.com/user/envault/internal/store"
)

func newVaultWithSecrets(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault("test-project", []byte("supersecretkey16"))
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PASS", `p@ss"word`)
	v.Set("API_KEY", "abc123")
	return v
}

func TestExportEnvBash(t *testing.T) {
	v := newVaultWithSecrets(t)
	out, err := shell.ExportEnv(v, shell.FormatBash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "export DB_HOST=") {
		t.Errorf("expected bash export statement, got:\n%s", out)
	}
	if !strings.Contains(out, "export API_KEY=") {
		t.Errorf("expected API_KEY export, got:\n%s", out)
	}
}

func TestExportEnvFish(t *testing.T) {
	v := newVaultWithSecrets(t)
	out, err := shell.ExportEnv(v, shell.FormatFish)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "set -x DB_HOST") {
		t.Errorf("expected fish set -x statement, got:\n%s", out)
	}
}

func TestExportEnvDotenv(t *testing.T) {
	v := newVaultWithSecrets(t)
	out, err := shell.ExportEnv(v, shell.FormatDotenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "export ") {
		t.Errorf("dotenv format should not contain 'export', got:\n%s", out)
	}
	if !strings.Contains(out, "DB_HOST=") {
		t.Errorf("expected KEY=VALUE format, got:\n%s", out)
	}
}

func TestExportEnvEmpty(t *testing.T) {
	v := store.NewVault("empty", []byte("key1234567890123"))
	out, err := shell.ExportEnv(v, shell.FormatBash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty output for empty vault, got: %q", out)
	}
}

func TestParseFormat(t *testing.T) {
	cases := []struct {
		input    string
		want     shell.Format
		wantErr  bool
	}{
		{"bash", shell.FormatBash, false},
		{"sh", shell.FormatBash, false},
		{"", shell.FormatBash, false},
		{"fish", shell.FormatFish, false},
		{"dotenv", shell.FormatDotenv, false},
		{"BASH", shell.FormatBash, false},
		{"zsh", shell.FormatBash, true},
	}
	for _, tc := range cases {
		got, err := shell.ParseFormat(tc.input)
		if tc.wantErr && err == nil {
			t.Errorf("ParseFormat(%q): expected error, got nil", tc.input)
			continue
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ParseFormat(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("ParseFormat(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

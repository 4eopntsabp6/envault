package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/envault/internal/cli"
	"github.com/yourusername/envault/internal/store"
)

func setupTemplateVault(t *testing.T) (vaultPath string) {
	t.Helper()
	dir := t.TempDir()
	vaultPath = filepath.Join(dir, "vault.enc")
	v := store.NewVault(vaultPath, "pass")
	v.Set("APP_HOST", "example.com")
	v.Set("APP_PORT", "8080")
	v.Set("SECRET_TOKEN", "tok_abc123")
	if err := v.Save(); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	return vaultPath
}

func writeTmpl(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "test.tmpl")
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("write tmpl: %v", err)
	}
	return p
}

func TestRunTemplateBasic(t *testing.T) {
	vaultPath := setupTemplateVault(t)
	tmplPath := writeTmpl(t, "HOST={{APP_HOST}}\nPORT={{APP_PORT}}\n")

	var buf bytes.Buffer
	if err := cli.RunTemplate(vaultPath, "pass", tmplPath, false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "HOST=example.com") {
		t.Errorf("expected HOST=example.com in output, got: %s", out)
	}
	if !strings.Contains(out, "PORT=8080") {
		t.Errorf("expected PORT=8080 in output, got: %s", out)
	}
}

func TestRunTemplateMissingKeyLenient(t *testing.T) {
	vaultPath := setupTemplateVault(t)
	tmplPath := writeTmpl(t, "X={{UNDEFINED_KEY}}\n")

	var buf bytes.Buffer
	if err := cli.RunTemplate(vaultPath, "pass", tmplPath, false, &buf); err != nil {
		t.Fatalf("unexpected error in lenient mode: %v", err)
	}
	if !strings.Contains(buf.String(), "warning") {
		t.Errorf("expected warning for missing key, got: %s", buf.String())
	}
}

func TestRunTemplateMissingKeyStrict(t *testing.T) {
	vaultPath := setupTemplateVault(t)
	tmplPath := writeTmpl(t, "X={{UNDEFINED_KEY}}\n")

	var buf bytes.Buffer
	err := cli.RunTemplate(vaultPath, "pass", tmplPath, true, &buf)
	if err == nil {
		t.Fatal("expected error in strict mode, got nil")
	}
}

func TestRunTemplateMissingTemplateFile(t *testing.T) {
	vaultPath := setupTemplateVault(t)
	var buf bytes.Buffer
	err := cli.RunTemplate(vaultPath, "pass", "/no/such/file.tmpl", false, &buf)
	if err == nil {
		t.Fatal("expected error for missing template file")
	}
}

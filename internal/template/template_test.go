package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/envault/internal/store"
	"github.com/yourusername/envault/internal/template"
)

func newTestVault(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault(filepath.Join(t.TempDir(), "vault.enc"), "testpass")
	v.Set("DB_HOST", "localhost")
	v.Set("DB_PORT", "5432")
	v.Set("API_KEY", "secret123")
	return v
}

func TestRenderBasic(t *testing.T) {
	v := newTestVault(t)
	res, err := template.Render("host={{DB_HOST}} port={{DB_PORT}}", v, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "host=localhost port=5432"
	if res.Output != want {
		t.Errorf("got %q, want %q", res.Output, want)
	}
	if len(res.Resolved) != 2 {
		t.Errorf("expected 2 resolved keys, got %d", len(res.Resolved))
	}
	if len(res.Missing) != 0 {
		t.Errorf("expected 0 missing keys, got %d", len(res.Missing))
	}
}

func TestRenderMissingKeyLenient(t *testing.T) {
	v := newTestVault(t)
	res, err := template.Render("key={{ MISSING_KEY }}", v, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Missing) != 1 || res.Missing[0] != "MISSING_KEY" {
		t.Errorf("expected MISSING_KEY in missing, got %v", res.Missing)
	}
	// Placeholder should remain unchanged
	if res.Output != "key={{ MISSING_KEY }}" {
		t.Errorf("unexpected output: %q", res.Output)
	}
}

func TestRenderMissingKeyStrict(t *testing.T) {
	v := newTestVault(t)
	_, err := template.Render("{{MISSING}}", v, true)
	if err == nil {
		t.Fatal("expected error in strict mode, got nil")
	}
}

func TestRenderWithSpacesInPlaceholder(t *testing.T) {
	v := newTestVault(t)
	res, err := template.Render("{{ API_KEY }}", v, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Output != "secret123" {
		t.Errorf("got %q, want %q", res.Output, "secret123")
	}
}

func TestRenderFile(t *testing.T) {
	v := newTestVault(t)
	tmpFile := filepath.Join(t.TempDir(), "config.tmpl")
	content := "DATABASE_URL=postgres://{{DB_HOST}}:{{DB_PORT}}/mydb"
	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	res, err := template.RenderFile(tmpFile, v, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "DATABASE_URL=postgres://localhost:5432/mydb"
	if res.Output != want {
		t.Errorf("got %q, want %q", res.Output, want)
	}
}

func TestRenderFileMissing(t *testing.T) {
	v := newTestVault(t)
	_, err := template.RenderFile("/nonexistent/file.tmpl", v, false)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

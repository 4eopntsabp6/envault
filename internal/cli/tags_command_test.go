package cli_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/user/envault/internal/cli"
	"github.com/user/envault/internal/store"
)

func setupTagsVault(t *testing.T) string {
	t.Helper()
	vp := filepath.Join(t.TempDir(), "tags_test.vault")
	v := store.NewVault()
	v.Set("DB_PASSWORD", "secret123")
	v.Set("API_KEY", "key-abc")
	v.Set("DEV_TOKEN", "dev-xyz")
	if err := store.Save(v, vp, "pass"); err != nil {
		t.Fatalf("setup vault: %v", err)
	}
	return vp
}

func TestRunTagsSet(t *testing.T) {
	vp := setupTagsVault(t)
	var buf bytes.Buffer
	err := cli.RunTags(vp, "pass", []string{"set", "DB_PASSWORD", "prod,database"}, &buf)
	if err != nil {
		t.Fatalf("RunTags set: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("tagged DB_PASSWORD")) {
		t.Errorf("expected tagged confirmation, got: %s", buf.String())
	}
}

func TestRunTagsGet(t *testing.T) {
	vp := setupTagsVault(t)
	// set first
	cli.RunTags(vp, "pass", []string{"set", "API_KEY", "prod"}, io.Discard)
	var buf bytes.Buffer
	err := cli.RunTags(vp, "pass", []string{"get", "API_KEY"}, &buf)
	if err != nil {
		t.Fatalf("RunTags get: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("prod")) {
		t.Errorf("expected prod tag, got: %s", buf.String())
	}
}

func TestRunTagsGetNoTags(t *testing.T) {
	vp := setupTagsVault(t)
	var buf bytes.Buffer
	err := cli.RunTags(vp, "pass", []string{"get", "DEV_TOKEN"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("no tags")) {
		t.Errorf("expected 'no tags' message, got: %s", buf.String())
	}
}

func TestRunTagsFilter(t *testing.T) {
	vp := setupTagsVault(t)
	cli.RunTags(vp, "pass", []string{"set", "DB_PASSWORD", "prod,database"}, io.Discard)
	cli.RunTags(vp, "pass", []string{"set", "API_KEY", "prod"}, io.Discard)
	cli.RunTags(vp, "pass", []string{"set", "DEV_TOKEN", "dev"}, io.Discard)

	var buf bytes.Buffer
	err := cli.RunTags(vp, "pass", []string{"filter", "prod"}, &buf)
	if err != nil {
		t.Fatalf("RunTags filter: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("DB_PASSWORD")) {
		t.Errorf("expected DB_PASSWORD in filter result, got: %s", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("API_KEY")) {
		t.Errorf("expected API_KEY in filter result, got: %s", buf.String())
	}
}

func TestRunTagsSetMissingKey(t *testing.T) {
	vp := setupTagsVault(t)
	var buf bytes.Buffer
	err := cli.RunTags(vp, "pass", []string{"set", "NONEXISTENT", "prod"}, &buf)
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

func TestRunTagsUnknownSubcommand(t *testing.T) {
	vp := setupTagsVault(t)
	err := cli.RunTags(vp, "pass", []string{"bogus"}, io.Discard)
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}

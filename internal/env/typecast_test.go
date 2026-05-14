package env

import (
	"testing"

	"github.com/envault/envault/internal/store"
)

func newTypecastVault(t *testing.T) *store.Vault {
	t.Helper()
	v := store.NewVault(t.TempDir()+"/test.vault", "password")
	v.Set("PORT", "8080")
	v.Set("RATIO", "3.14")
	v.Set("DEBUG", "true")
	v.Set("NAME", "envault")
	return v
}

func TestSetTypeHintValid(t *testing.T) {
	v := newTypecastVault(t)
	if err := SetTypeHint(v, "PORT", "int"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := GetTypeHint(v, "PORT"); got != "int" {
		t.Errorf("expected int, got %q", got)
	}
}

func TestSetTypeHintUnsupportedType(t *testing.T) {
	v := newTypecastVault(t)
	if err := SetTypeHint(v, "PORT", "json"); err == nil {
		t.Error("expected error for unsupported type")
	}
}

func TestSetTypeHintMissingKey(t *testing.T) {
	v := newTypecastVault(t)
	if err := SetTypeHint(v, "MISSING", "int"); err == nil {
		t.Error("expected error for missing key")
	}
}

func TestGetTypeHintDefault(t *testing.T) {
	v := newTypecastVault(t)
	if got := GetTypeHint(v, "NAME"); got != "string" {
		t.Errorf("expected default string, got %q", got)
	}
}

func TestCastValueInt(t *testing.T) {
	v := newTypecastVault(t)
	_ = SetTypeHint(v, "PORT", "int")
	val, err := CastValue(v, "PORT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "8080" {
		t.Errorf("expected 8080, got %q", val)
	}
}

func TestCastValueIntInvalid(t *testing.T) {
	v := newTypecastVault(t)
	v.Set("BAD_INT", "notanumber")
	_ = SetTypeHint(v, "BAD_INT", "int")
	if _, err := CastValue(v, "BAD_INT"); err == nil {
		t.Error("expected cast error for invalid int")
	}
}

func TestCastValueFloat(t *testing.T) {
	v := newTypecastVault(t)
	_ = SetTypeHint(v, "RATIO", "float")
	if _, err := CastValue(v, "RATIO"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCastValueBool(t *testing.T) {
	v := newTypecastVault(t)
	_ = SetTypeHint(v, "DEBUG", "bool")
	if _, err := CastValue(v, "DEBUG"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCastValueBoolInvalid(t *testing.T) {
	v := newTypecastVault(t)
	v.Set("FLAG", "yes")
	_ = SetTypeHint(v, "FLAG", "bool")
	if _, err := CastValue(v, "FLAG"); err == nil {
		t.Error("expected cast error for invalid bool")
	}
}

func TestValidateAllCasts(t *testing.T) {
	v := newTypecastVault(t)
	_ = SetTypeHint(v, "PORT", "int")
	_ = SetTypeHint(v, "RATIO", "float")
	_ = SetTypeHint(v, "DEBUG", "bool")
	errs := ValidateAllCasts(v)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateAllCastsFindsError(t *testing.T) {
	v := newTypecastVault(t)
	v.Set("BROKEN", "notanint")
	_ = SetTypeHint(v, "BROKEN", "int")
	errs := ValidateAllCasts(v)
	if len(errs) == 0 {
		t.Error("expected at least one validation error")
	}
}

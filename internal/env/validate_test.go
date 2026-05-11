package env

import (
	"testing"
)

func TestValidateKeyValid(t *testing.T) {
	validKeys := []string{
		"MY_VAR",
		"_PRIVATE",
		"DB_HOST_1",
		"A",
		"_",
	}
	for _, k := range validKeys {
		if err := ValidateKey(k); err != nil {
			t.Errorf("expected key %q to be valid, got: %v", k, err)
		}
	}
}

func TestValidateKeyEmpty(t *testing.T) {
	if err := ValidateKey(""); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestValidateKeyStartsWithDigit(t *testing.T) {
	if err := ValidateKey("1VAR"); err == nil {
		t.Error("expected error for key starting with digit")
	}
}

func TestValidateKeyContainsHyphen(t *testing.T) {
	if err := ValidateKey("MY-VAR"); err == nil {
		t.Error("expected error for key containing hyphen")
	}
}

func TestValidateKeyContainsSpace(t *testing.T) {
	if err := ValidateKey("MY VAR"); err == nil {
		t.Error("expected error for key containing space")
	}
}

func TestValidateKeysAllValid(t *testing.T) {
	keys := []string{"FOO", "BAR", "BAZ_1"}
	errs := ValidateKeys(keys)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateKeysSomeInvalid(t *testing.T) {
	keys := []string{"VALID", "1INVALID", "ALSO-BAD"}
	errs := ValidateKeys(keys)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}
}

func TestValidateValueValid(t *testing.T) {
	if err := ValidateValue("KEY", "hello world"); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateValueWithNewline(t *testing.T) {
	if err := ValidateValue("KEY", "line1\nline2"); err == nil {
		t.Error("expected error for value containing newline")
	}
}

func TestValidateValueWithCarriageReturn(t *testing.T) {
	if err := ValidateValue("KEY", "val\r"); err == nil {
		t.Error("expected error for value containing carriage return")
	}
}

func TestValidationErrorMessage(t *testing.T) {
	err := ValidationError{Key: "BAD KEY", Reason: "contains space"}
	got := err.Error()
	if got != `key "BAD KEY": contains space` {
		t.Errorf("unexpected error message: %s", got)
	}
}

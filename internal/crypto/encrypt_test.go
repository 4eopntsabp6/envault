package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	key := DeriveKey("my-secret-passphrase")
	if len(key) != 32 {
		t.Fatalf("expected key length 32, got %d", len(key))
	}

	// Same passphrase should always produce the same key
	key2 := DeriveKey("my-secret-passphrase")
	if !bytes.Equal(key, key2) {
		t.Fatal("expected deterministic key derivation")
	}

	// Different passphrases should produce different keys
	key3 := DeriveKey("different-passphrase")
	if bytes.Equal(key, key3) {
		t.Fatal("different passphrases should not produce the same key")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := DeriveKey("test-passphrase")
	plaintext := []byte("DATABASE_URL=postgres://user:pass@localhost/db")

	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	decrypted, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptProducesUniqueOutput(t *testing.T) {
	key := DeriveKey("test-passphrase")
	plaintext := []byte("SECRET=value")

	ct1, _ := Encrypt(key, plaintext)
	ct2, _ := Encrypt(key, plaintext)

	// Due to random nonce, two encryptions of the same plaintext should differ
	if bytes.Equal(ct1, ct2) {
		t.Fatal("expected unique ciphertexts for the same plaintext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key := DeriveKey("correct-passphrase")
	wrongKey := DeriveKey("wrong-passphrase")
	plaintext := []byte("API_KEY=supersecret")

	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(wrongKey, ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong key")
	}
}

func TestDecryptShortCiphertext(t *testing.T) {
	key := DeriveKey("passphrase")
	_, err := Decrypt(key, []byte("short"))
	if err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

package secrets

import "testing"

func TestEncryptDecryptString(t *testing.T) {
	secret := "test-secret"
	plaintext := "sk-test-value"

	encrypted, err := EncryptString(secret, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if encrypted == plaintext || encrypted == "" {
		t.Fatalf("unexpected encrypted value %q", encrypted)
	}

	decrypted, err := DecryptString(secret, encrypted)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("decrypted value = %q", decrypted)
	}
}

func TestEncryptRequiresSecret(t *testing.T) {
	if _, err := EncryptString("", "sk-test"); err == nil {
		t.Fatal("expected missing secret error")
	}
}

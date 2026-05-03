package crypto

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	service := New("test-master-key")

	encrypted, err := service.Encrypt("super-secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if encrypted == "super-secret" {
		t.Fatalf("Encrypt() returned plaintext")
	}

	decrypted, err := service.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if decrypted != "super-secret" {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, "super-secret")
	}
}

func TestGeneratePassword(t *testing.T) {
	password, err := GeneratePassword(24)
	if err != nil {
		t.Fatalf("GeneratePassword() error = %v", err)
	}
	if len(password) < 24 {
		t.Fatalf("GeneratePassword() length = %d, want at least 24", len(password))
	}
}

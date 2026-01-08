package utils

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyPasswordMD5Disabled(t *testing.T) {
	// MD5 fallback is disabled by default for security
	AllowLegacyMD5 = false
	hash := Md5("secret" + "rustdesk-api")
	ok, _, err := VerifyPassword(hash, "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("MD5 should be rejected when AllowLegacyMD5 is false")
	}
}

func TestVerifyPasswordMD5Enabled(t *testing.T) {
	// When explicitly enabled, MD5 fallback should work for migration
	AllowLegacyMD5 = true
	defer func() { AllowLegacyMD5 = false }()

	hash := Md5("secret" + "rustdesk-api")
	ok, newHash, err := VerifyPassword(hash, "secret")
	if err != nil {
		t.Fatalf("md5 verify failed: %v", err)
	}
	if !ok || newHash == "" {
		t.Fatalf("md5 migration failed")
	}
	if bcrypt.CompareHashAndPassword([]byte(newHash), []byte("secret")) != nil {
		t.Fatalf("invalid rehash")
	}
}

func TestVerifyPasswordBcrypt(t *testing.T) {
	b, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	ok, newHash, err := VerifyPassword(string(b), "pass")
	if err != nil || !ok || newHash != "" {
		t.Fatalf("bcrypt verify failed")
	}
}

func TestVerifyPasswordMigrate(t *testing.T) {
	// Enable MD5 for migration test
	AllowLegacyMD5 = true
	defer func() { AllowLegacyMD5 = false }()

	md5hash := Md5("mypass" + "rustdesk-api")
	ok, newHash, err := VerifyPassword(md5hash, "mypass")
	if err != nil || !ok || newHash == "" {
		t.Fatalf("expected bcrypt rehash")
	}
	if bcrypt.CompareHashAndPassword([]byte(newHash), []byte("mypass")) != nil {
		t.Fatalf("rehash not valid bcrypt")
	}
}

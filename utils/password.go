package utils

import (
	"errors"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// AllowLegacyMD5 controls whether legacy MD5 password fallback is enabled.
// Security: This should be disabled in production. Only enable temporarily for migration.
// Set environment variable ALLOW_LEGACY_MD5=true to enable.
var AllowLegacyMD5 = os.Getenv("ALLOW_LEGACY_MD5") == "true"

// EncryptPassword hashes the input password using bcrypt.
// An error is returned if hashing fails.
func EncryptPassword(password string) (string, error) {
	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// VerifyPassword checks the input password against the stored hash.
// Legacy MD5 fallback is disabled by default for security.
// Set ALLOW_LEGACY_MD5=true environment variable to enable migration mode.
func VerifyPassword(hash, input string) (bool, string, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input))
	if err == nil {
		return true, "", nil
	}

	var invalidPrefixErr bcrypt.InvalidHashPrefixError
	if errors.As(err, &invalidPrefixErr) || errors.Is(err, bcrypt.ErrHashTooShort) {
		// Security: Only allow MD5 fallback if explicitly enabled
		if AllowLegacyMD5 {
			if hash == Md5(input+"rustdesk-api") {
				log.Println("[SECURITY WARNING] Legacy MD5 password used. Please update to bcrypt.")
				newHash, err2 := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
				if err2 != nil {
					return true, "", err2
				}
				return true, string(newHash), nil
			}
		}
		// MD5 fallback disabled - treat as invalid password
		return false, "", nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, "", nil
	}
	return false, "", err
}

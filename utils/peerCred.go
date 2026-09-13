package utils

import (
	"crypto/subtle"
	"unicode"
)

func ValidPeerId(id string) bool {
	if len(id) < 1 || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func ValidPeerUuid(uuid string) bool {
	if len(uuid) < 8 || len(uuid) > 128 {
		return false
	}
	for _, r := range uuid {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != '=' && r != '+' && r != '/' {
			return false
		}
	}
	return true
}

func PeerUuidMatch(stored, got string) bool {
	if stored == "" || got == "" || len(stored) != len(got) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(got)) == 1
}

package utils

import "testing"

func TestValidPeerId(t *testing.T) {
	if !ValidPeerId("433249324") {
		t.Fatal("numeric rustdesk id")
	}
	if ValidPeerId("") || ValidPeerId("id;rm") || ValidPeerId("a/b") {
		t.Fatal("bad id must be rejected")
	}
}

func TestValidPeerUuid(t *testing.T) {
	ok := "Zjk4YmRkMDQtZmJiNC00MDQyLThmMjctNGYxMTBhM2Q0MDNh"
	if !ValidPeerUuid(ok) {
		t.Fatal("base64 uuid from clients")
	}
	if ValidPeerUuid("") || ValidPeerUuid("short") || ValidPeerUuid("abc\ndefghij") {
		t.Fatal("bad uuid must be rejected")
	}
}

func TestPeerUuidMatch(t *testing.T) {
	u := "Zjk4YmRkMDQtZmJiNC00MDQyLThmMjctNGYxMTBhM2Q0MDNh"
	if !PeerUuidMatch(u, u) {
		t.Fatal("same uuid")
	}
	if PeerUuidMatch(u, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA") {
		t.Fatal("different uuid")
	}
	if PeerUuidMatch(u, "") || PeerUuidMatch("", u) {
		t.Fatal("empty")
	}
}

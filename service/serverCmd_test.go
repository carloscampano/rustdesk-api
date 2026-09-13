package service

import (
	"testing"

	"github.com/lejianwen/rustdesk-api/v2/model"
)

func TestAllowedServerCmd(t *testing.T) {
	if !AllowedServerCmd("h", model.ServerCmdTargetIdServer) {
		t.Fatal("h should be allowed")
	}
	if !AllowedServerCmd("aur", model.ServerCmdTargetIdServer) {
		t.Fatal("aur should be allowed")
	}
	if !AllowedServerCmd("ml", model.ServerCmdTargetIdServer) {
		t.Fatal("ml should be allowed")
	}
	if !AllowedServerCmd("must-login", model.ServerCmdTargetIdServer) {
		t.Fatal("must-login should be allowed")
	}
	if !AllowedServerCmd("usage", model.ServerCmdTargetRelayServer) {
		t.Fatal("usage should be allowed on relay")
	}
	if AllowedServerCmd("h; rm", model.ServerCmdTargetIdServer) {
		t.Fatal("injected cmd must be rejected")
	}
	if AllowedServerCmd("not-a-cmd", model.ServerCmdTargetIdServer) {
		t.Fatal("unknown cmd must be rejected")
	}
	if AllowedServerCmd("aur", model.ServerCmdTargetRelayServer) {
		t.Fatal("aur is not a relay command")
	}
	if AllowedServerCmd("", model.ServerCmdTargetIdServer) {
		t.Fatal("empty cmd must be rejected")
	}
	if AllowedServerCmd("h\nrm", model.ServerCmdTargetIdServer) {
		t.Fatal("newline in cmd must be rejected")
	}
}

func TestSanitizeCmdArg(t *testing.T) {
	if _, err := sanitizeCmdArg("y"); err != nil {
		t.Fatal("plain option should be allowed")
	}
	if _, err := sanitizeCmdArg("a\nb"); err == nil {
		t.Fatal("newline in option must be rejected")
	}
	if _, err := sanitizeCmdArg(string(make([]byte, 513))); err == nil {
		t.Fatal("oversized option must be rejected")
	}
}

package utils

import "testing"

func TestNeedsPublicCompanion(t *testing.T) {
	cases := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.40", true},
		{"10.13.5.20", true},
		{"172.16.0.8", true},
		{"10.8.1.188", false},
		{"10.8.1.1", false},
		{"186.10.199.181", false},
		{"127.0.0.1", true},
		{"", false},
	}
	for _, c := range cases {
		if got := NeedsPublicCompanion(c.ip); got != c.want {
			t.Fatalf("%s: got %v want %v", c.ip, got, c.want)
		}
	}
}

func TestIsPublicIP(t *testing.T) {
	if !IsPublicIP("186.10.199.181") {
		t.Fatal("expected public")
	}
	if IsPublicIP("10.8.1.188") {
		t.Fatal("server lan is not public")
	}
}

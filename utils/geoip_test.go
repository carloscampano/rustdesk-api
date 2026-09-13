package utils

import "testing"

func TestLookupGeoBatchPrivateAndInvalid(t *testing.T) {
	got := LookupGeoBatch([]string{"10.8.1.188", "127.0.0.1", "not-an-ip", "192.168.1.1", "10.8.1.188"})
	if _, ok := got["not-an-ip"]; ok {
		t.Fatal("hostname/garbage must be dropped")
	}
	if !got["10.8.1.188"].Private || got["10.8.1.188"].Country != "Red local" {
		t.Fatalf("lan: %+v", got["10.8.1.188"])
	}
	if !got["127.0.0.1"].Private {
		t.Fatal("loopback should be private")
	}
	if !got["192.168.1.1"].Private {
		t.Fatal("rfc1918 should be private")
	}
}

func TestLookupGeoBatchCaps(t *testing.T) {
	ips := make([]string, 50)
	for i := range ips {
		ips[i] = "10.0.0.1"
	}
	got := LookupGeoBatch(ips)
	if len(got) != 1 {
		t.Fatalf("dedupe got %d", len(got))
	}
}

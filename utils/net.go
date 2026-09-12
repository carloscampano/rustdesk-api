package utils

import (
	"net"
	"strings"
)

// ServerLAN is the host network where rustdesk-api/hbbs live.
var ServerLAN = &net.IPNet{IP: net.ParseIP("10.8.1.0"), Mask: net.CIDRMask(24, 32)}

func ParseIP(s string) net.IP {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	return net.ParseIP(s)
}

func IsPrivateIP(s string) bool {
	ip := ParseIP(s)
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
}

func IsServerLAN(s string) bool {
	ip := ParseIP(s)
	if ip == nil || ServerLAN == nil {
		return false
	}
	return ServerLAN.Contains(ip)
}

func IsPublicIP(s string) bool {
	ip := ParseIP(s)
	return ip != nil && !IsPrivateIP(s)
}

// NeedsPublicCompanion is true for RFC1918/CGNAT addresses that are not the
// server LAN (10.8.1.0/24). Those are clinic/home LANs reported by the client.
func NeedsPublicCompanion(s string) bool {
	return IsPrivateIP(s) && !IsServerLAN(s)
}

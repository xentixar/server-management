package helpers

import (
	"net"
	"regexp"
	"strings"
)

var firewallNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _.-]{0,63}$`)

func IsValidFirewallName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return true
	}
	return firewallNameRe.MatchString(name)
}

func IsValidFirewallCIDR(cidr string) bool {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" || cidr == "any" {
		return true
	}
	_, network, err := net.ParseCIDR(cidr)
	if err == nil {
		return network.IP.To4() != nil
	}
	ip := net.ParseIP(cidr)
	return ip != nil && ip.To4() != nil
}

func NormalizeFirewallCIDR(cidr string) string {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" || cidr == "any" {
		return "0.0.0.0/0"
	}
	if _, _, err := net.ParseCIDR(cidr); err == nil {
		return cidr
	}
	ip := net.ParseIP(cidr)
	if ip != nil && ip.To4() != nil {
		return ip.String() + "/32"
	}
	return cidr
}

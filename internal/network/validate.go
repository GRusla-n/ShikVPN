package network

import (
	"fmt"
	"net"
	"regexp"
)

// validIfaceName matches safe interface names: alphanumeric, hyphens, underscores, dots, max 15 chars.
var validIfaceName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,14}$`)

// ValidateInterfaceName checks that the interface name is safe for use in OS commands.
func ValidateInterfaceName(name string) error {
	if !validIfaceName.MatchString(name) {
		return fmt.Errorf("invalid interface name %q: must be 1-15 alphanumeric characters, hyphens, underscores, or dots", name)
	}
	return nil
}

// ValidateCIDR checks that the address is a valid CIDR notation.
func ValidateCIDR(address string) error {
	_, _, err := net.ParseCIDR(address)
	if err != nil {
		return fmt.Errorf("invalid CIDR address %q: %w", address, err)
	}
	return nil
}

// ValidateIP checks that the string is a valid IP address.
func ValidateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address %q", ip)
	}
	return nil
}

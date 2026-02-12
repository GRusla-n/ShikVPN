package network

// InterfaceConfigurator provides platform-specific network interface configuration.
type InterfaceConfigurator interface {
	// AssignAddress assigns an IP address to a network interface.
	AssignAddress(ifaceName string, address string) error

	// SetInterfaceUp brings a network interface up.
	SetInterfaceUp(ifaceName string) error

	// SetMTU sets the MTU on a network interface.
	SetMTU(ifaceName string, mtu int) error

	// AddRoute adds a route via the specified interface.
	AddRoute(destination string, gateway string, ifaceName string) error

	// SetDefaultRoute sets the default route through the VPN tunnel.
	// It saves the current default route for restoration.
	SetDefaultRoute(ifaceName string, gateway string, serverEndpoint string) error

	// RemoveDefaultRoute restores the original default route.
	RemoveDefaultRoute(ifaceName string) error

	// EnableIPForwarding enables IP forwarding on the system (server-side).
	EnableIPForwarding() error

	// ConfigureNAT sets up NAT/masquerade for VPN traffic (server-side).
	ConfigureNAT(ifaceName string, vpnSubnet string) error

	// RemoveNAT removes NAT rules (cleanup).
	RemoveNAT(ifaceName string, vpnSubnet string) error
}

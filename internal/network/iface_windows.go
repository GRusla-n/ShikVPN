//go:build windows

package network

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

// WindowsConfigurator implements InterfaceConfigurator for Windows.
type WindowsConfigurator struct {
	savedGateway   string
	savedInterface string
}

func NewConfigurator() InterfaceConfigurator {
	return &WindowsConfigurator{}
}

func (c *WindowsConfigurator) AssignAddress(ifaceName string, address string) error {
	if err := ValidateInterfaceName(ifaceName); err != nil {
		return err
	}
	if err := ValidateCIDR(address); err != nil {
		return err
	}
	ip, mask := splitCIDR(address)
	return runCmd("netsh", "interface", "ip", "set", "address",
		fmt.Sprintf("name=%s", ifaceName), "static", ip, mask)
}

func (c *WindowsConfigurator) SetInterfaceUp(ifaceName string) error {
	// Windows interfaces come up automatically when configured via netsh
	return nil
}

func (c *WindowsConfigurator) SetMTU(ifaceName string, mtu int) error {
	return runCmd("netsh", "interface", "ipv4", "set", "subinterface",
		ifaceName, fmt.Sprintf("mtu=%d", mtu), "store=persistent")
}

func (c *WindowsConfigurator) AddRoute(destination string, gateway string, ifaceName string) error {
	dest, mask := splitCIDR(destination)
	if gateway != "" {
		return runCmd("route", "add", dest, "mask", mask, gateway)
	}
	idx, err := getInterfaceIndex(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface index for %s: %w", ifaceName, err)
	}
	return runCmd("route", "add", dest, "mask", mask, "0.0.0.0", "if", idx)
}

func (c *WindowsConfigurator) SetDefaultRoute(ifaceName string, gateway string, serverEndpoint string) error {
	// Save current default gateway
	out, err := exec.Command("cmd", "/c", "route", "print", "0.0.0.0").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 5 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
				c.savedGateway = fields[2]
				c.savedInterface = fields[4]
				break
			}
		}
	}

	// Route server endpoint via original gateway
	if c.savedGateway != "" && serverEndpoint != "" {
		host := strings.Split(serverEndpoint, ":")[0]
		_ = runCmd("route", "add", host, "mask", "255.255.255.255", c.savedGateway)
	}

	// Get the interface index for route commands
	idx, err := getInterfaceIndex(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface index for %s: %w", ifaceName, err)
	}

	// Delete current default route and add VPN default route
	_ = runCmd("route", "delete", "0.0.0.0", "mask", "0.0.0.0")
	return runCmd("route", "add", "0.0.0.0", "mask", "0.0.0.0", gateway, "if", idx)
}

func (c *WindowsConfigurator) RemoveDefaultRoute(ifaceName string) error {
	_ = runCmd("route", "delete", "0.0.0.0", "mask", "0.0.0.0")
	if c.savedGateway != "" {
		return runCmd("route", "add", "0.0.0.0", "mask", "0.0.0.0", c.savedGateway)
	}
	return nil
}

func (c *WindowsConfigurator) EnableIPForwarding() error {
	return runCmd("powershell", "-Command",
		"Set-NetIPInterface -Forwarding Enabled -AddressFamily IPv4")
}

func (c *WindowsConfigurator) ConfigureNAT(ifaceName string, vpnSubnet string) error {
	// Validate vpnSubnet is a proper CIDR before passing to PowerShell
	if _, _, err := net.ParseCIDR(vpnSubnet); err != nil {
		return fmt.Errorf("invalid VPN subnet CIDR %q: %w", vpnSubnet, err)
	}
	// Windows uses Internet Connection Sharing or netsh routing
	return runCmd("powershell", "-Command",
		fmt.Sprintf("New-NetNat -Name 'ShikVPN' -InternalIPInterfaceAddressPrefix '%s'", vpnSubnet))
}

func (c *WindowsConfigurator) RemoveNAT(ifaceName string, vpnSubnet string) error {
	return runCmd("powershell", "-Command", "Remove-NetNat -Name ShikVPN -Confirm:$false")
}

// getInterfaceIndex returns the numeric interface index for a named interface.
// Windows route commands require the numeric index, not the interface name.
func getInterfaceIndex(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", fmt.Errorf("interface %q not found: %w", ifaceName, err)
	}
	return strconv.Itoa(iface.Index), nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %q failed: %s: %w", name+" "+strings.Join(args, " "), string(output), err)
	}
	return nil
}

// splitCIDR splits "10.0.0.1/24" into IP and dotted netmask.
func splitCIDR(cidr string) (string, string) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Fallback: treat as plain IP with /24
		parts := strings.SplitN(cidr, "/", 2)
		return parts[0], "255.255.255.0"
	}
	mask := ipNet.Mask
	return ip.String(), fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

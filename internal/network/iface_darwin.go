//go:build darwin

package network

import (
	"fmt"
	"os/exec"
	"strings"
)

// DarwinConfigurator implements InterfaceConfigurator for macOS.
type DarwinConfigurator struct {
	savedGateway   string
	savedInterface string
}

func NewConfigurator() InterfaceConfigurator {
	return &DarwinConfigurator{}
}

func (c *DarwinConfigurator) AssignAddress(ifaceName string, address string) error {
	// macOS ifconfig wants "addr netmask mask" format
	ip, mask := splitCIDR(address)
	return runCmd("ifconfig", ifaceName, "inet", ip, ip, "netmask", mask)
}

func (c *DarwinConfigurator) SetInterfaceUp(ifaceName string) error {
	return runCmd("ifconfig", ifaceName, "up")
}

func (c *DarwinConfigurator) SetMTU(ifaceName string, mtu int) error {
	return runCmd("ifconfig", ifaceName, "mtu", fmt.Sprintf("%d", mtu))
}

func (c *DarwinConfigurator) AddRoute(destination string, gateway string, ifaceName string) error {
	if gateway != "" {
		return runCmd("route", "add", "-net", destination, gateway)
	}
	return runCmd("route", "add", "-net", destination, "-interface", ifaceName)
}

func (c *DarwinConfigurator) SetDefaultRoute(ifaceName string, gateway string, serverEndpoint string) error {
	// Save current default route
	out, err := exec.Command("route", "-n", "get", "default").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "gateway:") {
				c.savedGateway = strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
			}
			if strings.HasPrefix(line, "interface:") {
				c.savedInterface = strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
			}
		}
	}

	// Route server endpoint via original gateway
	if c.savedGateway != "" && serverEndpoint != "" {
		host := strings.Split(serverEndpoint, ":")[0]
		_ = runCmd("route", "add", "-host", host, c.savedGateway)
	}

	// Set default route through VPN
	_ = runCmd("route", "delete", "default")
	return runCmd("route", "add", "default", "-interface", ifaceName)
}

func (c *DarwinConfigurator) RemoveDefaultRoute(ifaceName string) error {
	_ = runCmd("route", "delete", "default")
	if c.savedGateway != "" {
		return runCmd("route", "add", "default", c.savedGateway)
	}
	return nil
}

func (c *DarwinConfigurator) EnableIPForwarding() error {
	return runCmd("sysctl", "-w", "net.inet.ip.forwarding=1")
}

func (c *DarwinConfigurator) ConfigureNAT(ifaceName string, vpnSubnet string) error {
	// macOS uses pfctl for NAT — write a minimal pf.conf snippet
	// For MVP, we rely on the user having PF configured or skip NAT on macOS
	return fmt.Errorf("NAT configuration on macOS requires manual pfctl setup")
}

func (c *DarwinConfigurator) RemoveNAT(ifaceName string, vpnSubnet string) error {
	return nil
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
	parts := strings.SplitN(cidr, "/", 2)
	ip := parts[0]
	mask := "255.255.255.0" // default /24
	if len(parts) == 2 {
		switch parts[1] {
		case "8":
			mask = "255.0.0.0"
		case "16":
			mask = "255.255.0.0"
		case "24":
			mask = "255.255.255.0"
		case "32":
			mask = "255.255.255.255"
		}
	}
	return ip, mask
}

//go:build linux

package network

import (
	"fmt"
	"os/exec"
	"strings"
)

// LinuxConfigurator implements InterfaceConfigurator for Linux.
type LinuxConfigurator struct {
	savedGateway   string
	savedInterface string
}

func NewConfigurator() InterfaceConfigurator {
	return &LinuxConfigurator{}
}

func (c *LinuxConfigurator) AssignAddress(ifaceName string, address string) error {
	if err := ValidateInterfaceName(ifaceName); err != nil {
		return err
	}
	if err := ValidateCIDR(address); err != nil {
		return err
	}
	return runCmd("ip", "addr", "add", address, "dev", ifaceName)
}

func (c *LinuxConfigurator) SetInterfaceUp(ifaceName string) error {
	if err := ValidateInterfaceName(ifaceName); err != nil {
		return err
	}
	return runCmd("ip", "link", "set", ifaceName, "up")
}

func (c *LinuxConfigurator) SetMTU(ifaceName string, mtu int) error {
	return runCmd("ip", "link", "set", ifaceName, "mtu", fmt.Sprintf("%d", mtu))
}

func (c *LinuxConfigurator) AddRoute(destination string, gateway string, ifaceName string) error {
	if gateway != "" {
		return runCmd("ip", "route", "add", destination, "via", gateway, "dev", ifaceName)
	}
	return runCmd("ip", "route", "add", destination, "dev", ifaceName)
}

func (c *LinuxConfigurator) SetDefaultRoute(ifaceName string, gateway string, serverEndpoint string) error {
	// Save current default route
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 5 {
			c.savedGateway = parts[2]
			c.savedInterface = parts[4]
		}
	}

	// Add route to server endpoint via original gateway
	if c.savedGateway != "" && serverEndpoint != "" {
		host := strings.Split(serverEndpoint, ":")[0]
		_ = runCmd("ip", "route", "add", host+"/32", "via", c.savedGateway, "dev", c.savedInterface)
	}

	// Replace default route to go through VPN
	_ = runCmd("ip", "route", "del", "default")
	return runCmd("ip", "route", "add", "default", "dev", ifaceName)
}

func (c *LinuxConfigurator) RemoveDefaultRoute(ifaceName string) error {
	_ = runCmd("ip", "route", "del", "default", "dev", ifaceName)

	if c.savedGateway != "" {
		return runCmd("ip", "route", "add", "default", "via", c.savedGateway, "dev", c.savedInterface)
	}
	return nil
}

func (c *LinuxConfigurator) EnableIPForwarding() error {
	return runCmd("sysctl", "-w", "net.ipv4.ip_forward=1")
}

func (c *LinuxConfigurator) ConfigureNAT(ifaceName string, vpnSubnet string) error {
	if err := ValidateCIDR(vpnSubnet); err != nil {
		return err
	}
	// Find the default outbound interface
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return fmt.Errorf("failed to find default route: %w", err)
	}
	parts := strings.Fields(string(out))
	outIface := "eth0"
	if len(parts) >= 5 {
		outIface = parts[4]
	}

	return runCmd("iptables", "-t", "nat", "-A", "POSTROUTING",
		"-s", vpnSubnet, "-o", outIface, "-j", "MASQUERADE")
}

func (c *LinuxConfigurator) RemoveNAT(ifaceName string, vpnSubnet string) error {
	out, _ := exec.Command("ip", "route", "show", "default").Output()
	parts := strings.Fields(string(out))
	outIface := "eth0"
	if len(parts) >= 5 {
		outIface = parts[4]
	}

	return runCmd("iptables", "-t", "nat", "-D", "POSTROUTING",
		"-s", vpnSubnet, "-o", outIface, "-j", "MASQUERADE")
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %q failed: %s: %w", name+" "+strings.Join(args, " "), string(output), err)
	}
	return nil
}

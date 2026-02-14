package client

import (
	"fmt"
	"log"
	"strings"

	"github.com/gavsh/ShikVPN/internal/config"
	"github.com/gavsh/ShikVPN/internal/crypto"
	"github.com/gavsh/ShikVPN/internal/network"
	"github.com/gavsh/ShikVPN/internal/tunnel"
)

// Client orchestrates the VPN client: registration, tunnel, and route management.
type Client struct {
	cfg       *config.ClientConfig
	tunnel    *tunnel.Tunnel
	netConfig network.InterfaceConfigurator
	connected bool
}

// New creates a new VPN client.
func New(cfg *config.ClientConfig) *Client {
	return &Client{
		cfg:       cfg,
		netConfig: network.NewConfigurator(),
	}
}

// Connect performs registration, creates the tunnel, and sets up routes.
func (c *Client) Connect() error {
	// Derive public key from private key for registration
	privKey, err := crypto.KeyFromBase64(c.cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}
	pubKey, err := crypto.PublicKeyFromPrivate(privKey)
	if err != nil {
		return fmt.Errorf("failed to derive public key: %w", err)
	}
	pubKeyB64 := crypto.KeyToBase64(pubKey)
	log.Printf("Client public key: %s", pubKeyB64)

	// Register with server
	apiURL := c.cfg.ServerAPIURL()
	log.Printf("Registering with server at %s...", apiURL)
	regResp, err := Register(apiURL, pubKeyB64, c.cfg.APIKey)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}
	log.Printf("Registered successfully. Assigned IP: %s", regResp.AssignedIP)

	// Store server public key and assigned address
	c.cfg.ServerPublicKey = regResp.ServerPublicKey
	c.cfg.Address = regResp.AssignedIP

	// Use the endpoint returned by the server's registration response
	serverEndpoint := regResp.ServerEndpoint

	// Create TUN device
	tun, err := tunnel.CreateTunnel(c.cfg.InterfaceName, c.cfg.MTU, c.cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}
	c.tunnel = tun
	log.Printf("Created TUN device: %s", tun.Name())

	// Convert keys to hex for UAPI
	privKeyHex := crypto.KeyToHex(privKey)
	serverPubKeyHex, err := crypto.Base64ToHex(c.cfg.ServerPublicKey)
	if err != nil {
		c.tunnel.Close()
		return fmt.Errorf("invalid server public key: %w", err)
	}

	// Configure WireGuard
	peer := tunnel.PeerConfig{
		PublicKeyHex:        serverPubKeyHex,
		Endpoint:            serverEndpoint,
		AllowedIPs:          []string{"0.0.0.0/0"},
		PersistentKeepalive: c.cfg.PersistentKeepalive,
	}

	uapi := tunnel.BuildClientUAPIConfig(privKeyHex, peer)
	if err := c.tunnel.Configure(uapi); err != nil {
		c.tunnel.Close()
		return fmt.Errorf("failed to configure WireGuard: %w", err)
	}

	// Bring device up
	if err := c.tunnel.Up(); err != nil {
		c.tunnel.Close()
		return fmt.Errorf("failed to bring up WireGuard device: %w", err)
	}
	log.Println("WireGuard device is up")

	// Configure network interface
	if err := c.configureNetwork(serverEndpoint); err != nil {
		c.tunnel.Close()
		return fmt.Errorf("failed to configure network: %w", err)
	}

	c.connected = true
	log.Println("VPN connected successfully")
	return nil
}

func (c *Client) configureNetwork(serverEndpoint string) error {
	ifaceName := c.tunnel.Name()

	// Assign the VPN IP to the interface
	if err := c.netConfig.AssignAddress(ifaceName, c.cfg.Address); err != nil {
		return fmt.Errorf("failed to assign address: %w", err)
	}
	log.Printf("Assigned address %s to %s", c.cfg.Address, ifaceName)

	// Set interface up
	if err := c.netConfig.SetInterfaceUp(ifaceName); err != nil {
		return fmt.Errorf("failed to set interface up: %w", err)
	}

	// Extract gateway IP (the server's VPN IP, typically x.x.x.1)
	gateway := extractGateway(c.cfg.Address)

	// Set default route through VPN
	if err := c.netConfig.SetDefaultRoute(ifaceName, gateway, serverEndpoint); err != nil {
		log.Printf("Warning: failed to set default route: %v", err)
		log.Println("VPN is connected but traffic may not be routed through it")
	}

	return nil
}

// Disconnect tears down the VPN tunnel and restores routes.
func (c *Client) Disconnect() {
	if !c.connected {
		return
	}
	log.Println("Disconnecting VPN...")

	if c.tunnel != nil {
		ifaceName := c.tunnel.Name()

		// Restore default route
		if err := c.netConfig.RemoveDefaultRoute(ifaceName); err != nil {
			log.Printf("Warning: failed to restore default route: %v", err)
		}

		c.tunnel.Close()
		log.Println("Tunnel closed")
	}

	c.connected = false
	log.Println("VPN disconnected")
}

// extractGateway derives the gateway IP from a CIDR address.
// e.g., "10.0.0.2/24" → "10.0.0.1"
func extractGateway(address string) string {
	ip := strings.SplitN(address, "/", 2)[0]
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.%s.1", parts[0], parts[1], parts[2])
	}
	return ""
}

package server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gavsh/ShikVPN/internal/config"
	"github.com/gavsh/ShikVPN/internal/crypto"
	"github.com/gavsh/ShikVPN/internal/network"
	"github.com/gavsh/ShikVPN/internal/tunnel"
)

// Server orchestrates the VPN server: tunnel, API, IPAM, and network config.
type Server struct {
	cfg       *config.ServerConfig
	tunnel    *tunnel.Tunnel
	api       *API
	ipam      *IPAM
	netConfig network.InterfaceConfigurator
}

// New creates a new VPN server.
func New(cfg *config.ServerConfig) *Server {
	return &Server{
		cfg:       cfg,
		netConfig: network.NewConfigurator(),
	}
}

// Start initializes and starts all server components.
func (s *Server) Start() error {
	// Initialize IPAM
	ipam, err := NewIPAM(s.cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to create IPAM: %w", err)
	}
	s.ipam = ipam

	// Create TUN device
	tun, err := tunnel.CreateTunnel(s.cfg.InterfaceName, s.cfg.MTU, s.cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}
	s.tunnel = tun
	log.Printf("Created TUN device: %s", tun.Name())

	// Convert private key to hex for UAPI
	privKeyHex, err := crypto.Base64ToHex(s.cfg.PrivateKey)
	if err != nil {
		s.tunnel.Close()
		return fmt.Errorf("invalid server private key: %w", err)
	}

	// Configure WireGuard device (no peers initially)
	uapi := tunnel.BuildServerUAPIConfig(privKeyHex, s.cfg.ListenPort, nil)
	if err := s.tunnel.Configure(uapi); err != nil {
		s.tunnel.Close()
		return fmt.Errorf("failed to configure WireGuard: %w", err)
	}

	// Bring device up
	if err := s.tunnel.Up(); err != nil {
		s.tunnel.Close()
		return fmt.Errorf("failed to bring up WireGuard device: %w", err)
	}
	log.Println("WireGuard device is up")

	// Configure network interface
	if err := s.configureNetwork(); err != nil {
		s.tunnel.Close()
		return fmt.Errorf("failed to configure network: %w", err)
	}

	// Build server endpoint string
	serverEndpoint := fmt.Sprintf("%s:%d", s.cfg.ExternalHost, s.cfg.ListenPort)

	// Create and start API
	s.api = NewAPI(s.ipam, s.cfg.PublicKey, serverEndpoint, s.cfg.DNSServers, s.cfg.MTU, s.cfg.APIKey, s.addPeer)

	apiAddr := fmt.Sprintf(":%d", s.cfg.APIPort)
	go func() {
		if err := s.api.ListenAndServe(apiAddr); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	log.Printf("VPN server started (WG port: %d, API port: %d)", s.cfg.ListenPort, s.cfg.APIPort)
	return nil
}

func (s *Server) configureNetwork() error {
	ifaceName := s.tunnel.Name()

	// Assign IP address
	if err := s.netConfig.AssignAddress(ifaceName, s.cfg.Address); err != nil {
		return fmt.Errorf("failed to assign address: %w", err)
	}
	log.Printf("Assigned address %s to %s", s.cfg.Address, ifaceName)

	// Set interface up
	if err := s.netConfig.SetInterfaceUp(ifaceName); err != nil {
		return fmt.Errorf("failed to set interface up: %w", err)
	}

	// Enable IP forwarding
	if err := s.netConfig.EnableIPForwarding(); err != nil {
		log.Printf("Warning: failed to enable IP forwarding: %v", err)
	}

	// Configure NAT
	subnet := s.cfg.Address
	// Extract subnet from address (e.g., "10.0.0.1/24" -> "10.0.0.0/24")
	if idx := strings.LastIndex(subnet, "."); idx != -1 {
		parts := strings.SplitN(subnet, "/", 2)
		if len(parts) == 2 {
			ipParts := strings.Split(parts[0], ".")
			if len(ipParts) == 4 {
				subnet = fmt.Sprintf("%s.%s.%s.0/%s", ipParts[0], ipParts[1], ipParts[2], parts[1])
			}
		}
	}

	if err := s.netConfig.ConfigureNAT(ifaceName, subnet); err != nil {
		log.Printf("Warning: failed to configure NAT: %v", err)
	}

	return nil
}

// addPeer adds a new peer to the WireGuard device dynamically.
func (s *Server) addPeer(peer tunnel.PeerConfig) error {
	uapi := tunnel.BuildAddPeerUAPI(peer)
	return s.tunnel.Configure(uapi)
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	log.Println("Stopping VPN server...")

	// Gracefully shut down the API server
	if s.api != nil {
		if err := s.api.Shutdown(5 * time.Second); err != nil {
			log.Printf("API shutdown error: %v", err)
		}
		log.Println("API server stopped")
	}

	if s.tunnel != nil {
		ifaceName := s.tunnel.Name()

		// Remove NAT
		subnet := s.cfg.Address
		if idx := strings.LastIndex(subnet, "."); idx != -1 {
			parts := strings.SplitN(subnet, "/", 2)
			if len(parts) == 2 {
				ipParts := strings.Split(parts[0], ".")
				if len(ipParts) == 4 {
					subnet = fmt.Sprintf("%s.%s.%s.0/%s", ipParts[0], ipParts[1], ipParts[2], parts[1])
				}
			}
		}
		_ = s.netConfig.RemoveNAT(ifaceName, subnet)

		s.tunnel.Close()
		log.Println("Tunnel closed")
	}

	log.Println("VPN server stopped")
}

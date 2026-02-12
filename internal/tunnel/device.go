package tunnel

import (
	"fmt"
	"strings"
	"sync"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

// PeerConfig holds configuration for a single WireGuard peer.
type PeerConfig struct {
	PublicKeyHex        string
	PresharedKeyHex     string
	Endpoint            string
	AllowedIPs          []string
	PersistentKeepalive int
}

// Tunnel wraps a WireGuard device with its TUN interface.
type Tunnel struct {
	device    *device.Device
	tunDevice tun.Device
	name      string
	mu        sync.Mutex
	closed    bool
}

// CreateTunnel creates a new TUN device and WireGuard device on top of it.
// logLevel controls WireGuard logging: "verbose", "error", or "silent".
func CreateTunnel(name string, mtu int, logLevel string) (*Tunnel, error) {
	tunDevice, err := tun.CreateTUN(name, mtu)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device %q: %w", name, err)
	}

	actualName, err := tunDevice.Name()
	if err != nil {
		tunDevice.Close()
		return nil, fmt.Errorf("failed to get TUN device name: %w", err)
	}

	level := device.LogLevelError
	switch logLevel {
	case "verbose":
		level = device.LogLevelVerbose
	case "silent":
		level = device.LogLevelSilent
	}

	log := device.NewLogger(level, fmt.Sprintf("(%s) ", actualName))

	wgDevice := device.NewDevice(tunDevice, conn.NewDefaultBind(), log)

	return &Tunnel{
		device:    wgDevice,
		tunDevice: tunDevice,
		name:      actualName,
	}, nil
}

// Configure applies a UAPI configuration string to the WireGuard device.
func (t *Tunnel) Configure(uapiConfig string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("tunnel is closed")
	}

	return t.device.IpcSet(uapiConfig)
}

// Up brings the WireGuard device up.
func (t *Tunnel) Up() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("tunnel is closed")
	}

	return t.device.Up()
}

// Down brings the WireGuard device down.
func (t *Tunnel) Down() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("tunnel is closed")
	}

	return t.device.Down()
}

// Close shuts down the WireGuard device and TUN interface.
func (t *Tunnel) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}
	t.closed = true

	t.device.Close()
	// tunDevice is closed by device.Close()
}

// Name returns the actual TUN interface name.
func (t *Tunnel) Name() string {
	return t.name
}

// Device returns the underlying wireguard-go device for direct IPC access.
func (t *Tunnel) Device() *device.Device {
	return t.device
}

// BuildServerUAPIConfig builds a UAPI config string for the server.
func BuildServerUAPIConfig(privateKeyHex string, listenPort int, peers []PeerConfig) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("private_key=%s\n", privateKeyHex))
	b.WriteString(fmt.Sprintf("listen_port=%d\n", listenPort))

	for _, peer := range peers {
		b.WriteString(fmt.Sprintf("public_key=%s\n", peer.PublicKeyHex))
		if peer.PresharedKeyHex != "" {
			b.WriteString(fmt.Sprintf("preshared_key=%s\n", peer.PresharedKeyHex))
		}
		if peer.Endpoint != "" {
			b.WriteString(fmt.Sprintf("endpoint=%s\n", peer.Endpoint))
		}
		for _, allowedIP := range peer.AllowedIPs {
			b.WriteString(fmt.Sprintf("allowed_ip=%s\n", allowedIP))
		}
		if peer.PersistentKeepalive > 0 {
			b.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", peer.PersistentKeepalive))
		}
	}

	return b.String()
}

// BuildClientUAPIConfig builds a UAPI config string for the client.
func BuildClientUAPIConfig(privateKeyHex string, peer PeerConfig) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("private_key=%s\n", privateKeyHex))

	b.WriteString(fmt.Sprintf("public_key=%s\n", peer.PublicKeyHex))
	if peer.PresharedKeyHex != "" {
		b.WriteString(fmt.Sprintf("preshared_key=%s\n", peer.PresharedKeyHex))
	}
	if peer.Endpoint != "" {
		b.WriteString(fmt.Sprintf("endpoint=%s\n", peer.Endpoint))
	}
	for _, allowedIP := range peer.AllowedIPs {
		b.WriteString(fmt.Sprintf("allowed_ip=%s\n", allowedIP))
	}
	if peer.PersistentKeepalive > 0 {
		b.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", peer.PersistentKeepalive))
	}

	return b.String()
}

// BuildAddPeerUAPI builds a UAPI config string to add a single peer (append mode).
func BuildAddPeerUAPI(peer PeerConfig) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("public_key=%s\n", peer.PublicKeyHex))
	if peer.PresharedKeyHex != "" {
		b.WriteString(fmt.Sprintf("preshared_key=%s\n", peer.PresharedKeyHex))
	}
	if peer.Endpoint != "" {
		b.WriteString(fmt.Sprintf("endpoint=%s\n", peer.Endpoint))
	}
	for _, allowedIP := range peer.AllowedIPs {
		b.WriteString(fmt.Sprintf("allowed_ip=%s\n", allowedIP))
	}
	if peer.PersistentKeepalive > 0 {
		b.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", peer.PersistentKeepalive))
	}

	return b.String()
}

package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// ServerConfig holds the VPN server configuration.
type ServerConfig struct {
	ListenPort   int      `toml:"listen_port"`
	Address      string   `toml:"address"`
	PrivateKey   string   `toml:"private_key"`
	PublicKey    string   `toml:"public_key"`
	APIPort      int      `toml:"api_port"`
	ExternalHost string   `toml:"external_host"`
	DNSServers   []string `toml:"dns_servers"`
	MTU          int      `toml:"mtu"`
	InterfaceName string  `toml:"interface_name"`
}

// ClientConfig holds the VPN client configuration.
type ClientConfig struct {
	ServerEndpoint      string `toml:"server_endpoint"`
	ServerAPIURL        string `toml:"server_api_url"`
	ServerPublicKey     string `toml:"server_public_key"`
	PrivateKey          string `toml:"private_key"`
	Address             string `toml:"address"`
	DNS                 string `toml:"dns"`
	MTU                 int    `toml:"mtu"`
	PersistentKeepalive int    `toml:"persistent_keepalive"`
	InterfaceName       string `toml:"interface_name"`
}

// LoadServerConfig reads and parses a server config from a TOML file.
func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return ParseServerConfig(string(data))
}

// ParseServerConfig parses a server config from a TOML string.
func ParseServerConfig(data string) (*ServerConfig, error) {
	cfg := &ServerConfig{}
	if _, err := toml.Decode(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}
	applyServerDefaults(cfg)
	return cfg, nil
}

// LoadClientConfig reads and parses a client config from a TOML file.
func LoadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return ParseClientConfig(string(data))
}

// ParseClientConfig parses a client config from a TOML string.
func ParseClientConfig(data string) (*ClientConfig, error) {
	cfg := &ClientConfig{}
	if _, err := toml.Decode(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse client config: %w", err)
	}
	applyClientDefaults(cfg)
	return cfg, nil
}

func applyServerDefaults(cfg *ServerConfig) {
	if cfg.ListenPort == 0 {
		cfg.ListenPort = DefaultListenPort
	}
	if cfg.Address == "" {
		cfg.Address = DefaultAddress
	}
	if cfg.APIPort == 0 {
		cfg.APIPort = DefaultAPIPort
	}
	if cfg.MTU == 0 {
		cfg.MTU = DefaultMTU
	}
	if len(cfg.DNSServers) == 0 {
		cfg.DNSServers = DefaultDNSServers
	}
	if cfg.InterfaceName == "" {
		cfg.InterfaceName = DefaultInterfaceName
	}
}

func applyClientDefaults(cfg *ClientConfig) {
	if cfg.MTU == 0 {
		cfg.MTU = DefaultMTU
	}
	if cfg.PersistentKeepalive == 0 {
		cfg.PersistentKeepalive = DefaultPersistentKeepalive
	}
	if cfg.InterfaceName == "" {
		cfg.InterfaceName = DefaultInterfaceName
	}
}

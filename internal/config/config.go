package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
)

// validIfaceNameRe matches safe interface names: alphanumeric, hyphens, underscores, dots, max 15 chars.
var validIfaceNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,14}$`)

// ServerConfig holds the VPN server configuration.
type ServerConfig struct {
	ListenPort    int      `toml:"listen_port"`
	Address       string   `toml:"address"`
	PrivateKey    string   `toml:"private_key"`
	PublicKey     string   `toml:"public_key"`
	APIPort       int      `toml:"api_port"`
	ExternalHost  string   `toml:"external_host"`
	DNSServers    []string `toml:"dns_servers"`
	MTU           int      `toml:"mtu"`
	InterfaceName string   `toml:"interface_name"`
	APIKey        string   `toml:"api_key"`
	LogLevel      string   `toml:"log_level"`
}

// ClientConfig holds the VPN client configuration.
type ClientConfig struct {
	Server              string `toml:"server" json:"server"`
	APIPort             int    `toml:"api_port" json:"api_port"`
	ServerPublicKey     string `toml:"server_public_key" json:"server_public_key"`
	PrivateKey          string `toml:"private_key" json:"private_key"`
	Address             string `toml:"address" json:"address"`
	DNS                 string `toml:"dns" json:"dns"`
	MTU                 int    `toml:"mtu" json:"mtu"`
	PersistentKeepalive int    `toml:"persistent_keepalive" json:"persistent_keepalive"`
	InterfaceName       string `toml:"interface_name" json:"interface_name"`
	APIKey              string `toml:"api_key" json:"api_key"`
	LogLevel            string `toml:"log_level" json:"log_level"`
}

// ServerAPIURL returns the full HTTP URL for the server's registration API.
func (c *ClientConfig) ServerAPIURL() string {
	return fmt.Sprintf("http://%s:%d", c.Server, c.APIPort)
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
	ApplyClientDefaults(cfg)
	return cfg, nil
}

// ValidateServerConfig checks that all required server fields are present and valid.
func ValidateServerConfig(cfg *ServerConfig) error {
	if cfg.PrivateKey == "" {
		return fmt.Errorf("private_key is required")
	}
	if err := validateBase64Key(cfg.PrivateKey, "private_key"); err != nil {
		return err
	}
	if cfg.PublicKey == "" {
		return fmt.Errorf("public_key is required")
	}
	if err := validateBase64Key(cfg.PublicKey, "public_key"); err != nil {
		return err
	}
	if cfg.ExternalHost == "" {
		return fmt.Errorf("external_host is required")
	}
	if _, _, err := net.ParseCIDR(cfg.Address); err != nil {
		return fmt.Errorf("address is not a valid CIDR: %w", err)
	}
	if cfg.ListenPort < 1 || cfg.ListenPort > 65535 {
		return fmt.Errorf("listen_port must be between 1 and 65535")
	}
	if cfg.APIPort < 1 || cfg.APIPort > 65535 {
		return fmt.Errorf("api_port must be between 1 and 65535")
	}
	if cfg.MTU < 576 || cfg.MTU > 65535 {
		return fmt.Errorf("mtu must be between 576 and 65535")
	}
	if cfg.InterfaceName != "" {
		if err := validateInterfaceName(cfg.InterfaceName); err != nil {
			return err
		}
	}
	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return err
	}
	return nil
}

// ValidateClientConfig checks that all required client fields are present and valid.
func ValidateClientConfig(cfg *ClientConfig) error {
	if cfg.PrivateKey == "" {
		return fmt.Errorf("private_key is required")
	}
	if err := validateBase64Key(cfg.PrivateKey, "private_key"); err != nil {
		return err
	}
	if cfg.Server == "" {
		return fmt.Errorf("server is required")
	}
	if cfg.MTU < 576 || cfg.MTU > 65535 {
		return fmt.Errorf("mtu must be between 576 and 65535")
	}
	if cfg.InterfaceName != "" {
		if err := validateInterfaceName(cfg.InterfaceName); err != nil {
			return err
		}
	}
	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return err
	}
	return nil
}

func validateBase64Key(key, field string) error {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("%s is not valid base64: %w", field, err)
	}
	if len(decoded) != 32 {
		return fmt.Errorf("%s must decode to 32 bytes (got %d)", field, len(decoded))
	}
	return nil
}

func validateLogLevel(level string) error {
	switch level {
	case "verbose", "error", "silent":
		return nil
	default:
		return fmt.Errorf("log_level must be one of: verbose, error, silent (got %q)", level)
	}
}

func validateInterfaceName(name string) error {
	if !validIfaceNameRe.MatchString(name) {
		return fmt.Errorf("interface_name %q is invalid: must be 1-15 alphanumeric characters, hyphens, underscores, or dots", name)
	}
	return nil
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
	if cfg.LogLevel == "" {
		cfg.LogLevel = DefaultLogLevel
	}
}

// ApplyClientDefaults fills in zero-value fields with sensible defaults.
func ApplyClientDefaults(cfg *ClientConfig) {
	if cfg.APIPort == 0 {
		cfg.APIPort = DefaultAPIPort
	}
	if cfg.MTU == 0 {
		cfg.MTU = DefaultMTU
	}
	if cfg.PersistentKeepalive == 0 {
		cfg.PersistentKeepalive = DefaultPersistentKeepalive
	}
	if cfg.InterfaceName == "" {
		cfg.InterfaceName = DefaultInterfaceName
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = DefaultLogLevel
	}
}

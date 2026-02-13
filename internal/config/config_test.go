package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

// validKey returns a valid base64-encoded 32-byte key for tests.
func validKey() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestParseServerConfig(t *testing.T) {
	tomlData := `
listen_port = 51820
address = "10.0.0.1/24"
private_key = "sErVerPrIvAtEkEyBase64EnCoDeDhErE="
public_key = "sErVerPuBlIcKeYBase64EnCoDeDhErE=="
api_port = 8080
external_host = "1.2.3.4"
dns_servers = ["1.1.1.1", "8.8.8.8"]
mtu = 1420
`
	cfg, err := ParseServerConfig(tomlData)
	if err != nil {
		t.Fatalf("ParseServerConfig() error: %v", err)
	}

	if cfg.ListenPort != 51820 {
		t.Errorf("ListenPort = %d, want 51820", cfg.ListenPort)
	}
	if cfg.Address != "10.0.0.1/24" {
		t.Errorf("Address = %s, want 10.0.0.1/24", cfg.Address)
	}
	if cfg.PrivateKey != "sErVerPrIvAtEkEyBase64EnCoDeDhErE=" {
		t.Errorf("PrivateKey = %s, unexpected", cfg.PrivateKey)
	}
	if cfg.APIPort != 8080 {
		t.Errorf("APIPort = %d, want 8080", cfg.APIPort)
	}
	if cfg.ExternalHost != "1.2.3.4" {
		t.Errorf("ExternalHost = %s, want 1.2.3.4", cfg.ExternalHost)
	}
	if cfg.MTU != 1420 {
		t.Errorf("MTU = %d, want 1420", cfg.MTU)
	}
	if len(cfg.DNSServers) != 2 {
		t.Errorf("DNSServers length = %d, want 2", len(cfg.DNSServers))
	}
}

func TestParseClientConfig(t *testing.T) {
	tomlData := `
server = "1.2.3.4"
api_port = 9090
server_public_key = "sErVerPuBlIcKeYBase64EnCoDeDhErE=="
private_key = "cLiEnTpRiVaTeKeYBase64EnCoDeDhErE="
mtu = 1400
persistent_keepalive = 30
`
	cfg, err := ParseClientConfig(tomlData)
	if err != nil {
		t.Fatalf("ParseClientConfig() error: %v", err)
	}

	if cfg.Server != "1.2.3.4" {
		t.Errorf("Server = %s, want 1.2.3.4", cfg.Server)
	}
	if cfg.APIPort != 9090 {
		t.Errorf("APIPort = %d, want 9090", cfg.APIPort)
	}
	if cfg.MTU != 1400 {
		t.Errorf("MTU = %d, want 1400", cfg.MTU)
	}
	if cfg.PersistentKeepalive != 30 {
		t.Errorf("PersistentKeepalive = %d, want 30", cfg.PersistentKeepalive)
	}
}

func TestServerConfigDefaults(t *testing.T) {
	cfg, err := ParseServerConfig("")
	if err != nil {
		t.Fatalf("ParseServerConfig() error: %v", err)
	}

	if cfg.ListenPort != DefaultListenPort {
		t.Errorf("default ListenPort = %d, want %d", cfg.ListenPort, DefaultListenPort)
	}
	if cfg.Address != DefaultAddress {
		t.Errorf("default Address = %s, want %s", cfg.Address, DefaultAddress)
	}
	if cfg.APIPort != DefaultAPIPort {
		t.Errorf("default APIPort = %d, want %d", cfg.APIPort, DefaultAPIPort)
	}
	if cfg.MTU != DefaultMTU {
		t.Errorf("default MTU = %d, want %d", cfg.MTU, DefaultMTU)
	}
	if len(cfg.DNSServers) != len(DefaultDNSServers) {
		t.Errorf("default DNSServers length = %d, want %d", len(cfg.DNSServers), len(DefaultDNSServers))
	}
	if cfg.InterfaceName != DefaultInterfaceName {
		t.Errorf("default InterfaceName = %s, want %s", cfg.InterfaceName, DefaultInterfaceName)
	}
	if cfg.LogLevel != DefaultLogLevel {
		t.Errorf("default LogLevel = %s, want %s", cfg.LogLevel, DefaultLogLevel)
	}
}

func TestClientConfigDefaults(t *testing.T) {
	cfg, err := ParseClientConfig("")
	if err != nil {
		t.Fatalf("ParseClientConfig() error: %v", err)
	}

	if cfg.APIPort != DefaultAPIPort {
		t.Errorf("default APIPort = %d, want %d", cfg.APIPort, DefaultAPIPort)
	}
	if cfg.MTU != DefaultMTU {
		t.Errorf("default MTU = %d, want %d", cfg.MTU, DefaultMTU)
	}
	if cfg.PersistentKeepalive != DefaultPersistentKeepalive {
		t.Errorf("default PersistentKeepalive = %d, want %d", cfg.PersistentKeepalive, DefaultPersistentKeepalive)
	}
	if cfg.InterfaceName != DefaultInterfaceName {
		t.Errorf("default InterfaceName = %s, want %s", cfg.InterfaceName, DefaultInterfaceName)
	}
	if cfg.LogLevel != DefaultLogLevel {
		t.Errorf("default LogLevel = %s, want %s", cfg.LogLevel, DefaultLogLevel)
	}
}

func TestInvalidTOMLReturnsError(t *testing.T) {
	_, err := ParseServerConfig("this is not [valid toml")
	if err == nil {
		t.Error("expected error for invalid TOML")
	}

	_, err = ParseClientConfig("this is not [valid toml")
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestValidateServerConfig_Valid(t *testing.T) {
	key := validKey()
	cfg := &ServerConfig{
		PrivateKey:   key,
		PublicKey:    key,
		ExternalHost: "1.2.3.4",
		Address:      "10.0.0.1/24",
		ListenPort:   51820,
		APIPort:      8080,
		MTU:          1420,
		LogLevel:     "error",
	}
	if err := ValidateServerConfig(cfg); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateServerConfig_MissingFields(t *testing.T) {
	key := validKey()
	tests := []struct {
		name string
		cfg  ServerConfig
		want string
	}{
		{
			name: "missing private_key",
			cfg:  ServerConfig{PublicKey: key, ExternalHost: "1.2.3.4", Address: "10.0.0.1/24", ListenPort: 51820, APIPort: 8080, MTU: 1420, LogLevel: "error"},
			want: "private_key is required",
		},
		{
			name: "missing public_key",
			cfg:  ServerConfig{PrivateKey: key, ExternalHost: "1.2.3.4", Address: "10.0.0.1/24", ListenPort: 51820, APIPort: 8080, MTU: 1420, LogLevel: "error"},
			want: "public_key is required",
		},
		{
			name: "missing external_host",
			cfg:  ServerConfig{PrivateKey: key, PublicKey: key, Address: "10.0.0.1/24", ListenPort: 51820, APIPort: 8080, MTU: 1420, LogLevel: "error"},
			want: "external_host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerConfig(&tt.cfg)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestValidateServerConfig_InvalidValues(t *testing.T) {
	key := validKey()
	base := ServerConfig{
		PrivateKey:   key,
		PublicKey:    key,
		ExternalHost: "1.2.3.4",
		Address:      "10.0.0.1/24",
		ListenPort:   51820,
		APIPort:      8080,
		MTU:          1420,
		LogLevel:     "error",
	}

	tests := []struct {
		name   string
		mutate func(c *ServerConfig)
		want   string
	}{
		{
			name:   "bad CIDR",
			mutate: func(c *ServerConfig) { c.Address = "not-a-cidr" },
			want:   "address is not a valid CIDR",
		},
		{
			name:   "port too high",
			mutate: func(c *ServerConfig) { c.ListenPort = 70000 },
			want:   "listen_port must be between",
		},
		{
			name:   "api_port zero",
			mutate: func(c *ServerConfig) { c.APIPort = 0 },
			want:   "api_port must be between",
		},
		{
			name:   "mtu too low",
			mutate: func(c *ServerConfig) { c.MTU = 100 },
			want:   "mtu must be between",
		},
		{
			name:   "bad base64 key",
			mutate: func(c *ServerConfig) { c.PrivateKey = "not-valid-base64!!!" },
			want:   "private_key is not valid base64",
		},
		{
			name:   "key wrong length",
			mutate: func(c *ServerConfig) { c.PrivateKey = base64.StdEncoding.EncodeToString([]byte("short")) },
			want:   "must decode to 32 bytes",
		},
		{
			name:   "bad log level",
			mutate: func(c *ServerConfig) { c.LogLevel = "debug" },
			want:   "log_level must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base // copy
			tt.mutate(&cfg)
			err := ValidateServerConfig(&cfg)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestValidateClientConfig_Valid(t *testing.T) {
	key := validKey()
	cfg := &ClientConfig{
		PrivateKey: key,
		Server:     "1.2.3.4",
		APIPort:    8080,
		MTU:        1420,
		LogLevel:   "error",
	}
	if err := ValidateClientConfig(cfg); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateClientConfig_InvalidValues(t *testing.T) {
	key := validKey()
	base := ClientConfig{
		PrivateKey: key,
		Server:     "1.2.3.4",
		APIPort:    8080,
		MTU:        1420,
		LogLevel:   "error",
	}

	tests := []struct {
		name   string
		mutate func(c *ClientConfig)
		want   string
	}{
		{
			name:   "missing private_key",
			mutate: func(c *ClientConfig) { c.PrivateKey = "" },
			want:   "private_key is required",
		},
		{
			name:   "missing server",
			mutate: func(c *ClientConfig) { c.Server = "" },
			want:   "server is required",
		},
		{
			name:   "mtu too low",
			mutate: func(c *ClientConfig) { c.MTU = 10 },
			want:   "mtu must be between",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base
			tt.mutate(&cfg)
			err := ValidateClientConfig(&cfg)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestParseServerConfigWithAPIKey(t *testing.T) {
	tomlData := `
private_key = "sErVerPrIvAtEkEyBase64EnCoDeDhErE="
public_key = "sErVerPuBlIcKeYBase64EnCoDeDhErE=="
external_host = "1.2.3.4"
api_key = "my-secret-key"
`
	cfg, err := ParseServerConfig(tomlData)
	if err != nil {
		t.Fatalf("ParseServerConfig() error: %v", err)
	}
	if cfg.APIKey != "my-secret-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "my-secret-key")
	}
}

func TestParseClientConfigWithAPIKey(t *testing.T) {
	tomlData := `
server = "1.2.3.4"
private_key = "cLiEnTpRiVaTeKeYBase64EnCoDeDhErE="
api_key = "my-secret-key"
`
	cfg, err := ParseClientConfig(tomlData)
	if err != nil {
		t.Fatalf("ParseClientConfig() error: %v", err)
	}
	if cfg.APIKey != "my-secret-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "my-secret-key")
	}
}

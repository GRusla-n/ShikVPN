package config

import (
	"testing"
)

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
server_endpoint = "1.2.3.4:51820"
server_api_url = "http://1.2.3.4:8080"
server_public_key = "sErVerPuBlIcKeYBase64EnCoDeDhErE=="
private_key = "cLiEnTpRiVaTeKeYBase64EnCoDeDhErE="
mtu = 1400
persistent_keepalive = 30
`
	cfg, err := ParseClientConfig(tomlData)
	if err != nil {
		t.Fatalf("ParseClientConfig() error: %v", err)
	}

	if cfg.ServerEndpoint != "1.2.3.4:51820" {
		t.Errorf("ServerEndpoint = %s, want 1.2.3.4:51820", cfg.ServerEndpoint)
	}
	if cfg.ServerAPIURL != "http://1.2.3.4:8080" {
		t.Errorf("ServerAPIURL = %s, unexpected", cfg.ServerAPIURL)
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
}

func TestClientConfigDefaults(t *testing.T) {
	cfg, err := ParseClientConfig("")
	if err != nil {
		t.Fatalf("ParseClientConfig() error: %v", err)
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

package config

const (
	DefaultListenPort        = 51820
	DefaultAPIPort           = 8080
	DefaultMTU               = 1420
	DefaultAddress           = "10.0.0.1/24"
	DefaultPersistentKeepalive = 25
	DefaultInterfaceName     = "wg0"
)

var DefaultDNSServers = []string{"1.1.1.1", "8.8.8.8"}

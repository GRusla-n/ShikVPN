# ShikVPN

A lightweight, cross-platform VPN built in Go using WireGuard for tunneling. Clients connect by registering with the server over HTTP and automatically receive an IP address and tunnel configuration.

## Quick Start

### Prerequisites

- Go 1.21+
- Root/admin privileges (required for creating TUN devices)
- Linux, macOS, or Windows

### Build

```bash
# Build all three binaries
make build

# Or build individually
go build -o build/vpn-keygen ./cmd/vpn-keygen
go build -o build/vpn-server ./cmd/vpn-server
go build -o build/vpn-client ./cmd/vpn-client
```

Cross-compile for another OS:

```bash
make linux    # Linux amd64
make darwin   # macOS amd64
make windows  # Windows amd64 (wintun.dll embedded)
```

Check version info:

```bash
./build/vpn-server -version
```

## Setup

### 1. Generate Keys

Run `vpn-keygen` twice — once for the server, once for the client:

```bash
./build/vpn-keygen
# Private Key: aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789abcdefg=
# Public Key:  XyZ0123456789abcdefgaBcDeFgHiJkLmNoPqRsTuVw=
```

Save both keypairs. The server needs its own private + public key. Each client needs its own private key.

### 2. Configure the Server

Create `server.toml` (see `deploy/server.toml.example` for a full template):

```toml
listen_port = 51820
address = "10.0.0.1/24"
api_port = 8080
external_host = "YOUR_SERVER_PUBLIC_IP"
private_key = "SERVER_PRIVATE_KEY"
public_key = "SERVER_PUBLIC_KEY"
dns_servers = ["1.1.1.1", "8.8.8.8"]
mtu = 1420
# api_key = "your-secret-key"
```

| Field | Description | Default |
|-------|-------------|---------|
| `listen_port` | WireGuard UDP port | `51820` |
| `address` | Server VPN address and subnet | `10.0.0.1/24` |
| `api_port` | HTTP registration API port | `8080` |
| `external_host` | Public IP/hostname clients connect to | *required* |
| `private_key` | Server private key from vpn-keygen | *required* |
| `public_key` | Server public key from vpn-keygen | *required* |
| `dns_servers` | DNS servers pushed to clients | `["1.1.1.1", "8.8.8.8"]` |
| `mtu` | Tunnel MTU | `1420` |
| `interface_name` | TUN interface name | `wg0` |
| `api_key` | Shared secret for client registration | *(empty = no auth)* |
| `log_level` | WireGuard log verbosity: `verbose`, `error`, `silent` | `error` |

### 3. Configure the Client

Create `client.toml` (see `deploy/client.toml.example` for a full template):

```toml
server = "YOUR_SERVER_PUBLIC_IP"
private_key = "CLIENT_PRIVATE_KEY"
mtu = 1420
persistent_keepalive = 25
# api_port = 8080
# api_key = "your-secret-key"
```

| Field | Description | Default |
|-------|-------------|---------|
| `server` | Server IP or hostname | *required* |
| `private_key` | Client private key from vpn-keygen | *required* |
| `api_port` | Server registration API port | `8080` |
| `mtu` | Tunnel MTU | `1420` |
| `persistent_keepalive` | Keepalive interval in seconds (helps with NAT) | `25` |
| `interface_name` | TUN interface name | `wg0` |
| `api_key` | Must match server's `api_key` if set | *(empty)* |
| `log_level` | WireGuard log verbosity: `verbose`, `error`, `silent` | `error` |

## Running

### Start the Server

```bash
sudo ./build/vpn-server -config server.toml
```

The server will:
- Create a WireGuard tunnel on `wg0`
- Listen for VPN traffic on UDP port 51820
- Listen for client registrations on HTTP port 8080
- Enable IP forwarding and NAT

### Connect a Client

```bash
sudo ./build/vpn-client -config client.toml
```

The client will:
1. Send its public key to the server's registration API
2. Receive an assigned IP address (e.g., `10.0.0.2/24`)
3. Create a WireGuard tunnel and configure routing
4. Route all traffic through the VPN

### Stop

Press `Ctrl+C` to gracefully shut down either the server or client. The client will restore original network routes on disconnect.

## How It Works

```
  Client                          Server
    |                               |
    |  POST /api/v1/register        |
    |  { "public_key": "..." }      |
    |  X-API-Key: <key>             |
    |------------------------------>|
    |                               |  Allocates IP (10.0.0.2)
    |                               |  Adds WireGuard peer
    |  { "assigned_ip": "10.0.0.2", |
    |    "server_public_key": "...",|
    |    "server_endpoint": "...",  |
    |    "dns_servers": [...],      |
    |    "mtu": 1420 }              |
    |<------------------------------|
    |                               |
    |  WireGuard tunnel established |
    |<=============================>|
    |  All traffic routed via VPN   |
```

## Production Deployment

### Linux Server Setup

1. **Install the binary:**

```bash
sudo cp build/vpn-server /usr/local/bin/
sudo mkdir -p /etc/shikvpn
sudo cp deploy/server.toml.example /etc/shikvpn/server.toml
# Edit /etc/shikvpn/server.toml with your keys and settings
```

2. **Generate an API key** (recommended):

```bash
openssl rand -hex 32
# Add the output as api_key in server.toml
```

3. **Install the systemd service:**

```bash
sudo cp deploy/shikvpn-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now shikvpn-server
sudo journalctl -u shikvpn-server -f   # view logs
```

4. **Open firewall ports:**

```bash
# UFW
sudo ufw allow 51820/udp   # WireGuard
sudo ufw allow 8080/tcp    # Registration API

# firewalld
sudo firewall-cmd --permanent --add-port=51820/udp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

5. **TLS for the registration API:**

The registration API speaks plain HTTP. In production, place it behind a reverse proxy (Caddy, nginx) that terminates TLS. This protects the API key in transit.

### Windows Client Setup

`wintun.dll` is embedded in the executables — no separate DLL download is needed.

1. **Create `client.toml`** next to the exe (copy from `deploy/client.toml.example`).

2. **Run as Administrator:**

```powershell
.\vpn-client.exe -config client.toml
```

3. Press `Ctrl+C` to disconnect and restore original routes.

## Firewall Notes

Open these ports on the server:

| Port | Protocol | Purpose |
|------|----------|---------|
| 51820 | UDP | WireGuard tunnel traffic |
| 8080 | TCP | Client registration API |

In production, consider placing the registration API behind HTTPS or restricting access.

## Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires TUN device access)
go test -tags integration ./...

# Full end-to-end test with Docker
make docker-test
```

## License

See [LICENSE](LICENSE) for details.

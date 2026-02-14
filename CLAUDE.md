# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ShikVPN is a lightweight, cross-platform VPN built in Go that uses WireGuard for tunneling. It provides server and client binaries with automatic IP allocation via an HTTP registration API.

## Build & Test Commands

```bash
make build              # Build all binaries (keygen, server, client) into build/
make test               # Run unit tests (go test ./...)
make test-integration   # Run integration tests (requires +build integration tag)
make docker-test        # Cross-compile for Linux and run Docker Compose integration tests
make clean              # Remove built binaries
```

Single test: `go test ./internal/server/ -run TestIPAM`

Cross-compile: `make linux`, `make darwin`, `make windows` (all amd64)

## Architecture

**Three binaries** in `cmd/`:
- `vpn-keygen` — generates WireGuard Curve25519 keypairs
- `vpn-server` — runs the VPN server (flag: `-config server.toml`)
- `vpn-client` — connects to a server (flag: `-config client.toml`)

**Internal packages** (`internal/`):

| Package | Purpose |
|---------|---------|
| `config` | TOML config parsing with defaults (port 51820, 10.0.0.1/24, MTU 1420) |
| `crypto` | Curve25519 key generation, base64/hex encoding |
| `tunnel` | WireGuard device management via UAPI; builds server/client configs, dynamic peer addition |
| `network` | Platform-specific network configuration behind `InterfaceConfigurator` interface — separate implementations in `iface_linux.go`, `iface_darwin.go`, `iface_windows.go` |
| `server` | Server orchestration, IPAM (IP address allocator from subnet), HTTP API (`POST /api/v1/register`) |
| `client` | Client orchestration, server registration, route management with state restoration |

**Client connection flow:**
1. Client sends its public key to `POST /api/v1/register` on the server's HTTP API
2. Server allocates an IP via IPAM, adds peer to WireGuard, returns config (assigned IP, server pubkey, endpoint, DNS, MTU)
3. Client creates TUN device, configures WireGuard, replaces default route

**Key design patterns:**
- Platform abstraction via `InterfaceConfigurator` interface with OS-specific build files
- WireGuard configured through user-space API (IPC/UAPI), not kernel sockets
- IPAM is idempotent — same public key always gets the same IP
- Graceful shutdown restores original network state (routes, NAT rules)
- Mutex-protected concurrent operations (IPAM allocations, tunnel state)

## Dependencies

- Go 1.21+
- `golang.zx2c4.com/wireguard` — WireGuard implementation
- `golang.org/x/crypto` — Curve25519
- `github.com/BurntSushi/toml` — config parsing
- Runtime requires elevated privileges (NET_ADMIN / admin) for TUN device and network configuration

package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

// IPAM manages IP address allocation within a VPN subnet.
type IPAM struct {
	mu        sync.Mutex
	network   *net.IPNet
	gateway   net.IP
	allocated map[string]net.IP // pubkey -> assigned IP
	used      map[string]string // IP string -> pubkey
	nextHost  uint32            // next host number to try (starts at 2)
}

// maxIPAMPrefix is the minimum prefix length allowed (prevents huge iteration).
const maxIPAMPrefix = 16

// NewIPAM creates a new IP allocator for the given CIDR (e.g., "10.0.0.1/24").
func NewIPAM(cidr string) (*IPAM, error) {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	ones, _ := network.Mask.Size()
	if ones < maxIPAMPrefix {
		return nil, fmt.Errorf("subnet /%d is too large; minimum prefix length is /%d", ones, maxIPAMPrefix)
	}

	return &IPAM{
		network:   network,
		gateway:   ip.To4(),
		allocated: make(map[string]net.IP),
		used:      make(map[string]string),
		nextHost:  2, // skip .0 (network) and .1 (gateway)
	}, nil
}

// Allocate assigns an IP address to the given public key.
// If the key already has an allocation, the same IP is returned (idempotent).
func (m *IPAM) Allocate(pubKey string) (net.IP, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ip, ok := m.allocated[pubKey]; ok {
		return ip, nil
	}

	ip, err := m.findAvailable()
	if err != nil {
		return nil, err
	}

	m.allocated[pubKey] = ip
	m.used[ip.String()] = pubKey
	return ip, nil
}

// Release frees the IP allocated to the given public key.
func (m *IPAM) Release(pubKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ip, ok := m.allocated[pubKey]; ok {
		delete(m.used, ip.String())
		delete(m.allocated, pubKey)
	}
}

// GetAllocation returns the IP allocated to the given public key, if any.
func (m *IPAM) GetAllocation(pubKey string) (net.IP, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ip, ok := m.allocated[pubKey]
	return ip, ok
}

func (m *IPAM) findAvailable() (net.IP, error) {
	ones, bits := m.network.Mask.Size()
	hostBits := uint(bits - ones)
	totalHosts := uint32(1) << hostBits // total addresses in subnet
	maxHost := totalHosts - 2           // last usable host number (excludes network + broadcast)

	baseIP := binary.BigEndian.Uint32(m.network.IP.To4())
	gatewayIP := binary.BigEndian.Uint32(m.gateway)

	// Try all usable host numbers (1..maxHost), wrapping from nextHost
	for i := uint32(0); i < maxHost; i++ {
		// Wrap host number within range [1, maxHost]
		hostNum := ((m.nextHost - 1 + i) % maxHost) + 1

		candidateIP := baseIP + hostNum
		candidate := make(net.IP, 4)
		binary.BigEndian.PutUint32(candidate, candidateIP)

		// Skip gateway IP
		if candidateIP == gatewayIP {
			continue
		}

		// Skip if already used
		if _, ok := m.used[candidate.String()]; ok {
			continue
		}

		// Advance nextHost for next allocation
		m.nextHost = hostNum + 1
		if m.nextHost > maxHost {
			m.nextHost = 1
		}

		return candidate, nil
	}

	return nil, fmt.Errorf("no available IP addresses in subnet %s", m.network.String())
}

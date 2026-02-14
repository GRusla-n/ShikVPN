package tunnel

import (
	"strings"
	"testing"

	"github.com/gavsh/ShikVPN/internal/crypto"
)

func TestBuildServerUAPIConfig(t *testing.T) {
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	privKeyHex := crypto.KeyToHex(kp.PrivateKey)

	peer1Kp, _ := crypto.GenerateKeyPair()
	peer2Kp, _ := crypto.GenerateKeyPair()

	peers := []PeerConfig{
		{
			PublicKeyHex: crypto.KeyToHex(peer1Kp.PublicKey),
			AllowedIPs:   []string{"10.0.0.2/32"},
		},
		{
			PublicKeyHex: crypto.KeyToHex(peer2Kp.PublicKey),
			AllowedIPs:   []string{"10.0.0.3/32"},
		},
	}

	result := BuildServerUAPIConfig(privKeyHex, 51820, peers)

	if !strings.Contains(result, "private_key="+privKeyHex) {
		t.Error("UAPI config missing private_key")
	}
	if !strings.Contains(result, "listen_port=51820") {
		t.Error("UAPI config missing listen_port")
	}
	if !strings.Contains(result, "public_key="+crypto.KeyToHex(peer1Kp.PublicKey)) {
		t.Error("UAPI config missing peer 1 public_key")
	}
	if !strings.Contains(result, "public_key="+crypto.KeyToHex(peer2Kp.PublicKey)) {
		t.Error("UAPI config missing peer 2 public_key")
	}
	if !strings.Contains(result, "allowed_ip=10.0.0.2/32") {
		t.Error("UAPI config missing peer 1 allowed_ip")
	}
	if !strings.Contains(result, "allowed_ip=10.0.0.3/32") {
		t.Error("UAPI config missing peer 2 allowed_ip")
	}
}

func TestBuildClientUAPIConfig(t *testing.T) {
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	serverKp, _ := crypto.GenerateKeyPair()

	privKeyHex := crypto.KeyToHex(kp.PrivateKey)
	peer := PeerConfig{
		PublicKeyHex:        crypto.KeyToHex(serverKp.PublicKey),
		Endpoint:            "1.2.3.4:51820",
		AllowedIPs:          []string{"0.0.0.0/0"},
		PersistentKeepalive: 25,
	}

	result := BuildClientUAPIConfig(privKeyHex, peer)

	if !strings.Contains(result, "private_key="+privKeyHex) {
		t.Error("UAPI config missing private_key")
	}
	if !strings.Contains(result, "public_key="+crypto.KeyToHex(serverKp.PublicKey)) {
		t.Error("UAPI config missing peer public_key")
	}
	if !strings.Contains(result, "endpoint=1.2.3.4:51820") {
		t.Error("UAPI config missing endpoint")
	}
	if !strings.Contains(result, "allowed_ip=0.0.0.0/0") {
		t.Error("UAPI config missing allowed_ip")
	}
	if !strings.Contains(result, "persistent_keepalive_interval=25") {
		t.Error("UAPI config missing persistent_keepalive_interval")
	}
}

func TestUAPIContainsHexKeys(t *testing.T) {
	kp, _ := crypto.GenerateKeyPair()
	privKeyHex := crypto.KeyToHex(kp.PrivateKey)

	// Hex-encoded keys should be 64 characters
	if len(privKeyHex) != 64 {
		t.Errorf("hex key length = %d, want 64", len(privKeyHex))
	}

	result := BuildServerUAPIConfig(privKeyHex, 51820, nil)
	// Verify the private key in the config is hex (64 hex chars)
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(line, "private_key=") {
			keyVal := strings.TrimPrefix(line, "private_key=")
			if len(keyVal) != 64 {
				t.Errorf("private_key in UAPI is not proper hex: length %d", len(keyVal))
			}
		}
	}
}

func TestBuildAddPeerUAPI(t *testing.T) {
	kp, _ := crypto.GenerateKeyPair()

	peer := PeerConfig{
		PublicKeyHex:        crypto.KeyToHex(kp.PublicKey),
		AllowedIPs:          []string{"10.0.0.5/32"},
		PersistentKeepalive: 15,
	}

	result := BuildAddPeerUAPI(peer)

	if !strings.Contains(result, "public_key="+crypto.KeyToHex(kp.PublicKey)) {
		t.Error("add peer UAPI missing public_key")
	}
	if !strings.Contains(result, "allowed_ip=10.0.0.5/32") {
		t.Error("add peer UAPI missing allowed_ip")
	}
	if !strings.Contains(result, "persistent_keepalive_interval=15") {
		t.Error("add peer UAPI missing persistent_keepalive_interval")
	}
	// Should NOT contain private_key
	if strings.Contains(result, "private_key=") {
		t.Error("add peer UAPI should not contain private_key")
	}
}

func TestBuildServerUAPIConfigNoPeers(t *testing.T) {
	kp, _ := crypto.GenerateKeyPair()
	privKeyHex := crypto.KeyToHex(kp.PrivateKey)

	result := BuildServerUAPIConfig(privKeyHex, 51820, nil)

	if !strings.Contains(result, "private_key=") {
		t.Error("UAPI config without peers should still have private_key")
	}
	if !strings.Contains(result, "listen_port=51820") {
		t.Error("UAPI config without peers should still have listen_port")
	}
	if strings.Contains(result, "public_key=") {
		t.Error("UAPI config with no peers should not contain public_key")
	}
}

func TestMultiplePeersSerialized(t *testing.T) {
	kp, _ := crypto.GenerateKeyPair()
	privKeyHex := crypto.KeyToHex(kp.PrivateKey)

	var peers []PeerConfig
	for i := 0; i < 5; i++ {
		peerKp, _ := crypto.GenerateKeyPair()
		peers = append(peers, PeerConfig{
			PublicKeyHex: crypto.KeyToHex(peerKp.PublicKey),
			AllowedIPs:   []string{strings.Replace("10.0.0.X/32", "X", string(rune('2'+i)), 1)},
		})
	}

	result := BuildServerUAPIConfig(privKeyHex, 51820, peers)

	// Count public_key occurrences — should be exactly 5
	count := strings.Count(result, "public_key=")
	if count != 5 {
		t.Errorf("expected 5 public_key entries, got %d", count)
	}
}

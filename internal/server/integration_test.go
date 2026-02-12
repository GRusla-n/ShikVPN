//go:build integration

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavsh/simplevpn/internal/config"
	"github.com/gavsh/simplevpn/internal/crypto"
	"github.com/gavsh/simplevpn/internal/tunnel"
)

func TestIntegrationRegisterAndVerifyPeer(t *testing.T) {
	// Generate server keys
	serverKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate server keys: %v", err)
	}

	cfg := &config.ServerConfig{
		ListenPort:    51821, // Use non-standard port for testing
		Address:       "10.99.0.1/24",
		PrivateKey:    crypto.KeyToBase64(serverKP.PrivateKey),
		PublicKey:     crypto.KeyToBase64(serverKP.PublicKey),
		APIPort:       0, // Not used directly
		ExternalHost:  "127.0.0.1",
		DNSServers:    []string{"1.1.1.1"},
		MTU:           1420,
		InterfaceName: "wgtest0",
	}

	srv := New(cfg)
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Stop()

	// Create test server for API
	ts := httptest.NewServer(srv.api.Handler())
	defer ts.Close()

	// Register a client
	clientKP, _ := crypto.GenerateKeyPair()
	reqBody, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(clientKP.PublicKey),
	})

	resp, err := http.Post(ts.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("registration returned %d", resp.StatusCode)
	}

	var regResp RegisterResponse
	json.NewDecoder(resp.Body).Decode(&regResp)

	if regResp.AssignedIP == "" {
		t.Fatal("no IP assigned")
	}

	t.Logf("Client registered with IP: %s", regResp.AssignedIP)
}

func TestIntegrationTwoClientsUniqueIPs(t *testing.T) {
	serverKP, _ := crypto.GenerateKeyPair()

	cfg := &config.ServerConfig{
		ListenPort:    51822,
		Address:       "10.99.1.1/24",
		PrivateKey:    crypto.KeyToBase64(serverKP.PrivateKey),
		PublicKey:     crypto.KeyToBase64(serverKP.PublicKey),
		ExternalHost:  "127.0.0.1",
		DNSServers:    []string{"1.1.1.1"},
		MTU:           1420,
		InterfaceName: "wgtest1",
	}

	srv := New(cfg)
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Stop()

	ts := httptest.NewServer(srv.api.Handler())
	defer ts.Close()

	var ips []string
	for i := 0; i < 2; i++ {
		clientKP, _ := crypto.GenerateKeyPair()
		reqBody, _ := json.Marshal(RegisterRequest{
			PublicKey: crypto.KeyToBase64(clientKP.PublicKey),
		})

		resp, err := http.Post(ts.URL+"/api/v1/register", "application/json",
			bytes.NewReader(reqBody))
		if err != nil {
			t.Fatalf("registration %d failed: %v", i, err)
		}

		var regResp RegisterResponse
		json.NewDecoder(resp.Body).Decode(&regResp)
		resp.Body.Close()

		ips = append(ips, regResp.AssignedIP)
	}

	if ips[0] == ips[1] {
		t.Errorf("two clients got same IP: %s", ips[0])
	}
	fmt.Printf("Client 1: %s, Client 2: %s\n", ips[0], ips[1])
}

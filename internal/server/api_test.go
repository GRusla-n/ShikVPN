package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gavsh/simplevpn/internal/crypto"
	"github.com/gavsh/simplevpn/internal/tunnel"
)

func setupTestAPI(t *testing.T) (*API, *httptest.Server) {
	t.Helper()
	return setupTestAPIWithKey(t, "")
}

func setupTestAPIWithKey(t *testing.T, apiKey string) (*API, *httptest.Server) {
	t.Helper()

	ipam, err := NewIPAM("10.0.0.1/24")
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	// Generate a server keypair for testing
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	noop := func(peer tunnel.PeerConfig) error { return nil }

	api := NewAPI(ipam, crypto.KeyToBase64(kp.PublicKey), "1.2.3.4:51820",
		[]string{"1.1.1.1"}, 1420, apiKey, noop)

	server := httptest.NewServer(api.Handler())
	return api, server
}

func TestRegisterValidRequest(t *testing.T) {
	_, server := setupTestAPI(t)
	defer server.Close()

	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	reqBody, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp.PublicKey),
	})

	resp, err := http.Post(server.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var regResp RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		t.Fatalf("decode response error: %v", err)
	}

	if regResp.AssignedIP == "" {
		t.Error("AssignedIP is empty")
	}
	if !strings.HasPrefix(regResp.AssignedIP, "10.0.0.") {
		t.Errorf("AssignedIP %s not in expected subnet", regResp.AssignedIP)
	}
	if regResp.ServerPublicKey == "" {
		t.Error("ServerPublicKey is empty")
	}
	if regResp.ServerEndpoint != "1.2.3.4:51820" {
		t.Errorf("ServerEndpoint = %s, want 1.2.3.4:51820", regResp.ServerEndpoint)
	}
}

func TestRegisterMissingPubkey(t *testing.T) {
	_, server := setupTestAPI(t)
	defer server.Close()

	reqBody, _ := json.Marshal(RegisterRequest{})

	resp, err := http.Post(server.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestRegisterMethodNotAllowed(t *testing.T) {
	_, server := setupTestAPI(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/register")
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestRegisterTwoDifferentClients(t *testing.T) {
	_, server := setupTestAPI(t)
	defer server.Close()

	// Register first client
	kp1, _ := crypto.GenerateKeyPair()
	reqBody1, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp1.PublicKey),
	})
	resp1, err := http.Post(server.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody1))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp1.Body.Close()

	var regResp1 RegisterResponse
	json.NewDecoder(resp1.Body).Decode(&regResp1)

	// Register second client
	kp2, _ := crypto.GenerateKeyPair()
	reqBody2, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp2.PublicKey),
	})
	resp2, err := http.Post(server.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody2))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp2.Body.Close()

	var regResp2 RegisterResponse
	json.NewDecoder(resp2.Body).Decode(&regResp2)

	if regResp1.AssignedIP == regResp2.AssignedIP {
		t.Errorf("two clients got same IP: %s", regResp1.AssignedIP)
	}
}

func TestRegisterReturnsIPInSubnet(t *testing.T) {
	_, server := setupTestAPI(t)
	defer server.Close()

	kp, _ := crypto.GenerateKeyPair()
	reqBody, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp.PublicKey),
	})
	resp, err := http.Post(server.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()

	var regResp RegisterResponse
	json.NewDecoder(resp.Body).Decode(&regResp)

	// Should be 10.0.0.X/24
	if !strings.HasPrefix(regResp.AssignedIP, "10.0.0.") {
		t.Errorf("AssignedIP %s not in 10.0.0.0/24 subnet", regResp.AssignedIP)
	}
	if !strings.HasSuffix(regResp.AssignedIP, "/24") {
		t.Errorf("AssignedIP %s should have /24 suffix", regResp.AssignedIP)
	}
}

func TestRegisterWithAPIKeyValid(t *testing.T) {
	_, server := setupTestAPIWithKey(t, "test-secret-key")
	defer server.Close()

	kp, _ := crypto.GenerateKeyPair()
	reqBody, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp.PublicKey),
	})

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-secret-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRegisterWithAPIKeyMissing(t *testing.T) {
	_, server := setupTestAPIWithKey(t, "test-secret-key")
	defer server.Close()

	kp, _ := crypto.GenerateKeyPair()
	reqBody, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp.PublicKey),
	})

	// No X-API-Key header
	resp, err := http.Post(server.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRegisterWithAPIKeyWrong(t *testing.T) {
	_, server := setupTestAPIWithKey(t, "test-secret-key")
	defer server.Close()

	kp, _ := crypto.GenerateKeyPair()
	reqBody, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp.PublicKey),
	})

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "wrong-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRegisterNoAPIKeyConfigured(t *testing.T) {
	// When no API key is configured, requests without a key should succeed
	_, server := setupTestAPI(t) // no API key
	defer server.Close()

	kp, _ := crypto.GenerateKeyPair()
	reqBody, _ := json.Marshal(RegisterRequest{
		PublicKey: crypto.KeyToBase64(kp.PublicKey),
	})

	resp, err := http.Post(server.URL+"/api/v1/register", "application/json",
		bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (no auth configured)", resp.StatusCode)
	}
}

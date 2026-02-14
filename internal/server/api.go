package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gavsh/ShikVPN/internal/crypto"
	"github.com/gavsh/ShikVPN/internal/tunnel"
)

// maxRequestBodySize limits registration request bodies to 4KB.
const maxRequestBodySize = 4096

// RegisterRequest is the JSON body for client registration.
type RegisterRequest struct {
	PublicKey string `json:"public_key"`
}

// RegisterResponse is returned to the client after successful registration.
type RegisterResponse struct {
	AssignedIP      string   `json:"assigned_ip"`
	ServerPublicKey string   `json:"server_public_key"`
	ServerEndpoint  string   `json:"server_endpoint"`
	DNSServers      []string `json:"dns_servers"`
	MTU             int      `json:"mtu"`
}

// PeerAddFunc is called when a new peer needs to be added to the WireGuard device.
type PeerAddFunc func(peer tunnel.PeerConfig) error

// API handles the HTTP registration endpoint.
type API struct {
	ipam            *IPAM
	serverPublicKey string
	serverEndpoint  string
	dnsServers      []string
	mtu             int
	apiKey          string
	onPeerAdd       PeerAddFunc
	mux             *http.ServeMux
	server          *http.Server
}

// NewAPI creates a new registration API handler.
func NewAPI(ipam *IPAM, serverPubKey, serverEndpoint string, dnsServers []string, mtu int, apiKey string, onPeerAdd PeerAddFunc) *API {
	api := &API{
		ipam:            ipam,
		serverPublicKey: serverPubKey,
		serverEndpoint:  serverEndpoint,
		dnsServers:      dnsServers,
		mtu:             mtu,
		apiKey:          apiKey,
		onPeerAdd:       onPeerAdd,
		mux:             http.NewServeMux(),
	}
	api.mux.HandleFunc("/api/v1/register", api.handleRegister)
	return api
}

// Handler returns the HTTP handler for the API.
func (a *API) Handler() http.Handler {
	return a.mux
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check API key if configured
	if a.apiKey != "" {
		provided := r.Header.Get("X-API-Key")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(a.apiKey)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Limit request body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PublicKey == "" {
		http.Error(w, "public_key is required", http.StatusBadRequest)
		return
	}

	// Validate the key is valid base64
	if _, err := crypto.KeyFromBase64(req.PublicKey); err != nil {
		log.Printf("Invalid public_key from client: %v", err)
		http.Error(w, "invalid public_key format", http.StatusBadRequest)
		return
	}

	// Allocate an IP for this peer
	assignedIP, err := a.ipam.Allocate(req.PublicKey)
	if err != nil {
		log.Printf("IPAM allocation failed: %v", err)
		http.Error(w, "failed to allocate IP address", http.StatusInternalServerError)
		return
	}

	// Convert pubkey to hex for WireGuard UAPI
	pubKeyHex, err := crypto.Base64ToHex(req.PublicKey)
	if err != nil {
		http.Error(w, "invalid public key encoding", http.StatusBadRequest)
		return
	}

	// Add peer to WireGuard device
	peer := tunnel.PeerConfig{
		PublicKeyHex: pubKeyHex,
		AllowedIPs:   []string{assignedIP.String() + "/32"},
	}

	if err := a.onPeerAdd(peer); err != nil {
		log.Printf("Failed to add peer: %v", err)
		a.ipam.Release(req.PublicKey)
		http.Error(w, "failed to configure peer", http.StatusInternalServerError)
		return
	}

	truncKey := req.PublicKey
	if len(truncKey) > 8 {
		truncKey = truncKey[:8]
	}
	log.Printf("Registered peer %s... with IP %s", truncKey, assignedIP.String())

	resp := RegisterResponse{
		AssignedIP:      assignedIP.String() + "/24",
		ServerPublicKey: a.serverPublicKey,
		ServerEndpoint:  a.serverEndpoint,
		DNSServers:      a.dnsServers,
		MTU:             a.mtu,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListenAndServe starts the API server.
func (a *API) ListenAndServe(addr string) error {
	if a.apiKey == "" {
		log.Println("WARNING: API server starting without authentication. Set api_key in config to require auth.")
	}
	a.server = &http.Server{
		Addr:              addr,
		Handler:           a.mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("API server listening on %s", addr)
	return a.server.ListenAndServe()
}

// Shutdown gracefully stops the API server.
func (a *API) Shutdown(timeout time.Duration) error {
	if a.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return a.server.Shutdown(ctx)
}

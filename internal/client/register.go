package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gavsh/ShikVPN/internal/server"
)

// retryDelays defines the backoff between registration attempts.
var retryDelays = []time.Duration{0, 2 * time.Second, 5 * time.Second}

// Register sends a registration request to the VPN server API with retry.
func Register(apiURL string, publicKey string, apiKey string) (*server.RegisterResponse, error) {
	reqBody := server.RegisterRequest{
		PublicKey: publicKey,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := apiURL + "/api/v1/register"
	client := &http.Client{Timeout: 10 * time.Second}

	var lastErr error
	for attempt, delay := range retryDelays {
		if delay > 0 {
			log.Printf("Retrying registration (attempt %d/%d) in %v...", attempt+1, len(retryDelays), delay)
			time.Sleep(delay)
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("registration request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("registration failed (HTTP %d): %s", resp.StatusCode, string(respBody))
			// Don't retry on auth errors
			if resp.StatusCode == http.StatusUnauthorized {
				return nil, lastErr
			}
			continue
		}

		var regResp server.RegisterResponse
		if err := json.Unmarshal(respBody, &regResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &regResp, nil
	}

	return nil, fmt.Errorf("registration failed after %d attempts: %w", len(retryDelays), lastErr)
}

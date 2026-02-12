package server

import (
	"fmt"
	"sync"
	"testing"
)

func TestIPAMAllocateFirst(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.1/24")
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	ip, err := ipam.Allocate("pubkey1")
	if err != nil {
		t.Fatalf("Allocate() error: %v", err)
	}

	if ip.String() != "10.0.0.2" {
		t.Errorf("first allocation = %s, want 10.0.0.2", ip.String())
	}
}

func TestIPAMAllocateSequential(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.1/24")
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	expected := []string{"10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5"}
	for i, want := range expected {
		ip, err := ipam.Allocate(fmt.Sprintf("pubkey%d", i))
		if err != nil {
			t.Fatalf("Allocate(%d) error: %v", i, err)
		}
		if ip.String() != want {
			t.Errorf("allocation %d = %s, want %s", i, ip.String(), want)
		}
	}
}

func TestIPAMIdempotent(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.1/24")
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	ip1, err := ipam.Allocate("pubkey1")
	if err != nil {
		t.Fatalf("Allocate() error: %v", err)
	}

	ip2, err := ipam.Allocate("pubkey1")
	if err != nil {
		t.Fatalf("Allocate() error: %v", err)
	}

	if ip1.String() != ip2.String() {
		t.Errorf("idempotent allocation failed: %s != %s", ip1.String(), ip2.String())
	}
}

func TestIPAMRelease(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.1/24")
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	ip1, err := ipam.Allocate("pubkey1")
	if err != nil {
		t.Fatalf("Allocate() error: %v", err)
	}

	ipam.Release("pubkey1")

	// After release, the IP should be available again
	ip2, err := ipam.Allocate("pubkey2")
	if err != nil {
		t.Fatalf("Allocate() error: %v", err)
	}

	// New allocation should succeed (may or may not reuse the same IP depending on implementation)
	if ip2 == nil {
		t.Error("allocation after release returned nil")
	}
	_ = ip1
}

func TestIPAMExhaust(t *testing.T) {
	// Use a /29 subnet: 10.0.0.0/29 has IPs .0-.7
	// .0 = network, .1 = gateway, .7 = broadcast → 5 usable (.2-.6)
	ipam, err := NewIPAM("10.0.0.1/29")
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	// Allocate all usable IPs
	var lastErr error
	allocated := 0
	for i := 0; i < 10; i++ {
		_, err := ipam.Allocate(fmt.Sprintf("pubkey%d", i))
		if err != nil {
			lastErr = err
			break
		}
		allocated++
	}

	if lastErr == nil {
		t.Error("expected error when exhausting subnet, got none")
	}
	if allocated < 1 {
		t.Error("should have allocated at least one IP before exhaustion")
	}
}

func TestIPAMConcurrent(t *testing.T) {
	ipam, err := NewIPAM("10.0.0.1/24")
	if err != nil {
		t.Fatalf("NewIPAM() error: %v", err)
	}

	const numGoroutines = 50
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ip, err := ipam.Allocate(fmt.Sprintf("pubkey%d", idx))
			if err != nil {
				errors[idx] = err
				return
			}
			results[idx] = ip.String()
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("goroutine %d error: %v", i, err)
		}
	}

	// Check all IPs are unique
	seen := make(map[string]bool)
	for i, ip := range results {
		if ip == "" {
			continue // error case already reported
		}
		if seen[ip] {
			t.Errorf("duplicate IP %s in goroutine %d", ip, i)
		}
		seen[ip] = true
	}
}

package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

const KeySize = 32

// KeyPair holds a WireGuard Curve25519 keypair.
type KeyPair struct {
	PrivateKey [KeySize]byte
	PublicKey  [KeySize]byte
}

// GenerateKeyPair generates a new WireGuard Curve25519 keypair with proper clamping.
func GenerateKeyPair() (*KeyPair, error) {
	var privateKey [KeySize]byte
	if _, err := rand.Read(privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	ClampPrivateKey(&privateKey)

	publicKey, err := curve25519.X25519(privateKey[:], curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to compute public key: %w", err)
	}

	var pubKey [KeySize]byte
	copy(pubKey[:], publicKey)

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  pubKey,
	}, nil
}

// ClampPrivateKey applies WireGuard clamping to a private key.
func ClampPrivateKey(key *[KeySize]byte) {
	key[0] &= 248  // Clear bits 0, 1, 2
	key[31] &= 127 // Clear bit 7
	key[31] |= 64  // Set bit 6
}

// PublicKeyFromPrivate derives the public key from a private key.
func PublicKeyFromPrivate(privateKey [KeySize]byte) ([KeySize]byte, error) {
	pub, err := curve25519.X25519(privateKey[:], curve25519.Basepoint)
	if err != nil {
		return [KeySize]byte{}, fmt.Errorf("failed to compute public key: %w", err)
	}
	var pubKey [KeySize]byte
	copy(pubKey[:], pub)
	return pubKey, nil
}

// KeyToBase64 encodes a key as base64.
func KeyToBase64(key [KeySize]byte) string {
	return base64.StdEncoding.EncodeToString(key[:])
}

// KeyFromBase64 decodes a base64-encoded key.
func KeyFromBase64(s string) ([KeySize]byte, error) {
	var key [KeySize]byte
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return key, fmt.Errorf("invalid base64 key: %w", err)
	}
	if len(b) != KeySize {
		return key, fmt.Errorf("invalid key length: got %d, want %d", len(b), KeySize)
	}
	copy(key[:], b)
	return key, nil
}

// KeyToHex encodes a key as hexadecimal (used in WireGuard UAPI).
func KeyToHex(key [KeySize]byte) string {
	return hex.EncodeToString(key[:])
}

// KeyFromHex decodes a hex-encoded key.
func KeyFromHex(s string) ([KeySize]byte, error) {
	var key [KeySize]byte
	b, err := hex.DecodeString(s)
	if err != nil {
		return key, fmt.Errorf("invalid hex key: %w", err)
	}
	if len(b) != KeySize {
		return key, fmt.Errorf("invalid key length: got %d, want %d", len(b), KeySize)
	}
	copy(key[:], b)
	return key, nil
}

// Base64ToHex converts a base64-encoded key to hex encoding.
func Base64ToHex(b64 string) (string, error) {
	key, err := KeyFromBase64(b64)
	if err != nil {
		return "", err
	}
	return KeyToHex(key), nil
}

// HexToBase64 converts a hex-encoded key to base64 encoding.
func HexToBase64(hexStr string) (string, error) {
	key, err := KeyFromHex(hexStr)
	if err != nil {
		return "", err
	}
	return KeyToBase64(key), nil
}

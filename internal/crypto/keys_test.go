package crypto

import (
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	if len(kp.PrivateKey) != KeySize {
		t.Errorf("private key length = %d, want %d", len(kp.PrivateKey), KeySize)
	}
	if len(kp.PublicKey) != KeySize {
		t.Errorf("public key length = %d, want %d", len(kp.PublicKey), KeySize)
	}

	// Verify not all zeros
	allZero := true
	for _, b := range kp.PrivateKey {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("private key is all zeros")
	}

	allZero = true
	for _, b := range kp.PublicKey {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("public key is all zeros")
	}
}

func TestClampPrivateKey(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	key := kp.PrivateKey

	// Bits 0, 1, 2 of byte[0] should be clear
	if key[0]&0x07 != 0 {
		t.Errorf("bits 0,1,2 of byte[0] not cleared: %08b", key[0])
	}

	// Bit 7 of byte[31] should be clear
	if key[31]&0x80 != 0 {
		t.Errorf("bit 7 of byte[31] not cleared: %08b", key[31])
	}

	// Bit 6 of byte[31] should be set
	if key[31]&0x40 == 0 {
		t.Errorf("bit 6 of byte[31] not set: %08b", key[31])
	}
}

func TestBase64HexRoundTrip(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	// Base64 round-trip
	b64 := KeyToBase64(kp.PrivateKey)
	decoded, err := KeyFromBase64(b64)
	if err != nil {
		t.Fatalf("KeyFromBase64() error: %v", err)
	}
	if decoded != kp.PrivateKey {
		t.Error("base64 round-trip failed: keys don't match")
	}

	// Hex round-trip
	hexStr := KeyToHex(kp.PublicKey)
	decodedHex, err := KeyFromHex(hexStr)
	if err != nil {
		t.Fatalf("KeyFromHex() error: %v", err)
	}
	if decodedHex != kp.PublicKey {
		t.Error("hex round-trip failed: keys don't match")
	}

	// Base64 → Hex → Base64
	hexFromB64, err := Base64ToHex(b64)
	if err != nil {
		t.Fatalf("Base64ToHex() error: %v", err)
	}
	b64FromHex, err := HexToBase64(hexFromB64)
	if err != nil {
		t.Fatalf("HexToBase64() error: %v", err)
	}
	if b64FromHex != b64 {
		t.Errorf("base64→hex→base64 round-trip failed: got %s, want %s", b64FromHex, b64)
	}
}

func TestUniqueKeyPairs(t *testing.T) {
	kp1, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}
	kp2, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	if kp1.PrivateKey == kp2.PrivateKey {
		t.Error("two generated private keys are identical")
	}
	if kp1.PublicKey == kp2.PublicKey {
		t.Error("two generated public keys are identical")
	}
}

func TestPublicKeyDeterministic(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	pub1, err := PublicKeyFromPrivate(kp.PrivateKey)
	if err != nil {
		t.Fatalf("PublicKeyFromPrivate() error: %v", err)
	}
	pub2, err := PublicKeyFromPrivate(kp.PrivateKey)
	if err != nil {
		t.Fatalf("PublicKeyFromPrivate() error: %v", err)
	}

	if pub1 != pub2 {
		t.Error("same private key produced different public keys")
	}
	if pub1 != kp.PublicKey {
		t.Error("PublicKeyFromPrivate result doesn't match GenerateKeyPair public key")
	}
}

func TestKeyFromBase64Invalid(t *testing.T) {
	_, err := KeyFromBase64("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	// Wrong length
	_, err = KeyFromBase64("AAAA")
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

func TestKeyFromHexInvalid(t *testing.T) {
	_, err := KeyFromHex("not-hex")
	if err == nil {
		t.Error("expected error for invalid hex")
	}

	// Wrong length
	_, err = KeyFromHex("aabb")
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

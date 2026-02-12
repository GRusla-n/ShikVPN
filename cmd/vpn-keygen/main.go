package main

import (
	"fmt"
	"os"

	"github.com/gavsh/simplevpn/internal/crypto"
)

func main() {
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating keypair: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Private Key: %s\n", crypto.KeyToBase64(kp.PrivateKey))
	fmt.Printf("Public Key:  %s\n", crypto.KeyToBase64(kp.PublicKey))
}

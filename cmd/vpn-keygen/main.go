package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gavsh/ShikVPN/internal/crypto"
	"github.com/gavsh/ShikVPN/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating keypair: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Private Key: %s\n", crypto.KeyToBase64(kp.PrivateKey))
	fmt.Printf("Public Key:  %s\n", crypto.KeyToBase64(kp.PublicKey))
}

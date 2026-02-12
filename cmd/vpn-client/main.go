package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gavsh/simplevpn/internal/client"
	"github.com/gavsh/simplevpn/internal/config"
)

func main() {
	configPath := flag.String("config", "client.toml", "path to client config file")
	flag.Parse()

	cfg, err := config.LoadClientConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.PrivateKey == "" {
		fmt.Fprintf(os.Stderr, "Error: private_key is required in config\n")
		os.Exit(1)
	}
	if cfg.ServerAPIURL == "" {
		fmt.Fprintf(os.Stderr, "Error: server_api_url is required in config\n")
		os.Exit(1)
	}

	vpnClient := client.New(cfg)

	if err := vpnClient.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("Received signal %v, disconnecting...", sig)

	vpnClient.Disconnect()
}

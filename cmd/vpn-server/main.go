package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gavsh/simplevpn/internal/config"
	"github.com/gavsh/simplevpn/internal/server"
)

func main() {
	configPath := flag.String("config", "server.toml", "path to server config file")
	flag.Parse()

	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.PrivateKey == "" {
		fmt.Fprintf(os.Stderr, "Error: private_key is required in config\n")
		os.Exit(1)
	}
	if cfg.PublicKey == "" {
		fmt.Fprintf(os.Stderr, "Error: public_key is required in config\n")
		os.Exit(1)
	}
	if cfg.ExternalHost == "" {
		fmt.Fprintf(os.Stderr, "Error: external_host is required in config\n")
		os.Exit(1)
	}

	srv := server.New(cfg)

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down...", sig)

	srv.Stop()
}

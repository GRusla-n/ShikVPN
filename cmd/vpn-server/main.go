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
	"github.com/gavsh/simplevpn/internal/version"
	"github.com/gavsh/simplevpn/internal/wintun"
)

func main() {
	if err := wintun.Extract(); err != nil {
		log.Printf("Warning: failed to extract wintun.dll: %v", err)
	}

	configPath := flag.String("config", "server.toml", "path to server config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := config.ValidateServerConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
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

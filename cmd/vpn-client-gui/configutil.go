package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gavsh/ShikVPN/internal/config"
)

// writeConfigToml serializes a ClientConfig to a TOML file.
// Uses restrictive permissions (0600) since config files contain private keys.
func writeConfigToml(path string, cfg *config.ClientConfig) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	return nil
}

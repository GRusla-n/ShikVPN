package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gavsh/ShikVPN/internal/config"
)

// writeConfigToml serializes a ClientConfig to a TOML file.
func writeConfigToml(path string, cfg *config.ClientConfig) error {
	f, err := os.Create(path)
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

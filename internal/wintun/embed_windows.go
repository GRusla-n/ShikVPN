//go:build windows

package wintun

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed wintun.dll
var wintunDLL []byte

// Extract writes the embedded wintun.dll next to the running executable.
// Safe to call multiple times — overwrites to ensure version consistency.
func Extract() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("wintun: cannot find executable path: %w", err)
	}
	dllPath := filepath.Join(filepath.Dir(exe), "wintun.dll")
	return os.WriteFile(dllPath, wintunDLL, 0644)
}

//go:build !windows

package wintun

// Extract is a no-op on non-Windows platforms.
func Extract() error { return nil }

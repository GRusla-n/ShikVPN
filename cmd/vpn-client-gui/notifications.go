//go:build windows

package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/toast.v1"
)

func sendNotification(title, message string) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	iconPath := filepath.Join(filepath.Dir(exe), "icon.ico")

	notification := toast.Notification{
		AppID:   "ShikVPN",
		Title:   title,
		Message: message,
	}
	// Only set icon if it exists on disk
	if _, err := os.Stat(iconPath); err == nil {
		notification.Icon = iconPath
	}

	if err := notification.Push(); err != nil {
		// Use stderr directly to avoid recursive log calls
		log.New(defaultStderr, "", 0).Printf("notification error: %v", err)
	}
}

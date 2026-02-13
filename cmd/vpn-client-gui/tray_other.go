//go:build !windows

package main

func initTray(app *App)          {}
func updateTrayStatus(bool)      {}
func cleanupTray()               {}

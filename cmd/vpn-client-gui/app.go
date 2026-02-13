package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gavsh/simplevpn/internal/client"
	"github.com/gavsh/simplevpn/internal/config"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StatusUpdate is emitted to the frontend via events.
type StatusUpdate struct {
	Status     string `json:"status"`
	AssignedIP string `json:"assignedIP"`
	Error      string `json:"error"`
}

// App struct holds the GUI application state.
type App struct {
	ctx        context.Context
	mu         sync.Mutex
	cfg        *config.ClientConfig
	cfgPath    string
	vpnClient  *client.Client
	status     string
	assignedIP string
	hidden     bool
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		status: "disconnected",
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Redirect log output to capture for the GUI
	logWriter := &EventLogWriter{ctx: ctx}
	log.SetOutput(logWriter)

	// Try to load a default config
	exe, err := os.Executable()
	if err == nil {
		defaultPath := filepath.Join(filepath.Dir(exe), "client.toml")
		if _, err := os.Stat(defaultPath); err == nil {
			a.loadConfigFromPath(defaultPath)
		}
	}

	initTray(a)
}

func (a *App) beforeClose(ctx context.Context) bool {
	// Hide to tray instead of quitting
	if !a.hidden {
		a.hidden = true
		runtime.WindowHide(ctx)
		return true // prevent close
	}
	return false
}

func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vpnClient != nil && a.status == "connected" {
		a.vpnClient.Disconnect()
	}
	cleanupTray()
}

// ShowWindow brings the window back from the tray.
func (a *App) ShowWindow() {
	a.hidden = false
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
}

// Quit exits the application.
func (a *App) Quit() {
	a.hidden = true // allow close to proceed
	a.shutdown(a.ctx)
	runtime.Quit(a.ctx)
}

// Connect starts the VPN connection asynchronously.
func (a *App) Connect() {
	a.mu.Lock()
	if a.status == "connecting" || a.status == "connected" {
		a.mu.Unlock()
		return
	}
	cfg := a.cfg
	a.mu.Unlock()

	if cfg == nil {
		a.emitStatus("error", "", "No configuration loaded. Please configure first.")
		return
	}

	if err := config.ValidateClientConfig(cfg); err != nil {
		a.emitStatus("error", "", fmt.Sprintf("Config error: %v", err))
		return
	}

	a.emitStatus("connecting", "", "")

	go a.connectAsync(cfg)
}

func (a *App) connectAsync(cfg *config.ClientConfig) {
	// Create a copy of the config so mutations during connect don't affect the saved one
	cfgCopy := *cfg
	vpnClient := client.New(&cfgCopy)

	err := vpnClient.Connect()
	a.mu.Lock()
	defer a.mu.Unlock()

	if err != nil {
		a.status = "error"
		a.emitStatusLocked("error", "", fmt.Sprintf("Connection failed: %v", err))
		sendNotification("SimpleVPN", fmt.Sprintf("Connection failed: %v", err))
		return
	}

	a.vpnClient = vpnClient
	a.assignedIP = cfgCopy.Address
	a.status = "connected"
	a.emitStatusLocked("connected", cfgCopy.Address, "")
	updateTrayStatus(true)
	sendNotification("SimpleVPN", fmt.Sprintf("Connected - %s", cfgCopy.Address))
}

// Disconnect tears down the VPN connection.
func (a *App) Disconnect() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.vpnClient == nil {
		return
	}

	a.vpnClient.Disconnect()
	a.vpnClient = nil
	a.status = "disconnected"
	a.assignedIP = ""
	a.emitStatusLocked("disconnected", "", "")
	updateTrayStatus(false)
	sendNotification("SimpleVPN", "Disconnected")
}

// GetStatus returns the current VPN status.
func (a *App) GetStatus() StatusUpdate {
	a.mu.Lock()
	defer a.mu.Unlock()
	return StatusUpdate{
		Status:     a.status,
		AssignedIP: a.assignedIP,
	}
}

// GetConfig returns the current client configuration or defaults.
func (a *App) GetConfig() *config.ClientConfig {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cfg != nil {
		return a.cfg
	}
	cfg := &config.ClientConfig{}
	config.ApplyClientDefaults(cfg)
	return cfg
}

// SaveConfig validates and saves the given config and writes it to the current config path.
func (a *App) SaveConfig(cfg config.ClientConfig) error {
	config.ApplyClientDefaults(&cfg)

	a.mu.Lock()
	a.cfg = &cfg
	path := a.cfgPath
	a.mu.Unlock()

	if path == "" {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot determine config path: %w", err)
		}
		path = filepath.Join(filepath.Dir(exe), "client.toml")
	}

	if err := writeConfigToml(path, &cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	a.mu.Lock()
	a.cfgPath = path
	a.mu.Unlock()

	log.Printf("Config saved to %s", path)
	return nil
}

// LoadConfigFile opens a file dialog and loads a TOML config.
func (a *App) LoadConfigFile() (*config.ClientConfig, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open VPN Config",
		Filters: []runtime.FileFilter{
			{DisplayName: "TOML Config (*.toml)", Pattern: "*.toml"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil // cancelled
	}

	return a.loadConfigFromPath(path)
}

func (a *App) loadConfigFromPath(path string) (*config.ClientConfig, error) {
	cfg, err := config.LoadClientConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	a.mu.Lock()
	a.cfg = cfg
	a.cfgPath = path
	a.mu.Unlock()

	log.Printf("Config loaded from %s", path)
	return cfg, nil
}

// SaveConfigFileAs opens a save dialog and writes the config.
func (a *App) SaveConfigFileAs(cfg config.ClientConfig) error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save VPN Config",
		DefaultFilename: "client.toml",
		Filters: []runtime.FileFilter{
			{DisplayName: "TOML Config (*.toml)", Pattern: "*.toml"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil // cancelled
	}

	config.ApplyClientDefaults(&cfg)
	if err := writeConfigToml(path, &cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	a.mu.Lock()
	a.cfg = &cfg
	a.cfgPath = path
	a.mu.Unlock()

	log.Printf("Config saved to %s", path)
	return nil
}

func (a *App) emitStatus(status, ip, errMsg string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.emitStatusLocked(status, ip, errMsg)
}

func (a *App) emitStatusLocked(status, ip, errMsg string) {
	a.status = status
	if ip != "" {
		a.assignedIP = ip
	}
	runtime.EventsEmit(a.ctx, "vpn:status", StatusUpdate{
		Status:     status,
		AssignedIP: ip,
		Error:      errMsg,
	})
}

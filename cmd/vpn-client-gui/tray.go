//go:build windows

package main

import (
	_ "embed"

	"github.com/energye/systray"
)

//go:embed icons/connected.ico
var connectedIcon []byte

//go:embed icons/disconnected.ico
var disconnectedIcon []byte

var (
	trayApp     *App
	mConnect    *systray.MenuItem
	mDisconnect *systray.MenuItem
)

func initTray(app *App) {
	trayApp = app
	go systray.Run(onTrayReady, nil)
}

func onTrayReady() {
	systray.SetIcon(disconnectedIcon)
	systray.SetTitle("SimpleVPN")
	systray.SetTooltip("SimpleVPN - Disconnected")

	mShow := systray.AddMenuItem("Show", "Show window")
	systray.AddSeparator()
	mConnect = systray.AddMenuItem("Connect", "Connect VPN")
	mDisconnect = systray.AddMenuItem("Disconnect", "Disconnect VPN")
	mDisconnect.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit SimpleVPN")

	mShow.Click(func() {
		if trayApp != nil {
			trayApp.ShowWindow()
		}
	})
	mConnect.Click(func() {
		if trayApp != nil {
			trayApp.Connect()
		}
	})
	mDisconnect.Click(func() {
		if trayApp != nil {
			trayApp.Disconnect()
		}
	})
	mQuit.Click(func() {
		if trayApp != nil {
			trayApp.Quit()
		}
	})
}

func updateTrayStatus(connected bool) {
	if connected {
		systray.SetIcon(connectedIcon)
		systray.SetTooltip("SimpleVPN - Connected")
		if mConnect != nil {
			mConnect.Disable()
		}
		if mDisconnect != nil {
			mDisconnect.Enable()
		}
	} else {
		systray.SetIcon(disconnectedIcon)
		systray.SetTooltip("SimpleVPN - Disconnected")
		if mConnect != nil {
			mConnect.Enable()
		}
		if mDisconnect != nil {
			mDisconnect.Disable()
		}
	}
}

func cleanupTray() {
	systray.Quit()
}

package main

import (
	"embed"
	"log"

	"github.com/gavsh/simplevpn/internal/wintun"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := wintun.Extract(); err != nil {
		log.Printf("Warning: failed to extract wintun.dll: %v", err)
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "SimpleVPN",
		Width:     900,
		Height:    600,
		MinWidth:  700,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 26, G: 26, B: 46, A: 255},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.Dark,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

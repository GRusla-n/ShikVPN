//go:build ignore

package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"bytes"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join("cmd", "vpn-client-gui", "icons")
	os.MkdirAll(dir, 0755)

	// Disconnected icon: grey shield
	writeICO(filepath.Join(dir, "disconnected.ico"), color.RGBA{128, 128, 128, 255}, color.RGBA{80, 80, 80, 255})
	// Connected icon: green shield
	writeICO(filepath.Join(dir, "connected.ico"), color.RGBA{0, 200, 83, 255}, color.RGBA{0, 150, 60, 255})
	// App icon: blue shield
	writeICO(filepath.Join(dir, "icon.ico"), color.RGBA{0, 210, 255, 255}, color.RGBA{0, 150, 200, 255})

	// Also copy app icon to build dir
	buildDir := filepath.Join("cmd", "vpn-client-gui", "build", "windows")
	os.MkdirAll(buildDir, 0755)
	src, _ := os.ReadFile(filepath.Join(dir, "icon.ico"))
	os.WriteFile(filepath.Join(buildDir, "icon.ico"), src, 0644)
}

func writeICO(path string, primary, secondary color.RGBA) {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Draw a simple shield shape
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Shield outline: a rounded top with pointed bottom
			cx, cy := float64(x)-float64(size)/2, float64(y)-float64(size)/2
			inShield := false

			if y < size*2/3 {
				// Top rectangle part
				if x >= 4 && x < size-4 && y >= 2 && y < size*2/3 {
					inShield = true
				}
			} else {
				// Bottom triangle part
				progress := float64(y-size*2/3) / float64(size/3)
				halfWidth := float64(size/2-4) * (1 - progress)
				if cx >= -halfWidth && cx <= halfWidth {
					inShield = true
				}
			}
			_ = cy

			if inShield {
				if y < size/3 {
					img.Set(x, y, primary)
				} else {
					img.Set(x, y, secondary)
				}
			} else {
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	// Encode PNG
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	pngData := pngBuf.Bytes()

	// Build ICO file
	var ico bytes.Buffer
	// ICONDIR header
	binary.Write(&ico, binary.LittleEndian, uint16(0))    // reserved
	binary.Write(&ico, binary.LittleEndian, uint16(1))    // type: icon
	binary.Write(&ico, binary.LittleEndian, uint16(1))    // count

	// ICONDIRENTRY
	ico.WriteByte(byte(size))        // width
	ico.WriteByte(byte(size))        // height
	ico.WriteByte(0)                 // color count
	ico.WriteByte(0)                 // reserved
	binary.Write(&ico, binary.LittleEndian, uint16(1))    // color planes
	binary.Write(&ico, binary.LittleEndian, uint16(32))   // bits per pixel
	binary.Write(&ico, binary.LittleEndian, uint32(len(pngData))) // size
	binary.Write(&ico, binary.LittleEndian, uint32(22))   // offset (6 + 16 = 22)

	ico.Write(pngData)

	os.WriteFile(path, ico.Bytes(), 0644)
}

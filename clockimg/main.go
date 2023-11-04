package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	// Bildgröße festlegen
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))

	// Hintergrundfarbe festlegen
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	// Aktuelle Uhrzeit und Datum
	now := time.Now()
	hour, min, _ := now.Clock()
	dateStr := now.Format("02.01.2006")
	timeStr := now.Format("15:04")

	// Zentrum der Uhr
	centerX, centerY := 64, 64

	// Zeichnen Sie die Zeiger
	drawHand(img, centerX, centerY, int(float64(hour)/12*360), 30, color.White) // Stundenzeiger
	drawHand(img, centerX, centerY, int(float64(min)/60*360), 40, color.White)  // Minutenzeiger

	// Datum unten mittig hinzufügen
	addLabel(img, centerX, 110, dateStr)
	addLabel(img, centerX, 125, timeStr)

	// Datei speichern
	f, err := os.Create("analog_clock.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

// drawHand zeichnet einen Zeiger der Uhr
func drawHand(img *image.RGBA, x, y, angle, length int, col color.Color) {
	// Umwandlung von Grad in Radiant
	rad := float64(angle-90) * math.Pi / 180

	// Berechnen Sie das Ende des Zeigers
	endX := x + int(float64(length)*math.Cos(rad))
	endY := y + int(float64(length)*math.Sin(rad))

	// Zeichnen Sie eine Linie vom Mittelpunkt zum berechneten Endpunkt
	drawLine(img, x, y, endX, endY, col)
}

// drawLine zeichnet eine einfache Linie von (x1, y1) zu (x2, y2)
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	sx := -1.0
	sy := -1.0
	if x1 < x2 {
		sx = 1.0
	}
	if y1 < y2 {
		sy = 1.0
	}
	err := dx - dy

	for {
		img.Set(int(x1), int(y1), col)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += int(sx)
		}
		if e2 < dx {
			err += dx
			y1 += int(sy)
		}
	}
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{255, 255, 255, 255} // Weiß
	point := fixed.P(x-(len(label)*3), y) // Zentriert das Datum
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

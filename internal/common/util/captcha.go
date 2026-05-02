package util

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"math/rand"
)

// CaptchaImage generates a captcha JPEG image from the given text.
// Returns the JPEG bytes. Width=120, Height=40, random noise lines and dots.
func CaptchaImage(text string) ([]byte, error) {
	width, height := 120, 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background: light gray
	bg := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	// Draw random noise lines
	for i := 0; i < 6; i++ {
		c := randomColor()
		x1, y1 := rand.Intn(width), rand.Intn(height)
		x2, y2 := rand.Intn(width), rand.Intn(height)
		drawLine(img, x1, y1, x2, y2, c)
	}

	// Draw random dots
	for i := 0; i < 50; i++ {
		x, y := rand.Intn(width), rand.Intn(height)
		img.Set(x, y, randomColor())
	}

	// Draw text characters as simple digit shapes
	startX := 10
	for _, ch := range text {
		drawDigit(img, startX, 8, byte(ch), randomDarkColor())
		startX += 25
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func randomColor() color.RGBA {
	return color.RGBA{
		R: uint8(rand.Intn(200) + 55),
		G: uint8(rand.Intn(200) + 55),
		B: uint8(rand.Intn(200) + 55),
		A: 255,
	}
}

func randomDarkColor() color.RGBA {
	return color.RGBA{
		R: uint8(rand.Intn(100)),
		G: uint8(rand.Intn(100)),
		B: uint8(rand.Intn(100)),
		A: 255,
	}
}

// drawLine draws a simple line using Bresenham's algorithm.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawDigit draws a simple 7-segment style digit at (x, y).
func drawDigit(img *image.RGBA, x, y int, ch byte, c color.RGBA) {
	// Simple pixel art for digits 0-9
	// Each digit is ~15x25 pixels
	segments := map[byte][]bool{
		'0': {true, true, true, true, true, true, false},
		'1': {false, true, true, false, false, false, false},
		'2': {true, true, false, true, true, false, true},
		'3': {true, true, true, true, false, false, true},
		'4': {false, true, true, false, false, true, true},
		'5': {true, false, true, true, false, true, true},
		'6': {true, false, true, true, true, true, true},
		'7': {true, true, true, false, false, false, false},
		'8': {true, true, true, true, true, true, true},
		'9': {true, true, true, true, false, true, true},
	}

	segs, ok := segments[ch-'0']
	if !ok {
		return
	}

	thick := 2
	w, h := 12, 20

	// Segment definitions: [x1, y1, x2, y2]
	type seg struct{ x1, y1, x2, y2 int }
	segCoords := []seg{
		{x, y, x + w, y},                   // top
		{x + w, y, x + w, y + h/2},         // top-right
		{x + w, y + h/2, x + w, y + h},     // bottom-right
		{x, y + h, x + w, y + h},           // bottom
		{x, y + h/2, x, y + h},             // bottom-left
		{x, y, x, y + h/2},                 // top-left
		{x, y + h/2, x + w, y + h/2},       // middle
	}

	for i, on := range segs {
		if !on {
			continue
		}
		s := segCoords[i]
		// Draw thick line by drawing multiple offset lines
		for t := -thick / 2; t <= thick/2; t++ {
			drawLine(img, s.x1, s.y1+t, s.x2, s.y2+t, c)
		}
	}
}

// SineWaveDistortion applies slight sine wave distortion to image pixels for anti-OCR.
func SineWaveDistortion(img *image.RGBA, amplitude float64, frequency float64) {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	result := image.NewRGBA(bounds)

	for y := 0; y < h; y++ {
		offset := int(amplitude * math.Sin(frequency*float64(y)))
		for x := 0; x < w; x++ {
			srcX := x + offset
			if srcX >= 0 && srcX < w {
				result.Set(x, y, img.At(srcX, y))
			}
		}
	}

	draw.Draw(img, bounds, result, image.Point{}, draw.Src)
}

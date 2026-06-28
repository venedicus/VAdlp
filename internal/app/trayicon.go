package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// generateTrayIconICO renders a small "V" glyph on the app's accent colors and
// wraps it in a minimal ICO container (PNG payload, supported since Vista).
// Generated at runtime so the tray icon needs no extra tracked asset file.
func generateTrayIconICO() []byte {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{R: 0x1a, G: 0x1a, B: 0x26, A: 0xff}
	fg := color.RGBA{R: 0xff, G: 0x79, B: 0xc6, A: 0xff}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
		}
	}
	thickness := 3
	for i := 0; i < size/2+2; i++ {
		for t := 0; t < thickness; t++ {
			setIfInBounds(img, 5+i+t, 4+i, fg)
			setIfInBounds(img, size-6-i+t, 4+i, fg)
		}
	}

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return nil
	}
	pngBytes := pngBuf.Bytes()

	var ico bytes.Buffer
	ico.Write([]byte{0, 0, 1, 0, 1, 0}) // ICONDIR: reserved, type=1 (icon), count=1
	ico.WriteByte(size)                 // width
	ico.WriteByte(size)                 // height
	ico.WriteByte(0)                    // color count (0 = use PNG)
	ico.WriteByte(0)                    // reserved
	writeUint16LE(&ico, 1)              // color planes
	writeUint16LE(&ico, 32)             // bits per pixel
	writeUint32LE(&ico, uint32(len(pngBytes)))
	writeUint32LE(&ico, 22) // offset: 6-byte header + 16-byte entry
	ico.Write(pngBytes)
	return ico.Bytes()
}

func setIfInBounds(img *image.RGBA, x, y int, c color.Color) {
	b := img.Bounds()
	if x >= b.Min.X && x < b.Max.X && y >= b.Min.Y && y < b.Max.Y {
		img.Set(x, y, c)
	}
}

func writeUint16LE(buf *bytes.Buffer, v uint16) {
	buf.WriteByte(byte(v))
	buf.WriteByte(byte(v >> 8))
}

func writeUint32LE(buf *bytes.Buffer, v uint32) {
	buf.WriteByte(byte(v))
	buf.WriteByte(byte(v >> 8))
	buf.WriteByte(byte(v >> 16))
	buf.WriteByte(byte(v >> 24))
}

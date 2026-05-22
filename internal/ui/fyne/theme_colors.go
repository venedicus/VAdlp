package fyneui

import (
	"image/color"
	"strings"

	"vadlp/internal/downloader"
)

func StatusColor(key string) color.Color {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "ready":
		return color.NRGBA{R: 0x3c, G: 0x40, B: 0x5a, A: 0xff}
	case "running":
		return color.NRGBA{R: 0x2f, G: 0x64, B: 0xb8, A: 0xff}
	case "completed", "complete":
		return color.NRGBA{R: 0x3d, G: 0x8c, B: 0x4a, A: 0xff}
	case "error":
		return color.NRGBA{R: 0xb8, G: 0x3d, B: 0x4f, A: 0xff}
	case "stopped", "cancelled":
		return color.NRGBA{R: 0xc6, G: 0x7a, B: 0x2f, A: 0xff}
	case "queued":
		return color.NRGBA{R: 0x6c, G: 0x4a, B: 0xb8, A: 0xff}
	default:
		return color.NRGBA{R: 0x3c, G: 0x40, B: 0x5a, A: 0xff}
	}
}

func PhaseColor(st downloader.Stage) color.Color {
	switch st {
	case downloader.StageExtracting:
		return color.NRGBA{R: 0x8c, G: 0x6a, B: 0x24, A: 0xff}
	case downloader.StageDownloading:
		return color.NRGBA{R: 0x2a, G: 0x55, B: 0xa8, A: 0xff}
	case downloader.StagePostProcess:
		return color.NRGBA{R: 0x6b, G: 0x3d, B: 0xa8, A: 0xff}
	default:
		return color.NRGBA{R: 0x34, G: 0x38, B: 0x4d, A: 0xff}
	}
}

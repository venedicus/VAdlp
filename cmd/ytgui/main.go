package main

import (
	"io"
	"log"

	fyneui "ytgui/internal/ui/fyne"
)

func main() {
	// Suppress noisy framework logs in terminal; issues are shown in-app.
	log.SetOutput(io.Discard)
	fyneui.Run()
}

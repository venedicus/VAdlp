package main

import (
	"io"
	"log"

	fyneui "vadlp/internal/ui/fyne"
)

func main() {
	log.SetOutput(io.Discard)
	fyneui.Run()
}

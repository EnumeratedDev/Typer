package main

import (
	"log"
	"os"
)

func main() {
	window, err := CreateWindow()
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	if len(os.Args) > 1 {
		for _, file := range os.Args[1:] {
			b, err := CreateFileBuffer(file, true)
			if err != nil {
				PrintMessage(window, "Could not open file: "+file)
				continue
			}

			if window.CurrentBuffer.Name == "New File 1" {
				delete(Buffers, window.CurrentBuffer.Id)
				window.CurrentBuffer = b
			}
		}
	}

	for window.screen != nil {
		window.Draw()
	}
}

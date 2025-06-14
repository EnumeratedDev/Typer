package main

import (
	"log"
	"os"
)

func main() {
	// Read config
	readConfig()

	// Read key bindings
	readKeybindings()

	// Read styles
	readStyles()

	// Initialize commands
	initCommands()

	window, err := CreateWindow()
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	if len(os.Args) > 1 {
		for i, file := range os.Args[1:] {
			b, err := CreateFileBuffer(file, true)
			if err != nil {
				PrintMessage(window, "Could not open file: "+file)
				continue
			}

			if i == 0 {
				window.CurrentBuffer = b
				Buffers = Buffers[1:]
			}
		}
	}

	for window.screen != nil {
		window.Draw()
		window.ProcessEvents()
	}

}

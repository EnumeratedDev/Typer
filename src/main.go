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

	var initialBuffer *Buffer = nil

	if len(os.Args) > 0 {
		for _, file := range os.Args[1:] {
			b, err := CreateFileBuffer(file)
			if err != nil {
				PrintMessage(window, "Could not open file: "+file)
				continue
			}
			Buffers[b.Id] = b
			if initialBuffer == nil {
				initialBuffer = b
			}
		}
	}

	if initialBuffer != nil {
		delete(Buffers, window.textArea.CurrentBuffer.Id)
		window.textArea.CurrentBuffer = initialBuffer
	}

	for window.screen != nil {
		window.Draw()
	}
}

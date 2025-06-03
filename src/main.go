package main

import (
	"log"
)

func main() {
	window, err := CreateWindow(nil)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	for window.screen != nil {
		window.Draw()
	}
}

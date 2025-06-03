package main

import (
	"github.com/gdamore/tcell"
)

type TopMenuButton struct {
	Name string
}

var TopMenuButtons = make([]TopMenuButton, 0)

func initTopMenu() {
	// Buttons
	fileButton := TopMenuButton{
		Name: "File",
	}
	EditButton := TopMenuButton{
		Name: "Edit",
	}
	Buffers := TopMenuButton{
		Name: "Buffers",
	}

	// Append buttons
	TopMenuButtons = append(TopMenuButtons, fileButton, EditButton, Buffers)
}

func drawTopMenu(window *Window) {
	screen := window.screen

	topMenuStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

	sizeX, _ := screen.Size()

	for x := 0; x < sizeX; x++ {
		screen.SetContent(x, 0, ' ', nil, topMenuStyle)
	}

	currentX := 1
	for _, button := range TopMenuButtons {
		drawText(screen, currentX, 0, currentX+len(button.Name), 0, topMenuStyle, button.Name)
		currentX += len(button.Name) + 1
	}

}

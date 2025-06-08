package main

import (
	"github.com/gdamore/tcell"
	"strconv"
	"strings"
)

func drawLineIndex(window *Window) {
	screen := window.screen
	buffer := window.CurrentBuffer

	lineIndexStyle := tcell.StyleDefault.Foreground(tcell.ColorDimGray).Background(tcell.Color237)

	_, sizeY := screen.Size()

	y := 0
	if window.ShowTopMenu {
		y = 1
	}

	for lineIndex := 1; lineIndex <= strings.Count(buffer.Contents, "\n")+1 && lineIndex < sizeY; lineIndex++ {
		drawText(screen, 0, y, 3, y, lineIndexStyle, strconv.Itoa(lineIndex)+". ")
		y++
	}
}

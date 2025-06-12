package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type TopMenuButton struct {
	Name   string
	Action func(w *Window)
}

var TopMenuButtons = make([]TopMenuButton, 0)

func initTopMenu() {
	// Buttons
	fileButton := TopMenuButton{
		Name: "File",
		Action: func(window *Window) {
			ClearDropdowns()
			d := CreateDropdownMenu([]string{"New", "Save", "Open", "Close", "Quit"}, 0, 1, 0, func(i int) {
				switch i {
				case 0:
					RunCommand(window, "new-buffer")
				case 1:
					RunCommand(window, "save")
				case 2:
					RunCommand(window, "open")
				case 3:
					RunCommand(window, "close-buffer")
				case 4:
					RunCommand(window, "quit")
				}
				ClearDropdowns()
			})
			ActiveDropdown = d
			window.CursorMode = CursorModeDropdown
		},
	}
	EditButton := TopMenuButton{
		Name: "Edit",
		Action: func(window *Window) {
			ClearDropdowns()
			d := CreateDropdownMenu([]string{"Copy", "Paste"}, 0, 1, 0, func(i int) {
				switch i {
				case 0:
					RunCommand(window, "copy")
				case 1:
					RunCommand(window, "paste")
				}
				ClearDropdowns()
				window.CursorMode = CursorModeBuffer
			})
			ActiveDropdown = d
			window.CursorMode = CursorModeDropdown
		},
	}
	Buffers := TopMenuButton{
		Name: "Buffers",
		Action: func(window *Window) {
			ClearDropdowns()
			buffersSlice := make([]string, 0)
			for _, buffer := range Buffers {
				if window.CurrentBuffer == buffer {
					buffersSlice = append(buffersSlice, fmt.Sprintf("[%d] * %s", buffer.Id, buffer.Name))
				} else {
					buffersSlice = append(buffersSlice, fmt.Sprintf("[%d] %s", buffer.Id, buffer.Name))
				}
			}

			slices.Sort(buffersSlice)

			d := CreateDropdownMenu(buffersSlice, 0, 1, 0, func(i int) {
				start := strings.Index(buffersSlice[i], "[")
				end := strings.Index(buffersSlice[i], "]")

				id, err := strconv.Atoi(buffersSlice[i][start+1 : end])
				if err != nil {
					PrintMessage(window, fmt.Sprintf("Cannot convert buffer id '%s' to int", buffersSlice[i][start:end]))
					return
				}

				window.CurrentBuffer = Buffers[id]
				ClearDropdowns()
				window.CursorMode = CursorModeBuffer
			})
			ActiveDropdown = d
			window.CursorMode = CursorModeDropdown
		},
	}

	// Append buttons
	TopMenuButtons = append(TopMenuButtons, fileButton, EditButton, Buffers)
}

func drawTopMenu(window *Window) {
	screen := window.screen

	topMenuStyle := tcell.StyleDefault.Background(CurrentStyle.TopMenuBg).Foreground(CurrentStyle.TopMenuFg)

	sizeX, _ := screen.Size()

	for x := 0; x < sizeX; x++ {
		screen.SetContent(x, 0, ' ', nil, topMenuStyle)
	}

	currentX := 1
	for _, button := range TopMenuButtons {
		drawText(screen, currentX, 0, currentX+len(button.Name), 0, topMenuStyle, button.Name)
		currentX += len(button.Name) + 1
	}

	// Draw buffer info
	filename := "Not set"
	if filepath.Base(window.CurrentBuffer.filename) != "." {
		filename = filepath.Base(window.CurrentBuffer.filename)
	}
	cursorX, cursorY := window.GetCursorPos2D()
	cursorInfo := fmt.Sprintf("File: %s Cursor: (%d,%d,%d) Words: %d", filename, cursorX+1, cursorY+1, window.CurrentBuffer.CursorPos+1, len(strings.Fields(window.CurrentBuffer.Contents)))
	drawText(screen, sizeX-len(cursorInfo)-1, 0, sizeX-1, 0, topMenuStyle, cursorInfo)
}

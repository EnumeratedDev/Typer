package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"maps"
	"slices"
	"strconv"
	"strings"
)

type TopMenuButton struct {
	Name   string
	Key    rune
	Action func(w *Window)
}

var TopMenuButtons = make([]TopMenuButton, 0)

func initTopMenu() {
	// Buttons
	fileButton := TopMenuButton{
		Name: "File",
		Key:  'f',
		Action: func(window *Window) {
			ClearDropdowns()
			d := CreateDropdownMenu([]string{"New", "Save", "Open", "Close", "Quit"}, 0, 1, 0, func(i int) {
				switch i {
				case 0:
					number := 1
					for _, buffer := range Buffers {
						if strings.HasPrefix(buffer.Name, "New File ") {
							number++
						}
					}
					buffer := CreateBuffer(fmt.Sprintf("New File %d", number))
					window.textArea.CurrentBuffer = buffer
					window.SetCursorPos(0)
					window.CursorMode = CursorModeBuffer
				case 1:
					_ = RequestInput(window, "Save buffer to:")
					PrintMessage(window, "Input requested...")
				case 2:
					inputChannel := RequestInput(window, "File to open:")
					go func() {
						input := <-inputChannel

						if input == "" {
							return
						}

						buffer, err := CreateFileBuffer(input)
						if err != nil {
							PrintMessage(window, fmt.Sprintf("Could not open file: %s", err.Error()))
							return
						}
						PrintMessage(window, fmt.Sprintf("Opening file: %s", input))
						window.textArea.CurrentBuffer = buffer
					}()
				case 3:
					delete(Buffers, window.textArea.CurrentBuffer.Id)
					buffersSlice := slices.Collect(maps.Values(Buffers))
					if len(buffersSlice) == 0 {
						window.Close()
						return
					}
					window.textArea.CurrentBuffer = buffersSlice[0]
					window.SetCursorPos(0)
					window.CursorMode = CursorModeBuffer
				case 4:
					window.Close()
					window.CursorMode = CursorModeBuffer
				}
				ClearDropdowns()
			})
			ActiveDropdown = d
			window.CursorMode = CursorModeDropdown
		},
	}
	EditButton := TopMenuButton{
		Name: "Edit",
		Key:  'e',
	}
	Buffers := TopMenuButton{
		Name: "Buffers",
		Key:  'b',
		Action: func(window *Window) {
			ClearDropdowns()
			buffersSlice := make([]string, 0)
			for _, buffer := range Buffers {
				if window.textArea.CurrentBuffer == buffer {
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

				window.textArea.CurrentBuffer = Buffers[id]
				window.SetCursorPos(0)
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

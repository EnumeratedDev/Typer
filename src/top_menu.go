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
					window.CurrentBuffer = buffer
					window.CursorMode = CursorModeBuffer
				case 1:
					if !window.CurrentBuffer.canSave {
						PrintMessage(window, "Cannot save this buffer!")
						return
					}

					inputChannel := RequestInput(window, "Save file [y\\N]:", "")
					go func() {
						input := <-inputChannel

						if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
							return
						}

						inputChannel = RequestInput(window, "Save buffer to:", window.CurrentBuffer.filename)

						input = <-inputChannel

						if strings.TrimSpace(input) == "" {
							PrintMessage(window, "No save location was given!")
							return
						}

						window.CurrentBuffer.filename = strings.TrimSpace(input)
						err := window.CurrentBuffer.Save()
						if err != nil {
							PrintMessage(window, fmt.Sprintf("Could not save file: %s", err))
							window.CurrentBuffer.filename = ""
							return
						}

						PrintMessage(window, "File saved.")
					}()
				case 2:
					inputChannel := RequestInput(window, "File to open:", "")
					go func() {
						input := <-inputChannel

						if input == "" {
							return
						}

						if openBuffer := GetOpenFileBuffer(input); openBuffer != nil {
							PrintMessage(window, fmt.Sprintf("File already open! Switching to buffer: %s", openBuffer.Name))
							window.CurrentBuffer = openBuffer
						} else {
							newBuffer, err := CreateFileBuffer(input, false)
							if err != nil {
								PrintMessage(window, fmt.Sprintf("Could not open file: %s", err.Error()))
								return
							}

							PrintMessage(window, fmt.Sprintf("Opening file at: %s", newBuffer.filename))
							window.CurrentBuffer = newBuffer
						}
					}()
				case 3:
					delete(Buffers, window.CurrentBuffer.Id)
					buffersSlice := slices.Collect(maps.Values(Buffers))
					if len(buffersSlice) == 0 {
						window.Close()
						return
					}
					window.CurrentBuffer = buffersSlice[0]
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
		Action: func(window *Window) {
			ClearDropdowns()
			d := CreateDropdownMenu([]string{"Copy", "Paste"}, 0, 1, 0, func(i int) {
				switch i {
				case 0:
					if window.CurrentBuffer.Selection == nil {
						// Copy line
						_, line := window.GetCursorPos2D()
						window.Clipboard = strings.SplitAfter(window.CurrentBuffer.Contents, "\n")[line]
						PrintMessage(window, "Copied line to clipboard.")
					} else {
						// Copy selection
						window.Clipboard = window.CurrentBuffer.GetSelectedText()
						PrintMessage(window, "Copied selection to clipboard.")
					}
				case 1:
					str := window.CurrentBuffer.Contents
					index := window.CurrentBuffer.CursorPos

					if index == len(str) {
						str += window.Clipboard
					} else {
						str = str[:index] + window.Clipboard + str[index:]
					}
					window.CurrentBuffer.CursorPos += len(window.Clipboard)
					window.CurrentBuffer.Contents = str
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
		Key:  'b',
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

	topMenuStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.Color236)

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

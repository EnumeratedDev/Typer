package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"slices"
	"strings"
)

type CursorMode uint8

const (
	CursorModeDisabled CursorMode = iota
	CursorModeBuffer
	CursorModeDropdown
	CursorModeInputBar
)

type Window struct {
	ShowTopMenu   bool
	ShowLineIndex bool
	CursorMode    CursorMode

	Clipboard string

	CurrentBuffer *Buffer

	screen tcell.Screen
}

var mouseHeld = false

func CreateWindow() (*Window, error) {
	window := Window{
		ShowTopMenu:   true,
		ShowLineIndex: true,
		CursorMode:    CursorModeBuffer,

		CurrentBuffer: nil,

		screen: nil,
	}

	// Create empty buffer if nil
	if window.CurrentBuffer == nil {
		window.CurrentBuffer = CreateBuffer("New File 1")
	}

	// Create tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to initialize tcell: %s", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %s", err)
	}

	// Set screen style
	screen.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.Color234))

	// Enable mouse
	screen.EnableMouse()

	// Set window screen field
	window.screen = screen

	// Initialize top menu
	initTopMenu()

	return &window, nil
}

func (window *Window) drawCurrentBuffer() {
	buffer := window.CurrentBuffer

	x, y, _, _ := window.GetTextAreaDimensions()

	bufferX, bufferY, _, _ := window.GetTextAreaDimensions()

	for i, r := range buffer.Contents + " " {
		if x-buffer.OffsetX >= bufferX && y-buffer.OffsetY >= bufferY {
			// Default style
			style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.Color234)

			// Change background if under cursor
			if i == buffer.CursorPos {
				style = style.Background(tcell.Color243)
			}

			// Change background if selected
			if buffer.Selection != nil {
				if buffer.Selection.selectionEnd >= buffer.Selection.selectionStart && i >= buffer.Selection.selectionStart && i <= buffer.Selection.selectionEnd {
					style = style.Background(tcell.Color243)

					// Show selection on entire tab space
					if r == '\t' {
						for j := 0; j < 4; j++ {
							window.screen.SetContent(x+j-buffer.OffsetX, y-buffer.OffsetY, r, nil, style)
						}
					}
				} else if i <= buffer.Selection.selectionStart && i >= buffer.Selection.selectionEnd {
					style = style.Background(tcell.Color243)

					// Show selection on entire tab space
					if r == '\t' {
						for j := 0; j < 4; j++ {
							window.screen.SetContent(x+j-buffer.OffsetX, y-buffer.OffsetY, r, nil, style)
						}
					}
				}
			}

			window.screen.SetContent(x-buffer.OffsetX, y-buffer.OffsetY, r, nil, style)
		}

		// Change position for next character
		if r == '\n' {
			x = bufferX
			y++
		} else if r == '\t' {
			x += 4
		} else {
			x++
		}
	}
}

func (window *Window) Draw() {
	// Clear screen
	window.screen.Clear()

	// Draw top menu
	if window.ShowTopMenu {
		drawTopMenu(window)
	}

	// Draw line index
	if window.ShowLineIndex {
		drawLineIndex(window)
	}

	// Draw current buffer
	if window.CurrentBuffer != nil {
		window.drawCurrentBuffer()
	}

	// Draw input bar
	if currentInputRequest != nil {
		drawInputBar(window)
	}

	// Draw message bar
	drawMessageBar(window)

	// Draw dropdowns
	drawDropdowns(window)

	// Draw cursor
	if window.CursorMode == CursorModeInputBar {
		_, sizeY := window.screen.Size()
		window.screen.ShowCursor(len(currentInputRequest.Text)+len(currentInputRequest.input)+1, sizeY-1)
	} else {
		window.screen.HideCursor()
	}

	// Update screen
	window.screen.Show()

	// Poll event
	ev := window.screen.PollEvent()

	// Process event
	switch ev := ev.(type) {
	case *tcell.EventResize:
		window.screen.Sync()
		window.SyncBufferOffset()
	case *tcell.EventMouse:
		window.mouseInput(ev)
	case *tcell.EventKey:
		window.input(ev)
	}
}

func (window *Window) input(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyRight { // Navigation Keys
		if window.CursorMode == CursorModeBuffer {
			// Add to selection
			if ev.Modifiers() == tcell.ModShift {
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: window.CurrentBuffer.CursorPos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
					return
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos + 1
				}
				// Prevent selecting dummy character at the end of the buffer
				if window.CurrentBuffer.Selection.selectionEnd >= len(window.CurrentBuffer.Contents) {
					window.CurrentBuffer.Selection.selectionEnd = len(window.CurrentBuffer.Contents) - 1
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
			// Move cursor
			window.SetCursorPos(window.CurrentBuffer.CursorPos + 1)
		}
	} else if ev.Key() == tcell.KeyLeft {
		if window.CursorMode == CursorModeBuffer {
			// Add to selection
			if ev.Modifiers() == tcell.ModShift {
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: window.CurrentBuffer.CursorPos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
					return
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos - 1
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
			// Move cursor
			window.SetCursorPos(window.CurrentBuffer.CursorPos - 1)
		}
	} else if ev.Key() == tcell.KeyUp {
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos
			// Move cursor
			x, y := window.GetCursorPos2D()
			window.SetCursorPos2D(x, y-1)
			// Add to selection
			if ev.Modifiers() == tcell.ModShift {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: window.CurrentBuffer.CursorPos,
						selectionEnd:   pos,
					}
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
		} else if window.CursorMode == CursorModeDropdown {
			dropdown := ActiveDropdown
			dropdown.Selected--
			if dropdown.Selected < 0 {
				dropdown.Selected = 0
			}
		} else if window.CursorMode == CursorModeInputBar {
			if len(inputHistory) == 0 {
				return
			}

			current := slices.Index(inputHistory, currentInputRequest.input)
			if current < 0 {
				current = len(inputHistory) - 1
			} else if current != 0 {
				current--
			}

			currentInputRequest.input = inputHistory[current]
			currentInputRequest.cursorPos = len(inputHistory[current])
		}
	} else if ev.Key() == tcell.KeyDown {
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos
			// Move cursor
			x, y := window.GetCursorPos2D()
			window.SetCursorPos2D(x, y+1)
			// Add to selection
			if ev.Modifiers() == tcell.ModShift {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: pos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos
				}
				// Prevent selecting dummy character at the end of the buffer
				if window.CurrentBuffer.Selection.selectionEnd >= len(window.CurrentBuffer.Contents) {
					window.CurrentBuffer.Selection.selectionEnd = len(window.CurrentBuffer.Contents) - 1
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
		} else if window.CursorMode == CursorModeDropdown {
			dropdown := ActiveDropdown
			dropdown.Selected++
			if dropdown.Selected >= len(dropdown.Options) {
				dropdown.Selected = len(dropdown.Options) - 1
			}
		} else if window.CursorMode == CursorModeInputBar {
			if len(inputHistory) == 0 {
				return
			}

			current := slices.Index(inputHistory, currentInputRequest.input)
			if current < 0 {
				return
			} else if current == len(inputHistory)-1 {
				currentInputRequest.input = ""
				return
			} else {
				current++
			}

			currentInputRequest.input = inputHistory[current]
			currentInputRequest.cursorPos = len(inputHistory[current])
		}
	} else if ev.Key() == tcell.KeyEscape {
		if window.CursorMode == CursorModeInputBar {
			currentInputRequest.inputChannel <- ""
			currentInputRequest = nil
			window.CursorMode = CursorModeBuffer
		} else {
			ClearDropdowns()
			window.CursorMode = CursorModeBuffer
		}
	}

	// Check key bindings
	for _, keybinding := range Keybinds {
		if keybinding.IsPressed(ev) && slices.Index(keybinding.cursorModes, window.CursorMode) != -1 {
			RunCommand(window, keybinding.command)
			return
		}
	}

	// Typing
	if ev.Key() == tcell.KeyBackspace2 {
		if window.CursorMode == CursorModeBuffer {
			str := window.CurrentBuffer.Contents
			index := window.CurrentBuffer.CursorPos

			if index != 0 {
				str = str[:index-1] + str[index:]
				window.CurrentBuffer.Contents = str
				window.SetCursorPos(window.CurrentBuffer.CursorPos - 1)
			}
		} else if window.CursorMode == CursorModeInputBar {
			str := currentInputRequest.input
			index := currentInputRequest.cursorPos

			if index != 0 {
				str = str[:index-1] + str[index:]
				currentInputRequest.cursorPos--
				currentInputRequest.input = str
			}
		}
	} else if ev.Key() == tcell.KeyTab {
		if window.CursorMode == CursorModeBuffer {
			str := window.CurrentBuffer.Contents
			index := window.CurrentBuffer.CursorPos

			if index == len(str) {
				str += "\t"
			} else {
				str = str[:index] + "\t" + str[index:]
			}
			window.CurrentBuffer.Contents = str
			window.SetCursorPos(window.CurrentBuffer.CursorPos + 1)
		}
	} else if ev.Key() == tcell.KeyEnter {
		if window.CursorMode == CursorModeBuffer {
			str := window.CurrentBuffer.Contents
			index := window.CurrentBuffer.CursorPos

			if index == len(str) {
				str += "\n"
			} else {
				str = str[:index] + "\n" + str[index:]
			}
			window.CurrentBuffer.Contents = str
			window.SetCursorPos(window.CurrentBuffer.CursorPos + 1)
		} else if window.CursorMode == CursorModeInputBar {
			if currentInputRequest.input == "" && slices.Index(inputHistory, currentInputRequest.input) == -1 {
				inputHistory = append(inputHistory, currentInputRequest.input)
			}
			currentInputRequest.inputChannel <- currentInputRequest.input
			currentInputRequest = nil
			window.CursorMode = CursorModeBuffer
		} else if window.CursorMode == CursorModeDropdown {
			d := ActiveDropdown
			d.Action(d.Selected)
		}
	} else if ev.Key() == tcell.KeyRune {
		if window.CursorMode == CursorModeBuffer {
			str := window.CurrentBuffer.Contents
			index := window.CurrentBuffer.CursorPos

			if index == len(str) {
				str += string(ev.Rune())
			} else {
				str = str[:index] + string(ev.Rune()) + str[index:]
			}
			window.CurrentBuffer.Contents = str
			window.SetCursorPos(window.CurrentBuffer.CursorPos + 1)
		} else if window.CursorMode == CursorModeInputBar {
			str := currentInputRequest.input
			index := currentInputRequest.cursorPos

			if index == len(str) {
				str += string(ev.Rune())
			} else {
				str = str[:index] + string(ev.Rune()) + str[index:]
			}

			currentInputRequest.cursorPos++
			currentInputRequest.input = str
		}
	}
}

func (window *Window) mouseInput(ev *tcell.EventMouse) {
	mouseX, mouseY := ev.Position()

	// Left click was pressed
	if ev.Buttons() == tcell.Button1 {
		// Ensure click was in buffer area
		x1, y1, x2, y2 := window.GetTextAreaDimensions()
		if mouseX >= x1 && mouseY >= y1 && mouseX <= x2 && mouseY <= y2 {
			bufferMouseX, bufferMouseY := window.AbsolutePosToCursorPos2D(mouseX, mouseY)
			if mouseHeld {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: window.CurrentBuffer.CursorPos,
						selectionEnd:   window.CursorPos2DToCursorPos(bufferMouseX, bufferMouseY),
					}
					return
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CursorPos2DToCursorPos(bufferMouseX, bufferMouseY)
				}
				// Prevent selecting dummy character at the end of the buffer
				if window.CurrentBuffer.Selection.selectionEnd >= len(window.CurrentBuffer.Contents) {
					window.CurrentBuffer.Selection.selectionEnd = len(window.CurrentBuffer.Contents) - 1
				}
			} else {
				// Clear selection
				if window.CurrentBuffer.Selection != nil {
					window.CurrentBuffer.Selection = nil
				}
			}
			// Move cursor
			window.SetCursorPos2D(bufferMouseX, bufferMouseY)
		}
		mouseHeld = true
	} else if ev.Buttons() == tcell.ButtonNone {
		if mouseHeld {
			mouseHeld = false
		}
	}
}

func (window *Window) Close() {
	window.screen.Fini()
	window.screen = nil
}

func (window *Window) GetTextAreaDimensions() (int, int, int, int) {
	x1, y1 := 0, 0
	x2, y2 := window.screen.Size()

	if window.ShowTopMenu {
		y1++
	}

	if window.ShowLineIndex {
		x1 += getLineIndexSize(window)
	}

	return x1, y1, x2 - 1, y2 - 2
}

func (window *Window) CursorPos2DToCursorPos(x, y int) int {
	// Ensure x and y are positive
	x = max(x, 0)
	y = max(y, 0)

	// Set cursor position to 0 buffer is empty
	if len(window.CurrentBuffer.Contents) == 0 {
		return 0
	}

	// Create line slice from buffer contents
	lines := make([]struct {
		charIndex int
		str       string
	}, 0)

	var str string
	for i, char := range window.CurrentBuffer.Contents {
		str += string(char)
		if char == '\n' || i == len(window.CurrentBuffer.Contents)-1 {
			lines = append(lines, struct {
				charIndex int
				str       string
			}{charIndex: i - len(str) + 1, str: str})
			str = ""
		}
	}

	// Append extra character or line
	if window.CurrentBuffer.Contents[len(window.CurrentBuffer.Contents)-1] == '\n' {
		lines = append(lines, struct {
			charIndex int
			str       string
		}{charIndex: len(window.CurrentBuffer.Contents), str: " "})
	} else {
		lines[len(lines)-1].str += " "
	}

	// Limit x and y
	y = min(y, len(lines)-1)
	x = min(x, len(lines[y].str)-1)

	return lines[y].charIndex + x
}

func (window *Window) AbsolutePosToCursorPos2D(x, y int) (int, int) {
	x1, y1, _, _ := window.GetTextAreaDimensions()

	x -= x1
	y -= y1

	x += window.CurrentBuffer.OffsetX
	y += window.CurrentBuffer.OffsetY

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	split := strings.SplitAfter(window.CurrentBuffer.Contents+" ", "\n")

	if y >= len(split) {
		y = len(split) - 1
	}
	line := split[y]

	posInLine := make([]int, 0)
	for i, char := range []rune(line) {
		if char == '\t' {
			for j := 0; j < 4; j++ {
				posInLine = append(posInLine, i)
			}
		} else {
			posInLine = append(posInLine, i)
		}
	}

	if len(posInLine) == 0 {
		x = 0
	} else if x >= len(posInLine) {
		x = posInLine[len(posInLine)-1]
	} else {
		x = posInLine[x]
	}

	return x, y
}

func (window *Window) GetAbsoluteCursorPos() (int, int) {
	cursorX, cursorY := window.GetCursorPos2D()

	x1, y1, _, _ := window.GetTextAreaDimensions()
	cursorX += x1
	cursorY += y1

	return cursorX, cursorY
}

func (window *Window) GetCursorPos2D() (int, int) {
	cursorX := 0
	cursorY := 0

	for i := 0; i < window.CurrentBuffer.CursorPos; i++ {
		char := window.CurrentBuffer.Contents[i]
		if char == '\n' {
			cursorY++
			cursorX = 0
		} else {
			cursorX++
		}
	}

	return cursorX, cursorY
}

func (window *Window) SetCursorPos(position int) {
	window.CurrentBuffer.CursorPos = position

	if window.CurrentBuffer.CursorPos < 0 {
		window.CurrentBuffer.CursorPos = 0
	}

	if window.CurrentBuffer.CursorPos > len(window.CurrentBuffer.Contents) {
		window.CurrentBuffer.CursorPos = len(window.CurrentBuffer.Contents)
	}

	window.SyncBufferOffset()
}

func (window *Window) SetCursorPos2D(x, y int) {
	// Ensure x and y are positive
	x = max(x, 0)
	y = max(y, 0)

	// Set cursor position to 0 buffer is empty
	if len(window.CurrentBuffer.Contents) == 0 {
		window.SetCursorPos(0)
		return
	}

	// Create line slice from buffer contents
	lines := make([]struct {
		charIndex int
		str       string
	}, 0)

	var str string
	for i, char := range window.CurrentBuffer.Contents {
		str += string(char)
		if char == '\n' || i == len(window.CurrentBuffer.Contents)-1 {
			lines = append(lines, struct {
				charIndex int
				str       string
			}{charIndex: i - len(str) + 1, str: str})
			str = ""
		}
	}

	// Append extra character or line
	if window.CurrentBuffer.Contents[len(window.CurrentBuffer.Contents)-1] == '\n' {
		lines = append(lines, struct {
			charIndex int
			str       string
		}{charIndex: len(window.CurrentBuffer.Contents), str: " "})
	} else {
		lines[len(lines)-1].str += " "
	}

	// Limit x and y
	y = min(y, len(lines)-1)
	x = min(x, len(lines[y].str)-1)

	window.SetCursorPos(lines[y].charIndex + x)
}

func (window *Window) SyncBufferOffset() {
	x, y := window.GetCursorPos2D()
	bufferX1, bufferY1, bufferX2, bufferY2 := window.GetTextAreaDimensions()

	if y < window.CurrentBuffer.OffsetY {
		window.CurrentBuffer.OffsetY = y
	} else if y > window.CurrentBuffer.OffsetY+(bufferY2-bufferY1) {
		window.CurrentBuffer.OffsetY = y - (bufferY2 - bufferY1)
	}

	if x < window.CurrentBuffer.OffsetX {
		window.CurrentBuffer.OffsetX = x
	} else if x > window.CurrentBuffer.OffsetX+(bufferX2-bufferX1) {
		window.CurrentBuffer.OffsetX = x - (bufferX2 - bufferX1)
	}
}

package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type CursorMode uint8

const (
	CursorModeDisabled CursorMode = iota
	CursorModeBuffer
	CursorModeDropdown
	CursorModeInputBar
)

var CursorModeNames = map[CursorMode]string{
	CursorModeDisabled: "disabled",
	CursorModeBuffer:   "buffer",
	CursorModeDropdown: "dropdown",
	CursorModeInputBar: "input_bar",
}

type Window struct {
	ShowTopMenu   bool
	ShowLineIndex bool
	CursorMode    CursorMode

	Clipboard string

	CurrentBuffer *Buffer

	screen tcell.Screen

	closed bool
}

var mouseHeld = false
var lastClick int64 = 0

func CreateWindow() (*Window, error) {
	window := Window{
		ShowTopMenu:   Config.ShowTopMenu,
		ShowLineIndex: Config.ShowLineIndex,
		CursorMode:    CursorModeBuffer,

		CurrentBuffer: nil,

		screen: nil,
	}

	// Create empty buffer if nil
	for i := 1; window.CurrentBuffer == nil; i++ {
		buffer, err := CreateBuffer("New Buffer " + strconv.Itoa(i))
		if err == nil {
			window.CurrentBuffer = buffer
		}
	}

	// Create tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to initialize tcell: %s", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %s", err)
	}

	// Enable mouse
	screen.EnableMouse()

	// Set window screen field
	window.screen = screen

	// Try to set screen style to selected one
	if ok := SetCurrentStyle(screen, Config.SelectedStyle); !ok {
		// Try to set screen style to selected fallback one
		if ok := SetCurrentStyle(screen, Config.FallbackStyle); !ok {
			// Use hard-coded fallback style
			screen.SetStyle(tcell.StyleDefault.Foreground(CurrentStyle.BufferAreaFg).Background(CurrentStyle.BufferAreaBg))
			PrintMessage(&window, "Could not set style either to selected one nor to fallback one!")
		}
	}

	// Initialize top menu
	initTopMenu()

	return &window, nil
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
		drawBuffer(window)
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
}

func (window *Window) ProcessEvents() {
	// Poll event
	ev := window.screen.PollEvent()

	// Process event
	switch ev := ev.(type) {
	case *tcell.EventResize:
		window.screen.Sync()
		window.SyncBufferOffset()
	case *tcell.EventMouse:
		window.handleMouseInput(ev)
	case *tcell.EventKey:
		window.handleKeyInput(ev)
	}
}

func (window *Window) handleKeyInput(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyRight { // Navigation Keys
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to start of word
				// Set variable to one character right of current position
				endOfWord := pos + 1
				if endOfWord >= len(window.CurrentBuffer.Contents) {
					endOfWord = len(window.CurrentBuffer.Contents)
				}

				// Skip all spaces
				for endOfWord < len(window.CurrentBuffer.Contents) && unicode.IsSpace(rune(window.CurrentBuffer.Contents[endOfWord])) {
					endOfWord++
				}

				// Find end of word
				for endOfWord < len(window.CurrentBuffer.Contents) && !unicode.IsSpace(rune(window.CurrentBuffer.Contents[endOfWord])) {
					endOfWord++
				}

				window.SetCursorPos(endOfWord)
			} else {
				// Move cursor one character backwards
				window.SetCursorPos(window.CurrentBuffer.CursorPos + 1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
				if window.CurrentBuffer.Selection == nil {
					// Cancel cursor movement when creating selection without holding ctrl
					if ev.Modifiers()&tcell.ModCtrl == 0 {
						window.SetCursorPos(pos)
					}

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
			}
		}
	} else if ev.Key() == tcell.KeyLeft {
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to start of word
				// Set variable to one character left of current position
				startOfWord := pos - 1
				if startOfWord < 0 {
					startOfWord = 0
				}

				// Skip all spaces
				for startOfWord >= 0 && len(window.CurrentBuffer.Contents) != 0 && unicode.IsSpace(rune(window.CurrentBuffer.Contents[startOfWord])) {
					startOfWord--
				}

				// Find start of word
				for startOfWord >= 0 && len(window.CurrentBuffer.Contents) != 0 && !unicode.IsSpace(rune(window.CurrentBuffer.Contents[startOfWord])) {
					startOfWord--
				}

				// Move one character to the right
				startOfWord++

				window.SetCursorPos(startOfWord)
			} else {
				// Move cursor one character backwards
				window.SetCursorPos(window.CurrentBuffer.CursorPos - 1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
				if window.CurrentBuffer.Selection == nil {
					// Cancel cursor movement when creating selection without holding ctrl
					if ev.Modifiers()&tcell.ModCtrl == 0 {
						window.SetCursorPos(pos)
					}

					window.CurrentBuffer.Selection = &Selection{
						selectionStart: pos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
					}
					return
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CurrentBuffer.CursorPos
				}
			} else if window.CurrentBuffer.Selection != nil {
				// Unset selection
				window.CurrentBuffer.Selection = nil
				return
			}
		}
	} else if ev.Key() == tcell.KeyUp {
		if window.CursorMode == CursorModeBuffer {
			// Get original cursor position
			pos := window.CurrentBuffer.CursorPos

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to top of buffer
				window.SetCursorPos(0)
			} else {
				// Move cursor one line up
				x, y := window.GetCursorPos2D()
				window.SetCursorPos2D(x, y-1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: pos,
						selectionEnd:   window.CurrentBuffer.CursorPos,
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

			if ev.Modifiers()&tcell.ModCtrl != 0 {
				// Move cursor to bottom of buffer
				window.SetCursorPos(len(window.CurrentBuffer.Contents))
			} else {
				// Move cursor one line down
				x, y := window.GetCursorPos2D()
				window.SetCursorPos2D(x, y+1)
			}

			// Add to selection
			if ev.Modifiers()&tcell.ModShift != 0 {
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
	for _, keybinding := range Keybindings.Keybindings {
		if keybinding.IsPressed(ev) && slices.Index(keybinding.GetCursorModes(), window.CursorMode) != -1 {
			RunCommand(window, keybinding.Command)
			return
		}
	}

	// Typing
	if ev.Key() == tcell.KeyBackspace2 {
		if window.CursorMode == CursorModeBuffer {
			str := window.CurrentBuffer.Contents
			index := window.CurrentBuffer.CursorPos

			if window.CurrentBuffer.Selection != nil {
				edge1, edge2 := window.CurrentBuffer.GetSelectionEdges()
				if edge2 == len(window.CurrentBuffer.Contents) {
					edge2 = len(window.CurrentBuffer.Contents) - 1
				}

				str = str[:edge1] + str[edge2+1:]
				window.CurrentBuffer.Contents = str
				window.SetCursorPos(edge1)
				window.CurrentBuffer.Selection = nil
			} else if index != 0 {
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

			// Remove selected text
			if window.CurrentBuffer.Selection != nil {
				edge1, edge2 := window.CurrentBuffer.GetSelectionEdges()
				if edge2 == len(window.CurrentBuffer.Contents) {
					edge2 = len(window.CurrentBuffer.Contents) - 1
				}

				str = str[:edge1] + str[edge2+1:]
				window.CurrentBuffer.Contents = str
				window.SetCursorPos(edge1)
				window.CurrentBuffer.Selection = nil
			}

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

			// Remove selected text
			if window.CurrentBuffer.Selection != nil {
				edge1, edge2 := window.CurrentBuffer.GetSelectionEdges()
				if edge2 == len(window.CurrentBuffer.Contents) {
					edge2 = len(window.CurrentBuffer.Contents) - 1
				}

				str = str[:edge1] + str[edge2+1:]
				window.CurrentBuffer.Contents = str
				window.SetCursorPos(edge1)
				window.CurrentBuffer.Selection = nil
			}

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

			// Remove selected text
			if window.CurrentBuffer.Selection != nil {
				edge1, edge2 := window.CurrentBuffer.GetSelectionEdges()
				if edge2 == len(window.CurrentBuffer.Contents) {
					edge2 = len(window.CurrentBuffer.Contents) - 1
				}

				str = str[:edge1] + str[edge2+1:]
				window.CurrentBuffer.Contents = str
				window.SetCursorPos(edge1)
				window.CurrentBuffer.Selection = nil
			}

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

func (window *Window) handleMouseInput(ev *tcell.EventMouse) {
	mouseX, mouseY := ev.Position()

	// Left click was pressed
	if ev.Buttons() == tcell.Button1 {
		// Get last click time
		lastClickTime := time.UnixMilli(lastClick)
		// Ensure click was in buffer area
		x1, y1, x2, y2 := window.GetTextAreaDimensions()
		if mouseX >= x1 && mouseY >= y1 && mouseX <= x2 && mouseY <= y2 {
			currentX, currentY := window.GetCursorPos2D()
			bufferMouseX, bufferMouseY := window.AbsolutePosToCursorPos2D(mouseX, mouseY)
			if mouseHeld {
				// Add to selection
				if window.CurrentBuffer.Selection == nil {
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: window.CurrentBuffer.CursorPos,
						selectionEnd:   window.CursorPos2DToCursorPos(bufferMouseX, bufferMouseY),
					}

					// Set last click time
					lastClick = time.Now().UnixMilli()

					return
				} else {
					window.CurrentBuffer.Selection.selectionEnd = window.CursorPos2DToCursorPos(bufferMouseX, bufferMouseY)
				}
				// Prevent selecting dummy character at the end of the buffer
				if window.CurrentBuffer.Selection.selectionEnd >= len(window.CurrentBuffer.Contents) {
					window.CurrentBuffer.Selection.selectionEnd = len(window.CurrentBuffer.Contents) - 1
				}
			} else if currentX == bufferMouseX && currentY == bufferMouseY && window.CurrentBuffer.CursorPos < len(window.CurrentBuffer.Contents) && time.Since(lastClickTime).Milliseconds() < 300 {
				selectedText := window.CurrentBuffer.GetSelectedText()
				if window.CurrentBuffer.Selection == nil || strings.HasSuffix(selectedText, "\n") {
					// Select word
					startOfWord := window.CurrentBuffer.CursorPos
					endOfWord := window.CurrentBuffer.CursorPos

					// Find end of word
					for i := window.CurrentBuffer.CursorPos + 1; i < len(window.CurrentBuffer.Contents); i++ {
						currentRune := rune(window.CurrentBuffer.Contents[i])
						if unicode.IsLetter(currentRune) || unicode.IsDigit(currentRune) || currentRune == '_' {
							endOfWord++
						} else {
							break
						}
					}

					// Find start of word
					for i := window.CurrentBuffer.CursorPos - 1; i >= 0; i-- {
						currentRune := rune(window.CurrentBuffer.Contents[i])
						if unicode.IsLetter(currentRune) || unicode.IsDigit(currentRune) || currentRune == '_' {
							startOfWord--
						} else {
							break
						}
					}

					// Add to selection
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: startOfWord,
						selectionEnd:   endOfWord,
					}
				} else {
					// Select line
					startOfLine := window.CurrentBuffer.CursorPos
					endOfLine := window.CurrentBuffer.CursorPos

					// Find end of line
					for i := window.CurrentBuffer.CursorPos + 1; i < len(window.CurrentBuffer.Contents); i++ {
						currentLetter := window.CurrentBuffer.Contents[i]

						endOfLine++
						if currentLetter == '\n' {
							break
						}
					}

					// Find start of line
					for i := window.CurrentBuffer.CursorPos - 1; i >= 0; i-- {
						currentLetter := window.CurrentBuffer.Contents[i]
						if currentLetter != '\n' {
							startOfLine--
						} else {
							break
						}
					}

					// Add to selection
					window.CurrentBuffer.Selection = &Selection{
						selectionStart: startOfLine,
						selectionEnd:   endOfLine,
					}
				}

				// Set last click time
				lastClick = time.Now().UnixMilli()

				return
			} else {
				// Clear selection
				if window.CurrentBuffer.Selection != nil {
					window.CurrentBuffer.Selection = nil
				}
			}
			// Move cursor
			window.SetCursorPos2D(bufferMouseX, bufferMouseY)

			// Set last click time
			lastClick = time.Now().UnixMilli()
		}
		mouseHeld = true
	} else if ev.Buttons() == tcell.ButtonNone {
		if mouseHeld {
			mouseHeld = false
		}
	}
}

func (window *Window) Close() {
	window.closed = true
	err := window.screen.PostEvent(tcell.NewEventInterrupt(nil))
	if err != nil {
		return
	}
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
			for j := 0; j < Config.TabIndentation; j++ {
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

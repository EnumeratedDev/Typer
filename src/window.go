package main

import (
	"github.com/gdamore/tcell"
	"log"
	"maps"
	"slices"
)

type CursorMode uint8

const (
	CursorModeDisabled CursorMode = iota
	CursorModeBuffer
	CursorModeDropdown
	CursorModeMessageBar
)

type Window struct {
	ShowTopMenu   bool
	ShowLineIndex bool
	CursorMode    CursorMode

	textArea TextArea

	screen tcell.Screen
}

type TextArea struct {
	CursorPos     int
	CurrentBuffer *Buffer
}

func CreateWindow() (*Window, error) {
	window := Window{
		ShowTopMenu:   true,
		ShowLineIndex: true,
		CursorMode:    CursorModeBuffer,

		textArea: TextArea{
			CursorPos:     0,
			CurrentBuffer: nil,
		},

		screen: nil,
	}

	// Create empty buffer if nil
	window.textArea.CurrentBuffer = CreateBuffer("New File 1")

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

	// Set window screen field
	window.screen = screen

	// Initialize top menu
	initTopMenu()

	return &window, nil
}

func (window *Window) drawCurrentBuffer() {
	buffer := window.textArea.CurrentBuffer

	x, y := 0, 0
	if window.ShowTopMenu {
		y++
	}
	if window.ShowLineIndex {
		x += 3
	}

	width, _ := window.screen.Size()

	for _, r := range []rune(buffer.Contents) {
		if x >= width || r == '\n' {
			x = 0
			if window.ShowLineIndex {
				x += 3
			}
			y++
		}

		window.screen.SetContent(x, y, r, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.Color234))

		if r != '\n' {
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
	if window.textArea.CurrentBuffer != nil {
		window.drawCurrentBuffer()
	}

	// Draw message bar
	drawMessageBar(window)

	// Draw dropdowns
	drawDropdowns(window)

	// Draw cursor
	if window.CursorMode == CursorModeBuffer {
		window.screen.ShowCursor(window.GetAbsoluteCursorPos())
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
	case *tcell.EventKey:
		window.input(ev)
	}
}

func (window *Window) input(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyRight { // Navigation Keys
		if window.CursorMode == CursorModeBuffer {
			window.SetCursorPos(window.textArea.CursorPos + 1)
		}
	} else if ev.Key() == tcell.KeyLeft {
		if window.CursorMode == CursorModeBuffer {
			window.SetCursorPos(window.textArea.CursorPos - 1)
		}
	} else if ev.Key() == tcell.KeyUp {
		if window.CursorMode == CursorModeBuffer {
			x, y := window.GetCursorPos2D()
			window.SetCursorPos2D(x, y-1)
		} else if ActiveDropdown != nil {
			dropdown := ActiveDropdown
			dropdown.Selected--
			if dropdown.Selected < 0 {
				dropdown.Selected = 0
			}
		}
	} else if ev.Key() == tcell.KeyDown {
		if window.CursorMode == CursorModeBuffer {
			x, y := window.GetCursorPos2D()
			window.SetCursorPos2D(x, y+1)
		} else if ActiveDropdown != nil {
			dropdown := ActiveDropdown
			dropdown.Selected++
			if dropdown.Selected >= len(dropdown.Options) {
				dropdown.Selected = len(dropdown.Options) - 1
			}
		}
	} else if ev.Key() == tcell.KeyEscape {
		ClearDropdowns()
		window.CursorMode = CursorModeBuffer
	} else if ev.Key() == tcell.KeyCtrlC { // Close buffer key
		delete(Buffers, window.textArea.CurrentBuffer.Id)
		buffersSlice := slices.Collect(maps.Values(Buffers))
		if len(buffersSlice) == 0 {
			window.Close()
			return
		}
		window.textArea.CurrentBuffer = buffersSlice[0]
		window.SetCursorPos(0)
		ClearDropdowns()
		window.CursorMode = CursorModeBuffer
	} else if ev.Key() == tcell.KeyCtrlQ { // Exit key
		window.Close()
	} else if ev.Modifiers()&tcell.ModAlt != 0 { // Menu Bar
		for _, button := range TopMenuButtons {
			if ev.Rune() == button.Key {
				button.Action(window)
				break
			}
		}
	} else if ev.Key() == tcell.KeyBackspace2 { // Typing
		str := window.textArea.CurrentBuffer.Contents
		index := window.textArea.CursorPos

		if index != 0 {
			str = str[:index-1] + str[index:]
			window.textArea.CursorPos--
			window.textArea.CurrentBuffer.Contents = str
		}
	} else if ev.Key() == tcell.KeyTab {
		if ActiveDropdown != nil {
			return
		}

		str := window.textArea.CurrentBuffer.Contents
		index := window.textArea.CursorPos

		if index == len(str) {
			str += "\t"
		} else {
			str = str[:index] + "\t" + str[index:]
		}
		window.textArea.CursorPos++
		window.textArea.CurrentBuffer.Contents = str
	} else if ev.Key() == tcell.KeyEnter {
		if ActiveDropdown != nil {
			d := ActiveDropdown
			d.Action(d.Selected)
		} else {
			str := window.textArea.CurrentBuffer.Contents
			index := window.textArea.CursorPos

			if index == len(str) {
				str += "\n"
			} else {
				str = str[:index] + "\n" + str[index:]
			}
			window.textArea.CursorPos++
			window.textArea.CurrentBuffer.Contents = str
		}
	} else if ev.Key() == tcell.KeyRune {
		if ActiveDropdown != nil {
			return
		}

		str := window.textArea.CurrentBuffer.Contents
		index := window.textArea.CursorPos

		if index == len(str) {
			str += string(ev.Rune())
		} else {
			str = str[:index] + string(ev.Rune()) + str[index:]
		}
		window.textArea.CursorPos++
		window.textArea.CurrentBuffer.Contents = str
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
		x1 += 3
	}

	return x1, y1, x2, y2
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

	for i := 0; i < window.textArea.CursorPos; i++ {
		char := window.textArea.CurrentBuffer.Contents[i]
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
	window.textArea.CursorPos = position

	if window.textArea.CursorPos < 0 {
		window.textArea.CursorPos = 0
	}

	if window.textArea.CursorPos > len(window.textArea.CurrentBuffer.Contents) {
		window.textArea.CursorPos = len(window.textArea.CurrentBuffer.Contents)
	}
}

func (window *Window) SetCursorPos2D(x, y int) {
	// Ensure x and y are positive
	x = max(x, 0)
	y = max(y, 0)

	// Set cursor position to 0 buffer is empty
	if len(window.textArea.CurrentBuffer.Contents) == 0 {
		window.SetCursorPos(0)
		return
	}

	// Create line slice from buffer contents
	lines := make([]struct {
		charIndex int
		str       string
	}, 0)

	var str string
	for i, char := range window.textArea.CurrentBuffer.Contents {
		str += string(char)
		if char == '\n' || i == len(window.textArea.CurrentBuffer.Contents)-1 {
			lines = append(lines, struct {
				charIndex int
				str       string
			}{charIndex: i - len(str) + 1, str: str})
			str = ""
		}
	}

	// Append extra character or line
	if window.textArea.CurrentBuffer.Contents[len(window.textArea.CurrentBuffer.Contents)-1] == '\n' {
		lines = append(lines, struct {
			charIndex int
			str       string
		}{charIndex: len(window.textArea.CurrentBuffer.Contents), str: " "})
	} else {
		lines[len(lines)-1].str += " "
	}

	// Limit x and y
	y = min(y, len(lines)-1)
	x = min(x, len(lines[y].str)-1)

	window.SetCursorPos(lines[y].charIndex + x)
}

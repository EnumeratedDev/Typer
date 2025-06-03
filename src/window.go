package main

import (
	"github.com/gdamore/tcell"
	"log"
)

type Window struct {
	ShowTopMenu   bool
	ShowLineIndex bool

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

		textArea: TextArea{
			CursorPos:     0,
			CurrentBuffer: nil,
		},

		screen: nil,
	}

	// Create empty buffer if nil
	window.textArea.CurrentBuffer = CreateBuffer("New File")

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

	// Draw cursor
	window.screen.ShowCursor(window.GetAbsoluteCursorPos())

	// Update screen
	window.screen.Show()

	// Poll event
	ev := window.screen.PollEvent()

	// Process event
	switch ev := ev.(type) {
	case *tcell.EventResize:
		window.screen.Sync()
	case *tcell.EventKey:
		// Navigation Keys
		if ev.Key() == tcell.KeyRight {
			window.SetCursorPos(window.textArea.CursorPos + 1)
		} else if ev.Key() == tcell.KeyLeft {
			window.SetCursorPos(window.textArea.CursorPos - 1)
		} else if ev.Key() == tcell.KeyUp {
			x, y := window.GetCursorPos2D()
			window.SetCursorPos2D(x, y-1)
		} else if ev.Key() == tcell.KeyDown {
			x, y := window.GetCursorPos2D()
			window.SetCursorPos2D(x, y+1)
		}

		// Exit key
		if ev.Key() == tcell.KeyCtrlC {
			window.Close()
		}

		// Typing
		if ev.Key() == tcell.KeyBackspace2 {
			str := window.textArea.CurrentBuffer.Contents
			index := window.textArea.CursorPos

			if index != 0 {
				str = str[:index-1] + str[index:]
				window.textArea.CursorPos--
				window.textArea.CurrentBuffer.Contents = str
			}
		} else if ev.Key() == tcell.KeyTab {
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
			str := window.textArea.CurrentBuffer.Contents
			index := window.textArea.CursorPos

			if index == len(str) {
				str += "\n"
			} else {
				str = str[:index] + "\n" + str[index:]
			}
			window.textArea.CursorPos++
			window.textArea.CurrentBuffer.Contents = str
		} else if ev.Key() == tcell.KeyRune {
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
	x = max(x, 0)
	y = max(y, 0)

	lines := make([]struct {
		charIndex int
		str       string
	}, 0)

	var str string
	for i, char := range window.textArea.CurrentBuffer.Contents {
		str += string(char)
		if char == '\n' || i == len(window.textArea.CurrentBuffer.Contents) {
			lines = append(lines, struct {
				charIndex int
				str       string
			}{charIndex: i - len(str) + 1, str: str})
			str = ""
		}
	}

	y = min(y, len(lines))

	if y == len(lines) {
		x = 0
		window.SetCursorPos(lines[y-1].charIndex + len(lines[y-1].str) + 1)
	} else {
		x = min(x, len(lines[y].str)-1)
		window.SetCursorPos(lines[y].charIndex + x)
	}

}

package main

import (
	"github.com/gdamore/tcell/v2"
)

type TyperInputRequest struct {
	Text         string
	input        string
	cursorPos    int
	inputChannel chan string
}

var inputHistory = make([]string, 0)
var currentInputRequest *TyperInputRequest

func RequestInput(window *Window, text string, defaultInput string) chan string {
	request := &TyperInputRequest{
		Text:         text,
		input:        defaultInput,
		cursorPos:    0,
		inputChannel: make(chan string),
	}

	currentInputRequest = request

	window.CursorMode = CursorModeInputBar

	_ = window.screen.PostEvent(tcell.NewEventInterrupt(nil))

	return request.inputChannel
}

func IsRequestingInput() bool {
	return currentInputRequest != nil
}

func drawInputBar(window *Window) {
	if currentInputRequest == nil {
		return
	}

	screen := window.screen

	inputBarStyle := tcell.StyleDefault.Background(CurrentStyle.InputBarBg).Foreground(CurrentStyle.InputBarFg)

	sizeX, sizeY := screen.Size()

	// Draw bar
	for x := 0; x < sizeX; x++ {
		char := ' '
		screen.SetContent(x, sizeY-1, char, nil, inputBarStyle)
	}

	// Write text
	for x := 0; x < len(currentInputRequest.Text); x++ {
		screen.SetContent(x, sizeY-1, rune(currentInputRequest.Text[x]), nil, inputBarStyle)
	}
	for x := 0; x < len(currentInputRequest.input); x++ {
		screen.SetContent(x+len(currentInputRequest.Text)+1, sizeY-1, rune(currentInputRequest.input[x]), nil, inputBarStyle)
	}
}

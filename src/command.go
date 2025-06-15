package main

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
)

type Command struct {
	cmd          string
	run          func(window *Window, args ...string)
	autocomplete func(window *Window, args ...string) []string
}

var commands = make(map[string]*Command)

func initCommands() {
	// Setup commands
	copyCmd := Command{
		cmd: "copy",
		run: func(window *Window, args ...string) {
			// Copy text from buffer
			copiedText, copyingMethod := window.CurrentBuffer.CopyText()

			// Put copied text to clipboard
			window.Clipboard = copiedText

			// Send appropriate message depending on copying method
			if copyingMethod == 0 {
				PrintMessage(window, "Copied line to clipboard.")
			} else {
				PrintMessage(window, "Copied selection to clipboard.")
			}
		},
	}

	pasteCmd := Command{
		cmd: "paste",
		run: func(window *Window, args ...string) {
			if window.Clipboard != "" {
				window.CurrentBuffer.PasteText(window, window.Clipboard)
				PrintMessage(window, "Pasted text to buffer.")
			}
		},
	}

	saveCmd := Command{
		cmd: "save",
		run: func(window *Window, args ...string) {
			if !window.CurrentBuffer.canSave {
				PrintMessage(window, "Cannot save buffer!")
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
		},
		autocomplete: func(window *Window, args ...string) []string {
			return nil
		},
	}

	openCmd := Command{
		cmd: "open",
		run: func(window *Window, args ...string) {
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
		},
	}

	reloadCmd := Command{
		cmd: "reload",
		run: func(window *Window, args ...string) {
			err := window.CurrentBuffer.Load()
			if err != nil {
				log.Fatalf("Could not reload buffer: %s", err)
			}

			window.SetCursorPos(window.CurrentBuffer.CursorPos)
			PrintMessage(window, "Buffer reloaded.")
		},
	}

	prevBufferCmd := Command{
		cmd: "prev-buffer",
		run: func(window *Window, args ...string) {
			if window.CursorMode != CursorModeBuffer {
				return
			}

			index := slices.Index(Buffers, window.CurrentBuffer)

			index--
			if index < 0 {
				index = 0
			}

			window.CurrentBuffer = Buffers[index]
			PrintMessage(window, fmt.Sprintf("Set current buffer to '%s'.", window.CurrentBuffer.Name))
		},
	}

	nextBufferCmd := Command{
		cmd: "next-buffer",
		run: func(window *Window, args ...string) {
			if window.CursorMode != CursorModeBuffer {
				return
			}

			index := slices.Index(Buffers, window.CurrentBuffer)

			index++
			if index >= len(Buffers) {
				index = len(Buffers) - 1
			}

			window.CurrentBuffer = Buffers[index]
			PrintMessage(window, fmt.Sprintf("Set current buffer to '%s'.", window.CurrentBuffer.Name))
		},
	}

	newBufferCmd := Command{
		cmd: "new-buffer",
		run: func(window *Window, args ...string) {
			for i := 1; true; i++ {
				buffer, err := CreateBuffer("New Buffer " + strconv.Itoa(i))
				if err == nil {
					window.CurrentBuffer = buffer
					break
				}
			}

			window.CursorMode = CursorModeBuffer
			PrintMessage(window, fmt.Sprintf("New buffer created with the name '%s'.", window.CurrentBuffer.Name))
		},
	}

	closeBufferCmd := Command{
		cmd: "close-buffer",
		run: func(window *Window, args ...string) {
			bufferIndex := slices.Index(Buffers, window.CurrentBuffer)
			Buffers = DeleteFromSlice(Buffers, bufferIndex)
			if len(Buffers) == 0 {
				window.Close()
				return
			}
			if bufferIndex >= len(Buffers) {
				window.CurrentBuffer = Buffers[bufferIndex-1]
			} else {
				window.CurrentBuffer = Buffers[bufferIndex]
			}
			window.CursorMode = CursorModeBuffer
			PrintMessage(window, "Buffer closed.")
		},
	}

	menuFileCmd := Command{
		cmd: "menu-file",
		run: func(window *Window, args ...string) {
			for _, button := range TopMenuButtons {
				if button.Name == "File" {
					button.Action(window)
					break
				}
			}
		},
	}

	menuEditCmd := Command{
		cmd: "menu-edit",
		run: func(window *Window, args ...string) {
			for _, button := range TopMenuButtons {
				if button.Name == "Edit" {
					button.Action(window)
					break
				}
			}
		},
	}

	menuBuffersCmd := Command{
		cmd: "menu-buffers",
		run: func(window *Window, args ...string) {
			for _, button := range TopMenuButtons {
				if button.Name == "Buffers" {
					button.Action(window)
					break
				}
			}
		},
	}

	quitCmd := Command{
		cmd: "quit",
		run: func(window *Window, args ...string) {
			window.Close()
			window.CursorMode = CursorModeBuffer
		},
	}

	executeCmd := Command{
		cmd: "execute",
		run: func(window *Window, args ...string) {
			inputChannel := RequestInput(window, "Run:", "")

			go func() {
				input := strings.TrimSpace(<-inputChannel)

				if input == "" {
					return
				}

				var arguments []string

				builder := &strings.Builder{}
				quoted := false
				for _, r := range input {
					if r == '"' {
						quoted = !quoted
					} else if !quoted && r == ' ' {
						arguments = append(arguments, builder.String())
						builder.Reset()
					} else {
						builder.WriteRune(r)
					}
				}
				if builder.Len() > 0 {
					arguments = append(arguments, builder.String())
				}

				window.CursorMode = CursorModeBuffer

				if len(arguments) == 1 {
					RunCommand(window, arguments[0])
				} else {
					RunCommand(window, arguments[0], arguments[1:]...)
				}
			}()
		},
	}

	// Register commands
	commands["copy"] = &copyCmd
	commands["paste"] = &pasteCmd
	commands["save"] = &saveCmd
	commands["open"] = &openCmd
	commands["reload"] = &reloadCmd
	commands["prev-buffer"] = &prevBufferCmd
	commands["next-buffer"] = &nextBufferCmd
	commands["new-buffer"] = &newBufferCmd
	commands["close-buffer"] = &closeBufferCmd
	commands["menu-file"] = &menuFileCmd
	commands["menu-edit"] = &menuEditCmd
	commands["menu-buffers"] = &menuBuffersCmd
	commands["quit"] = &quitCmd
	commands["execute"] = &executeCmd
}

func RunCommand(window *Window, cmd string, args ...string) bool {
	if command, ok := commands[cmd]; ok {
		command.run(window, args...)
		return true
	} else {
		PrintMessage(window, fmt.Sprintf("Could not find command '%s'!", cmd))
		return false
	}
}

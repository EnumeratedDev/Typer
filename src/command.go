package main

import (
	"fmt"
	"log"
	"maps"
	"slices"
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
		},
	}

	pasteCmd := Command{
		cmd: "paste",
		run: func(window *Window, args ...string) {
			str := window.CurrentBuffer.Contents
			index := window.CurrentBuffer.CursorPos

			if index == len(str) {
				str += window.Clipboard
			} else {
				str = str[:index] + window.Clipboard + str[index:]
			}
			window.CurrentBuffer.CursorPos += len(window.Clipboard)
			window.CurrentBuffer.Contents = str
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

			buffers := slices.Collect(maps.Values(Buffers))
			index := slices.Index(buffers, window.CurrentBuffer)

			index--
			if index < 0 {
				index = 0
			}

			window.CurrentBuffer = buffers[index]
		},
	}

	nextBufferCmd := Command{
		cmd: "next-buffer",
		run: func(window *Window, args ...string) {
			if window.CursorMode != CursorModeBuffer {
				return
			}

			buffers := slices.Collect(maps.Values(Buffers))
			index := slices.Index(buffers, window.CurrentBuffer)

			index++
			if index >= len(buffers) {
				index = len(buffers) - 1
			}

			window.CurrentBuffer = buffers[index]
		},
	}

	newBufferCmd := Command{
		cmd: "new-buffer",
		run: func(window *Window, args ...string) {
			number := 1
			for _, buffer := range Buffers {
				if strings.HasPrefix(buffer.Name, "New File ") {
					number++
				}
			}
			buffer := CreateBuffer(fmt.Sprintf("New File %d", number))
			window.CurrentBuffer = buffer
			window.CursorMode = CursorModeBuffer
		},
	}

	closeBufferCmd := Command{
		cmd: "close-buffer",
		run: func(window *Window, args ...string) {
			delete(Buffers, window.CurrentBuffer.Id)
			buffersSlice := slices.Collect(maps.Values(Buffers))
			if len(buffersSlice) == 0 {
				window.Close()
				return
			}
			window.CurrentBuffer = buffersSlice[0]
			window.CursorMode = CursorModeBuffer
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

package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"os"
	"path/filepath"
	"strings"
)

type Buffer struct {
	Name     string
	Contents string

	CursorPos        int
	OffsetX, OffsetY int

	Selection *Selection

	canSave  bool
	filename string
}

type Selection struct {
	selectionStart int
	selectionEnd   int
}

var Buffers = make([]*Buffer, 0)

func GetBufferByName(name string) *Buffer {
	for _, buffer := range Buffers {
		if buffer.Name == name {
			return buffer
		}
	}
	return nil
}

func GetBufferByFilename(filename string) *Buffer {
	for _, buffer := range Buffers {
		if buffer.filename == filename {
			return buffer
		}
	}
	return nil
}

func drawBuffer(window *Window) {
	buffer := window.CurrentBuffer

	x, y, _, _ := window.GetTextAreaDimensions()

	bufferX, bufferY, _, _ := window.GetTextAreaDimensions()

	for i, r := range buffer.Contents + " " {
		if x-buffer.OffsetX >= bufferX && y-buffer.OffsetY >= bufferY {
			// Default style
			style := tcell.StyleDefault.Background(CurrentStyle.BufferAreaBg).Foreground(CurrentStyle.BufferAreaFg)

			// Change background if under cursor
			if i == buffer.CursorPos {
				style = style.Background(CurrentStyle.BufferAreaSel)
			}

			// Change background if selected
			if buffer.Selection != nil {
				if edge1, edge2 := buffer.GetSelectionEdges(); i >= edge1 && i <= edge2 {
					style = style.Background(CurrentStyle.BufferAreaSel)

					// Show selection on entire tab space
					if r == '\t' {
						for j := 0; j < int(Config.TabIndentation); j++ {
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
			x += int(Config.TabIndentation)
		} else {
			x++
		}
	}
}

func (buffer *Buffer) Load() error {
	// Do not load if canSave is false or filename is not set
	if !buffer.canSave || buffer.filename == "" {
		return nil
	}

	content, err := os.ReadFile(buffer.filename)
	if err != nil {
		return err
	}

	buffer.Contents = string(content)
	return nil
}

func (buffer *Buffer) Save() error {
	// Do not save if canSave is false or filename is not set
	if !buffer.canSave || buffer.filename == "" {
		return nil
	}

	// Append new line character at end of buffer contents if not present
	if buffer.Contents[len(buffer.Contents)-1] != '\n' {
		buffer.Contents += "\n"
	}

	err := os.WriteFile(buffer.filename, []byte(buffer.Contents), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (buffer *Buffer) GetSelectionEdges() (int, int) {
	if buffer.Selection == nil {
		return -1, -1
	}

	if buffer.Selection.selectionStart < buffer.Selection.selectionEnd {
		return buffer.Selection.selectionStart, buffer.Selection.selectionEnd
	} else {
		return buffer.Selection.selectionEnd, buffer.Selection.selectionStart
	}
}

func (buffer *Buffer) GetSelectedText() string {
	if buffer.Selection == nil {
		return ""
	}

	if len(buffer.Contents) == 0 {
		return ""
	}

	start := buffer.Selection.selectionStart
	end := buffer.Selection.selectionEnd

	if start >= len(buffer.Contents) {
		start = len(buffer.Contents) - 1
	}
	if end >= len(buffer.Contents) {
		end = len(buffer.Contents) - 1
	}

	if start <= end {
		return buffer.Contents[start : end+1]
	} else {
		return buffer.Contents[end : start+1]
	}
}

func GetOpenFileBuffer(filename string) *Buffer {
	// Replace tilde with home directory
	if filename != "~" && strings.HasPrefix(filename, "~/") {
		homedir, err := os.UserHomeDir()

		if err != nil {
			return nil
		}

		filename = filepath.Join(homedir, filename[2:])
	}

	// Get absolute path of file
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return nil
	}

	for _, buffer := range Buffers {
		if buffer.filename == absFilename {
			return buffer
		}
	}

	return nil
}

func CreateFileBuffer(filename string, openNonExistentFile bool) (*Buffer, error) {
	// Replace tilde with home directory
	if filename != "~" && strings.HasPrefix(filename, "~/") {
		homedir, err := os.UserHomeDir()

		if err != nil {
			return nil, err
		}

		filename = filepath.Join(homedir, filename[2:])
	}

	// Get absolute path of file
	abs, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(abs)
	if !openNonExistentFile {
		if err != nil {
			return nil, err
		}

		if !stat.Mode().IsRegular() {
			return nil, fmt.Errorf("%s is not a regular file", filename)
		}
	}

	if GetBufferByName(filename) != nil {
		return nil, fmt.Errorf("a buffer with the name (%s) is already open", filename)
	}

	if GetBufferByFilename(abs) != nil {
		return nil, fmt.Errorf("%s is already open in another buffer", filename)
	}

	buffer := Buffer{
		Name:      filename,
		Contents:  "",
		CursorPos: 0,
		canSave:   true,
		filename:  abs,
	}

	// Load file contents if no error was encountered in stat call
	if err == nil {
		err = buffer.Load()

		if err != nil {
			return nil, err
		}
	}

	Buffers = append(Buffers, &buffer)

	return &buffer, nil
}

func CreateBuffer(bufferName string) (*Buffer, error) {
	buffer := Buffer{
		Name:      bufferName,
		Contents:  "",
		CursorPos: 0,
		canSave:   true,
		filename:  "",
	}

	if GetBufferByName(bufferName) != nil {
		return nil, fmt.Errorf("a buffer with the name (%s) is already open", bufferName)
	}

	Buffers = append(Buffers, &buffer)

	return &buffer, nil
}

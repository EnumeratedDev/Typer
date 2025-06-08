package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Buffer struct {
	Id        int
	Name      string
	Contents  string
	CursorPos int

	canSave  bool
	filename string
}

var Buffers = make(map[int]*Buffer)
var LastBufferId int

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

	buffer := Buffer{
		Id:        LastBufferId + 1,
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

	Buffers[buffer.Id] = &buffer
	LastBufferId++

	return &buffer, nil
}

func CreateBuffer(bufferName string) *Buffer {
	buffer := Buffer{
		Id:        LastBufferId + 1,
		Name:      bufferName,
		Contents:  "",
		CursorPos: 0,
		canSave:   true,
		filename:  "",
	}

	Buffers[buffer.Id] = &buffer
	LastBufferId++

	return &buffer
}

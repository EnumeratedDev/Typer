package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Buffer struct {
	Id       int
	Name     string
	Contents string
	LoadFunc func(buffer *Buffer) error
	SaveFunc func(buffer *Buffer) error
}

var Buffers = make(map[int]*Buffer)
var LastBufferId int

func CreateFileBuffer(filename string) (*Buffer, error) {
	// Replace tilde with home directory
	if filename != "~" && strings.HasPrefix(filename, "~/") {
		homedir, err := os.UserHomeDir()

		if err != nil {
			return nil, err
		}

		filename = filepath.Join(homedir, filename[2:])
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if !stat.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", filename)
	}

	buffer := Buffer{
		Id:       LastBufferId + 1,
		Name:     filename,
		Contents: "",
		LoadFunc: func(buffer *Buffer) error {
			content, err := os.ReadFile(filename)
			if err != nil {
				return err
			}

			buffer.Contents = string(content)
			return nil
		},
		SaveFunc: func(buffer *Buffer) error {
			err := os.WriteFile(filename, []byte(buffer.Contents), 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}

	err = buffer.LoadFunc(&buffer)
	if err != nil {
		return nil, err
	}

	Buffers[buffer.Id] = &buffer
	LastBufferId++

	return &buffer, nil
}

func CreateBuffer(bufferName string) *Buffer {
	buffer := Buffer{
		Id:       LastBufferId + 1,
		Name:     bufferName,
		Contents: "",
		LoadFunc: func(buffer *Buffer) error { return nil },
		SaveFunc: func(buffer *Buffer) error { return nil },
	}

	Buffers[buffer.Id] = &buffer
	LastBufferId++

	return &buffer
}

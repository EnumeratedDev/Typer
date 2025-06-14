package main

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
)

type TyperConfig struct {
	SelectedStyle     string `yaml:"selected_style,omitempty"`
	FallbackStyle     string `yaml:"fallback_style,omitempty"`
	ShowTopMenu       bool   `yaml:"show_top_menu,omitempty"`
	ShowLineIndex     bool   `yaml:"show_line_index,omitempty"`
	ExtendLineIndex   bool   `yaml:"extend_line_index,omitempty"`
	BufferInfoMessage string `yaml:"buffer_info_message,omitempty"`
	TabIndentation    int    `yaml:"tab_indentation,omitempty"`
}

var Config TyperConfig

func readConfig() {
	Config = TyperConfig{
		SelectedStyle:     "default",
		FallbackStyle:     "default-fallback",
		ShowTopMenu:       true,
		ShowLineIndex:     true,
		ExtendLineIndex:   false,
		BufferInfoMessage: "File: %f Cursor: (%x, %y, %p) Chars: %c",
		TabIndentation:    4,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not get home directory: %s", err)
	}

	if _, err := os.Stat(path.Join(homeDir, ".config/typer/config.yml")); err == nil {
		data, err := os.ReadFile(path.Join(homeDir, ".config/typer/config.yml"))
		if err != nil {
			log.Fatalf("Could not read config.yml: %s", err)
		}
		err = yaml.Unmarshal(data, &Config)
		if err != nil {
			log.Fatalf("Could not unmarshal config.yml: %s", err)
		}
	} else if _, err := os.Stat("/etc/typer/config.yml"); err == nil {
		reader, err := os.Open("/etc/typer/config.yml")
		if err != nil {
			log.Fatalf("Could not read config.yml: %s", err)
		}
		err = yaml.NewDecoder(reader).Decode(&Config)
		if err != nil {
			log.Fatalf("Could not read config.yml: %s", err)
		}
		reader.Close()
	}

	// Validate config options
	if Config.TabIndentation < 1 {
		Config.TabIndentation = 1
	}
}

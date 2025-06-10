package main

import (
	"github.com/gdamore/tcell/v2"
	"strings"
)

type Keybinding struct {
	keybind     string
	cursorModes []CursorMode
	command     string
}

var Keybinds = make([]Keybinding, 0)

func initKeybindings() {
	// Add key bindings
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-Q",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "quit",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-C",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "copy",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-V",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "paste",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-S",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "save",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-O",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "open",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-R",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "reload",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "PgUp",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "prev-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "PgDn",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "next-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-N",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "new-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Delete",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "close-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "Ctrl-Q",
		cursorModes: []CursorMode{CursorModeBuffer},
		command:     "quit",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "F1",
		cursorModes: []CursorMode{CursorModeBuffer, CursorModeDropdown},
		command:     "menu-file",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "F2",
		cursorModes: []CursorMode{CursorModeBuffer, CursorModeDropdown},
		command:     "menu-edit",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind:     "F3",
		cursorModes: []CursorMode{CursorModeBuffer, CursorModeDropdown},
		command:     "menu-buffers",
	})
}

func (keybind *Keybinding) IsPressed(ev *tcell.EventKey) bool {
	keys := strings.SplitN(keybind.keybind, "+", 2)

	if len(keys) == 0 {
		return false
	} else if len(keys) == 1 {
		for k, v := range tcell.KeyNames {
			if k != tcell.KeyRune {
				if keybind.keybind == v {
					if ev.Key() == k {
						return true
					}
				}
			} else {
				if keybind.keybind == string(ev.Rune()) {
					return true
				}
			}
		}
	} else {
		modKey := keys[0]
		key := keys[1]

		switch modKey {
		case "Shift":
			if ev.Modifiers() != tcell.ModShift {
				return false
			}
		case "Alt":
			if ev.Modifiers() != tcell.ModAlt {
				return false
			}
		case "Ctrl":
			if ev.Modifiers() != tcell.ModCtrl {
				return false
			}
		case "Meta":
			if ev.Modifiers() != tcell.ModMeta {
				return false
			}
		}

		for k, v := range tcell.KeyNames {
			if k != tcell.KeyRune {
				if key == v {
					if ev.Key() == k {
						return true
					}
				}
			}
		}

		if strings.ToLower(key) == string(ev.Rune()) {
			return true
		}
	}

	return false
}

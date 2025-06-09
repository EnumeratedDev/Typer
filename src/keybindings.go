package main

import (
	"github.com/gdamore/tcell/v2"
	"strings"
)

type Keybinding struct {
	keybind string
	command string
}

var Keybinds = make([]Keybinding, 0)

func initKeybindings() {
	// Add key bindings
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-Q",
		command: "quit",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-C",
		command: "copy",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-V",
		command: "paste",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-S",
		command: "save",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-O",
		command: "open",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-R",
		command: "reload",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "PgUp",
		command: "prev-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "PgDn",
		command: "next-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-N",
		command: "new-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Delete",
		command: "close-buffer",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "Ctrl-Q",
		command: "quit",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "F1",
		command: "menu-file",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "F2",
		command: "menu-edit",
	})
	Keybinds = append(Keybinds, Keybinding{
		keybind: "F3",
		command: "menu-buffers",
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

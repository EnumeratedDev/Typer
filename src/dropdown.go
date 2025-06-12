package main

import (
	"github.com/gdamore/tcell/v2"
)

type Dropdown struct {
	Selected   int
	Options    []string
	PosX, PosY int
	Width      int
	Action     func(int)
}

var dropdowns = make([]*Dropdown, 0)
var ActiveDropdown *Dropdown

func CreateDropdownMenu(options []string, posX, posY, dropdownWidth int, action func(int)) *Dropdown {
	if len(options) == 0 {
		return nil
	}

	width := 0

	if dropdownWidth <= 0 {
		for _, option := range options {
			if len(option) > width {
				width = len(option)
			}
		}
	}

	d := &Dropdown{
		Selected: 0,
		Options:  options,
		PosX:     posX,
		PosY:     posY,
		Width:    width,
		Action:   action,
	}

	dropdowns = append(dropdowns, d)

	return d
}

func ClearDropdowns() {
	dropdowns = make([]*Dropdown, 0)
	ActiveDropdown = nil
}

func drawDropdowns(window *Window) {
	dropdownStyle := tcell.StyleDefault.Background(CurrentStyle.DropdownBg).Foreground(CurrentStyle.DropdownFg)
	for _, d := range dropdowns {
		drawBox(window.screen, d.PosX, d.PosY, d.PosX+d.Width+1, d.PosY+len(d.Options)+1, dropdownStyle)
		line := 1
		for i, option := range d.Options {
			if d.Selected == i {
				drawText(window.screen, d.PosX+1, d.PosY+line, d.PosX+d.Width+1, d.PosY+line, dropdownStyle.Background(CurrentStyle.DropdownSel), option)
			} else {
				drawText(window.screen, d.PosX+1, d.PosY+line, d.PosX+d.Width+1, d.PosY+line, dropdownStyle, option)
			}

			line++
		}
	}
}

package main

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/oxplot/raspberrypi-archlinux-installer/disk"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	title = "Raspberry Pi Arch Linux Installer"
)

var (
	sdCards = widget.NewSelect(nil, nil)
)

func refreshSDCards() {
Outer:
	for {
		time.Sleep(time.Second)
		dd, err := disk.Get()
		if err != nil {
			log.Printf("error: cannot get list of disks: %s", err)
			continue
		}

		opts := make([]string, len(dd))
		for i, d := range dd {
			opts[i] = d.Name()
		}
		sort.Slice(opts, func(i, j int) bool {
			return strings.ToLower(opts[i]) < strings.ToLower(opts[j])
		})
		sdCards.Options = opts
		for _, o := range opts {
			if o == sdCards.Selected {
				continue Outer
			}
		}
		sdCards.ClearSelected()
	}
}

func main() {
	log.SetFlags(0)
	a := app.New()
	w := a.NewWindow(title)

	install := widget.NewButton("Install", func() {
	})
	install.Disable()
	wifiName := widget.NewEntry()
	wifiPassword := widget.NewPasswordEntry()
	wifiEnable := widget.NewCheck("Enable WiFi", func(checked bool) {
		if checked {
			wifiName.Enable()
			wifiPassword.Enable()
		} else {
			wifiName.Disable()
			wifiPassword.Disable()
		}
	})
	sdCards.OnChanged = func(s string) {
		if sdCards.SelectedIndex() == -1 {
			install.Disable()
		} else {
			install.Enable()
		}
	}

	w.SetContent(container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("SD Card:"), nil, sdCards),
		widget.NewSeparator(),
		wifiEnable,
		container.NewBorder(nil, nil, widget.NewLabel("WiFi Name:"), nil, wifiName),
		container.NewBorder(nil, nil, widget.NewLabel("WiFi Password:"), nil, wifiPassword),
		widget.NewSeparator(),
		layout.NewSpacer(),
		install,
	)))

	wifiEnable.OnChanged(false)

	go refreshSDCards()
	w.ShowAndRun()
}

package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/oxplot/raspberrypi-archlinux-installer/disk"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	title      = "Raspberry Pi Arch Linux Installer"
	confirmMsg = "You are about to DESTROY ALL DATA on\n\n%s\n\nThere is no going back.\nAre you sure?"
)

var (
	mainWin fyne.Window
	sdCards = widget.NewSelect(nil, nil)

	disks = struct {
		mu    sync.Mutex
		items []disk.Disk
	}{}
)

func refreshSDCards() {
Outer:
	for {
		time.Sleep(time.Second)

		ds, err := disk.Get()
		if err != nil {
			log.Printf("error: cannot get list of disks: %s", err)
		}

		sort.Slice(ds, func(i, j int) bool {
			return strings.ToLower(ds[i].Name()) < strings.ToLower(ds[j].Name())
		})
		names := make([]string, len(ds))
		for i, d := range ds {
			names[i] = d.Name()
		}

		disks.mu.Lock()
		disks.items = ds
		sdCards.Options = names
		for _, o := range names {
			if o == sdCards.Selected {
				disks.mu.Unlock()
				continue Outer
			}
		}
		sdCards.ClearSelected()
		disks.mu.Unlock()
	}
}

func install(d disk.Disk) error {
	cancelled := make(chan bool)
	prog := widget.NewProgressBar()
	prog.Min, prog.Max = 0, 100
	progDiag := dialog.NewCustom("Installing", "Cancel", prog, mainWin)
	progDiag.SetOnClosed(func() {
		close(cancelled)
	})
	progDiag.Show()

	<-cancelled

	return fmt.Errorf("shit went to hell")
}

func main() {
	log.SetFlags(0)
	a := app.New()
	mainWin = a.NewWindow(title)

	install := widget.NewButton("Install", func() {
		disks.mu.Lock()
		selIdx := sdCards.SelectedIndex()
		if selIdx == -1 {
			disks.mu.Unlock()
			return
		}
		d := disks.items[selIdx]
		disks.mu.Unlock()

		dialog.ShowConfirm("Are you sure?", fmt.Sprintf(confirmMsg, d.Name()), func(y bool) {
			if !y {
				return
			}
			go func() {
				if err := install(d); err != nil {
					dialog.ShowInformation("Error", err.Error(), mainWin)
				}
			}()
		}, mainWin)
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
	wifiEnable.OnChanged(false)

	sdCards.OnChanged = func(s string) {
		if s == "" {
			install.Disable()
		} else {
			install.Enable()
		}
	}

	mainWin.SetContent(container.NewPadded(container.NewVBox(
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

	contentSize := mainWin.Content().Size()
	mainWin.Resize(fyne.Size{contentSize.Width * 2, contentSize.Height * 2})

	go refreshSDCards()
	mainWin.ShowAndRun()
}

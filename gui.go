package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/docker/go-units"

	"github.com/oxplot/raspberrypi-archlinux-installer/disk"
)

var (
	mainWin  fyne.Window
	diskList = widget.NewSelect(nil, nil)

	disks = struct {
		mu    sync.Mutex
		items []disk.Disk
	}{}
)

func refreshDiskList() {
Outer:
	for {
		time.Sleep(time.Second)

		ds, err := disk.Get()
		if err != nil {
			log.Printf("error: cannot get list of disks: %s", err)
		}

		// Filter small drives
		n := 0
		for _, d := range ds {
			if d.Size() >= imgSize+configSize {
				ds[n] = d
				n++
			}
		}
		ds = ds[:n]

		sort.Slice(ds, func(i, j int) bool {
			return strings.ToLower(ds[i].Name()) < strings.ToLower(ds[j].Name())
		})
		names := make([]string, len(ds))
		for i, d := range ds {
			names[i] = fmt.Sprintf("%s %s (%s)", d.Name(), units.BytesSize(float64(d.Size())), d.Path())
		}

		disks.mu.Lock()
		disks.items = ds
		diskList.Options = names
		for _, o := range names {
			if o == diskList.Selected {
				disks.mu.Unlock()
				continue Outer
			}
		}
		diskList.ClearSelected()
		disks.mu.Unlock()
	}
}

func installImgWithGUI(d disk.Disk, cfg imgConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelling := dialog.NewInformation("Cancelling", "Cancelling ...", mainWin)
	prog := widget.NewProgressBar()
	prog.Min, prog.Max = 0, 100
	progDiag := dialog.NewCustom("Installing", "Cancel", prog, mainWin)
	progDiag.SetOnClosed(func() {
		cancel()
		cancelling.Show()
	})
	progDiag.Show()
	err := installImg(ctx, d, cfg, func(percent float64) {
		prog.SetValue(percent)
	})
	progDiag.Hide()
	cancelling.Hide()
	return err
}

func runInGUIMode() error {
	a := app.New()
	mainWin = a.NewWindow(title)

	piName := widget.NewEntry()
	piName.SetText("alarmpi")
	piName.Validator = func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("your Pi must have a non-blank name")
		}
		return nil
	}
	wifiName := widget.NewEntry()
	wifiName.SetPlaceHolder("(optional)")
	wifiPassword := widget.NewPasswordEntry()
	wifiPassword.SetPlaceHolder("(optional)")

	install := widget.NewButton("Install", func() {
		disks.mu.Lock()
		selIdx := diskList.SelectedIndex()
		if selIdx == -1 {
			disks.mu.Unlock()
			return
		}
		d := disks.items[selIdx]
		disks.mu.Unlock()

		confirmMsg := "You are about to DESTROY ALL DATA on\n\n%s\n\nAre you sure?"
		dialog.ShowConfirm("Are you sure?", fmt.Sprintf(confirmMsg, d.Name()), func(y bool) {
			if !y {
				return
			}
			go func() {
				mainWin.Content().Hide()
				cfg := imgConfig{
					hostname:     piName.Text,
					wifiSSID:     wifiName.Text,
					wifiPassword: wifiPassword.Text,
				}
				if err := installImgWithGUI(d, cfg); err != nil {
					dialog.ShowError(err, mainWin)
				} else {
					dialog.ShowInformation("Success", "Done!", mainWin)
				}
				mainWin.Content().Show()
			}()
		}, mainWin)
	})
	install.Disable()

	diskList.OnChanged = func(s string) {
		if s == "" {
			install.Disable()
		} else {
			install.Enable()
		}
	}

	// We only support these for now - will add more in the future
	rpiModel := widget.NewSelect([]string{"Zero/Zero W"}, nil)
	rpiModel.SetSelectedIndex(0)

	mainWin.SetContent(container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, widget.NewLabel("Pi Model:"), nil, rpiModel),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("Install to:"), nil, diskList),
		widget.NewLabel("Only appropriate drives are shown\n(i.e. external with large enough size)"),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("Pi Name:"), nil, piName),
		container.NewBorder(nil, nil, widget.NewLabel("WiFi Network:"), nil, wifiName),
		container.NewBorder(nil, nil, widget.NewLabel("WiFi Password:"), nil, wifiPassword),
		widget.NewSeparator(),
		layout.NewSpacer(),
		install,
	)))

	contentSize := mainWin.Content().Size()
	mainWin.Resize(fyne.Size{
		Width:  contentSize.Width * 1.5,
		Height: contentSize.Height + 50,
	})

	go refreshDiskList()
	mainWin.ShowAndRun()
	return nil
}

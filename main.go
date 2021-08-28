package main

import (
	crand "crypto/rand"
	_ "embed"
	"fmt"
	"log"
	mrand "math/rand"
	"regexp"
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
	title        = "Raspberry Pi Arch Linux Installer"
	confirmMsg   = "You are about to DESTROY ALL DATA on\n\n%s\n\nAre you sure?"
	diskNote     = "Only appropriate drives are shown\n(i.e. external with large enough size)"
	minDriveSize = 2_097_152_000 // uncompressed size of blank_img.xz
	bufSize      = 0x100_000
	bootUUIDLen  = 4
	rootUUIDLen  = 16
)

var (
	// `xz -d < blank_img.xz | hexdump -C | less` to find these
	bootUUIDPat = regexp.MustCompile(`(\x00)\x29\x0c\x4f\x4b(\x95\x50\x49\x42\x4f\x4f\x54\x20\x20\x20\x20\x20)`)
	rootUUIDPat = regexp.MustCompile(`\x37\x31\x04\xbf\xbd\xed\x42\xb2\x87\xe0\x63\x2b\x07\x25\xe6\xa7`)

	mainWin  fyne.Window
	diskList = widget.NewSelect(nil, nil)

	disks = struct {
		mu    sync.Mutex
		items []disk.Disk
	}{}

	//go:embed arch_img.xz
	archImg []byte
)

func genUUID(length int) []byte {
	b := make([]byte, length)
	if _, err := crand.Read(b); err != nil {
		mrand.Seed(time.Now().Unix())
		_, _ = mrand.Read(b)
	}
	return b
}

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
			if d.Size() >= minDriveSize {
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
			names[i] = d.Name()
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

func install(d disk.Disk) error {
	cancelled := make(chan bool)
	prog := widget.NewProgressBar()
	prog.Min, prog.Max = 0, 100
	progDiag := dialog.NewCustom("Installing", "Cancel", prog, mainWin)
	progDiag.SetOnClosed(func() {
		close(cancelled)
	})
	progDiag.Show()

	w, err := d.OpenForWrite()
	if err != nil {
		progDiag.Hide()
		return err
	}
	defer w.Close()

	if _, err := w.Write([]byte("luck")); err != nil {
		return err
	}

	<-cancelled
	return nil
}

func main() {
	log.SetFlags(0)
	a := app.New()
	mainWin = a.NewWindow(title)

	install := widget.NewButton("Install", func() {
		disks.mu.Lock()
		selIdx := diskList.SelectedIndex()
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
		widget.NewLabel(diskNote),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("Pi Name:"), nil, piName),
		container.NewBorder(nil, nil, widget.NewLabel("WiFi Network:"), nil, wifiName),
		container.NewBorder(nil, nil, widget.NewLabel("WiFi Password:"), nil, wifiPassword),
		widget.NewSeparator(),
		layout.NewSpacer(),
		install,
	)))

	contentSize := mainWin.Content().Size()
	mainWin.Resize(fyne.Size{contentSize.Width * 1.5, contentSize.Height + 50})

	go refreshDiskList()
	mainWin.ShowAndRun()
}

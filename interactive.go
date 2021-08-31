package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"

	"github.com/oxplot/raspberrypi-archlinux-installer/disk"
)

func runInInteractiveMode() error {
	if !term.IsTerminal(int(os.Stderr.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("--interactive must be used inside of a terminal, use --batch if scripting")
	}

	drives, err := disk.Get()
	if err != nil {
		return fmt.Errorf("failed to get list of drives: %s", err)
	}
	if len(drives) == 0 {
		return fmt.Errorf("no suitable drives are available to install to")
	}

	fmt.Printf("%s\n\n", title)

	selPt := promptui.Select{
		Label: "Select your Pi Model",
		Items: []string{"Zero/Zero W"},
	}
	_, _, err = selPt.Run()
	if err != nil {
		return err
	}

	selPt = promptui.Select{
		Label: "Select drive to install to",
		Items: drives,
	}
	driveI, _, err := selPt.Run()
	if err != nil {
		return err
	}
	selectedDrive := drives[driveI]

	pt := promptui.Prompt{
		Label:   "Enter a name for your Pi",
		Default: "alarmpi",
		Validate: func(v string) error {
			v = strings.TrimSpace(v)
			if v == "" {
				return fmt.Errorf("name cannot be blank")
			}
			return nil
		},
		Pointer: promptui.PipeCursor,
	}
	piName, err := pt.Run()
	if err != nil {
		return err
	}

	pt = promptui.Prompt{
		Label:   "Enter you WiFi network name (leave blank if not needed)",
		Pointer: promptui.PipeCursor,
	}
	wifiNetwork, err := pt.Run()
	if err != nil {
		return err
	}

	var wifiPassword string

	if wifiNetwork != "" {
		pt := promptui.Prompt{
			Label:   "Enter you WiFi password (leave blank for open networks)",
			Mask:    '*',
			Pointer: promptui.PipeCursor,
		}
		wifiPassword, err = pt.Run()
		if err != nil {
			return err
		}
		pt = promptui.Prompt{
			Label: "Confirm you WiFi password",
			Validate: func(v string) error {
				if v != wifiPassword {
					return fmt.Errorf("passwords must be the same")
				}
				return nil
			},
			Mask:    '*',
			Pointer: promptui.PipeCursor,
		}
		_, err = pt.Run()
		if err != nil {
			return err
		}
	}

	pt = promptui.Prompt{
		Label:     fmt.Sprintf("You are about to DESTROY ALL DATA on %s - Are you sure", selectedDrive),
		IsConfirm: true,
	}
	if _, err := pt.Run(); err != nil {
		return fmt.Errorf("cancelled")
	}

	cfg := imgConfig{
		hostname:     piName,
		wifiSSID:     wifiNetwork,
		wifiPassword: wifiPassword,
	}

	prog := progressbar.Default(100)
	err = installImg(context.Background(), selectedDrive, cfg, func(percent float64) {
		prog.Set(int(percent))
	})

	return nil
}

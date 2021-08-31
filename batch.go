package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/docker/go-units"

	"github.com/oxplot/raspberrypi-archlinux-installer/disk"
)

var (
	batchFlags = struct {
		showDrives      *bool
		drivePath       *string
		wifiNetwork     *string
		wifiHasPassword *bool
		piName          *string
		showPiModels    *bool
		piModel         *string
	}{
		showDrives:      flag.Bool("show-drives", false, "show list of available drives to install to"),
		drivePath:       flag.String("drive-path", "", "drive path to install to"),
		wifiNetwork:     flag.String("wifi-network", "", "name of the wifi network to configure on Pi"),
		wifiHasPassword: flag.Bool("wifi-has-password", false, "if set, RAI_WIFI_PASSWORD environmental variable must be set to the wifi password"),
		piName:          flag.String("pi-name", "alarmpi", "installation name"),
		showPiModels:    flag.Bool("show-pi-models", false, "show list of supported Pi models"),
		piModel:         flag.String("pi-model", "", "RaspberryPi model"),
	}
)

func runInBatchMode() error {
	if *batchFlags.showDrives && *batchFlags.showPiModels {
		return fmt.Errorf("only one of --show-drives or --show-pi-models can be specified, not both")
	}

	if *batchFlags.showDrives {
		ds, err := disk.Get()
		if err != nil {
			return fmt.Errorf("failed to get list of drives: %s", err)
		}
		for _, d := range ds {
			fmt.Printf("%s (size = %s) (path = %s)\n", d.Name(), units.BytesSize(float64(d.Size())), d.Path())
		}
		return nil
	}

	if *batchFlags.showPiModels {
		fmt.Print("zero\nzero-w\n")
		return nil
	}

	for _, s := range []*string{batchFlags.piName, batchFlags.piModel, batchFlags.drivePath} {
		*s = strings.TrimSpace(*s)
		if *s == "" {
			return fmt.Errorf("--pi-name, --pi-model and --drive-path are required")
		}
	}

	if *batchFlags.piModel != "zero" && *batchFlags.piModel != "zero-w" {
		return fmt.Errorf("invalid --pi-model, run with --show-pi-models to see list of supported models")
	}

	ds, err := disk.Get()
	if err != nil {
		return fmt.Errorf("failed to get list of drives: %s", err)
	}
	var td disk.Disk
	for _, d := range ds {
		if d.Path() == *batchFlags.drivePath {
			td = d
			break
		}
	}

	if td == nil {
		return fmt.Errorf("'%s' drive was not found, run with --show-drives to see a list of available drives",
			*batchFlags.drivePath)
	}

	return nil
}

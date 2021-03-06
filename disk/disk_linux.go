package disk

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/go-units"
	"github.com/godbus/dbus/v5"
)

func nativeGet() (disks []Disk, err error) {
	if disks, err = getDBUSDisks(); err == nil {
		return
	}
	log.Printf("info: failed to get disks from dbus, trying other methods: %s", err)
	return nil, nil
}

type udisk struct {
	dev  string
	path string
	name string
	size uint64
}

func (d *udisk) Path() string {
	return "udisk" + d.dev
}

func (d *udisk) Name() string {
	return d.name
}

func (d *udisk) Size() uint64 {
	return d.size
}

func (d *udisk) String() string {
	return fmt.Sprintf("%s %s (%s)", d.name, units.BytesSize(float64(d.size)), d.dev)
}

type udiskWriter struct {
	f *os.File
}

func (w *udiskWriter) Write(b []byte) (int, error) {
	return w.f.Write(b)
}

func (w *udiskWriter) Close() error {
	err := w.f.Sync()
	if err != nil {
		_ = w.f.Close()
		return err
	}
	return w.f.Close()
}

func (d *udisk) OpenForWrite() (io.WriteCloser, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	o := conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath(d.path))
	c := o.Call("org.freedesktop.UDisks2.Block.OpenDevice", 0, "w", map[string]dbus.Variant{})
	if c.Err != nil {
		return nil, c.Err
	}
	var fd dbus.UnixFD
	if err := c.Store(&fd); err != nil {
		return nil, err
	}
	return &udiskWriter{os.NewFile(uintptr(fd), "")}, nil
}

func getUDisksProp(path, prop string) (interface{}, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return "", err
	}
	o := conn.Object("org.freedesktop.UDisks",
		dbus.ObjectPath(strings.Replace(path, "UDisks2/block_devices", "UDisks/devices", 1)))
	p, err := o.GetProperty("org.freedesktop.UDisks.Device." + prop)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s of %s: %s", prop, path, err)
	}
	return p.Value(), nil
}

func getUDisksStringProp(path, prop string) (string, error) {
	v, err := getUDisksProp(path, prop)
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expected string %s, got %T", prop, v)
	}
	return s, nil
}

func getUDisksBoolProp(path, prop string) (bool, error) {
	v, err := getUDisksProp(path, prop)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("expected bool %s, got %T", prop, v)
	}
	return b, nil
}

func getUDisksUint64Prop(path, prop string) (uint64, error) {
	v, err := getUDisksProp(path, prop)
	if err != nil {
		return 0, err
	}
	i, ok := v.(uint64)
	if !ok {
		return 0, fmt.Errorf("expected uint64 %s, got %T", prop, v)
	}
	return i, nil
}

func getDBUSDisks() ([]Disk, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	if !conn.SupportsUnixFDs() {
		return nil, fmt.Errorf("DBUS conn does not support Unix FDs")
	}
	o := conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath("/org/freedesktop/UDisks2/Manager"))
	c := o.Call("org.freedesktop.UDisks2.Manager.GetBlockDevices", 0, map[string]dbus.Variant{})
	if c.Err != nil {
		return nil, c.Err
	}
	var blocks []string
	if err := c.Store(&blocks); err != nil {
		return nil, err
	}
	disks := make([]Disk, 0, len(blocks))
Outer:
	for _, b := range blocks {
		for _, t := range []string{
			"Partition", "ReadOnly", "LuksClearText", "LinuxLoop",
			"LinuxMd", "OpticalDisk", "SystemInternal",
		} {
			if isBadType, _ := getUDisksBoolProp(b, "DeviceIs"+t); isBadType {
				continue Outer
			}
		}
		if isDrive, _ := getUDisksBoolProp(b, "DeviceIsDrive"); !isDrive {
			continue
		}
		model, err := getUDisksStringProp(b, "DriveModel")
		if err != nil {
			log.Print(err)
			continue
		}
		size, err := getUDisksUint64Prop(b, "DeviceSize")
		if err != nil {
			log.Print(err)
			continue
		}
		devFile, err := getUDisksStringProp(b, "DeviceFile")
		if err != nil {
			log.Print(err)
			continue
		}
		name := model
		disks = append(disks, &udisk{dev: devFile, path: b, name: name, size: size})
	}
	return disks, nil
}

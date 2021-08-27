package disk

import (
	"fmt"
	"io"
)

type linuxDisk struct {
}

func (d *linuxDisk) Name() string {
	return "TODO"
}

func (d *linuxDisk) IsRemovable() bool {
	return true
}

func (d *linuxDisk) Size() uint64 {
	return 0
}

func (d *linuxDisk) OpenForWrite() (io.WriteCloser, error) {
	return nil, fmt.Errorf("boo")
}

func nativeGet() ([]Disk, error) {
	return []Disk{
		&linuxDisk{},
	}, nil
}

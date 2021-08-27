package disk

import (
	"io"
)

type Disk interface {
	Name() string
	IsRemovable() bool
	Size() uint64
	OpenForWrite() (io.WriteCloser, error)
}

func Get() ([]Disk, error) {
	return nativeGet()
}

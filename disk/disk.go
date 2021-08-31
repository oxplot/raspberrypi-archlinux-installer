package disk

import (
	"io"
)

type Disk interface {
	Path() string
	Name() string
	Size() uint64
	OpenForWrite() (io.WriteCloser, error)
}

func Get() ([]Disk, error) {
	return nativeGet()
}

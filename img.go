package main

import (
	crand "crypto/rand"
	"embed"
	"io"
	"io/fs"
	mrand "math/rand"
	"regexp"
	"time"

	"github.com/ulikunitz/xz"
)

const (
	imgSize          = 2_097_152_000 // uncompressed size of blank_img.xz
	imgReadBlockSize = 0x100_000
	bootUUIDLen      = 4
	rootUUIDLen      = 16
)

type archImgReader struct {
	bootUUID        []byte
	patchedBootUUID bool
	rootUUID        []byte
	curBlock        []byte
	bytePos         int
	xzReader        io.Reader
	file            fs.File
}

func newArchImgReader() *archImgReader {
	f, _ := files.Open("arch_img.xz")
	r, err := xz.NewReader(f)
	if err != nil {
		panic(err)
	}
	return &archImgReader{
		bootUUID: genUUID(bootUUIDLen),
		rootUUID: genUUID(rootUUIDLen),
		xzReader: r,
		file:     f,
	}
}

func (a *archImgReader) Read(b []byte) (int, error) {
	return a.xzReader.Read(b)
}

func (a *archImgReader) Close() error {
	return a.file.Close()
}

var (
	// `xz -d < blank_img.xz | hexdump -C | less` to find these
	bootUUIDPat = regexp.MustCompile(`(\x00)\x29\x0c\x4f\x4b(\x95\x50\x49\x42\x4f\x4f\x54\x20\x20\x20\x20\x20)`)
	rootUUIDPat = regexp.MustCompile(`\x37\x31\x04\xbf\xbd\xed\x42\xb2\x87\xe0\x63\x2b\x07\x25\xe6\xa7`)

	//go:embed arch_img.xz
	files embed.FS
)

func genUUID(length int) []byte {
	b := make([]byte, length)
	if _, err := crand.Read(b); err != nil {
		mrand.Seed(time.Now().Unix())
		_, _ = mrand.Read(b)
	}
	return b
}

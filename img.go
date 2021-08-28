package main

import (
	"bytes"
	crand "crypto/rand"
	_ "embed"
	"io"
	mrand "math/rand"
	"regexp"
	"time"

	"github.com/ulikunitz/xz"
)

const (
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
}

func newArchImgReader() *archImgReader {
	r, err := xz.NewReader(bytes.NewReader(archImg))
	if err != nil {
		panic(err)
	}
	return &archImgReader{
		bootUUID: genUUID(bootUUIDLen),
		rootUUID: genUUID(rootUUIDLen),
		xzReader: r,
	}
}

func (a *archImgReader) Read(b []byte) (int, error) {
	return a.xzReader.Read(b)
}

var (
	// `xz -d < blank_img.xz | hexdump -C | less` to find these
	bootUUIDPat = regexp.MustCompile(`(\x00)\x29\x0c\x4f\x4b(\x95\x50\x49\x42\x4f\x4f\x54\x20\x20\x20\x20\x20)`)
	rootUUIDPat = regexp.MustCompile(`\x37\x31\x04\xbf\xbd\xed\x42\xb2\x87\xe0\x63\x2b\x07\x25\xe6\xa7`)

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

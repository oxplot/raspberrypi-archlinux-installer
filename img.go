package main

import (
	"archive/tar"
	"crypto/sha1"
	"embed"
	"encoding/hex"
	"io"
	"io/fs"

	"github.com/ulikunitz/xz"
	"golang.org/x/crypto/pbkdf2"
)

const (
	imgSize    = 1572864000 // uncompressed size of blank_img.xz
	configSize = 1000000    // length of trailing config tar
)

var (
	//go:embed arch_img.xz
	files embed.FS
)

type archImgReader struct {
	curBlock []byte
	bytePos  int
	xzReader io.Reader
	file     fs.File
}

func newArchImgReader() *archImgReader {
	f, _ := files.Open("arch_img.xz")
	r, err := xz.NewReader(f)
	if err != nil {
		panic(err)
	}
	return &archImgReader{
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

type imgConfig struct {
	wifiSSID     string
	wifiPassword string
	hostname     string
}

func calcWifiPSK(ssid, password string) string {
	return hex.EncodeToString(pbkdf2.Key([]byte(password), []byte(ssid), 4096, 256, sha1.New))
}

func writeTarFile(w *tar.Writer, name string, content []byte) error {
	if err := w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     name,
		Size:     int64(len(content)),
		Mode:     0666,
	}); err != nil {
		return err
	}
	if _, err := w.Write(content); err != nil {
		return err
	}
	return nil
}

func (c imgConfig) writeTo(w io.Writer) error {
	tw := tar.NewWriter(w)
	if err := writeTarFile(tw, "hostname", []byte(c.hostname)); err != nil {
		return err
	}

	if len(c.wifiSSID) > 0 {

		if err := writeTarFile(tw, "wifi_ssid", []byte(c.wifiSSID)); err != nil {
			return err
		}

		if len(c.wifiPassword) > 0 {
			psk := calcWifiPSK(c.wifiSSID, c.wifiPassword)
			if err := writeTarFile(tw, "wifi_psk", []byte(psk)); err != nil {
				return err
			}
		}

	}

	tw.Flush()
	buf := make([]byte, 4096)
	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}

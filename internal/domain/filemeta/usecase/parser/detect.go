package parser

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"strings"
)

var (
	zipMagic = []byte("PK\x03\x04")
	exeMagic = []byte("MZ")
	arMagic  = []byte("!<arch>\n")
	rpmMagic = []byte{0xED, 0xAB, 0xEE, 0xDB}
	oleMagic = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	dmgMagic = []byte("koly")
)

func Detect(path string) (FileType, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileTypeUnknown, err
	}
	defer f.Close()

	header := make([]byte, 8)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return FileTypeUnknown, err
	}
	header = header[:n]

	switch {
	case bytes.HasPrefix(header, zipMagic):
		if isAPK(path) {
			return FileTypeAPK, nil
		}
		return FileTypeUnknown, nil
	case bytes.HasPrefix(header, exeMagic):
		return FileTypeEXE, nil
	case bytes.HasPrefix(header, arMagic):
		if isDEB(path) {
			return FileTypeDEB, nil
		}
		return FileTypeUnknown, nil
	case bytes.HasPrefix(header, rpmMagic):
		return FileTypeRPM, nil
	case bytes.HasPrefix(header, oleMagic):
		return FileTypeMSI, nil
	}

	if isDMG(path) {
		return FileTypeDMG, nil
	}

	return FileTypeUnknown, nil
}

func isAPK(path string) bool {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name == "AndroidManifest.xml" {
			return true
		}
	}
	return false
}

func isDEB(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := f.Seek(int64(len(arMagic)), io.SeekStart); err != nil {
		return false
	}

	nameField := make([]byte, 16)
	if _, err := io.ReadFull(f, nameField); err != nil {
		return false
	}

	name := strings.TrimRight(string(nameField), " /")
	return name == "debian-binary"
}

func isDMG(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.Size() < 512 {
		return false
	}

	if _, err := f.Seek(-512, io.SeekEnd); err != nil {
		return false
	}

	trailer := make([]byte, len(dmgMagic))
	if _, err := io.ReadFull(f, trailer); err != nil {
		return false
	}

	return bytes.Equal(trailer, dmgMagic)
}

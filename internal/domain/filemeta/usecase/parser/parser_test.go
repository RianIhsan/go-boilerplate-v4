package parser

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func buildTestDeb(t *testing.T) string {
	t.Helper()

	controlTarGz := buildTarGz(t, map[string]string{
		"./control": "Package: hello-test\n" +
			"Version: 1.2.3\n" +
			"Architecture: amd64\n" +
			"Maintainer: Jane Doe <jane@example.com>\n" +
			"Description: a tiny test package\n",
	})
	dataTarGz := buildTarGz(t, map[string]string{})

	var ar bytes.Buffer
	ar.WriteString("!<arch>\n")
	writeArMember(&ar, "debian-binary", []byte("2.0\n"))
	writeArMember(&ar, "control.tar.gz", controlTarGz)
	writeArMember(&ar, "data.tar.gz", dataTarGz)

	path := filepath.Join(t.TempDir(), "test.deb")
	if err := os.WriteFile(path, ar.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeArMember(buf *bytes.Buffer, name string, data []byte) {
	header := make([]byte, 60)
	copy(header, padRight(name, 16))
	copy(header[16:], padRight("0", 12))
	copy(header[28:], padRight("0", 6))
	copy(header[34:], padRight("0", 6))
	copy(header[40:], padRight("100644", 8))
	copy(header[48:], padRight(strconv.Itoa(len(data)), 10))
	header[58] = '`'
	header[59] = '\n'

	buf.Write(header)
	buf.Write(data)
	if len(data)%2 != 0 {
		buf.WriteByte('\n')
	}
}

func padRight(s string, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	copy(b, s)
	return b
}

func writeZipWithManifest(path string, manifest []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	w, err := zw.Create("AndroidManifest.xml")
	if err != nil {
		return err
	}
	if _, err := w.Write(manifest); err != nil {
		return err
	}
	return zw.Close()
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		path string
		want FileType
	}{
		{"deb", buildTestDeb(t), FileTypeDEB},
		{"rpm", "testdata/sample.rpm", FileTypeRPM},
		{"exe", "testdata/sample.exe", FileTypeEXE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Detect(tt.path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Detect() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetect_MSI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.msi")
	data := append([]byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, make([]byte, 512)...)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	got, err := Detect(path)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if got != FileTypeMSI {
		t.Errorf("Detect() = %q, want %q", got, FileTypeMSI)
	}
}

func TestDetect_Unknown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.bin")
	if err := os.WriteFile(path, []byte("just some random garbage bytes"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := Detect(path)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if got != FileTypeUnknown {
		t.Errorf("Detect() = %q, want %q", got, FileTypeUnknown)
	}
}

func TestParseDEB(t *testing.T) {
	meta, err := ParseDEB(buildTestDeb(t))
	if err != nil {
		t.Fatal(err)
	}
	if meta["Package"] != "hello-test" || meta["Version"] != "1.2.3" || meta["Maintainer"] != "Jane Doe <jane@example.com>" {
		t.Errorf("ParseDEB() = %+v, want Package=hello-test Version=1.2.3 Maintainer=\"Jane Doe <jane@example.com>\"", meta)
	}
}

func TestParseRPM(t *testing.T) {
	meta, err := ParseRPM("testdata/sample.rpm")
	if err != nil {
		t.Fatal(err)
	}
	if meta["name"] != "epel-release" {
		t.Errorf("ParseRPM() name = %q, want %q", meta["name"], "epel-release")
	}
	if meta["version"] == "" {
		t.Error("ParseRPM() version is empty, want a non-empty version")
	}
}

func TestParseEXE(t *testing.T) {
	meta, err := ParseEXE("testdata/sample.exe")
	if err != nil {
		t.Fatal(err)
	}
	if len(meta) == 0 {
		t.Error("ParseEXE() returned entirely empty version info for a fixture with known VERSIONINFO")
	}
}

func TestParseAPK_RejectsPlainTextManifest(t *testing.T) {
	dir := t.TempDir()
	apkPath := filepath.Join(dir, "fake.apk")

	if err := writeZipWithManifest(apkPath, []byte("<manifest package=\"com.example\"/>")); err != nil {
		t.Fatal(err)
	}

	ft, err := Detect(apkPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if ft != FileTypeAPK {
		t.Fatalf("Detect() = %q, want %q", ft, FileTypeAPK)
	}

	if _, err := ParseAPK(apkPath); err == nil {
		t.Error("ParseAPK() error = nil, want an error for a non-binary manifest")
	}
}

func TestParseDMG_NonAPFS(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.dmg")
	data := append(make([]byte, 1024), []byte("koly")...)
	data = append(data, make([]byte, 508)...)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	meta, err := ParseDMG(path)
	if err != nil {
		t.Fatalf("ParseDMG() error = %v, want nil (graceful empty fallback)", err)
	}
	if len(meta) != 0 {
		t.Errorf("ParseDMG() = %+v, want an empty map for a DMG with no real partitions", meta)
	}
}

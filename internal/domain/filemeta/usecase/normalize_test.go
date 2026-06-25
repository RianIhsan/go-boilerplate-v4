package usecase

import (
	"testing"

	"github.com/avast/apkverifier"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/usecase/parser"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		fileType parser.FileType
		meta     any
		want     normalized
	}{
		{
			name:     "apk",
			fileType: parser.FileTypeAPK,
			meta: map[string]any{
				"package":     "me.pou.app",
				"versionName": "1.4.135",
				"signer":      &apkverifier.CertInfo{Subject: "CN=Paul"},
			},
			want: normalized{PackageName: "me.pou.app", Version: "1.4.135", Publisher: "CN=Paul"},
		},
		{
			name:     "apk - no signer",
			fileType: parser.FileTypeAPK,
			meta: map[string]any{
				"package":     "com.example",
				"versionName": "1.0",
			},
			want: normalized{PackageName: "com.example", Version: "1.0", Publisher: ""},
		},
		{
			name:     "exe",
			fileType: parser.FileTypeEXE,
			meta: map[string]string{
				"ProductName":    "PuTTY suite",
				"ProductVersion": "Release 0.73",
				"CompanyName":    "Simon Tatham",
			},
			want: normalized{PackageName: "PuTTY suite", Version: "Release 0.73", Publisher: "Simon Tatham"},
		},
		{
			name:     "exe - falls back to InternalName and FileVersion",
			fileType: parser.FileTypeEXE,
			meta: map[string]string{
				"InternalName": "app.exe",
				"FileVersion":  "1.2.3",
			},
			want: normalized{PackageName: "app.exe", Version: "1.2.3", Publisher: ""},
		},
		{
			name:     "deb",
			fileType: parser.FileTypeDEB,
			meta: map[string]string{
				"Package":    "klikaku",
				"Version":    "1.1.0",
				"Maintainer": "Julian <juliyanto160784@gmail.com>",
			},
			want: normalized{PackageName: "klikaku", Version: "1.1.0", Publisher: "Julian <juliyanto160784@gmail.com>"},
		},
		{
			name:     "rpm",
			fileType: parser.FileTypeRPM,
			meta: map[string]any{
				"name":    "epel-release",
				"version": "7",
				"release": "5",
				"vendor":  "Fedora Project",
			},
			want: normalized{PackageName: "epel-release", Version: "7-5", Publisher: "Fedora Project"},
		},
		{
			name:     "rpm - falls back to packager when vendor empty",
			fileType: parser.FileTypeRPM,
			meta: map[string]any{
				"name":     "foo",
				"version":  "1",
				"packager": "Some Packager",
			},
			want: normalized{PackageName: "foo", Version: "1", Publisher: "Some Packager"},
		},
		{
			name:     "dmg",
			fileType: parser.FileTypeDMG,
			meta: map[string]any{
				"CFBundleIdentifier":         "com.insomnia.app",
				"CFBundleShortVersionString": "13.0.2",
				"NSHumanReadableCopyright":   "Copyright © 2026 Kong",
			},
			want: normalized{PackageName: "com.insomnia.app", Version: "13.0.2", Publisher: "Copyright © 2026 Kong"},
		},
		{
			name:     "dmg - empty metadata",
			fileType: parser.FileTypeDMG,
			meta:     map[string]any{},
			want:     normalized{},
		},
		{
			name:     "unknown file type",
			fileType: parser.FileTypeUnknown,
			meta:     map[string]any{"foo": "bar"},
			want:     normalized{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalize(tt.fileType, tt.meta)
			if got != tt.want {
				t.Errorf("normalize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

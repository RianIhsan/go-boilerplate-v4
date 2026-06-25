package parser

type FileType string

const (
	FileTypeAPK     FileType = "apk"
	FileTypeEXE     FileType = "exe"
	FileTypeDEB     FileType = "deb"
	FileTypeRPM     FileType = "rpm"
	FileTypeDMG     FileType = "dmg"
	FileTypeMSI     FileType = "msi"
	FileTypeUnknown FileType = ""
)

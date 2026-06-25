package usecase

import (
	"github.com/avast/apkverifier"

	"github.com/RianIhsan/go-boilerplate-v4/internal/domain/filemeta/usecase/parser"
)

type normalized struct {
	PackageName string
	Version     string
	Publisher   string
}

func normalize(fileType parser.FileType, meta any) normalized {
	switch fileType {
	case parser.FileTypeAPK:
		return normalizeAPK(asMapAny(meta))
	case parser.FileTypeEXE:
		return normalizeEXE(asMapString(meta))
	case parser.FileTypeDEB:
		return normalizeDEB(asMapString(meta))
	case parser.FileTypeRPM:
		return normalizeRPM(asMapAny(meta))
	case parser.FileTypeDMG:
		return normalizeDMG(asMapAny(meta))
	default:
		return normalized{}
	}
}

func normalizeAPK(m map[string]any) normalized {
	publisher := ""
	if signer, ok := m["signer"].(*apkverifier.CertInfo); ok && signer != nil {
		publisher = signer.Subject
	}
	return normalized{
		PackageName: strVal(m["package"]),
		Version:     strVal(m["versionName"]),
		Publisher:   publisher,
	}
}

func normalizeEXE(m map[string]string) normalized {
	return normalized{
		PackageName: firstNonEmpty(m["ProductName"], m["InternalName"], m["OriginalFilename"]),
		Version:     firstNonEmpty(m["ProductVersion"], m["FileVersion"]),
		Publisher:   m["CompanyName"],
	}
}

func normalizeDEB(m map[string]string) normalized {
	return normalized{
		PackageName: m["Package"],
		Version:     m["Version"],
		Publisher:   m["Maintainer"],
	}
}

func normalizeRPM(m map[string]any) normalized {
	version := strVal(m["version"])
	if release := strVal(m["release"]); release != "" {
		version += "-" + release
	}
	return normalized{
		PackageName: strVal(m["name"]),
		Version:     version,
		Publisher:   firstNonEmpty(strVal(m["vendor"]), strVal(m["packager"])),
	}
}

func normalizeDMG(m map[string]any) normalized {
	return normalized{
		PackageName: strVal(m["CFBundleIdentifier"]),
		Version:     firstNonEmpty(strVal(m["CFBundleShortVersionString"]), strVal(m["CFBundleVersion"])),

		Publisher: strVal(m["NSHumanReadableCopyright"]),
	}
}

func asMapAny(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func asMapString(v any) map[string]string {
	m, _ := v.(map[string]string)
	return m
}

func strVal(v any) string {
	s, _ := v.(string)
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

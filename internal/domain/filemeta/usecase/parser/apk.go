package parser

import (
	"bytes"
	"encoding/xml"

	"github.com/avast/apkparser"
	"github.com/avast/apkverifier"
)

type manifestAttrs map[string]any

func (m *manifestAttrs) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	attrs := make(map[string]any, len(start.Attr))
	for _, a := range start.Attr {
		attrs[a.Name.Local] = a.Value
	}
	*m = attrs
	return d.Skip()
}

func ParseAPK(path string) (map[string]any, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)

	if zipErr, _, manErr := apkparser.ParseApk(path, enc); zipErr != nil {
		return nil, zipErr
	} else if manErr != nil {
		return nil, manErr
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}

	var attrs manifestAttrs
	if err := xml.Unmarshal(buf.Bytes(), &attrs); err != nil {
		return nil, err
	}

	result := map[string]any(attrs)
	if signer := apkSigner(path); signer != nil {
		result["signer"] = signer
	}
	return result, nil
}

func apkSigner(path string) *apkverifier.CertInfo {
	chains, err := apkverifier.ExtractCerts(path, nil)
	if err != nil || len(chains) == 0 {
		return nil
	}

	certInfo, _ := apkverifier.PickBestApkCert(chains)
	return certInfo
}

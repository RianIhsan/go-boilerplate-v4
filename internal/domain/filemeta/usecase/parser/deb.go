package parser

import (
	debpkg "pault.ag/go/debian/deb"
)

func ParseDEB(path string) (map[string]string, error) {
	d, closer, err := debpkg.LoadFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = closer() }()

	return d.Control.Values, nil
}

package parser

import (
	rpmpkg "github.com/cavaliergopher/rpm"
)

func ParseRPM(path string) (map[string]any, error) {
	pkg, err := rpmpkg.Open(path)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"name":             pkg.Name(),
		"version":          pkg.Version(),
		"release":          pkg.Release(),
		"epoch":            pkg.Epoch(),
		"architecture":     pkg.Architecture(),
		"vendor":           pkg.Vendor(),
		"packager":         pkg.Packager(),
		"summary":          pkg.Summary(),
		"description":      pkg.Description(),
		"license":          pkg.License(),
		"groups":           pkg.Groups(),
		"url":              pkg.URL(),
		"distribution":     pkg.Distribution(),
		"operating_system": pkg.OperatingSystem(),
		"platform":         pkg.Platform(),
		"build_time":       pkg.BuildTime(),
		"build_host":       pkg.BuildHost(),
		"source_rpm":       pkg.SourceRPM(),
		"provides":         dependencyNames(pkg.Provides()),
		"requires":         dependencyNames(pkg.Requires()),
		"conflicts":        dependencyNames(pkg.Conflicts()),
		"obsoletes":        dependencyNames(pkg.Obsoletes()),
		"recommends":       dependencyNames(pkg.Recommends()),
		"suggests":         dependencyNames(pkg.Suggests()),
	}, nil
}

func dependencyNames(deps []rpmpkg.Dependency) []string {
	names := make([]string, len(deps))
	for i, d := range deps {
		names[i] = d.Name()
	}
	return names
}

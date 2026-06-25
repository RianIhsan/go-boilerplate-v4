package parser

import (
	pe "github.com/saferwall/pe"
)

func ParseEXE(path string) (map[string]string, error) {
	file, err := pe.New(path, &pe.Options{
		OmitExportDirectory:       true,
		OmitImportDirectory:       true,
		OmitExceptionDirectory:    true,
		OmitSecurityDirectory:     true,
		OmitRelocDirectory:        true,
		OmitDebugDirectory:        true,
		OmitArchitectureDirectory: true,
		OmitGlobalPtrDirectory:    true,
		OmitTLSDirectory:          true,
		OmitLoadConfigDirectory:   true,
		OmitBoundImportDirectory:  true,
		OmitIATDirectory:          true,
		OmitDelayImportDirectory:  true,
		OmitCLRHeaderDirectory:    true,
		OmitCLRMetadata:           true,
	})
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := file.Parse(); err != nil {
		return nil, err
	}

	return file.ParseVersionResources()
}

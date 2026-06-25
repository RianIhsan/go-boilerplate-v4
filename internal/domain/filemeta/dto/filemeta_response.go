package dto

type FileMetadataResponse struct {
	FileType    string `json:"file_type"`
	PackageName string `json:"package_name"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
	Metadata    any    `json:"metadata"`
}

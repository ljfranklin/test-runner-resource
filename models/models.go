package models

type Version struct {
	Key string `json:"key"`
}

type Source struct {
	StorageType   string
	StorageConfig map[string]interface{}
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

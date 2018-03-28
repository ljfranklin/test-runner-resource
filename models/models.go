package models

type Version struct {
	Key string `json:"key"`
}

type Source struct {
	StorageType   string                 `json:"storage_type"`
	StorageConfig map[string]interface{} `json:"storage_config"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

type InRequest struct {
	Source    Source   `json:"source"`
	Version   Version  `json:"version"`
	Params    InParams `json:"params"`
	OutputDir string   `json:"-"`
}

type InParams struct {
	Summaries []Summary `json:"summaries"`
}

type InResponse struct {
	Version  Version           `json:"version"`
	Metadata map[string]string `json:"metadata"`
}

type Summary struct {
	Type  string `json:"type"`
	Limit int    `json:"limit"`
}

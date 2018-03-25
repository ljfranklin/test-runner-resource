package storage

import (
	"fmt"
	"io"
)

type FileNotFound struct {
	Key string
}

func (f FileNotFound) Error() string {
	return fmt.Sprintf("could not find file with key '%s'", f.Key)
}

// go:generate counterfeiter . Storage

type Storage interface {
	Get(string, io.Writer) error
	Put(string, io.Reader) error
	Delete(string) error
	List() ([]string, error)
}

func New(configType string, config map[string]interface{}) (Storage, error) {
	switch configType {
	case "s3":
		if err := ValidateS3Config(config); err != nil {
			return nil, err
		}
		return NewS3(config), nil
	default:
		return nil, fmt.Errorf("unrecognized storage_type '%s'; set storage_type to one of the following: 's3'", configType)
	}
}

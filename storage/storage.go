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

func CreateFromJSON(configType string, config map[string]interface{}) (Storage, error) {
	return NewS3(config), nil
	// TODO: add validate()
}

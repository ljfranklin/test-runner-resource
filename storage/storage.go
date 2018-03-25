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
	// TODO: add validate()
	return NewS3(config), nil
}

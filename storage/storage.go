package storage

import (
	"fmt"
	"io"
)

type Version interface {
	Compare(interface{}) int
}

type Result struct {
	Key     string
	Version Version
}

type Results []Result

func (r Results) Len() int           { return len(r) }
func (r Results) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Results) Less(i, j int) bool { return r[i].Version.Compare(r[j].Version) < 0 }

type FileNotFound struct {
	Key string
}

func (f FileNotFound) Error() string {
	return fmt.Sprintf("could not find file with key '%s'", f.Key)
}

type Storage interface {
	Get(string, io.Writer) (Result, error)
	Put(string, io.Reader) (Result, error)
	Delete(string) error
	List(string) (Results, error)
}

func CreateFromJSON(configType string, config map[string]interface{}) (Storage, error) {
	return NewS3(config), nil
	// TODO: add validate()
}

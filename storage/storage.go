package storage

import (
	"io"
)

type Version interface {
	Compare(Version) int
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(b []byte) error
}

type Result struct {
	Key     string
	Version Version
}

type Storage interface {
	Get(string, io.Writer) (Result, error)
	Put(string, io.Reader) (Result, error)
	Delete(string) error
	List(string) ([]Result, error)
}

package helpers

import (
	"path"
	"runtime"
)

func ProjectRoot() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(filename), "..")
}

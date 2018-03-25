package main

import (
	"encoding/json"
	"os"

	"github.com/ljfranklin/test-runner-resource/check"
	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/storage"
)

func main() {
	var request models.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		panic(err) // TODO
	}

	storage, err := storage.New(request.Source.StorageType, request.Source.StorageConfig)
	if err != nil {
		panic(err) // TODO
	}

	checker := check.Checker{
		Storage: storage,
	}

	results, err := checker.Check(request.Version)
	if err != nil {
		panic(err) // TODO
	}

	err = json.NewEncoder(os.Stdout).Encode(results)
	if err != nil {
		panic(err) // TODO
	}
}

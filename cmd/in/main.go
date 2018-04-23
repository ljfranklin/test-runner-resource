package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/ljfranklin/test-runner-resource/in"
	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/storage"
	"github.com/ljfranklin/test-runner-resource/viewer"
)

func main() {
	var request models.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		log.Fatalf("failed to decode input JSON: %s", err)
	}

	request.OutputDir = os.Args[1]

	storage, err := storage.New(request.Source.StorageType, request.Source.StorageConfig)
	if err != nil {
		log.Fatalf("failed to initialize storage: %s", err)
	}

	viewer := viewer.JunitCLI{
		OutputWriter: os.Stderr,
		ResultsDir:   request.OutputDir,
	}

	getter := in.Getter{
		Storage:     storage,
		JunitViewer: viewer,
	}

	results, err := getter.Get(request)
	if err != nil {
		log.Fatalf("failed to get requested version: %s", err)
	}

	err = json.NewEncoder(os.Stdout).Encode(results)
	if err != nil {
		log.Fatalf("failed to encode output JSON: %s", err)
	}
}

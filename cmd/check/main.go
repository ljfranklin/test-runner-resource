package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/ljfranklin/test-runner-resource/check"
	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/storage"
)

func main() {
	var request models.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		log.Fatalf("failed to decode input JSON: %s", err)
	}

	storage, err := storage.New(request.Source.StorageType, request.Source.StorageConfig)
	if err != nil {
		log.Fatalf("failed to initialize storage: %s", err)
	}

	checker := check.Checker{
		Storage: storage,
	}

	results, err := checker.Check(request.Version)
	if err != nil {
		log.Fatalf("failed to check for new versions: %s", err)
	}

	err = json.NewEncoder(os.Stdout).Encode(results)
	if err != nil {
		log.Fatalf("failed to encode output JSON: %s", err)
	}
}

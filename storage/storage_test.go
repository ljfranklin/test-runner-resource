package storage_test

import (
	"strings"
	"testing"

	"github.com/ljfranklin/test-runner-resource/storage"
)

func TestErrorOnInvalidType(t *testing.T) {
	t.Parallel()

	_, err := storage.New("invalid-type", nil)
	if err == nil {
		t.Fatal("expected error on invalid type but none occurred")
	}
	if !strings.Contains(err.Error(), "invalid-type") {
		t.Fatalf("expected error to contain 'invalid-type' but it did not: %s", err)
	}
}

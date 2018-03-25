package helpers

import (
	"encoding/json"
	"reflect"
	"testing"
)

func AssertEquals(t *testing.T, actual interface{}, expected interface{}) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected '%v' to deep equal '%v'", actual, expected)
	}
}

func AssertJSONEquals(t *testing.T, actual, expected string) {
	t.Helper()

	var actualStruct interface{}
	err := json.Unmarshal([]byte(actual), &actualStruct)
	if err != nil {
		t.Fatal(err)
	}

	var expectedStruct interface{}
	err = json.Unmarshal([]byte(expected), &expectedStruct)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actualStruct, expectedStruct) {
		t.Fatalf("expected '%s' to deep equal '%s'", actual, expected)
	}
}

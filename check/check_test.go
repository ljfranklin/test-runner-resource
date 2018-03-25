package check_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ljfranklin/test-runner-resource/check"
	"github.com/ljfranklin/test-runner-resource/storage/storagefakes"
	"github.com/ljfranklin/test-runner-resource/test/helpers"
)

func TestCheckWithNoInputVersion(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-2018-01-02T15:04:05Z.xml",
		"test-results-2018-01-01T15:04:05Z.xml",
		"test-results-2018-01-03T15:04:05Z.xml",
	}, nil)

	checker := check.Checker{
		Storage: fakeStorage,
	}

	versions, err := checker.Check(check.Version{})
	if err != nil {
		t.Fatal(err)
	}

	resultsJSON, err := json.Marshal(versions)
	if err != nil {
		t.Fatal(err)
	}
	helpers.AssertJSONEquals(t, string(resultsJSON), `[
	  {"key": "test-results-2018-01-01T15:04:05Z.xml"},
	  {"key": "test-results-2018-01-02T15:04:05Z.xml"},
	  {"key": "test-results-2018-01-03T15:04:05Z.xml"}
	]`)
}

func TestCheckWithInputVersion(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-2018-01-02T15:04:05Z.xml",
		"test-results-2018-01-01T15:04:05Z.xml",
		"test-results-2018-01-03T15:04:05Z.xml",
	}, nil)

	checker := check.Checker{
		Storage: fakeStorage,
	}

	versions, err := checker.Check(check.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	})
	if err != nil {
		t.Fatal(err)
	}

	resultsJSON, err := json.Marshal(versions)
	if err != nil {
		t.Fatal(err)
	}
	helpers.AssertJSONEquals(t, string(resultsJSON), `[
	  {"key": "test-results-2018-01-02T15:04:05Z.xml"},
	  {"key": "test-results-2018-01-03T15:04:05Z.xml"}
	]`)
}

func TestCheckWithPathPrefix(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"some-path/test-results-2018-01-02T15:04:05Z.xml",
		"some-path/test-results-2018-01-01T15:04:05Z.xml",
		"some-path/test-results-2018-01-03T15:04:05Z.xml",
	}, nil)

	checker := check.Checker{
		Storage: fakeStorage,
	}

	versions, err := checker.Check(check.Version{})
	if err != nil {
		t.Fatal(err)
	}

	resultsJSON, err := json.Marshal(versions)
	if err != nil {
		t.Fatal(err)
	}
	helpers.AssertJSONEquals(t, string(resultsJSON), `[
	  {"key": "some-path/test-results-2018-01-01T15:04:05Z.xml"},
	  {"key": "some-path/test-results-2018-01-02T15:04:05Z.xml"},
	  {"key": "some-path/test-results-2018-01-03T15:04:05Z.xml"}
	]`)
}

func TestCheckErrorWithInvalidStartingVersion(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{}, nil)

	checker := check.Checker{
		Storage: fakeStorage,
	}

	_, err := checker.Check(check.Version{
		Key: "test-results-invalid-time.xml",
	})
	if err == nil {
		t.Fatal("expected invalid time to err but it did not")
	}
	if !strings.Contains(err.Error(), "invalid-time") {
		t.Fatalf("expected '%s' to contain substring 'invalid-time'", err.Error())
	}
}

func TestCheckErrorWhenListErrors(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{}, errors.New("some-error"))

	checker := check.Checker{
		Storage: fakeStorage,
	}

	_, err := checker.Check(check.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	})
	if err == nil {
		t.Fatal("expected Check to err but it did not")
	}
	if !strings.Contains(err.Error(), "some-error") {
		t.Fatalf("expected '%s' to contain substring 'some-error'", err.Error())
	}
}

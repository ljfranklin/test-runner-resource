package in_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ljfranklin/test-runner-resource/in"
	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/storage/storagefakes"
	"github.com/ljfranklin/test-runner-resource/test/helpers"
	"github.com/ljfranklin/test-runner-resource/viewer/viewerfakes"
)

func TestGet(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-2018-01-02T15:04:05Z.xml",
		"test-results-2018-01-01T15:04:05Z.xml",
		"test-results-2018-01-03T15:04:05Z.xml",
	}, nil)
	fakeStorage.GetStub = func(key string, writer io.Writer) error {
		switch key {
		case "test-results-2018-01-01T15:04:05Z.xml":
			f, err := os.Open(filepath.Join("..", "fixtures", "junit", "success.xml"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if _, err = io.Copy(writer, f); err != nil {
				t.Fatal(err)
			}
		case "test-results-2018-01-02T15:04:05Z.xml":
			f, err := os.Open(filepath.Join("..", "fixtures", "junit", "failures.xml"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if _, err = io.Copy(writer, f); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected Get call with '%s'", key)
		}
		return nil
	}
	fakeJunit := &viewerfakes.FakeJunit{}

	tmpDir, err := ioutil.TempDir("", "get-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	}
	result, err := getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: tmpDir,
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 10,
				},
				{
					Type:  "frequent-failures",
					Limit: 5,
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	helpers.AssertEquals(t, fakeJunit.PrintSummaryCallCount(), 2)
	firstSummaryType := fakeJunit.PrintSummaryArgsForCall(0)
	helpers.AssertEquals(t, firstSummaryType, models.Summary{
		Type:  "pass-fail",
		Limit: 10,
	})
	secondSummaryType := fakeJunit.PrintSummaryArgsForCall(1)
	helpers.AssertEquals(t, secondSummaryType, models.Summary{
		Type:  "frequent-failures",
		Limit: 5,
	})

	helpers.AssertEquals(t, result.Version, models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	})
	helpers.AssertEquals(t, result.Metadata, map[string]string{
		"test_suite_count": "2",
	})

	firstOutputXML := filepath.Join(tmpDir, "test-results-2018-01-01T15:04:05Z.xml")
	if _, err := os.Stat(firstOutputXML); err == nil {
		expectedContent, err := ioutil.ReadFile(filepath.Join("..", "fixtures", "junit", "success.xml"))
		if err != nil {
			t.Fatal(err)
		}
		actualContent, err := ioutil.ReadFile(firstOutputXML)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actualContent, expectedContent) {
			t.Fatalf("expected '%s' to equal '%s' but it did not", string(actualContent), string(expectedContent))
		}
	} else {
		t.Fatalf("expected '%s' to exist but it does not: %s", firstOutputXML, err)
	}

	secondOutputXML := filepath.Join(tmpDir, "test-results-2018-01-02T15:04:05Z.xml")
	if _, err := os.Stat(secondOutputXML); err == nil {
		expectedContent, err := ioutil.ReadFile(filepath.Join("..", "fixtures", "junit", "failures.xml"))
		if err != nil {
			t.Fatal(err)
		}
		actualContent, err := ioutil.ReadFile(secondOutputXML)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actualContent, expectedContent) {
			t.Fatalf("expected '%s' to equal '%s' but it did not", string(actualContent), string(expectedContent))
		}
	} else {
		t.Fatalf("expected '%s' to exist but it does not: %s", secondOutputXML, err)
	}
}

func TestGetLimitFileDownloads(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-2018-01-02T15:04:05Z.xml",
		"test-results-2018-01-01T15:04:05Z.xml",
		"test-results-2018-01-03T15:04:05Z.xml",
	}, nil)
	fakeStorage.GetStub = func(key string, writer io.Writer) error {
		switch key {
		case "test-results-2018-01-02T15:04:05Z.xml":
			f, err := os.Open(filepath.Join("..", "fixtures", "junit", "failures.xml"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if _, err = io.Copy(writer, f); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected Get call with '%s'", key)
		}
		return nil
	}
	fakeJunit := &viewerfakes.FakeJunit{}

	tmpDir, err := ioutil.TempDir("", "get-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	}
	_, err = getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: tmpDir,
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 1,
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	helpers.AssertEquals(t, fakeStorage.GetCallCount(), 1)
}

func TestGetErrorOnInvalidStartingVersion(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeJunit := &viewerfakes.FakeJunit{}
	tmpDir, err := ioutil.TempDir("", "get-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-invalid-date.xml",
	}
	_, err = getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: tmpDir,
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 10,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected err to occur but it did not")
	}
	if !strings.Contains(err.Error(), "invalid-date") {
		t.Fatalf("expected err to contain 'invalid-date', but it did not: %s", err)
	}
}

func TestGetErrorOnListFailure(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns(nil, errors.New("some-error"))
	fakeJunit := &viewerfakes.FakeJunit{}
	tmpDir, err := ioutil.TempDir("", "get-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	}
	_, err = getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: tmpDir,
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 10,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected err to occur but it did not")
	}
	if !strings.Contains(err.Error(), "some-error") {
		t.Fatalf("expected err to contain 'some-error', but it did not: %s", err)
	}
}

func TestGetErrorOnListBadTimestamp(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-invalid-date.xml",
	}, nil)
	fakeJunit := &viewerfakes.FakeJunit{}
	tmpDir, err := ioutil.TempDir("", "get-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	}
	_, err = getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: tmpDir,
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 10,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected err to occur but it did not")
	}
	if !strings.Contains(err.Error(), "invalid-date") {
		t.Fatalf("expected err to contain 'invalid-date', but it did not: %s", err)
	}
}

func TestGetErrorOnInvalidOutputDir(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-2018-01-02T15:04:05Z.xml",
	}, nil)
	fakeJunit := &viewerfakes.FakeJunit{}

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	}
	_, err := getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: "fake-dir",
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 10,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected err to occur but it did not")
	}
	if !strings.Contains(err.Error(), "fake-dir") {
		t.Fatalf("expected err to contain 'fake-dir', but it did not: %s", err)
	}
}

func TestGetErrorOnGetFail(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-2018-01-02T15:04:05Z.xml",
	}, nil)
	fakeStorage.GetReturns(errors.New("some-error"))

	fakeJunit := &viewerfakes.FakeJunit{}
	tmpDir, err := ioutil.TempDir("", "get-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	}
	_, err = getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: tmpDir,
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 10,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected err to occur but it did not")
	}
	if !strings.Contains(err.Error(), "some-error") {
		t.Fatalf("expected err to contain 'some-error', but it did not: %s", err)
	}
}

func TestGetErrorOnPrintSummaryFail(t *testing.T) {
	fakeStorage := &storagefakes.FakeStorage{}
	fakeStorage.ListReturns([]string{
		"test-results-2018-01-02T15:04:05Z.xml",
	}, nil)

	fakeJunit := &viewerfakes.FakeJunit{}
	fakeJunit.PrintSummaryReturns(errors.New("some-error"))

	tmpDir, err := ioutil.TempDir("", "get-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	getter := in.Getter{
		Storage:     fakeStorage,
		JunitViewer: fakeJunit,
	}

	requestedVersion := models.Version{
		Key: "test-results-2018-01-02T15:04:05Z.xml",
	}
	_, err = getter.Get(models.InRequest{
		Version:   requestedVersion,
		OutputDir: tmpDir,
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 10,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected err to occur but it did not")
	}
	if !strings.Contains(err.Error(), "some-error") {
		t.Fatalf("expected err to contain 'some-error', but it did not: %s", err)
	}
}

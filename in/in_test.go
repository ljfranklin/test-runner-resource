package in_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

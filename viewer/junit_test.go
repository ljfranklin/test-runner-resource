package viewer_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/viewer"
)

func TestJunitPrintSummary(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "junit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fixtureContents, err := ioutil.ReadFile(filepath.Join("..", "fixtures", "junit", "success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "success.xml"), fixtureContents, 0755)
	if err != nil {
		t.Fatal(err)
	}

	output := bytes.Buffer{}
	junit := viewer.JunitCLI{
		OutputWriter: &output,
		ResultsDir:   tmpDir,
	}

	err = junit.PrintSummary(models.Summary{
		Type:  "pass-fail",
		Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "Summary") {
		t.Fatalf("expected output to contain 'Summary' but it did not: %s", output.String())
	}
}

func TestJunitEnforceLimit(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "junit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	newerFixture, err := ioutil.ReadFile(filepath.Join("..", "fixtures", "junit", "success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "test-results-2018-03-15T14:22:46Z.xml"), newerFixture, 0755)
	if err != nil {
		t.Fatal(err)
	}
	olderFixture, err := ioutil.ReadFile(filepath.Join("..", "fixtures", "junit", "failures.xml"))
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "test-results-2018-03-14T14:22:46Z.xml"), olderFixture, 0755)
	if err != nil {
		t.Fatal(err)
	}

	output := bytes.Buffer{}
	junit := viewer.JunitCLI{
		OutputWriter: &output,
		ResultsDir:   tmpDir,
	}

	err = junit.PrintSummary(models.Summary{
		Type:  "pass-fail",
		Limit: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "2018-03-15") {
		t.Fatalf("expected output to contain '2018-03-15' but it did not: %s", output.String())
	}
	if strings.Contains(output.String(), "2018-03-14") {
		t.Fatalf("expected output to not contain '2018-03-14' but it did: %s", output.String())
	}
}

func TestErrorOnInvalidPath(t *testing.T) {
	output := bytes.Buffer{}
	junit := viewer.JunitCLI{
		OutputWriter: &output,
		ResultsDir:   "some-fake-dir",
	}

	err := junit.PrintSummary(models.Summary{
		Type: "pass-fail",
	})
	if err == nil {
		t.Fatal("expected error on invalid path but it succeeded")
	}

	if !strings.Contains(err.Error(), "some-fake-dir") {
		t.Fatalf("expected output to contain 'some-fake-dir' but it did not: %s", err.Error())
	}
}

func TestErrorOnInvalidType(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "junit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fixtureContents, err := ioutil.ReadFile(filepath.Join("..", "fixtures", "junit", "success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "success.xml"), fixtureContents, 0755)
	if err != nil {
		t.Fatal(err)
	}

	output := bytes.Buffer{}
	junit := viewer.JunitCLI{
		OutputWriter: &output,
		ResultsDir:   tmpDir,
	}

	err = junit.PrintSummary(models.Summary{
		Type: "some-invalid-type",
	})
	if err == nil {
		t.Fatal("expected error on invalid type but it succeeded")
	}

	if !strings.Contains(output.String(), "some-invalid-type") {
		t.Fatalf("expected output to contain 'some-invalid-type' but it did not: %s", output.String())
	}
}

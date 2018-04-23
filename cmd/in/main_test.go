package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/test/helpers"
)

var (
	mainPath string
)

func TestMain(m *testing.M) {
	tmpDir, err := ioutil.TempDir("", "in")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	mainPath = buildMain(tmpDir)

	os.Exit(m.Run())
}

func TestInCmd(t *testing.T) {
	t.Parallel()

	config := buildStorageConfig(t)

	tmpDir, err := ioutil.TempDir("", "in")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	awsVerifier := helpers.NewAWSVerifier(config["access_key_id"], config["secret_access_key"], config["region_name"], "")
	nestedBucketPath := filepath.Join(config["path_prefix"], helpers.RandomString("check"))
	uploadedFixtures := []string{
		"test-results-2018-01-02T15:04:05Z.xml",
		"test-results-2018-01-01T15:04:05Z.xml",
		"test-results-2018-01-03T15:04:05Z.xml",
	}

	for _, remotePath := range uploadedFixtures {
		fixture, err := os.Open(fixturePath("junit/success.xml"))
		if err != nil {
			t.Fatal(err)
		}
		defer fixture.Close()

		awsVerifier.UploadObjectToS3(t, config["bucket"], filepath.Join(nestedBucketPath, remotePath), fixture)
		defer awsVerifier.DeleteObjectFromS3(t, config["bucket"], filepath.Join(nestedBucketPath, remotePath))
	}

	inRequest := models.InRequest{
		Version: models.Version{
			Key: "test-results-2018-01-02T15:04:05Z.xml",
		},
		Params: models.InParams{
			Summaries: []models.Summary{
				{
					Type:  "pass-fail",
					Limit: 100,
				},
			},
		},
		Source: models.Source{
			StorageType: "s3",
			StorageConfig: map[string]interface{}{
				"access_key_id":     config["access_key_id"],
				"secret_access_key": config["secret_access_key"],
				"region_name":       config["region_name"],
				"bucket":            config["bucket"],
				"path_prefix":       nestedBucketPath,
			},
		},
	}

	inJSON, err := json.Marshal(inRequest)
	if err != nil {
		t.Fatal(err)
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.Command(mainPath, tmpDir)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		_, err = stdin.Write(inJSON)
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to run in: %s, %s, %s", err, stdout.String(), stderr.String())
	}

	var inOutput models.InResponse
	err = json.Unmarshal(stdout.Bytes(), &inOutput)
	if err != nil {
		t.Fatal(err)
	}

	helpers.AssertEquals(t, inOutput, models.InResponse{
		Version: models.Version{
			Key: "test-results-2018-01-02T15:04:05Z.xml",
		},
		Metadata: map[string]string{
			"test_suite_count": "2",
		},
	})

	if !strings.Contains(stderr.String(), "Summary") {
		t.Fatalf("expected output to contain 'Summary' but it did not: %s", stderr.String())
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "test-results-2018-01-02T15:04:05Z.xml")); os.IsNotExist(err) {
		t.Fatalf("expected '%s' to exist but it does not", "test-results-2018-01-02T15:04:05Z.xml")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "test-results-2018-01-01T15:04:05Z.xml")); os.IsNotExist(err) {
		t.Fatalf("expected '%s' to exist but it does not", "test-results-2018-01-01T15:04:05Z.xml")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "test-results-2018-01-03T15:04:05Z.xml")); !os.IsNotExist(err) {
		t.Fatalf("expected '%s' to NOT exist but it does", "test-results-2018-01-03T15:04:05Z.xml")
	}
}

func TestInCmdErrorOnInvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", "in")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.Command(mainPath, tmpDir)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		_, err = stdin.Write([]byte(`{{{`))
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = cmd.Wait()
	if err == nil {
		t.Fatalf("expected check to err but it did not: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "input JSON") {
		t.Fatalf("expected error to contain 'input JSON' but it did not: %s", stderr.String())
	}
}

func TestInCmdErrorOnInvalidStorageType(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", "in")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	inRequest := models.InRequest{
		Version: models.Version{
			Key: "test-results-2018-01-02T15:04:05Z.xml",
		},
		Source: models.Source{
			StorageType:   "invalid-type",
			StorageConfig: nil,
		},
	}

	inJSON, err := json.Marshal(inRequest)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(mainPath, tmpDir)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		_, err = stdin.Write(inJSON)
		if err != nil {
			t.Fatal(err)
		}
	}()

	combinedOutput, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected check to err but it did not: %s", string(combinedOutput))
	}
	if !strings.Contains(string(combinedOutput), "invalid-type") {
		t.Fatalf("expected error to contain 'invalid-type' but it did not: %s", string(combinedOutput))
	}
}

func TestInCmdErrorOnInvalidCreds(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", "in")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	inRequest := models.InRequest{
		Version: models.Version{
			Key: "test-results-2018-01-02T15:04:05Z.xml",
		},
		Source: models.Source{
			StorageType: "s3",
			StorageConfig: map[string]interface{}{
				"access_key_id":     "fake-keys",
				"secret_access_key": "fake-keys",
				"region_name":       "us-east-1",
				"bucket":            "fake-bucket",
				"path_prefix":       "fake-path",
			},
		},
	}

	inJSON, err := json.Marshal(inRequest)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(mainPath, tmpDir)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		_, err = stdin.Write(inJSON)
		if err != nil {
			t.Fatal(err)
		}
	}()

	combinedOutput, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected in to err but it did not: %s", string(combinedOutput))
	}
	if !strings.Contains(string(combinedOutput), "failed to get") {
		t.Fatalf("expected error to contain 'failed to get' but it did not: %s", string(combinedOutput))
	}
}

func buildMain(tmpDir string) string {
	mainPath := filepath.Join(tmpDir, "in")
	cmd := exec.Command("go", "build", "-o", mainPath, "github.com/ljfranklin/test-runner-resource/cmd/in")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed to build main.go: %s, %s", err, string(output)))
	}

	return mainPath
}

func buildStorageConfig(t *testing.T) map[string]string {
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	if accessKey == "" {
		t.Fatalf("AWS_ACCESS_KEY must be set")
	}
	secretKey := os.Getenv("AWS_SECRET_KEY")
	if secretKey == "" {
		t.Fatalf("AWS_SECRET_KEY must be set")
	}
	bucket := os.Getenv("AWS_BUCKET")
	if bucket == "" {
		t.Fatalf("AWS_BUCKET must be set")
	}
	bucketPath := os.Getenv("AWS_BUCKET_SUBFOLDER")
	if bucketPath == "" {
		t.Fatalf("AWS_BUCKET_SUBFOLDER must be set")
	}
	region := os.Getenv("AWS_REGION") // optional
	if region == "" {
		region = "us-east-1"
	}

	s3Config := map[string]string{
		"access_key_id":     accessKey,
		"secret_access_key": secretKey,
		"region_name":       region,
		"bucket":            bucket,
		"path_prefix":       bucketPath,
	}
	return s3Config
}

func fixturePath(fixture string) string {
	return filepath.Join("..", "..", "fixtures", fixture)
}

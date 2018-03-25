package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/test/helpers"
)

var (
	mainPath string
)

func TestMain(m *testing.M) {
	tmpDir, err := ioutil.TempDir("", "check")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	mainPath = buildMain(tmpDir)

	os.Exit(m.Run())
}

func TestCheckCmd(t *testing.T) {
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

	awsVerifier := helpers.NewAWSVerifier(accessKey, secretKey, region, "")
	nestedBucketPath := filepath.Join(bucketPath, helpers.RandomString("check"))
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

		awsVerifier.UploadObjectToS3(t, bucket, filepath.Join(nestedBucketPath, remotePath), fixture)
		defer awsVerifier.DeleteObjectFromS3(t, bucket, filepath.Join(nestedBucketPath, remotePath))
	}

	checkRequest := models.CheckRequest{
		Source: models.Source{
			StorageType: "s3",
			StorageConfig: map[string]interface{}{
				"access_key_id":     accessKey,
				"secret_access_key": secretKey,
				"region_name":       region,
				"bucket":            bucket,
				"path_prefix":       bucketPath,
			},
		},
	}

	checkJSON, err := json.Marshal(checkRequest)
	if err != nil {
		t.Fatal(err)
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.Command(mainPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		_, err = stdin.Write(checkJSON)
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to run check: %s, %s, %s", err, stdout.String(), stderr.String())
	}

	var checkOutput models.CheckResponse
	err = json.Unmarshal(stdout.Bytes(), &checkOutput)
	if err != nil {
		t.Fatal(err)
	}

	helpers.AssertEquals(t, checkOutput, models.CheckResponse{
		{
			Key: "test-results-2018-01-01T15:04:05Z.xml",
		},
		{
			Key: "test-results-2018-01-02T15:04:05Z.xml",
		},
		{
			Key: "test-results-2018-01-03T15:04:05Z.xml",
		},
	})
}

func TestCheckCmdWithInitialVersion(t *testing.T) {
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

	awsVerifier := helpers.NewAWSVerifier(accessKey, secretKey, region, "")
	nestedBucketPath := filepath.Join(bucketPath, helpers.RandomString("check"))
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

		awsVerifier.UploadObjectToS3(t, bucket, filepath.Join(nestedBucketPath, remotePath), fixture)
		defer awsVerifier.DeleteObjectFromS3(t, bucket, filepath.Join(nestedBucketPath, remotePath))
	}

	checkRequest := models.CheckRequest{
		Source: models.Source{
			StorageType: "s3",
			StorageConfig: map[string]interface{}{
				"access_key_id":     accessKey,
				"secret_access_key": secretKey,
				"region_name":       region,
				"bucket":            bucket,
				"path_prefix":       bucketPath,
			},
		},
		Version: models.Version{
			Key: "test-results-2018-01-02T15:04:05Z.xml",
		},
	}

	checkJSON, err := json.Marshal(checkRequest)
	if err != nil {
		t.Fatal(err)
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.Command(mainPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		_, err = stdin.Write(checkJSON)
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to run check: %s, %s, %s", err, stdout.String(), stderr.String())
	}

	var checkOutput models.CheckResponse
	err = json.Unmarshal(stdout.Bytes(), &checkOutput)
	if err != nil {
		t.Fatal(err)
	}

	helpers.AssertEquals(t, checkOutput, models.CheckResponse{
		{
			Key: "test-results-2018-01-02T15:04:05Z.xml",
		},
		{
			Key: "test-results-2018-01-03T15:04:05Z.xml",
		},
	})
}

func buildMain(tmpDir string) string {
	mainPath := filepath.Join(tmpDir, "check")
	cmd := exec.Command("go", "build", "-o", mainPath, "github.com/ljfranklin/test-runner-resource/cmd/check")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed to build main.go: %s, %s", err, string(output)))
	}

	return mainPath
}

func fixturePath(fixture string) string {
	return filepath.Join("..", "..", "fixtures", fixture)
}

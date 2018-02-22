package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ljfranklin/test-runner-resource/test/helpers"
)

func TestGet(t *testing.T) {
	t.Parallel()

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

	awsVerifier := helpers.NewAWSVerifier(t, accessKey, secretKey, region, "")

	fixture, err := os.Open(filepath.Join(helpers.ProjectRoot(), "storage", "fixtures", "some-file"))
	if err != nil {
		t.Error(err)
	}
	defer fixture.Close()
	s3RemotePath := filepath.Join(bucketPath, helpers.RandomString("s3-get-test"))
	awsVerifier.UploadObjectToS3(bucket, s3RemotePath, fixture)
	defer awsVerifier.DeleteObjectFromS3(bucket, s3RemotePath)

	t.Run("Parallel", func(t *testing.T) {
		t.Run("downloads the file", func(t *testing.T) {
			t.Parallel()
			awsVerifier.ExpectS3ObjectToExist(bucket, s3RemotePath)
		})
	})
}

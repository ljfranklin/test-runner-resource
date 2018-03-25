package storage_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ljfranklin/test-runner-resource/storage"
	"github.com/ljfranklin/test-runner-resource/test/helpers"
)

const (
	// e.g. "2006-01-02T15:04:05Z"
	timeFormat = time.RFC3339
)

type testConfig struct {
	Bucket          string
	BucketPath      string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Endpoint        string
}

func buildTestConfig(t *testing.T) testConfig {
	t.Helper()

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

	return testConfig{
		Bucket:          bucket,
		BucketPath:      bucketPath,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Region:          region,
	}
}

func buildS3CompatibleTestConfig(t *testing.T) testConfig {
	t.Helper()

	accessKey := os.Getenv("S3_COMPATIBLE_ACCESS_KEY")
	if accessKey == "" {
		t.Fatalf("S3_COMPATIBLE_ACCESS_KEY must be set")
	}
	secretKey := os.Getenv("S3_COMPATIBLE_SECRET_KEY")
	if secretKey == "" {
		t.Fatalf("S3_COMPATIBLE_SECRET_KEY must be set")
	}
	bucket := os.Getenv("S3_COMPATIBLE_BUCKET")
	if bucket == "" {
		t.Fatalf("S3_COMPATIBLE_BUCKET must be set")
	}
	bucketPath := os.Getenv("S3_COMPATIBLE_BUCKET_SUBFOLDER")
	if bucketPath == "" {
		t.Fatalf("S3_COMPATIBLE_BUCKET_SUBFOLDER must be set")
	}
	endpoint := os.Getenv("S3_COMPATIBLE_ENDPOINT")
	if bucketPath == "" {
		t.Fatalf("S3_COMPATIBLE_ENDPOINT must be set")
	}

	return testConfig{
		Bucket:          bucket,
		BucketPath:      bucketPath,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
	}
}

func TestS3Get(t *testing.T) {
	t.Parallel()

	c := buildTestConfig(t)

	testGet(t, c)
}

func TestS3CompatibleGet(t *testing.T) {
	t.Parallel()

	c := buildS3CompatibleTestConfig(t)

	testGet(t, c)
}

func testGet(t *testing.T, c testConfig) {
	awsVerifier := helpers.NewAWSVerifier(c.AccessKeyID, c.SecretAccessKey, c.Region, c.Endpoint)

	fixture, err := os.Open(fixturePath("some-file"))
	if err != nil {
		t.Fatal(err)
	}
	defer fixture.Close()
	s3RemotePath := filepath.Join(c.BucketPath, helpers.RandomString("s3-get-test"))
	awsVerifier.UploadObjectToS3(t, c.Bucket, s3RemotePath, fixture)
	defer awsVerifier.DeleteObjectFromS3(t, c.Bucket, s3RemotePath)

	s3, err := storage.New("s3", buildS3Config(c))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Get", func(t *testing.T) {
		t.Run("downloads the file", func(t *testing.T) {
			t.Parallel()

			awsVerifier.ExpectS3ObjectToExist(t, c.Bucket, s3RemotePath)

			fileContents := bytes.Buffer{}
			err := s3.Get(s3RemotePath, &fileContents)
			if err != nil {
				t.Fatal(err)
			}

			expectedFileContents := []byte("some-file-contents\n")
			if !bytes.Equal(fileContents.Bytes(), expectedFileContents) {
				t.Fatalf("expected '%s' to equal '%s'", fileContents.Bytes(), expectedFileContents)
			}
		})

		t.Run("errors on non-existent key", func(t *testing.T) {
			t.Parallel()

			awsVerifier.ExpectS3ObjectToExist(t, c.Bucket, s3RemotePath)

			fileContents := bytes.Buffer{}
			err := s3.Get("key-that-does-not-exist", &fileContents)
			if err == nil {
				t.Fatal("expected error to occur on missing file")
			}

			expectedErr := "key-that-does-not-exist"
			if !strings.Contains(err.Error(), expectedErr) {
				t.Fatalf("expected '%s' to contain '%s'", err.Error(), expectedErr)
			}
		})
	})
}

func TestS3Delete(t *testing.T) {
	t.Parallel()

	c := buildTestConfig(t)

	testDelete(t, c)
}

func TestS3CompatibleDelete(t *testing.T) {
	t.Parallel()

	c := buildS3CompatibleTestConfig(t)

	testDelete(t, c)
}

func testDelete(t *testing.T, c testConfig) {
	awsVerifier := helpers.NewAWSVerifier(c.AccessKeyID, c.SecretAccessKey, c.Region, c.Endpoint)

	fixture, err := os.Open(fixturePath("some-file"))
	if err != nil {
		t.Fatal(err)
	}
	defer fixture.Close()

	s3RemotePath := filepath.Join(c.BucketPath, helpers.RandomString("s3-get-test"))
	awsVerifier.UploadObjectToS3(t, c.Bucket, s3RemotePath, fixture)

	s3, err := storage.New("s3", buildS3Config(c))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Delete", func(t *testing.T) {
		t.Run("deletes the file", func(t *testing.T) {
			t.Parallel()

			awsVerifier.ExpectS3ObjectToExist(t, c.Bucket, s3RemotePath)

			err := s3.Delete(s3RemotePath)
			if err != nil {
				t.Fatal(err)
			}

			awsVerifier.ExpectS3ObjectToNotExist(t, c.Bucket, s3RemotePath)
		})
	})
}

func TestS3Put(t *testing.T) {
	t.Parallel()

	c := buildTestConfig(t)

	testPut(t, c)
}

func TestS3CompatiblePut(t *testing.T) {
	t.Parallel()

	c := buildS3CompatibleTestConfig(t)

	testPut(t, c)
}

func testPut(t *testing.T, c testConfig) {
	awsVerifier := helpers.NewAWSVerifier(c.AccessKeyID, c.SecretAccessKey, c.Region, c.Endpoint)

	s3, err := storage.New("s3", buildS3Config(c))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Delete", func(t *testing.T) {
		t.Run("uploads the file", func(t *testing.T) {
			t.Parallel()

			s3RemotePath := filepath.Join(c.BucketPath, helpers.RandomString("s3-upload-test"))
			awsVerifier.ExpectS3ObjectToNotExist(t, c.Bucket, s3RemotePath)

			fixture, err := os.Open(fixturePath("some-file"))
			if err != nil {
				t.Fatal(err)
			}
			defer fixture.Close()

			err = s3.Put(s3RemotePath, fixture)
			if err != nil {
				t.Fatal(err)
			}
			defer awsVerifier.DeleteObjectFromS3(t, c.Bucket, s3RemotePath)

			awsVerifier.ExpectS3ObjectToExist(t, c.Bucket, s3RemotePath)
		})

		t.Run("returns error on invalid input file", func(t *testing.T) {
			t.Parallel()

			badReader := errReader{}
			err := s3.Put("some-upload-path", badReader)
			if err == nil {
				t.Fatal("expected an error on `nil` file param")
			}
			if !strings.Contains(err.Error(), "some-read-error") {
				t.Fatalf("expected '%s' to contain 'some-read-error'", err.Error())
			}
		})
	})
}

func TestS3List(t *testing.T) {
	t.Parallel()

	c := buildTestConfig(t)

	testList(t, c)
}

func TestS3CompatibleList(t *testing.T) {
	t.Parallel()

	c := buildS3CompatibleTestConfig(t)

	testList(t, c)
}

func testList(t *testing.T, c testConfig) {
	awsVerifier := helpers.NewAWSVerifier(c.AccessKeyID, c.SecretAccessKey, c.Region, c.Endpoint)

	nestedBucketPath := filepath.Join(c.BucketPath, helpers.RandomString("s3-list"))
	uploadedFixtures := []string{
		helpers.RandomString("s3-list-test"),
		helpers.RandomString("s3-list-test"),
		helpers.RandomString("s3-list-test"),
	}

	for _, remotePath := range uploadedFixtures {
		fixture, err := os.Open(fixturePath("some-file"))
		if err != nil {
			t.Fatal(err)
		}
		defer fixture.Close()

		awsVerifier.UploadObjectToS3(t, c.Bucket, filepath.Join(nestedBucketPath, remotePath), fixture)
		defer awsVerifier.DeleteObjectFromS3(t, c.Bucket, filepath.Join(nestedBucketPath, remotePath))
	}

	t.Run("List", func(t *testing.T) {
		t.Run("returns all files in bucket path", func(t *testing.T) {
			t.Parallel()

			config := buildS3Config(c)
			config["path_prefix"] = nestedBucketPath

			s3, err := storage.New("s3", config)
			if err != nil {
				t.Fatal(err)
			}

			results, err := s3.List()
			if err != nil {
				t.Fatal(err)
			}

			if len(results) != 3 {
				t.Fatalf("expected '%#v' to have length 3", results)
			}

			sort.Strings(results)
			sort.Strings(uploadedFixtures)

			for i := range uploadedFixtures {
				if results[i] != uploadedFixtures[i] {
					t.Fatalf("expected '%s' to equal '%s'", results[i], uploadedFixtures[i])
				}
			}
		})

		t.Run("returns empty list if prefix is empty", func(t *testing.T) {
			t.Parallel()

			config := buildS3Config(c)
			config["path_prefix"] = "path-that-does-not-exist"

			s3, err := storage.New("s3", config)
			if err != nil {
				t.Fatal(err)
			}

			results, err := s3.List()
			if err != nil {
				t.Fatal(err)
			}

			if len(results) != 0 {
				t.Fatalf("expected '%#v' to be empty", results)
			}
		})
	})
}

func TestErrorOnInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := storage.New("s3", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error to occur but it did not")
	}

	requiredProps := []string{
		"access_key_id",
		"secret_access_key",
		"bucket",
	}
	for _, prop := range requiredProps {
		if !strings.Contains(err.Error(), prop) {
			t.Fatalf("expected error to contain '%s' but it does not: %s", prop, err)
		}
	}
}

func buildS3Config(c testConfig) map[string]interface{} {
	s3Config := map[string]interface{}{
		"access_key_id":     c.AccessKeyID,
		"secret_access_key": c.SecretAccessKey,
		"region_name":       c.Region,
		"bucket":            c.Bucket,
		"path_prefix":       c.BucketPath,
	}
	if len(c.Region) > 0 {
		s3Config["region_name"] = c.Region
	}
	if len(c.Endpoint) > 0 {
		s3Config["endpoint"] = c.Endpoint
	}
	return s3Config
}

func fixturePath(fixtureName string) string {
	return filepath.Join(helpers.ProjectRoot(), "storage", "fixtures", fixtureName)
}

type errReader struct{}

func (e errReader) Read(p []byte) (int, error) {
	return 0, errors.New("some-read-error")
}

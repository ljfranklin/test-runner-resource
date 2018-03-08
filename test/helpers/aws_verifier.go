package helpers

import (
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ljfranklin/test-runner-resource/storage"
)

type AWSVerifier struct {
	s3 *awss3.S3
}

func NewAWSVerifier(accessKey string, secretKey string, region string, endpoint string) *AWSVerifier {
	if len(region) == 0 {
		region = " " // aws sdk complains if region is empty
	}
	awsConfig := &aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
		MaxRetries:       aws.Int(10),
	}
	if len(endpoint) > 0 {
		awsConfig.Endpoint = aws.String(endpoint)
	}

	s3 := awss3.New(awsSession.New(awsConfig))
	if len(endpoint) > 0 {
		// many s3-compatible endpoints only support v2 signing
		storage.Setv2Handlers(s3)
	}

	return &AWSVerifier{
		s3: s3,
	}
}

func (a AWSVerifier) ExpectS3ObjectToExist(t *testing.T, bucketName string, key string) {
	t.Helper()

	params := &awss3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err := a.s3.HeadObject(params)
	if err != nil {
		t.Fatalf(
			"Expected S3 file '%s' to exist in bucket '%s', but it does not",
			key,
			bucketName)
	}
}

func (a AWSVerifier) ExpectS3ObjectToNotExist(t *testing.T, bucketName string, key string) {
	t.Helper()

	params := &awss3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err := a.s3.HeadObject(params)
	if err == nil {
		t.Fatalf(
			"Expected S3 file '%s' to not exist in bucket '%s', but it does",
			key,
			bucketName)
	}

	reqErr, ok := err.(awserr.RequestFailure)
	if !ok {
		t.Fatalf("Invalid AWS error type: %s", err)
	}
	if reqErr.StatusCode() != 404 {
		t.Fatalf("Expected req to return 404 but was %s", reqErr.StatusCode())
	}
}

func (a AWSVerifier) UploadObjectToS3(t *testing.T, bucketName string, key string, content io.Reader) {
	t.Helper()

	uploader := s3manager.NewUploaderWithClient(a.s3)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   content,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (a AWSVerifier) DeleteObjectFromS3(t *testing.T, bucketName string, key string) {
	t.Helper()

	deleteInput := &awss3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	_, err := a.s3.DeleteObject(deleteInput)
	if err != nil {
		t.Fatal(err)
	}
}

func (a AWSVerifier) GetS3ObjectLastModified(t *testing.T, bucketName string, key string, timeFormat string) string {
	t.Helper()

	params := &awss3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	resp, err := a.s3.HeadObject(params)
	if err != nil {
		t.Fatalf(
			"Expected S3 file '%s' to exist in bucket '%s', but it does not",
			key,
			bucketName)
	}

	return resp.LastModified.Format(timeFormat)
}

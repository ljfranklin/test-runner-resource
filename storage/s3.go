package storage

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3 struct {
	bucket          string
	accessKeyID     string
	secretAccessKey string
	regionName      string
	endpoint        string
	prefix          string
	useV4Signing    bool
	client          *awss3.S3
	uploader        *s3manager.Uploader
}

const (
	maxRetries    = 10
	defaultRegion = "us-east-1"
)

func NewS3(config map[string]interface{}) Storage {
	s3 := &s3{
		bucket:          config["bucket"].(string),
		accessKeyID:     config["access_key_id"].(string),
		secretAccessKey: config["secret_access_key"].(string),
		regionName:      config["region_name"].(string),
		prefix:          config["path_prefix"].(string),
	}
	if endpoint, ok := config["endpoint"]; ok {
		s3.endpoint = endpoint.(string)
	}
	if useV4Signing, ok := config["use_v4_signing"]; ok {
		s3.useV4Signing = useV4Signing.(bool)
	}

	creds := credentials.NewStaticCredentials(s3.accessKeyID, s3.secretAccessKey, "")

	regionName := s3.regionName
	if len(regionName) == 0 {
		regionName = defaultRegion
	}

	awsConfig := &aws.Config{
		Region:           aws.String(regionName),
		Credentials:      creds,
		S3ForcePathStyle: aws.Bool(true),
		MaxRetries:       aws.Int(maxRetries),
		Logger:           nil,
	}
	if len(s3.endpoint) > 0 {
		awsConfig.Endpoint = aws.String(s3.endpoint)
	}

	session := awsSession.New(awsConfig)
	s3.client = awss3.New(session, awsConfig)
	if len(s3.endpoint) > 0 && !s3.useV4Signing {
		Setv2Handlers(s3.client)
	}
	s3.uploader = s3manager.NewUploaderWithClient(s3.client)
	if s3.isGCSHost() {
		// GCS returns `InvalidArgument` on multipart uploads
		s3.uploader.MaxUploadParts = 1
	}

	return s3
}

func (s *s3) Get(key string, destination io.Writer) error {
	params := &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	resp, err := s.client.GetObject(params)
	if err != nil {
		return fmt.Errorf("unable to fetch '%s': %s", key, err.Error())
	}
	defer resp.Body.Close()

	_, err = io.Copy(destination, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy download to local file: %s", err)
	}

	return nil
}

func (s *s3) Put(key string, source io.Reader) error {
	params := &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   source,
	}

	_, err := s.uploader.Upload(params)
	if err != nil {
		return fmt.Errorf("unable to upload '%s': %s", key, err.Error())
	}

	return nil
}

func (s *s3) Delete(key string) error {
	params := &awss3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(params)
	if err != nil {
		return fmt.Errorf("unable to delete '%s': %s", key, err.Error())
	}

	return nil
}

func (s *s3) List() ([]string, error) {
	params := &awss3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.prefix),
	}

	objects := []*awss3.Object{}
	err := s.client.ListObjectsPages(params,
		func(page *awss3.ListObjectsOutput, lastPage bool) bool {
			objects = append(objects, page.Contents...)
			return true
		})
	if err != nil {
		return nil, fmt.Errorf("unable to list bucket '%s' with '%s': %s", s.bucket, s.prefix, err.Error())
	}

	results := []string{}
	for _, obj := range objects {
		key := path.Base(*obj.Key)
		results = append(results, key)
	}

	return results, nil
}

func (s *s3) isGCSHost() bool {
	return (s.endpoint != "" && strings.Contains(s.endpoint, "storage.googleapis.com"))
}

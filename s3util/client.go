package s3util

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

type S3WrapperClient struct {
	*s3.S3
}

var instance *S3WrapperClient

// Instance Retrieves the instance of the client
func Instance() *S3WrapperClient {
	return instance
}

// Initialize Initializes a new s3 client
func Initialize() {
	if instance != nil {
		logrus.Error("S3 client is already initialized")
		return
	}

	key := config.Instance.S3.AccessKey
	secret := config.Instance.S3.Secret

	cfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String(config.Instance.S3.Endpoint),
		Region:           aws.String(config.Instance.S3.Region),
		S3ForcePathStyle: aws.Bool(false),
	}

	sess, err := session.NewSession(cfg)

	if err != nil {
		panic(err)
	}

	instance = &S3WrapperClient{s3.New(sess)}
	logrus.Info("S3 client has been initialized!")
}

// UploadFile Uploads a file to s3.
func (client *S3WrapperClient) UploadFile(folder string, fileName string, path string) error {
	sess := session.Must(session.NewSession(&client.Config))
	uploader := s3manager.NewUploader(sess)

	uploader.PartSize = 5 * 1024 * 1024
	uploader.Concurrency = 16

	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.Instance.S3.Bucket),
		Key:    aws.String(fmt.Sprintf("%v/%v", folder, fileName)),
		Body:   file,
	})

	if err != nil {
		return err
	}

	return nil
}

// ListFiles Lists files in a given bucket & folder
func (client *S3WrapperClient) ListFiles(folder string) ([]string, error) {
	output, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(config.Instance.S3.Bucket),
	})

	if err != nil {
		return nil, err
	}

	var files []string

	for _, object := range output.Contents {
		directory := fmt.Sprintf("%v/", folder)
		fileName := aws.StringValue(object.Key)

		if strings.HasPrefix(fileName, directory) && fileName != directory {
			files = append(files, fileName)
		}
	}

	return files, nil
}

// DeleteFile Delete a file from s3
func (client *S3WrapperClient) DeleteFile(folder string, file string) error {
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.Instance.S3.Bucket),
		Key:    aws.String(fmt.Sprintf("%v/%v", folder, file)),
	})

	if err != nil {
		return err
	}

	return nil
}

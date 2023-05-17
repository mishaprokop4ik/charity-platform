package repository

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type AWSConfig struct {
	AccessKey       string
	SecretAccessKey string
	Region          string
	BucketName      string
}

type AWSFile struct {
	config AWSConfig
}

func NewFile(config AWSConfig) *AWSFile {
	return &AWSFile{config: config}
}

func (f *AWSFile) Get(ctx context.Context, identifier string) (io.Reader, error) {
	return nil, nil
}

func (f *AWSFile) Upload(ctx context.Context, fileName string, fileData io.Reader) (string, error) {
	awsSession, err := f.connectToAWS()
	if err != nil {
		return "", err
	}

	uploader := s3manager.NewUploader(awsSession)
	uploadedFile, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(f.config.BucketName),
		Key:    aws.String("images/" + fileName),
		Body:   fileData,
	})

	if err != nil {
		return "", err
	}

	return uploadedFile.Location, nil
}

func (f *AWSFile) Delete(ctx context.Context, identifier string) error {
	awsSession, err := f.connectToAWS()
	if err != nil {
		return err
	}

	objects := []s3manager.BatchDeleteObject{
		{
			Object: &s3.DeleteObjectInput{
				Key:    aws.String(identifier),
				Bucket: aws.String(f.config.BucketName),
			},
		},
	}

	deleter := s3manager.NewBatchDelete(awsSession)

	return deleter.Delete(ctx, &s3manager.DeleteObjectsIterator{Objects: objects})
}

func (f *AWSFile) connectToAWS() (*session.Session, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(f.config.Region),
			Credentials: credentials.NewStaticCredentials(
				f.config.AccessKey,
				f.config.SecretAccessKey,
				"",
			),
		})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

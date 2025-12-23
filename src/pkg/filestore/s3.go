package filestore

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	Client     *s3.Client
	BucketName string
}

func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
	return &S3Storage{
		Client:     client,
		BucketName: bucket,
	}
}

func (s *S3Storage) Save(
	ctx context.Context,
	file io.Reader,
	filename string,
	grade string,
	subject string,
) (string, error) {

	key := fmt.Sprintf("grade-%s/%s/%s", grade, subject, filename)

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) ListByGradeSubject(
	ctx context.Context,
	grade string,
	subject string,
) ([]string, error) {

	prefix := fmt.Sprintf("grade-%s/%s/", grade, subject)

	resp, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.BucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, obj := range resp.Contents {
		keys = append(keys, *obj.Key)
	}
	return keys, nil
}

func (s *S3Storage) GenerateSignedURL(
	ctx context.Context,
	key string,
) (string, error) {

	presigner := s3.NewPresignClient(s.Client)

	resp, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", err
	}

	return resp.URL, nil
}

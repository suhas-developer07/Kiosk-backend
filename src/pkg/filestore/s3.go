package filestore

// import (
// 	"context"
// 	"io"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/service/s3"
// )

// type S3Storage struct {
// 	Client *s3.Client
// 	Bucket string
// }

// func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
// 	return &S3Storage{Client: client, Bucket: bucket}
// }

// func (s *S3Storage) Save(file io.Reader, filename string) (string, error) {
// 	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
// 		Bucket: aws.String(s.Bucket),
// 		Key:    aws.String(filename),
// 		Body:   file,
// 	})
// 	if err != nil {
// 		return "", err
// 	}

// 	return "https://" + s.Bucket + ".s3.amazonaws.com/" + filename, nil
// }

// func (s *S3Storage) Delete(path string) error {
// 	// extract file key
// 	key := path[strings.LastIndex(path, "/")+1:]

// 	_, err := s.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
// 		Bucket: aws.String(s.Bucket),
// 		Key:    aws.String(key),
// 	})

// 	return err
// }

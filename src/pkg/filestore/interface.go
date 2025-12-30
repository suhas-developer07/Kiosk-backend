package filestore

import (
	"context"
	"io"
)

type FileStorage interface {
	Save(
		ctx context.Context,
		file io.Reader,
		filename string,
		grade string,
		subject string,
	) (string, string, error)

	Delete(ctx context.Context, key string) error

	ListByGradeSubject(
		ctx context.Context,
		grade string,
		subject string,
	) ([]string, error)

	GenerateSignedURL(
		ctx context.Context,
		key string,
	) (string, error)
}

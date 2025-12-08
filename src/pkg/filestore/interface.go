package filestore

import "io"

type FileStorage interface {
	Save(file io.Reader, filename string) (string, error)
	Delete(path string) error
}

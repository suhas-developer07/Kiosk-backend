package filestore

import (
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BasePath string
}

func NewLocalStorage(base string) *LocalStorage {
	return &LocalStorage{BasePath: base}
}

func (l *LocalStorage) Save(file io.Reader, filename string) (string, error) {
	if err := os.MkdirAll(l.BasePath, 0755); err != nil {
		return "", err
	}

	fullPath := filepath.Join(l.BasePath, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return fullPath, nil
}

func (l *LocalStorage) Delete(path string) error {
	return os.Remove(path)
}

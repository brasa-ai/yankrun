package services

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	ReadDir(path string) ([]os.FileInfo, error)
	Stat(path string) (os.FileInfo, error)
	EnsureDir(path string) error
	Join(elem ...string) string
}

type OsFileSystem struct{}

func (o *OsFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (o *OsFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (o *OsFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (o *OsFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (o *OsFileSystem) EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %s", err)
		}
	}
	return nil
}

func (o *OsFileSystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}

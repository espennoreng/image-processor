package adapter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type LocalStorageAdapter struct {
	BasePath string
}

func NewLocalStorageAdapter(basePath string) (*LocalStorageAdapter, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("could not create base path '%s': %w", basePath, err)
	}
	return &LocalStorageAdapter{BasePath: basePath}, nil
}

func (l *LocalStorageAdapter) Download(ctx context.Context, bucket, object string) ([]byte, error) {
	return os.ReadFile(object)
}

func (l *LocalStorageAdapter) Upload(ctx context.Context, bucket, object string, data []byte) error {
	fullPath := filepath.Join(l.BasePath, object)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory '%s': %w", dir, err)
	}

	return os.WriteFile(fullPath, data, 0644)
}

func (l *LocalStorageAdapter) Delete(ctx context.Context, bucket, object string) error {
	return os.Remove(object)
}

package image_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/espennoreng/image-processor/internal/config"
	"github.com/espennoreng/image-processor/internal/image"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) Download(ctx context.Context, bucket, object string) ([]byte, error) {
	args := m.Called(ctx, bucket, object)
	if data, ok := args.Get(0).([]byte); ok {
		return data, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorageService) Upload(ctx context.Context, bucket, object string, data []byte) error {
	args := m.Called(ctx, bucket, object, data)
	return args.Error(0)
}

func (m *MockStorageService) Delete(ctx context.Context, bucket, object string) error {
	args := m.Called(ctx, bucket, object)
	return args.Error(0)
}

func TestProcessImage(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Load the sample data
	testImageData, err := os.ReadFile("testdata/gopher.png")
	require.NoError(t, err, "failed to read test image")

	t.Run("successful processing", func(t *testing.T) {
		mockStorage := new(MockStorageService)
		event := config.GCSEvent{Bucket: "test-bucket", Name: "uploads/gopher.png"}
		config := config.Config{SmallImgDir: "small/", MedImgDir: "medium/", OrgImgDir: "original/", FileDir: "uploads/"}
		processor := image.NewProcessor(mockStorage, &config, log)

		mockStorage.On("Download", ctx, event.Bucket, event.Name).Return(testImageData, nil)

		mockStorage.On("Upload", ctx, event.Bucket, "small/gopher.png", mock.Anything).Return(nil)
		mockStorage.On("Upload", ctx, event.Bucket, "medium/gopher.png", mock.Anything).Return(nil)
		mockStorage.On("Upload", ctx, event.Bucket, "original/gopher.png", testImageData).Return(nil)

		mockStorage.On("Delete", ctx, event.Bucket, event.Name).Return(nil)

		err := processor.Process(ctx, event, true)
		require.NoError(t, err, "expected no error during processing")

		mockStorage.AssertExpectations(t)
	})

	t.Run("file not in uploads directory", func(t *testing.T) {
		mockStorage := new(MockStorageService)
		event := config.GCSEvent{Bucket: "test-bucket", Name: "gopher.png"}
		config := config.Config{SmallImgDir: "small/", MedImgDir: "medium/", OrgImgDir: "original/", FileDir: "uploads/"}
		processor := image.NewProcessor(mockStorage, &config, log)

		err := processor.Process(ctx, event, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not in the uploads/ directory")

		mockStorage.AssertExpectations(t)
	})

	t.Run("upload failure", func(t *testing.T) {
		mockStorage := new(MockStorageService)
		event := config.GCSEvent{Bucket: "test-bucket", Name: "uploads/gopher.png"}
		config := config.Config{SmallImgDir: "small/", MedImgDir: "medium/", OrgImgDir: "original/", FileDir: "uploads/"}
		processor := image.NewProcessor(mockStorage, &config, log)

		mockStorage.On("Download", ctx, event.Bucket, event.Name).Return(testImageData, nil)

		mockStorage.On("Upload", ctx, event.Bucket, "small/gopher.png", mock.Anything).Return(nil)
		mockStorage.On("Upload", ctx, event.Bucket, "medium/gopher.png", mock.Anything).Return(nil)
		mockStorage.On("Upload", ctx, event.Bucket, "original/gopher.png", testImageData).Return(os.ErrInvalid)

		err := processor.Process(ctx, event, true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to process and upload original/gopher.png")

		mockStorage.AssertExpectations(t)
	})

	t.Run("download failure", func(t *testing.T) {
		mockStorage := new(MockStorageService)
		event := config.GCSEvent{Bucket: "test-bucket", Name: "uploads/gopher.png"}
		config := config.Config{SmallImgDir: "small/", MedImgDir: "medium/", OrgImgDir: "original/", FileDir: "uploads/"}
		processor := image.NewProcessor(mockStorage, &config, log)

		mockStorage.On("Download", ctx, event.Bucket, event.Name).Return(nil, os.ErrNotExist)

		err := processor.Process(ctx, event, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to download image")

		mockStorage.AssertExpectations(t)
	})
}
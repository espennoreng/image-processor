package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/espennoreng/image-processor/internal/config"
)

type StorageService interface {
	Download(ctx context.Context, bucket, object string) ([]byte, error)
	Upload(ctx context.Context, bucket, object string, data []byte) error
	Delete(ctx context.Context, bucket, object string) error
}

type processor struct {
	storage StorageService
	config  *config.Config
	log     *slog.Logger
}

type Processor interface {
	Process(ctx context.Context, event config.GCSEvent, deleteOriginal bool) error
}

func NewProcessor(storage StorageService, cfg *config.Config, log *slog.Logger) Processor {
	return &processor{
		storage: storage,
		config:  cfg,
		log:     log.With("component", "image-processor"),
	}
}

func (p *processor) Process(ctx context.Context, event config.GCSEvent, deleteOriginal bool) error {
	log := p.log.With("bucket", event.Bucket, "object", event.Name)
	log.Info("Processing image")
	// Validate the trigger event

	if !strings.HasPrefix(event.Name, p.config.FileDir) {
		log.Error("File not in the correct directory", "file", event.Name, "expectedDir", p.config.FileDir)
		return fmt.Errorf("file %s is not in the %s directory", event.Name, p.config.FileDir)
	}

	if !p.isSupportedFormat(event.Name) {
		log.Error("Unsupported file format", "file", event.Name)
		return fmt.Errorf("unsupported file format for %s", event.Name)
	}

	// Download the original image
	imageData, err := p.storage.Download(ctx, event.Bucket, event.Name)
	if err != nil {
		log.Error("Failed to download image", "error", err)
		return fmt.Errorf("failed to download image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Error("Failed to decode image", "error", err)
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Concurrently process and upload images
	baseFilename := filepath.Base(event.Name)
	target := []config.ImageTarget{
		{Path: filepath.Join(p.config.SmallImgDir, baseFilename), Width: 400, Quality: 65},
		{Path: filepath.Join(p.config.MedImgDir, baseFilename), Width: 800, Quality: 75},
		{Path: filepath.Join(p.config.OrgImgDir, baseFilename), IsOrg: true},
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(target))

	for _, target := range target {
		wg.Add(1)
		go func(t config.ImageTarget) {
			defer wg.Done()
			if err := p.processAndUpload(ctx, event.Bucket, img, imageData, t); err != nil {
				errs <- fmt.Errorf("failed to process and upload %s: %w", t.Path, err)
			}
		}(target)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			log.Error("Error during processing", "error", err)
			return err
		}
	}

	// Delete the original upload
	log.Info("Successfully processed images")
	
	if deleteOriginal {
		if err := p.storage.Delete(ctx, event.Bucket, event.Name); err != nil {
			log.Error("Failed to delete original image", "error", err)
			return fmt.Errorf("failed to delete original image: %w", err)
		}
		log.Info("Original image deleted successfully")
	} else {
		log.Info("Skipping deletion of original image as per configuration")
	}

	log.Info("Image processing completed successfully")
	return nil
}

func (p *processor) processAndUpload(ctx context.Context, bucket string, img image.Image, originalData []byte, target config.ImageTarget) error {
	var outputBuffer bytes.Buffer

	if target.IsOrg {
		outputBuffer.Write(originalData)
	} else {
		resizedImg := imaging.Resize(img, target.Width, 0, imaging.Lanczos)
		encodeOptions := imaging.JPEGQuality(target.Quality)
		if err := imaging.Encode(&outputBuffer, resizedImg, imaging.JPEG, encodeOptions); err != nil {
			return fmt.Errorf("failed to encode image: %w", err)
		}
	}

	return p.storage.Upload(ctx, bucket, target.Path, outputBuffer.Bytes())
}

func (p *processor) isSupportedFormat(filename string) bool {
	supportedExtensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	return false
}
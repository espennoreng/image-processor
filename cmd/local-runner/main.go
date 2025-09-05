// Package main provides a command-line tool to run the image processor locally.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/espennoreng/image-processor/internal/adapter"
	"github.com/espennoreng/image-processor/internal/config"
	"github.com/espennoreng/image-processor/internal/image"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log := slog.New(slog.NewTextHandler(log.Writer(), nil))
	ctx := context.Background()

	inputFile := flag.String("file", "", "Path to the input image file (e.g., ./uploads/my-image.jpg)")
	outputDir := flag.String("output", "./output", "The root directory for processed images.")
	flag.Parse()

	if *inputFile == "" {
		log.Error("Input file is required. Use -file to specify the path.")
		return fmt.Errorf("input file is required")
	}

	localAdapter, err := adapter.NewLocalStorageAdapter(*outputDir)
	if err != nil {
		log.Error("Failed to initialize local storage adapter", "error", err)
		return fmt.Errorf("failed to initialize local storage adapter: %w", err)
	}

	imgConfig := &config.Config{
		SmallImgDir: "small/",
		MedImgDir:   "medium/",
		OrgImgDir:   "original/",
		FileDir:     "./uploads/",
	}
	log.Info("Using configuration", "smallDir", imgConfig.SmallImgDir, "medDir", imgConfig.MedImgDir, "orgDir", imgConfig.OrgImgDir)

	imageProcessor := image.NewProcessor(localAdapter, imgConfig, log)

	// Create a fake event object based on the input file.
	event := config.GCSEvent{
		Bucket: "local-bucket", // Bucket name is ignored by the local adapter
		Name:   *inputFile,     // e.g., "upload/my-image.jpg"
	}

	log.Info("Processing file", "file", *inputFile)

	if err := imageProcessor.Process(ctx, event, false); err != nil {
		log.Error("Processor failed", "error", err)
		return fmt.Errorf("processor failed: %w", err)
	}

	log.Info("Successfully processed file", "file", *inputFile, "outputDir", *outputDir)

	return nil
}

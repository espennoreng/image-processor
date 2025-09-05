# image-processor

Lightweight Go image processing utility used as a local runner or Cloud Function to resize images and write different sizes to storage.

This repository contains:

- A Cloud Function entrypoint in `cmd/cloud-function`.
- A local CLI runner in `cmd/local-runner` for testing using a local storage adapter.
- Image processing logic in `internal/image/processor.go`.
- A small local storage adapter in `internal/adapter` (used by the local runner).
- Configuration and tests in `internal/config` and `internal/image`.

The processor resizes uploaded images into small and medium JPEGs and writes an original copy to the configured output locations.

## Requirements

- Go 1.25.1 (from `go.mod`)
- The project uses `github.com/disintegration/imaging` for image resize/encode operations.

## Quick start — run locally

This project includes a small CLI (`cmd/local-runner`) that simulates an upload event and writes processed images to a local directory tree.

1. Build the local runner:

```bash
go build -o bin/local-runner ./cmd/local-runner
```

2. Run it against a file inside `./uploads` (the code expects uploaded files under `./uploads/` by default):

```bash
./bin/local-runner -file ./uploads/gopher.png -output ./output
```
or with Makefile
```bash
make local-runner
```

3. Output layout

By default the processor uses these directories (relative to the configured output root):

- `small/` — resized small images (width 400, quality 65)
- `medium/` — resized medium images (width 800, quality 75)
- `original/` — original file copy

So for the example above you'll find processed images under `./output/small/`, `./output/medium/` and `./output/original/`.

## How it works

- The Cloud Function or local runner produces a `config.GCSEvent` (bucket + object name) and calls the `Processor.Process` method.
- `Processor` validates the event (file must be in the configured `FileDir` and have a supported extension), downloads the original bytes from the configured `StorageService`, decodes the image, and concurrently produces target images:
  - Small (400px wide, JPEG quality 65)
  - Medium (800px wide, JPEG quality 75)
  - Original copy (no change)
- Each produced image is uploaded using the provided `StorageService` implementation.

Supported input formats: `.jpg`, `.jpeg`, `.png`, `.webp`.

## Configuration

Configuration is defined by `internal/config.Config` and used by the processor. The local runner sets an example configuration in `cmd/local-runner/main.go`:

- `SmallImgDir` — directory path for small images (e.g. `small/`).
- `MedImgDir` — directory path for medium images (e.g. `medium/`).
- `OrgImgDir` — directory path for original images (e.g. `original/`).
- `FileDir` — the expected directory prefix for incoming uploads (e.g. `./uploads/`).

When running as a Cloud Function, wire your config values from environment variables or your preferred config loader.

## Tests

There is a basic unit test for the image processor in `internal/image/processor_test.go` that uses the `internal/image/testdata/gopher.png` fixture.

Run tests with:

```bash
go test ./...
```
or with Makefile
```bash
make test
```

## Development notes & suggestions

- The processor writes JPEG for resized outputs via `github.com/disintegration/imaging`.
- The local storage adapter in `internal/adapter` implements a small, filesystem-based `StorageService` for local testing. When deploying to cloud storage, provide an implementation that talks to your object store (GCS, S3, etc.).
- The processor currently decodes the image once and reuses the decoded image for resize operations; original bytes are preserved and uploaded for the original copy target.

## License

This project is licensed under the MIT License — see the [LICENSE](./LICENSE) file for details.
# image-processor

Lightweight Go image processing utility used as a local runner or Cloud Function to resize images and write different sizes to storage.

This repository contains:

- A Cloud Function entrypoint in `cmd/cloud-function`.
- A local CLI runner in `cmd/local-runner` for testing using a local storage adapter.
- Image processing logic in `internal/image/processor.go`.
- A small local storage adapter in `internal/adapter` (used by the local runner).
- Configuration and tests in `internal/config` and `internal/image`.

The processor resizes uploaded images into small and medium JPEGs and writes an original copy to the configured output locations.

## Requirements

- Go 1.25.1 (from `go.mod`)
- The project uses `github.com/disintegration/imaging` for image resize/encode operations.

## Quick start — run locally

This project includes a small CLI (`cmd/local-runner`) that simulates an upload event and writes processed images to a local directory tree.

1. Build the local runner:

```bash
go build -o bin/local-runner ./cmd/local-runner
```

2. Run it against a file inside `./uploads` (the code expects uploaded files under `./uploads/` by default):

```bash
./bin/local-runner -file ./uploads/gopher.png -output ./output
```

3. Output layout

By default the processor uses these directories (relative to the configured output root):

- `small/` — resized small images (width 400, quality 65)
- `medium/` — resized medium images (width 800, quality 75)
- `original/` — original file copy

So for the example above you'll find processed images under `./output/small/`, `./output/medium/` and `./output/original/`.

## How it works

- The Cloud Function or local runner produces a `config.GCSEvent` (bucket + object name) and calls the `Processor.Process` method.
- `Processor` validates the event (file must be in the configured `FileDir` and have a supported extension), downloads the original bytes from the configured `StorageService`, decodes the image, and concurrently produces target images:
  - Small (400px wide, JPEG quality 65)
  - Medium (800px wide, JPEG quality 75)
  - Original copy (no change)
- Each produced image is uploaded using the provided `StorageService` implementation.

Supported input formats: `.jpg`, `.jpeg`, `.png`, `.webp`.

## Configuration

Configuration is defined by `internal/config.Config` and used by the processor. The local runner sets an example configuration in `cmd/local-runner/main.go`:

- `SmallImgDir` — directory path for small images (e.g. `small/`).
- `MedImgDir` — directory path for medium images (e.g. `medium/`).
- `OrgImgDir` — directory path for original images (e.g. `original/`).
- `FileDir` — the expected directory prefix for incoming uploads (e.g. `./uploads/`).

When running as a Cloud Function, wire your config values from environment variables or your preferred config loader.

## Tests

There is a basic unit test for the image processor in `internal/image/processor_test.go` that uses the `internal/image/testdata/gopher.png` fixture.

Run tests with:

```bash
go test ./...
```

## Development notes & suggestions

- The processor writes JPEG for resized outputs via `github.com/disintegration/imaging`.
- The local storage adapter in `internal/adapter` implements a small, filesystem-based `StorageService` for local testing. When deploying to cloud storage, provide an implementation that talks to your object store (GCS, S3, etc.).
- The processor currently decodes the image once and reuses the decoded image for resize operations; original bytes are preserved and uploaded for the original copy target.

## License

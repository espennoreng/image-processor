# Run tests, and run the local and cloud runners
.PHONY: test local-runner cloud-runner
test:
	go test -v ./...

local-runner:
	go run ./cmd/local-runner --file ./uploads/gopher.png --output ./output
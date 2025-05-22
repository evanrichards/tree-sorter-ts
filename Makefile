.PHONY: build test install clean fmt lint run help

# Default target
all: build

# Build the binary
build:
	go build -o bin/tree-sorter-ts ./cmd/tree-sorter-ts

# Run tests
test:
	go test -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./internal/processor

# Install the binary to GOPATH/bin
install:
	go install ./cmd/tree-sorter-ts

# Clean build artifacts
clean:
	rm -rf bin/
	go clean -cache

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run ./...

# Run the tool with sample files
run: build
	./bin/tree-sorter-ts testdata/fixtures/

# Run with check mode
check: build
	./bin/tree-sorter-ts --check testdata/fixtures/

# Show help
help:
	@echo "Available targets:"
	@echo "  make build    - Build the binary"
	@echo "  make test     - Run tests"
	@echo "  make bench    - Run benchmarks"
	@echo "  make install  - Install to GOPATH/bin"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Run linter (requires golangci-lint)"
	@echo "  make run      - Build and run with sample files"
	@echo "  make check    - Build and run in check mode"
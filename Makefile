.PHONY: build test lint clean

VERSION := 0.10.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DIRTY := $(shell git diff --quiet 2>/dev/null || echo "-dirty")
LDFLAGS := -X main.version=$(VERSION)-$(COMMIT)$(DIRTY)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/akb ./cmd/akb/

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
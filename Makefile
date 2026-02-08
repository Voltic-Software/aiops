.PHONY: build install clean test

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X github.com/voltic-software/aiops/internal/config.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/aiops ./cmd/aiops

install:
	go install $(LDFLAGS) ./cmd/aiops

clean:
	rm -rf bin/ dist/

test:
	go test ./...

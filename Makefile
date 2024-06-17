VERSION = $(shell git rev-parse --abbrev-ref HEAD)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
VERBOSE_FLAG = $(if $(VERBOSE),-v)
EXECUTABLE = aws-google-login

.PHONY: test
test:
	go test $(VERBOSE_FLAG) ./...

.PHONY: lint
lint:
	go list -f '{{.Dir}}/...' -m | xargs golangci-lint run $(VERBOSE_FLAG) --timeout=5m

.PHONY: build
build:
	go build $(VERBOSE_FLAG) -ldflags=$(BUILD_LDFLAGS) -o $(EXECUTABLE) cmd/*

.PHONY: goreleaser
goreleaser:
	goreleaser release --snapshot --clean

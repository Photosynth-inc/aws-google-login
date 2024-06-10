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

DIST_DIR = dist
.PHONY: cross
cross:
	GOOS=linux  GOARCH=arm   go build --tags release -ldflags=$(BUILD_LDFLAGS) -trimpath -o ../$(DIST_DIR)/$(EXECUTABLE).linux.arm cmd/*
	GOOS=linux  GOARCH=amd64 go build --tags release -ldflags=$(BUILD_LDFLAGS) -trimpath -o ../$(DIST_DIR)/$(EXECUTABLE).linux.amd64 cmd/*
	GOOS=linux  GOARCH=arm64 go build --tags release -ldflags=$(BUILD_LDFLAGS) -trimpath -o ../$(DIST_DIR)/$(EXECUTABLE).linux.arm64 cmd/*
	GOOS=darwin GOARCH=amd64 go build --tags release -ldflags=$(BUILD_LDFLAGS) -trimpath -o ../$(DIST_DIR)/$(EXECUTABLE).darwin.amd64 cmd/*
	GOOS=darwin GOARCH=arm64 go build --tags release -ldflags=$(BUILD_LDFLAGS) -trimpath -o ../$(DIST_DIR)/$(EXECUTABLE).darwin.arm64 cmd/*

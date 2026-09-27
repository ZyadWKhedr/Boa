.PHONY: all build test clean run lint cross-compile

BINARY_NAME=compressor
BUILD_DIR=bin
VERSION?=v1.0.0
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE?=$(shell date -u +%Y-%m-%d)
LDFLAGS=-ldflags "-X compressor/cmd.Version=$(VERSION) -X compressor/cmd.Commit=$(COMMIT) -X compressor/cmd.Date=$(DATE) -s -w"

all: test build

build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

test:
	go test -v -race ./...

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) dist *.zip *.out

cross-compile:
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "Cross-compilation completed in dist/"

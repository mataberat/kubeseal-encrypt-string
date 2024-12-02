BINARY_NAME=kubeseal-encrypt-string
VERSION=v1.0.1
BUILD_TIME=$(shell date +%FT%T%z)

# Build directories
BUILD_DIR=build
DARWIN_AMD64=$(BUILD_DIR)/darwin-amd64
DARWIN_ARM64=$(BUILD_DIR)/darwin-arm64
LINUX_AMD64=$(BUILD_DIR)/linux-amd64

.PHONY: test clean build all

test:
	go test -v ./internal/encrypt

clean:
	rm -rf $(BUILD_DIR)

build: clean
	mkdir -p $(DARWIN_AMD64) $(DARWIN_ARM64) $(LINUX_AMD64)
	GOOS=darwin GOARCH=amd64 go build -o $(DARWIN_AMD64)/$(BINARY_NAME) cmd/main.go
	GOOS=darwin GOARCH=arm64 go build -o $(DARWIN_ARM64)/$(BINARY_NAME) cmd/main.go
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_AMD64)/$(BINARY_NAME) cmd/main.go

all: test build

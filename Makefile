.PHONY: all build check-deps clean help lint proto reset setup test update

.DEFAULT_GOAL := all

GOARCH := $(shell uname -m)
GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
OUTPUT := translation-service
BUILD_FLAGS := -a -ldflags="-s -w -extldflags '-static'"
CMD_DIR := ./cmd
PROTOC_GEN_GO := $(shell go tool -n protoc-gen-go)
PROTOC_GEN_GO_GRPC := $(shell go tool -n protoc-gen-go-grpc)
PROTO_SRC_DIR := proto
GO_OUT_DIR := pb

all: help

help:
	@echo "Usage: make [target]"
	@echo "Targets: help, setup, reset, lint, update, test, build, clean"

check-deps:
	@echo "Checking dependencies..."
	@command -v go >/dev/null 2>&1 || { echo >&2 "Go is not installed."; exit 1; }
	@command -v pre-commit >/dev/null 2>&1 || { echo >&2 "pre-commit is not installed."; exit 1; }
	@command -v goose >/dev/null 2>&1 || { echo >&2 "goose is not installed."; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo >&2 "Docker is not installed."; exit 1; }

proto:
	@echo "Generating protobuf model..."
	@command rm -r $(GO_OUT_DIR)
	@command mkdir -p $(GO_OUT_DIR)
	@command protoc -I $(PROTO_SRC_DIR) \
	  --plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
	  --plugin=protoc-gen-go-grpc=$(PROTOC_GEN_GO_GRPC) \
	  --go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
	  --go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
	  $(shell find $(PROTO_SRC_DIR) -name '*.proto')

setup: check-deps
	@echo "Setting up commit hooks and local database..."
	./scripts/setup.sh

reset:
	@echo "Uninstall pre-commit hooks..."
	pre-commit uninstall
	@echo "Cleanup pre-commit cache..."
	pre-commit clean
	@echo "Cleanup local docker database..."
	docker compose down --volumes --remove-orphans
	@echo "Delete generated protobuf sources ..."
	rm -r pb

lint:
	@echo "Running linter..."
	pre-commit run --all-files

update:
	@echo "Updating Linter and Go modules..."
	pre-commit autoupdate
	pre-commit migrate-config
	go get -u ./...
	go mod tidy

test: lint
	@echo "Running tests..."
	go test -v ./...

build: proto
	@echo "Building the service..."
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build $(BUILD_FLAGS) -o $(OUTPUT) $(CMD_DIR)/

clean:
	@echo "Cleaning build artifacts..."
	go clean
	@rm -f $(OUTPUT)

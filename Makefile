.PHONY: all build check-deps clean help lint proto reset setup test update

.DEFAULT_GOAL := all

GOARCH := $(shell uname -m)
GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
OUTPUT := translation-service
BUILD_FLAGS := -a -ldflags="-s -w -extldflags '-static'"
CMD_DIR := ./cmd

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
	@command buf generate

openapi:
	@echo "Generate openapi spec ..."
	@command go tool oapi-codegen --config=cfg.yaml api.yaml

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
	rm -r gen

lint: proto openapi
	@echo "Running linter..."
	pre-commit run --all-files

update:
	@echo "Updating Linter and Go modules..."
	pre-commit autoupdate
	pre-commit migrate-config
	go get -u ./...
	go mod tidy

test: proto openapi
	@echo "Running tests..."
	go test -v ./...

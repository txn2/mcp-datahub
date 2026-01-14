.PHONY: all build test lint clean coverage security help tidy verify fmt lint-fix test-integration

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=mcp-datahub
COVERAGE_FILE=coverage.out

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

all: lint test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/mcp-datahub

test:
	$(GOTEST) -v -race ./...

test-integration:
	$(GOTEST) -v -tags=integration ./pkg/client/...

coverage:
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

coverage-html: coverage
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html

lint:
	golangci-lint run --timeout=5m

lint-fix:
	golangci-lint run --fix --timeout=5m

fmt:
	$(GOCMD) fmt ./...
	goimports -w -local github.com/txn2/mcp-datahub .

security:
	gosec ./...
	govulncheck ./...

tidy:
	$(GOCMD) mod tidy
	$(GOCMD) mod verify

clean:
	rm -f $(BINARY_NAME) $(COVERAGE_FILE) coverage.html
	$(GOCMD) clean -cache -testcache

verify: tidy lint test

help:
	@echo "Available targets:"
	@echo "  all            - Run lint, test, and build (default)"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests with race detection"
	@echo "  test-integration - Run integration tests"
	@echo "  coverage       - Generate coverage report"
	@echo "  coverage-html  - Generate HTML coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  lint-fix       - Run golangci-lint with auto-fix"
	@echo "  fmt            - Format code"
	@echo "  security       - Run security scans"
	@echo "  tidy           - Tidy and verify modules"
	@echo "  clean          - Remove build artifacts"
	@echo "  verify         - Run tidy, lint, and test"
	@echo "  help           - Show this help"

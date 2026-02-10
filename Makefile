.PHONY: all build test lint clean coverage security help tidy verify fmt lint-fix test-integration \
       patch-coverage mutation deadcode bench profile build-check

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=mcp-datahub
COVERAGE_FILE=coverage.out
COVERAGE_THRESHOLD := 80
MUTATION_THRESHOLD := 60

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

all: lint test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/mcp-datahub

test:
	$(GOTEST) -v -race -shuffle=on -count=1 ./...

test-integration:
	$(GOTEST) -v -tags=integration ./pkg/client/...

coverage:
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@COVERAGE=$$($(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$NF}' | sed 's/%//'); \
	echo "Coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "FAIL: Coverage $${COVERAGE}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi

coverage-html: coverage
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html

## Patch coverage (only changed lines vs main branch)
PATCH_THRESHOLD := 80
patch-coverage:
	@MERGE_BASE=$$(git merge-base main HEAD 2>/dev/null || echo "HEAD"); \
	if [ "$$MERGE_BASE" = "$$(git rev-parse HEAD)" ]; then \
		echo "On main branch, skipping patch coverage"; exit 0; \
	fi; \
	CHANGED_FILES=$$(git diff --name-only "$$MERGE_BASE"...HEAD -- '*.go' | grep -v '_test.go' || true); \
	if [ -z "$$CHANGED_FILES" ]; then \
		echo "No non-test Go files changed, skipping patch coverage"; exit 0; \
	fi; \
	echo "Changed files: $$CHANGED_FILES"; \
	TOTAL=0; COVERED=0; \
	for FILE in $$CHANGED_FILES; do \
		if [ ! -f "$$FILE" ]; then continue; fi; \
		PKG=$$(dirname "$$FILE"); \
		$(GOTEST) -coverprofile=patch_cov.tmp -covermode=atomic "./$$PKG" > /dev/null 2>&1 || true; \
		if [ -f patch_cov.tmp ]; then \
			FILE_COV=$$($(GOCMD) tool cover -func=patch_cov.tmp 2>/dev/null | grep "$$FILE" || true); \
			rm -f patch_cov.tmp; \
		fi; \
	done; \
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./... > /dev/null 2>&1; \
	for FILE in $$CHANGED_FILES; do \
		LINES=$$(git diff --unified=0 "$$MERGE_BASE"...HEAD -- "$$FILE" | \
			grep '^@@' | sed 's/.*+\([0-9]*\).*/\1/' || true); \
		for LINE in $$LINES; do \
			TOTAL=$$((TOTAL + 1)); \
			if grep -q "$$FILE:$$LINE" $(COVERAGE_FILE) 2>/dev/null; then \
				COVERED=$$((COVERED + 1)); \
			fi; \
		done; \
	done; \
	if [ "$$TOTAL" -eq 0 ]; then \
		echo "No executable changed lines detected"; exit 0; \
	fi; \
	PCT=$$((COVERED * 100 / TOTAL)); \
	echo "Patch coverage: $$COVERED/$$TOTAL lines = $$PCT%"; \
	if [ "$$PCT" -lt "$(PATCH_THRESHOLD)" ]; then \
		echo "FAIL: Patch coverage $$PCT% is below threshold $(PATCH_THRESHOLD)%"; \
		exit 1; \
	fi

lint:
	golangci-lint run --timeout=5m
	$(GOCMD) vet ./...

lint-fix:
	golangci-lint run --fix --timeout=5m

fmt:
	$(GOCMD) fmt ./...
	goimports -w -local github.com/txn2/mcp-datahub .

security:
	gosec ./...
	govulncheck ./...

## Mutation testing (requires gremlins: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest)
mutation:
	gremlins unleash --workers 1 --timeout-coefficient 3 --threshold-efficacy $(MUTATION_THRESHOLD)

## Dead code detection
deadcode:
	deadcode ./...

## Benchmarking
bench:
	$(GOTEST) -bench=. -benchmem -count=3 -run='^$$' ./... | tee bench.txt

## Profiling (CPU and memory)
profile:
	$(GOTEST) -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof -run='^$$' ./...
	@echo "CPU profile: go tool pprof cpu.prof"
	@echo "Memory profile: go tool pprof mem.prof"

## Build validation
build-check:
	$(GOCMD) build ./...
	$(GOCMD) mod verify

tidy:
	$(GOCMD) mod tidy
	$(GOCMD) mod verify

clean:
	rm -f $(BINARY_NAME) $(COVERAGE_FILE) coverage.html bench.txt cpu.prof mem.prof
	$(GOCMD) clean -cache -testcache

verify: tidy lint test coverage security deadcode build-check
	@echo "All verification checks passed."

help:
	@echo "Available targets:"
	@echo "  all              - Run lint, test, and build (default)"
	@echo "  build            - Build the binary"
	@echo "  test             - Run tests with race detection"
	@echo "  test-integration - Run integration tests"
	@echo "  coverage         - Generate coverage report (threshold: $(COVERAGE_THRESHOLD)%)"
	@echo "  coverage-html    - Generate HTML coverage report"
	@echo "  patch-coverage   - Coverage of changed lines only (threshold: $(PATCH_THRESHOLD)%)"
	@echo "  lint             - Run golangci-lint + go vet"
	@echo "  lint-fix         - Run golangci-lint with auto-fix"
	@echo "  fmt              - Format code"
	@echo "  security         - Run gosec + govulncheck"
	@echo "  mutation         - Run mutation testing with gremlins"
	@echo "  deadcode         - Detect unreachable functions"
	@echo "  bench            - Run benchmarks with memory reporting"
	@echo "  profile          - Generate CPU and memory profiles"
	@echo "  build-check      - Verify build and modules"
	@echo "  tidy             - Tidy and verify modules"
	@echo "  clean            - Remove build artifacts"
	@echo "  verify           - Run full verification suite"
	@echo "  help             - Show this help"

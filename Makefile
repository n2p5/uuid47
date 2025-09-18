.PHONY: test
test:
	go test -v ./...

.PHONY: test-short
test-short:
	go test -short ./...

.PHONY: test-race
test-race:
	go test -race -v ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: bench
bench:
	go test -bench=. -benchmem ./...

.PHONY: bench-compare
bench-compare:
	go test -bench=. -benchmem -count=5 ./... | tee bench.txt

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	go build ./...

.PHONY: example
example:
	go run example/main.go

.PHONY: clean
clean:
	go clean
	rm -f coverage.out coverage.html bench.txt
	rm -f uuid47.test
	rm -f c_validation/test_vectors_gen

.PHONY: clean-all
clean-all: clean
	rm -f c_validation/uuidv47.h

# C header URL and commit
C_HEADER_URL := https://raw.githubusercontent.com/stateless-me/uuidv47/main/uuidv47.h
C_HEADER_PATH := c_validation/uuidv47.h

.PHONY: download-c-header
download-c-header:
	@if [ ! -f $(C_HEADER_PATH) ]; then \
		echo "Downloading uuidv47.h from upstream..."; \
		curl -s $(C_HEADER_URL) -o $(C_HEADER_PATH); \
		echo "Downloaded $(C_HEADER_PATH)"; \
	else \
		echo "$(C_HEADER_PATH) already exists"; \
	fi

.PHONY: verify-c-compat
verify-c-compat: download-c-header
	@echo "Compiling C test vector generator..."
	@gcc -o c_validation/test_vectors_gen c_validation/test_vectors_gen.c
	@echo "Running C test vectors..."
	@./c_validation/test_vectors_gen
	@echo ""
	@echo "Running Go tests for C compatibility..."
	@go test -v -run TestExactCCompatibility

.PHONY: all
all: fmt vet test

.PHONY: ci
ci: fmt vet test-race bench

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  test           - Run tests with verbose output"
	@echo "  test-short     - Run short tests"
	@echo "  test-race      - Run tests with race detector"
	@echo "  test-coverage  - Generate coverage report"
	@echo "  bench          - Run benchmarks"
	@echo "  bench-compare  - Run benchmarks multiple times for comparison"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  build          - Build the package"
	@echo "  example        - Run the example program"
	@echo "  clean          - Clean build artifacts"
	@echo "  clean-all      - Clean everything including downloaded C header"
	@echo "  download-c-header - Download uuidv47.h from upstream"
	@echo "  verify-c-compat - Verify compatibility with C implementation"
	@echo "  all            - Run fmt, vet, and test"
	@echo "  ci             - Run CI checks (fmt, vet, test-race, bench)"
	@echo "  help           - Show this help message"
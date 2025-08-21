# The @ before a command in a Makefile suppresses the command itself from being echoed to the terminal.
# Target to run when no target is specified

.DEFAULT_GOAL := test

.PHONY: fmt vet test test-local clean

TEST_FLAGS ?=

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test $(TEST_FLAGS) ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out 
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tr -d [[:space:]] | grep -q "total:(statements)100.0%" || (echo "Coverage is less than 100%. Test failed!" && exit 1)

clean: 
	go clean

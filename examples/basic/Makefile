# Target to run when no target is specified
.DEFAULT_GOAL := run

.PHONY: fmt vet run clean

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

run: 
	go run .

clean: 
	go clean

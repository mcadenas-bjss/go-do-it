.PHONY: default all help fmt vet lint
default: all
all: fmt vet lint

help:
	@echo "Available phony targets:"
	@echo "help		: prints out this targets information"
	@echo "default		: runs 'all' target"
	@echo "build	: builds binary to goDoIt"
	@echo "all			: runs the phony targets - fmt vet lint test bench benchstat"
	@echo "fmt			: Formats Go code using go fmt"
	@echo "vet			: Validate Go code using go vet"

fmt: *.go
	go fmt

vet: *.go
	go vet

lint: *.go
	golangci-lint run

build:
	@echo "Building executable..."
	go build -o goDoIt
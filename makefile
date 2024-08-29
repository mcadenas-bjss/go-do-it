.PHONY: default all help fmt vet lint test bench benchstat fuzz tunnel
default: all
all: fmt vet lint test benchstat 

help:
	@echo "Available phony targets:"
	@echo "help			: prints out this targets information"
	@echo "default		: runs 'all' target"
	@echo "all			: runs the phony targets - fmt vet lint test bench benchstat"
	@echo "fmt			: Formats Go code using go fmt"
	@echo "vet			: Validate Go code using go vet"
	@echo "lint			: Validates Go code using golangci-lint"
	@echo "test			: Tests the solution"
	@echo "bench		: Benchmark tests the solution"
	@echo "benchstat	: A/B comparions of benchmark results"
	@echo "Fuzz			: Fuzzing tests the solution"
	@echo "run-api		: Runs API server"

fmt: *.go
	go fmt

vet: *.go
	go vet

lint: *.go
	golangci-lint run

test: *_test.go
	go test -coverprofile coverage.out ./...
	go tool cover -html coverage.out -o coverage.html

bench: *_test.go
	go test -bench=. -benchmem -count=10 > benchstat.txt

benchstat.old.txt: benchstat.txt
	cp -f benchstat.txt benchstat.old.txt

benchstat: bench benchstat.old.txt
	benchstat benchstat.old.txt benchstat.txt

fuzz: *_test.go
	go test -fuzz FuzzProcessExpression

# App commands
run-api:
	go run ./api/main.go

tunnel:
	ngrok http --domain=cricket-rational-pika.ngrok-free.app 4321
.PHONY: fmt vet lint sec test build all

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

sec:
	gosec ./...

test:
	go test ./...

build:
	go build ./...

all: fmt vet lint sec test build

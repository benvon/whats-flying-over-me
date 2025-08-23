.PHONY: fmt vet lint sec test build clean all

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

sec:
	@if ! command -v gosec &> /dev/null; then \
		echo "Installing gosec..."; \
		curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.19.0; \
		export PATH=$$PATH:$$(go env GOPATH)/bin; \
	fi
	gosec -no-fail ./...

test:
	go test ./...

build:
	go build -o whats-flying-over-me ./cmd/whats-flying-over-me

clean:
	rm -f whats-flying-over-me

check: fmt vet lint sec test build

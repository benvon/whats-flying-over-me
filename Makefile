.PHONY: fmt vet lint sec test build clean all

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

sec:
	@if ! command -v govulncheck &> /dev/null; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	govulncheck ./...

test:
	go test ./...

build:
	go build -o whats-flying-over-me ./cmd/whats-flying-over-me

clean:
	rm -f whats-flying-over-me

check: fmt vet lint sec test build

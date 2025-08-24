.PHONY: fmt vet lint sec test build clean all goreleaser-check goreleaser-test goreleaser-test-docker release-test help

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
	rm -rf dist/

# GoReleaser commands
goreleaser-check:
	@echo "📋 Checking GoReleaser configuration..."
	goreleaser check

goreleaser-test:
	@echo "🧪 Testing GoReleaser locally (no Docker)..."
	@echo "Creating temporary local config..."
	@echo "version: 2" > .goreleaser.local.yml
	@echo "project_name: whats-flying-over-me" >> .goreleaser.local.yml
	@echo "" >> .goreleaser.local.yml
	@echo "builds:" >> .goreleaser.local.yml
	@echo "  - id: wfo" >> .goreleaser.local.yml
	@echo "    main: ./cmd/whats-flying-over-me" >> .goreleaser.local.yml
	@echo "    binary: whats-flying-over-me" >> .goreleaser.local.yml
	@echo "    env:" >> .goreleaser.local.yml
	@echo "      - CGO_ENABLED=0" >> .goreleaser.local.yml
	@echo "    goos:" >> .goreleaser.local.yml
	@echo "      - linux" >> .goreleaser.local.yml
	@echo "      - windows" >> .goreleaser.local.yml
	@echo "      - darwin" >> .goreleaser.local.yml
	@echo "    goarch:" >> .goreleaser.local.yml
	@echo "      - amd64" >> .goreleaser.local.yml
	@echo "      - arm64" >> .goreleaser.local.yml
	@echo "" >> .goreleaser.local.yml
	@echo "archives:" >> .goreleaser.local.yml
	@echo "  - name_template: \"{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}\"" >> .goreleaser.local.yml
	@echo "" >> .goreleaser.local.yml
	@echo "checksum:" >> .goreleaser.local.yml
	@echo "  name_template: \"{{ .ProjectName }}_{{ .Version }}_checksums.txt\"" >> .goreleaser.local.yml
	@echo "" >> .goreleaser.local.yml
	@echo "# No Docker builds for local testing" >> .goreleaser.local.yml
	goreleaser release --snapshot --clean --config .goreleaser.local.yml
	@echo "✅ Local test completed successfully!"
	@echo "📁 Generated files in 'dist/' directory:"
	@ls -la dist/
	@echo "Cleaning up temporary config..."
	@rm .goreleaser.local.yml

goreleaser-test-docker:
	@echo "🐳 Testing GoReleaser with Docker builds..."
	@if [ -z "$$GITHUB_REPOSITORY" ]; then \
		echo "Setting GITHUB_REPOSITORY for local testing..."; \
		GITHUB_REPOSITORY=benvon/whats-flying-over-me goreleaser release --snapshot --clean; \
	else \
		goreleaser release --snapshot --clean; \
	fi

# Add GoReleaser testing to the main check target
check: fmt vet lint sec test build goreleaser-check

# Convenience target for full release testing
release-test: goreleaser-check goreleaser-test
	@echo ""
	@echo "🎉 Release testing completed successfully!"
	@echo "🚀 Ready to push to GitHub for CI testing!"

# Help target
help:
	@echo "Available targets:"
	@echo "  fmt                    - Format Go code"
	@echo "  vet                    - Vet Go code"
	@echo "  lint                   - Run linter"
	@echo "  sec                    - Run security scanner"
	@echo "  test                   - Run tests"
	@echo "  build                  - Build binary"
	@echo "  clean                  - Clean build artifacts"
	@echo "  check                  - Run all checks (fmt, vet, lint, sec, test, build, goreleaser-check)"
	@echo ""
	@echo "GoReleaser targets:"
	@echo "  goreleaser-check       - Validate GoReleaser configuration"
	@echo "  goreleaser-test        - Test GoReleaser locally (no Docker)"
	@echo "  goreleaser-test-docker - Test GoReleaser with Docker builds"
	@echo "  release-test           - Full release testing (check + local test)"
	@echo ""
	@echo "Examples:"
	@echo "  make check             - Run all standard checks"
	@echo "  make release-test      - Test release process locally"
	@echo "  make goreleaser-test-docker - Test with Docker (requires Docker daemon)"

GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOVULNCHECK ?= govulncheck

.PHONY: fmt build test test-race vet lint vuln verify

fmt:
	$(GO) fmt ./...

build:
	$(GO) build ./cmd/kaiten-mcp

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

lint:
	$(GOLANGCI_LINT) run ./...

vuln:
	$(GOVULNCHECK) -C internal/app -scan=module

verify: fmt test test-race vet lint vuln build

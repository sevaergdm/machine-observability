# Config
GO ?= go
GOPATH = $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin

# Tools
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint

# Commands
.PHONY: build
build:
	$(GO) build -o bin/agent ./cmd/agent

.PHONY: test
test:
	$(GO) test ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: fmt-check
fmt-check:
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "unformatted files:"; echo "$$out"; exit 1; \
	fi

.PHONY: lint
lint:
	@$(GOLANGCI_LINT) run --fix

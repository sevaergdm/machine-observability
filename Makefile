# Config
GO ?= go
GOPATH = $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin

# Tools
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint

# Commands
build:
	go build -o bin/agent ./cmd/agent

test:
	go test ./...

vet:
	go vet ./...

fmt-check:
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "unformatted files:"; echo "$$out"; exit 1; \
	fi

lint:
	@$(GOLANGCI_LINT) run --fix

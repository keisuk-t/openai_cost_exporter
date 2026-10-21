GO ?= go
VERSION ?= dev
BINARY := bin/openai_cost_exporter

.PHONY: setup build run fmt lint test check docker

setup:
	$(GO) mod download

build:
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/openai_cost_exporter

run:
	$(GO) run ./cmd/openai_cost_exporter

fmt:
	gofmt -w cmd internal

lint:
	@test -z "$$(gofmt -l cmd internal)" || (gofmt -l cmd internal; exit 1)
	$(GO) vet ./...

test:
	$(GO) test -race -count=1 -timeout=60s ./...

check: lint test build

docker:
	docker build --build-arg VERSION=$(VERSION) -t openai-cost-exporter:$(VERSION) .

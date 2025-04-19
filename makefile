.PHONY: all

VERSION := $(shell cz version -p)
COMMIT := $(shell git rev-parse --short HEAD)
TAG ?= latest

-include .env
export

# Run unit tests with coverage
test:
	@echo "Running unit tests with coverage for free@home monitor v$(VERSION)-$(COMMIT)"
	@go test -timeout 5s -coverprofile coverage.out -v ./...
	@go tool cover -html=coverage.out -o coverage.html

# Run the free@home monitor
monitor-run:
	@echo "Starting free@home monitor v$(VERSION)-$(COMMIT)"
	@go run -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" ./cmd/monitor/main.go

# Build the Docker image for the free@home monitor
monitor-build:
	@echo "Building Docker image for free@home monitor v$(VERSION) from $(COMMIT) with tag $(TAG)."
	@docker build --build-arg version=$(VERSION) --build-arg commit=$(COMMIT) -t ghcr.io/pgerke/freeathome-monitor:${TAG} -f ./cmd/monitor/Dockerfile .

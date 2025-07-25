.PHONY: all

VERSION := $(shell cz version -p)
COMMIT := $(shell git rev-parse --short HEAD)
TAG ?= latest

# Run unit tests with coverage
unittest:
	@echo "Running unit tests with coverage for v$(VERSION)-$(COMMIT)"
	@go test -timeout 5s -covermode atomic -coverprofile unittest.coverage.out -v ./...
	@go tool cover -html=unittest.coverage.out -o unittest.coverage.html

# Run the free@home CLI locally
cli-run-local:
	@echo "Starting free@home CLI v$(VERSION)-$(COMMIT)"
	@go run -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" ./cmd/cli/main.go

# Build the free@home CLI
cli-build:
	@echo "Building free@home CLI v$(VERSION)-$(COMMIT)"
	@go build -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" -o fh ./cmd/cli/main.go
	@chmod +x fh

# Run the free@home monitor locally
monitor-run-local:
	@echo "Starting free@home monitor v$(VERSION)-$(COMMIT)"
	@set -a; [ -f .env ] && . .env; set +a; go run ./cmd/monitor
	@go run -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" ./cmd/monitor/main.go

# Build the Docker image for the free@home monitor
monitor-build-docker:
	@echo "Building Docker image for free@home monitor v$(VERSION) from $(COMMIT) with tag $(TAG)."
	@docker build --build-arg version=$(VERSION) --build-arg commit=$(COMMIT) -t ghcr.io/pgerke/freeathome-monitor:${TAG} -f ./cmd/monitor/Dockerfile .

monitor-build-docker-multiarch:
	@echo "Building multi-arch Docker image for free@home monitor v$(VERSION) from $(COMMIT) with tag $(TAG)."
	@docker buildx build --platform linux/amd64,linux/arm64 --build-arg version=$(VERSION) --build-arg commit=$(COMMIT) -t ghcr.io/pgerke/freeathome-monitor:${TAG} -f ./cmd/monitor/Dockerfile --push .

# Run the free@home monitor integration tests
monitor-integration-test:
	@echo "Running integration tests for free@home monitor v$(VERSION)-$(COMMIT)"
	@rm -rf ./coverage-monitor && mkdir -p ./coverage-monitor
	@go test -c -o monitor-integration.test -covermode atomic ./cmd/monitor
	@go test -timeout 5s -tags integration integration/monitor_test.go
	@go tool covdata textfmt -i coverage-monitor -o monitor-integration.coverage.out
	@go tool cover -html=monitor-integration.coverage.out -o monitor-integration.coverage.html

# Run all unit and integration tests and aggregate the coverage reports
test-ci: unittest monitor-integration-test
	@$(shell go env GOPATH)/bin/gocovmerge unittest.coverage.out monitor-integration.coverage.out > coverage.out
	@go tool cover -html=coverage.out -o coverage.html

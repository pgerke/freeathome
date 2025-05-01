.PHONY: all

VERSION := $(shell cz version -p)
COMMIT := $(shell git rev-parse --short HEAD)
TAG ?= latest

# Run unit tests with coverage
unittest:
	@echo "Running unit tests with coverage for v$(VERSION)-$(COMMIT)"
	@go test -timeout 5s -covermode atomic -coverprofile unittest.coverage.out -v ./...
	@go tool cover -html=unittest.coverage.out -o unittest.coverage.html

# Run the free@home monitor locally
monitor-run-local:
	@echo "Starting free@home monitor v$(VERSION)-$(COMMIT)"
	@set -a; [ -f .env ] && . .env; set +a; go run ./cmd/monitor
	@go run -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" ./cmd/monitor/main.go

# Build the Docker image for the free@home monitor
monitor-build-docker:
	@echo "Building Docker image for free@home monitor v$(VERSION) from $(COMMIT) with tag $(TAG)."
	@set -a; [ -f .env ] && . .env; set +a; go run ./cmd/monitor
	@docker build --build-arg version=$(VERSION) --build-arg commit=$(COMMIT) -t ghcr.io/pgerke/freeathome-monitor:${TAG} -f ./cmd/monitor/Dockerfile .

# Run the free@home monitor integration tests
monitor-integration-test:
	@echo "Running integration tests for free@home monitor v$(VERSION)-$(COMMIT)"
	@rm -rf ./coverage-monitor && mkdir -p ./coverage-monitor
	@go test -c -o monitor-integration.test -covermode atomic ./cmd/monitor
	@GOCOVERDIR=./coverage-monitor RUN_MAIN=1 ./monitor-integration.test -test.run=TestMonitor_Main || true
	@go tool covdata textfmt -i coverage-monitor -o monitor-integration.coverage.out
	@go tool cover -html=monitor-integration.coverage.out -o monitor-integration.coverage.html

# Run all unit and integration tests and aggregate the coverage reports
test-ci: unittest monitor-integration-test
	@$(shell go env GOPATH)/bin/gocovmerge unittest.coverage.out monitor-integration.coverage.out > coverage.out
	@go tool cover -html=coverage.out -o coverage.html

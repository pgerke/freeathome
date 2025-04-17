.PHONY: all

VERSION := $(shell cz version -p)
COMMIT := $(shell git rev-parse --short HEAD)
TAG ?= latest

-include .env
export

# Run unit tests with coverage
test:
	@echo "Running unit tests with coverage for free@home monitor v$(VERSION)-$(COMMIT)"
	@go test -coverprofile coverage.out -v ./...
	@go tool cover -html=coverage.out -o coverage.html

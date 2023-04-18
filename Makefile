.DEFAULT_GOAL := build
VERSION ?= $(shell git describe --tags --always --match=v* 2> /dev/null || echo v0)
COMMIT = $(shell git rev-parse HEAD)
BUILD_TS = $(shell date +%FT%T%:z)
MODULEPATH := $(shell go mod edit -json 2> /dev/null | jq -r '.Module.Path')

BIN = $(CURDIR)/bin
$(BIN):
	@mkdir -p $@

PLATFORM=local

.PHONY: bin/k8s-inventory-client
bin/k8s-inventory-client:
	@DOCKER_BUILDKIT=1 docker build --target bin \
		--output bin/ \
		--platform ${PLATFORM} \
		--tag neticdk-k8s/k8s-inventory-client \
		.
	@DOCKER_BUILDKIT=1 docker build --platform ${PLATFORM} \
		--tag neticdk-k8s/k8s-inventory-client \
		.

# Runs go lint
.PHONY: lint
lint:
	@echo "Linting..."
	@golangci-lint run

# Runs go clean
.PHONY: clean
clean:
	@echo "Cleaning..."
	@go clean

# Runs go fmt
.PHONY: fmt
fmt:
	@echo "Formatting..."
	@go fmt ./...

# Runs go build
.PHONY: build
build: clean fmt lint | $(BIN)
	@echo "Building k8s-inventory-client..."
	CGO_ENABLED=0 go build -o $(BIN)/k8s-inventory-client \
		-v \
		-a \
		-tags release \
		-ldflags '-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TS)'

# Runs go build
.PHONY: build2
build2: clean fmt | $(BIN)
	@echo "Building k8s-inventory-client..."
	CGO_ENABLED=0 go build -o $(BIN)/k8s-inventory-client \
		-v \
		-tags release \
		-ldflags '-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TS)'

# Build docker client
.PHONY: docker-build
docker-build-client:
	@echo "Building k8s-inventory-client image..."
	DOCKER_BUILDKIT=1 docker build --progress=plain --no-cache -t neticdk-k8s/k8s-inventory-client -f dist/Dockerfile.client .

# Tag and push docker client
.PHONY: docker-push
docker-push:
	docker tag neticdk-k8s/k8s-inventory-client:latest ghcr.io/neticdk-k8s/k8s-inventory-client:latest
	docker push ghcr.io/neticdk-k8s/k8s-inventory-client:latest
	docker tag neticdk-k8s/k8s-inventory-client:latest ghcr.io/neticdk-k8s/k8s-inventory-client:${VERSION}
	docker push ghcr.io/neticdk-k8s/k8s-inventory-client:${VERSION}

# Build, tag and push docker images
.PHONY: docker-all
docker-all: docker-build docker-push

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
		--tag netic/k8s-inventory-collector \
		.
	@DOCKER_BUILDKIT=1 docker build --platform ${PLATFORM} \
		--tag netic/k8s-inventory-collector \
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
	DOCKER_BUILDKIT=1 docker build -t netic/k8s-inventory-client -f dist/Dockerfile.client .

# Tag and push docker client
.PHONY: docker-push
docker-push:
	docker tag netic/k8s-inventory-client:latest registry.netic.dk/k8s-inventory-collector/client:latest
	docker push registry.netic.dk/k8s-inventory-collector/client:latest
	docker tag netic/k8s-inventory-client:latest registry.netic.dk/k8s-inventory-collector/client:${VERSION}
	docker push registry.netic.dk/k8s-inventory-collector/client:${VERSION}

# Build, tag and push docker images
.PHONY: docker-all
docker-all: docker-build docker-push

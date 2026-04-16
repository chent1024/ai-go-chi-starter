include Makefile.rules

BIN_DIR ?= $(CURDIR)/bin
DIST_DIR ?= $(CURDIR)/dist
DEV_COMPOSE_FILE ?= deploy/docker-compose.dev.yaml
DEV_ENV_FILE ?= deploy/.env.dev.example
DOCKER_COMPOSE ?= docker compose
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
TARGET_OS ?= $(shell go env GOOS)
TARGET_ARCH ?= $(shell go env GOARCH)
RELEASE_DIR ?= $(DIST_DIR)/$(VERSION)/$(TARGET_OS)-$(TARGET_ARCH)
GO_LDFLAGS ?= -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

.PHONY: build build-api build-worker build-migrate
.PHONY: release-build release-build-api release-build-worker release-build-migrate
.PHONY: run-api run-worker migrate migrate-up migrate-version
.PHONY: dev-up dev-down dev-logs dev-ps test-integration smoke

build: build-api build-worker build-migrate

build-api:
	mkdir -p $(BIN_DIR)
	go build -ldflags "$(GO_LDFLAGS)" -o $(BIN_DIR)/api ./cmd/api

build-worker:
	mkdir -p $(BIN_DIR)
	go build -ldflags "$(GO_LDFLAGS)" -o $(BIN_DIR)/worker ./cmd/worker

build-migrate:
	mkdir -p $(BIN_DIR)
	go build -ldflags "$(GO_LDFLAGS)" -o $(BIN_DIR)/migrate ./cmd/migrate

release-build: release-build-api release-build-worker release-build-migrate

release-build-api:
	mkdir -p $(RELEASE_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -trimpath -ldflags "$(GO_LDFLAGS)" -o $(RELEASE_DIR)/api ./cmd/api

release-build-worker:
	mkdir -p $(RELEASE_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -trimpath -ldflags "$(GO_LDFLAGS)" -o $(RELEASE_DIR)/worker ./cmd/worker

release-build-migrate:
	mkdir -p $(RELEASE_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -trimpath -ldflags "$(GO_LDFLAGS)" -o $(RELEASE_DIR)/migrate ./cmd/migrate

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

migrate:
	go run ./cmd/migrate -action up

migrate-up:
	go run ./cmd/migrate -action up

migrate-version:
	go run ./cmd/migrate -action version

dev-up:
	$(DOCKER_COMPOSE) -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) up -d

dev-down:
	$(DOCKER_COMPOSE) -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) down

dev-logs:
	$(DOCKER_COMPOSE) -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) logs -f

dev-ps:
	$(DOCKER_COMPOSE) -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) ps

test-integration:
	bash scripts/test-integration.sh

smoke:
	bash scripts/smoke.sh

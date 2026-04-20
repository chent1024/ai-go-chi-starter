include Makefile.rules

APP_DIR ?= app
APP_BIN_DIR ?= $(CURDIR)/bin
APP_DIST_DIR ?= $(CURDIR)/$(APP_DIR)/dist
APP_DEV_COMPOSE_FILE ?= $(CURDIR)/deploy/docker-compose.dev.yaml
APP_DEV_ENV_FILE ?= $(CURDIR)/deploy/.env.dev.example
DOCKER_COMPOSE ?= docker compose
SQLC ?= $(TOOLS_BIN)/sqlc
SQLC_VERSION ?= 1.30.0
SQLC_OS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
SQLC_ARCH ?= $(shell uname -m | sed 's/^x86_64$$/amd64/; s/^amd64$$/amd64/; s/^aarch64$$/arm64/; s/^arm64$$/arm64/')
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
TARGET_OS ?= $(shell cd $(APP_DIR) && go env GOOS)
TARGET_ARCH ?= $(shell cd $(APP_DIR) && go env GOARCH)
APP_RELEASE_DIR ?= $(APP_DIST_DIR)/$(VERSION)/$(TARGET_OS)-$(TARGET_ARCH)
GO_LDFLAGS ?= -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

.PHONY: build build-api build-worker build-migrate
.PHONY: release-build release-build-api release-build-worker release-build-migrate
.PHONY: run-api run-worker migrate migrate-up migrate-version
.PHONY: dev-up dev-down dev-logs dev-ps test-integration smoke sqlc-generate sqlc-verify

$(SQLC):
	mkdir -p $(TOOLS_BIN)
	test "$(SQLC_OS)" = "darwin" -o "$(SQLC_OS)" = "linux"
	test "$(SQLC_ARCH)" = "amd64" -o "$(SQLC_ARCH)" = "arm64"
	curl -fsSL -o $(TOOLS_BIN)/sqlc.tar.gz https://github.com/sqlc-dev/sqlc/releases/download/v$(SQLC_VERSION)/sqlc_$(SQLC_VERSION)_$(SQLC_OS)_$(SQLC_ARCH).tar.gz
	tar -xzf $(TOOLS_BIN)/sqlc.tar.gz -C $(TOOLS_BIN) sqlc
	rm -f $(TOOLS_BIN)/sqlc.tar.gz
	chmod +x $(SQLC)

build:
	$(MAKE) build-api
	$(MAKE) build-worker
	$(MAKE) build-migrate

build-api:
	mkdir -p $(APP_BIN_DIR)
	cd $(APP_DIR) && go build -ldflags "$(GO_LDFLAGS)" -o $(APP_BIN_DIR)/api ./cmd/api

build-worker:
	mkdir -p $(APP_BIN_DIR)
	cd $(APP_DIR) && go build -ldflags "$(GO_LDFLAGS)" -o $(APP_BIN_DIR)/worker ./cmd/worker

build-migrate:
	mkdir -p $(APP_BIN_DIR)
	cd $(APP_DIR) && go build -ldflags "$(GO_LDFLAGS)" -o $(APP_BIN_DIR)/migrate ./cmd/migrate

release-build:
	$(MAKE) release-build-api
	$(MAKE) release-build-worker
	$(MAKE) release-build-migrate

release-build-api:
	mkdir -p $(APP_RELEASE_DIR)
	cd $(APP_DIR) && GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -trimpath -ldflags "$(GO_LDFLAGS)" -o $(APP_RELEASE_DIR)/api ./cmd/api

release-build-worker:
	mkdir -p $(APP_RELEASE_DIR)
	cd $(APP_DIR) && GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -trimpath -ldflags "$(GO_LDFLAGS)" -o $(APP_RELEASE_DIR)/worker ./cmd/worker

release-build-migrate:
	mkdir -p $(APP_RELEASE_DIR)
	cd $(APP_DIR) && GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -trimpath -ldflags "$(GO_LDFLAGS)" -o $(APP_RELEASE_DIR)/migrate ./cmd/migrate

run-api:
	cd $(APP_DIR) && go run ./cmd/api

run-worker:
	cd $(APP_DIR) && go run ./cmd/worker

migrate:
	cd $(APP_DIR) && go run ./cmd/migrate -action up

migrate-up:
	cd $(APP_DIR) && go run ./cmd/migrate -action up

migrate-version:
	cd $(APP_DIR) && go run ./cmd/migrate -action version

dev-up:
	$(DOCKER_COMPOSE) -f $(APP_DEV_COMPOSE_FILE) --env-file $(APP_DEV_ENV_FILE) up -d

dev-down:
	$(DOCKER_COMPOSE) -f $(APP_DEV_COMPOSE_FILE) --env-file $(APP_DEV_ENV_FILE) down

dev-logs:
	$(DOCKER_COMPOSE) -f $(APP_DEV_COMPOSE_FILE) --env-file $(APP_DEV_ENV_FILE) logs -f

dev-ps:
	$(DOCKER_COMPOSE) -f $(APP_DEV_COMPOSE_FILE) --env-file $(APP_DEV_ENV_FILE) ps

test-integration:
	bash scripts/app/test-integration.sh

smoke:
	bash scripts/app/smoke.sh

sqlc-generate: $(SQLC)
	cd $(APP_DIR) && $(SQLC) generate

sqlc-verify: $(SQLC)
	tmpdir=$$(mktemp -d); \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	outdir="$(CURDIR)/$(APP_DIR)/internal/infra/store/postgres/sqlc"; \
	mkdir -p "$$tmpdir/before"; \
	if [ -d "$$outdir" ]; then cp -R "$$outdir/." "$$tmpdir/before/"; fi; \
	cd $(APP_DIR) && $(SQLC) generate; \
	diff -ru "$$tmpdir/before" "$$outdir"

verify-app: sqlc-verify

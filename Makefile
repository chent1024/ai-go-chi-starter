include Makefile.rules

BIN_DIR ?= $(CURDIR)/bin
DEV_COMPOSE_FILE ?= deploy/docker-compose.dev.yaml
DEV_ENV_FILE ?= deploy/.env.dev.example
DOCKER_COMPOSE ?= docker compose

.PHONY: build build-api build-worker build-migrate
.PHONY: run-api run-worker migrate migrate-up migrate-version
.PHONY: dev-up dev-down dev-logs dev-ps

build: build-api build-worker build-migrate

build-api:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/api ./cmd/api

build-worker:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/worker ./cmd/worker

build-migrate:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/migrate ./cmd/migrate

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

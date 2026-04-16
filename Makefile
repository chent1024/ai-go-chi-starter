include Makefile.rules

.PHONY: run-api run-worker migrate-up

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

migrate-up:
	go run ./cmd/migrate -action up

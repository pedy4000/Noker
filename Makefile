# Makefile — Noker

.PHONY: all build run dev test test-coverage test-race migrate-up migrate-down sqlc generate seed clean db-up db-down run-prod run-dev air docker-build docker-up docker-down lint

include .env
export $(shell sed 's/=.*//' .env)

# ===================================================================
# Core
# ===================================================================
all: generate build

build:
	go build -trimpath -ldflags="-s -w" -o bin/noker ./cmd/noker

run: build
	./bin/noker

# ===================================================================
# Database & Migrations
# ===================================================================
DATABASE_URL ?= postgres://noker:noker@localhost:5432/noker?sslmode=disable
DATABASE_TEST_URL ?= postgres://noker:noker@localhost:5432/noker_test?sslmode=disable

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down-to 0

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir migrations create $$name sql

migrate-up-test:
	goose -dir migrations postgres "$(DATABASE_TEST_URL)" up


# ===================================================================
# Code Generation
# ===================================================================
sqlc:
	sqlc generate

# ===================================================================
# Testing
# ===================================================================
test: migrate-up-test
	go test -v -count=1 ./tests/

test-coverage: migrate-up-test
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out

test-integration: migrate-up-test
	go test -tags=integration -v ./tests/...

# ===================================================================
# Docker
# ===================================================================
docker-build:
	docker build -t noker:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f noker

docker-rebuild: docker-down docker-build docker-up

# ===================================================================
# Utilities
# ===================================================================
seed:
	python scripts/seed.py

clean:
	rm -rf bin/ coverage.out

# ===================================================================
# One-liners for live demo
# ===================================================================
demo-up: docker-up migrate-up seed
	@echo "Noker is ready!"
	@echo "Graph: http://localhost:8080/graph"
	@echo "API:   curl -H 'X-API-Key: noker-dev-key-2025' ..."

run-prod: docker-build docker-up migrate-up
	@echo "Noker running in production mode"

# ===================================================================
# PostgreSQL
# ===================================================================
postgres-up:
	@echo "Starting PostgreSQL..."
	docker-compose up -d postgres

postgres-down:
	docker-compose down postgres

postgres-restart: postgres-down postgres-up

postgres-logs:
	docker-compose logs -f postgres

postgres-shell:
	docker-compose exec postgres psql -U noker -d noker

postgres-reset: postgres-down
	docker volume rm noker_postgres_data || true
	$(MAKE) postgres-up
	sleep 10
	$(MAKE) migrate-up

postgres-check:
	@docker-compose ps postgres | grep "Up" || (echo "PostgreSQL is DOWN" && exit 1)
	@echo "PostgreSQL is UP and running"
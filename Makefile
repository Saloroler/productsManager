DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/products?sslmode=disable
TEST_DATABASE_URL ?= $(DATABASE_URL)
GOCACHE ?= $(CURDIR)/.gocache
GOMODCACHE ?= $(CURDIR)/.gomodcache
GO_ENV := env GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)

.PHONY: up down logs migrate migrate-up migrate-down migrate-reset test test-integration run-products run-notifications

up:
	docker compose up -d --build

down:
	docker compose down -v

logs:
	docker compose logs -f

migrate: migrate-up

migrate-up:
	$(GO_ENV) go run ./cmd/migrate -database-url "$(DATABASE_URL)" -command up

migrate-down:
	$(GO_ENV) go run ./cmd/migrate -database-url "$(DATABASE_URL)" -command down -steps 1

migrate-reset:
	$(GO_ENV) go run ./cmd/migrate -database-url "$(DATABASE_URL)" -command down
	$(GO_ENV) go run ./cmd/migrate -database-url "$(DATABASE_URL)" -command up

test:
	$(GO_ENV) go test ./...

test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" SQS_ENABLED=false $(GO_ENV) go test ./...

run-products:
	$(GO_ENV) go run ./cmd/products

run-notifications:
	$(GO_ENV) go run ./cmd/notifications

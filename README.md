# productsManager

Go monorepo with two microservices:
- `products`: HTTP API on `:3000`, PostgreSQL persistence, Prometheus metrics, SQS publishing
- `notifications`: SQS consumer that logs product events and acknowledges messages

Local infrastructure is provided by Docker Compose:
- PostgreSQL on `localhost:5432`
- LocalStack SQS on `localhost:4566`
- Prometheus on `localhost:9090`

## Requirements

- Go `1.26+`
- Docker Desktop or Docker Engine with Compose v2

## Project layout

- `cmd/products/main.go`
- `cmd/notifications/main.go`
- `cmd/migrate/main.go`
- `internal/products/...`
- `internal/notifications/...`
- `migrations/`
- `docker-compose.yml`
- `prometheus.yml`
- `init-localstack.sh`
- `Makefile`

## Quick start

Start everything:

```bash
make up
```

Apply database migrations:

```bash
make migrate
```

Run tests:

```bash
go test ./...
```

The `Makefile` already uses repo-local Go caches. If you want to run the same command manually inside this repo without touching the host cache, use:

```bash
env GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go test ./...
```

## Service URLs

- Products API: `http://localhost:3000`
- Products metrics: `http://localhost:3000/metrics`
- Prometheus UI: `http://localhost:9090`
- LocalStack endpoint: `http://localhost:4566`

## API

Create a product:

```bash
curl -X POST http://localhost:3000/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Desk","price":19999}'
```

List products:

```bash
curl 'http://localhost:3000/products?page=1&limit=20'
```

Delete a product:

```bash
curl -i -X DELETE http://localhost:3000/products/1
```

Inspect notifications logs:

```bash
docker compose logs -f notifications
```

## Make targets

- `make up`: build and start all containers
- `make down`: stop containers and remove volumes
- `make logs`: follow all Compose logs
- `make migrate`: apply migrations up
- `make migrate-up`: apply migrations up
- `make migrate-down`: rollback one migration step
- `make migrate-reset`: drop all migrations and re-apply
- `make test`: run all Go tests
- `make test-integration`: run tests with `TEST_DATABASE_URL` and `SQS_ENABLED=false`
- `make run-products`: run products service from the host
- `make run-notifications`: run notifications service from the host

## Configuration

Defaults are tuned for Docker Compose:

- `DATABASE_URL=postgres://postgres:postgres@psql:5432/products?sslmode=disable`
- `HTTP_PORT=3000` for products, `3001` for notifications
- `AWS_REGION=us-east-1`
- `AWS_ENDPOINT_URL=http://localstack:4566`
- `AWS_ACCESS_KEY_ID=test`
- `AWS_SECRET_ACCESS_KEY=test`
- `SQS_QUEUE_NAME=product-events`
- `SQS_QUEUE_URL=` optional explicit queue URL
- `SQS_ENABLED=true`
- `SQS_WAIT_TIME_SECONDS=10`
- `SQS_MAX_MESSAGES=10`
- `TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/products?sslmode=disable`

## Testing

The products tests:
- use `httptest`
- use a real PostgreSQL database
- truncate `products` between tests
- disable SQS publishing by injecting a no-op publisher

Run the integration tests after the stack is up and migrations have been applied:

```bash
make up
make migrate
make test
```

## Caveats

- Product writes and SQS publishes are not atomic. If the DB write succeeds and SQS publish fails, the API returns `500` but the row change may already exist.
- The notifications service is intentionally minimal and only logs received events before deleting them from the queue.

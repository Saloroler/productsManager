# ProductsManager

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
make test
```

## Migrations

Migrations are executed by the in-repo executable at [`cmd/migrate/main.go`](/Users/olegmaster/Work/CodeActivity/guruappsProductsManager/productsManager/cmd/migrate/main.go), not by a separately installed CLI tool.

It uses:
- `github.com/golang-migrate/migrate/v4`

Migration files live in [`migrations/`](/Users/olegmaster/Work/CodeActivity/guruappsProductsManager/productsManager/migrations).

The Make targets wrap this executable:
- `make migrate`
- `make migrate-up`
- `make migrate-down`
- `make migrate-reset`

Equivalent direct command:

```bash
go run ./cmd/migrate -database-url "postgres://postgres:postgres@localhost:5432/products?sslmode=disable" -command up
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

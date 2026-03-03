package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTestDatabaseURL = "postgres://postgres:postgres@localhost:5432/products?sslmode=disable"

func OpenTestConn(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultTestDatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test connection: %v", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		t.Fatalf("ping test postgres: %v", err)
	}

	t.Cleanup(conn.Close)
	return conn
}

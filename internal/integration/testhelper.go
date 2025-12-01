package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	var pool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				break
			}
			pool.Close()
		}
		if i < 4 {
			time.Sleep(time.Second)
		}
	}
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("failed to create connection pool: %v", err)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		pool.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("failed to apply migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	sql := `
CREATE TABLE IF NOT EXISTS couriers (
    id                  BIGSERIAL PRIMARY KEY,
    name                TEXT NOT NULL,
    phone               TEXT NOT NULL UNIQUE,
    status              TEXT NOT NULL DEFAULT 'available',
    transport_type      TEXT NOT NULL DEFAULT 'on_foot',
    created_at          TIMESTAMP DEFAULT now(),
    updated_at          TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS delivery (
    id                  BIGSERIAL PRIMARY KEY,
    courier_id          BIGINT NOT NULL,
    order_id            VARCHAR(255) NOT NULL,
    status              VARCHAR(50) NOT NULL DEFAULT 'active',
    assigned_at         TIMESTAMP NOT NULL DEFAULT NOW(),
    deadline            TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP DEFAULT NULL
);
`
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	return nil
}

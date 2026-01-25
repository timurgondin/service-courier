package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustInitDB() *pgxpool.Pool {
	var dbPool *pgxpool.Pool

	config, err := pgxpool.ParseConfig(getConnectionString())
	if err != nil {
		log.Fatalf("Unable to parse connection string: %v\n", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Panicf("Unable to create connection pool: %v\n", err)
	}

	err = pingDatabaseWithRetry(ctx, dbPool, 5, 2*time.Second)
	if err != nil {
		dbPool.Close()
		log.Panicf("Unable to ping database: %v\n", err)
	}

	log.Println("Database connection pool established")
	return dbPool
}

func getConnectionString() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func pingDatabaseWithRetry(ctx context.Context, dbPool *pgxpool.Pool, maxRetries int, retryDelay time.Duration) error {
	for i := range maxRetries {
		err := dbPool.Ping(ctx)
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			log.Printf("db ping attempt %d failed: %v", i+1, err)
			time.Sleep(retryDelay)
		}
	}
	return fmt.Errorf("failed to ping database after %d attempts", maxRetries)
}

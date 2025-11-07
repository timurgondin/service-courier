package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"service-courier/internal/handler"
	"service-courier/internal/repository"
	"service-courier/internal/service"
	"syscall"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	flag "github.com/spf13/pflag"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	dbPool := initDBPool()
	courierRepository := repository.NewCourierRepository(dbPool)
	courierService := service.NewCourierService(courierRepository)
	courier := handler.NewCourierHandler(courierService)

	srv := &http.Server{
		Addr:    ":" + resolvePort(),
		Handler: initRouter(courier),
	}

	go gracefulShutdown(srv, dbPool)

	log.Printf("Server started on %s\n", srv.Addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server start error: %v\n", err)
	}

}

func resolvePort() string {
	port := os.Getenv("PORT")

	var portFlag = flag.String("port", "", "укажите порт")
	flag.Parse()

	if portFlag != nil && *portFlag != "" {
		port = *portFlag
	}

	if port == "" {
		log.Fatalf("Config error")
	}

	return port
}

func gracefulShutdown(srv *http.Server, dbPool *pgxpool.Pool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}
	log.Println("Shutting down service-courier")

}

func initRouter(courier *handler.CourierHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/ping", handler.Ping)
	r.Head("/healthcheck", handler.HealthCheck)

	r.Get("/couriers", courier.GetAll)

	r.Route("/courier", func(r chi.Router) {
		r.Get("/{id}", courier.Get)
		r.Post("/", courier.Create)
		r.Put("/", courier.Update)
	})

	return r
}

func initDBPool() *pgxpool.Pool {
	var dbPool *pgxpool.Pool

	connString := getConnectionString()
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	config, err := pgxpool.ParseConfig(connString)
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
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"service-courier/internal/handler/common"
	courierHandler "service-courier/internal/handler/courier"
	deliveryHandler "service-courier/internal/handler/delivery"
	courierRepo "service-courier/internal/repository/courier"
	deliveryRepo "service-courier/internal/repository/delivery"
	courierService "service-courier/internal/service/courier"
	deliveryService "service-courier/internal/service/delivery"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	flag "github.com/spf13/pflag"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	dbPool := mustInitDB()

	courierRepository := courierRepo.NewCourierRepository(dbPool)
	courierService := courierService.NewCourierService(courierRepository)
	courier := courierHandler.NewCourierHandler(courierService)

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(dbPool, ctxGetter)
	deliveryTimeFactory := deliveryService.NewDeliveryTimeFactory()

	txManager := manager.Must(trmpgx.NewDefaultFactory(dbPool))

	deliveryService := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		deliveryTimeFactory,
		txManager,
	)
	delivery := deliveryHandler.NewDeliveryHandler(deliveryService)

	srv := &http.Server{
		Addr:    ":" + resolvePort(),
		Handler: initRouter(courier, delivery),
	}

	serverErr := make(chan error, 1)
	go func() {
		defer close(serverErr)
		log.Printf("Server started on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	waitGracefulShutdown(srv, dbPool, serverErr)

	log.Println("Shutting down service-courier")
}

func resolvePort() string {
	port := os.Getenv("PORT")

	var portFlag = flag.String("port", "", "укажите порт")
	flag.Parse()

	if portFlag != nil && *portFlag != "" {
		port = *portFlag
	}

	if port == "" {
		log.Fatalf("Server port is not specified")
	}

	return port
}

func waitGracefulShutdown(srv *http.Server, dbPool *pgxpool.Pool, serverErr <-chan error) {
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		log.Printf("Server error occurred: %v\n", err)
	case <-sigCtx.Done():
		log.Println("Shutdown initiated by signal")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v\n", err)
	} else {
		log.Println("HTTP server stoped")
	}

	log.Println("Closing DB pool...")
	dbPool.Close()
	log.Println("DB pool closed")
}

func initRouter(courier *courierHandler.Handler, delivery *deliveryHandler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/ping", common.Ping)
	r.Head("/healthcheck", common.HealthCheck)

	r.Get("/couriers", courier.GetAll)

	r.Route("/courier", func(r chi.Router) {
		r.Get("/{id}", courier.Get)
		r.Post("/", courier.Create)
		r.Put("/", courier.Update)
	})

	r.Route("/delivery", func(r chi.Router) {
		r.Post("/assign", delivery.Assign)
		r.Post("/unassign", delivery.Unassign)
	})

	return r
}

func mustInitDB() *pgxpool.Pool {
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
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	err = pingDatabaseWithRetry(ctx, dbPool, 2, 2*time.Second)
	if err != nil {
		dbPool.Close()
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

func pingDatabaseWithRetry(ctx context.Context, dbPool *pgxpool.Pool, maxRetries int, retryDelay time.Duration) error {
	for i := range maxRetries {
		err := dbPool.Ping(ctx)
		if err == nil {
			return nil
		}

		if i < maxRetries {
			log.Printf("db ping attempt %d failed: %v", i+1, err)
			time.Sleep(retryDelay)
		}
	}
	return fmt.Errorf("failed to ping database after %d attempts", maxRetries)
}

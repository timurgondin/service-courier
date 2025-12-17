package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	orderGateway "service-courier/internal/gateway/order"
	"service-courier/internal/handler/common"
	courierHandler "service-courier/internal/handler/courier"
	deliveryHandler "service-courier/internal/handler/delivery"
	db "service-courier/internal/pkg/db"
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

	dbPool := db.MustInitDB()

	courierRepository := courierRepo.NewCourierRepository(dbPool)
	courierSvc := courierService.NewCourierService(courierRepository)
	courier := courierHandler.NewCourierHandler(courierSvc)

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(dbPool, ctxGetter)
	deliveryTransportFactory := deliveryService.NewTransportFactory()

	txManager := manager.Must(trmpgx.NewDefaultFactory(dbPool))

	clock := deliveryService.RealClock{}

	deliverySvc := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		deliveryTransportFactory,
		txManager,
		clock,
	)
	delivery := deliveryHandler.NewDeliveryHandler(deliverySvc)

	orderCfg := orderGateway.LoadConfig()
	orderClient, err := orderGateway.NewClient(orderCfg)
	if err != nil {
		log.Fatalf("Failed to init order gateway: %v", err)
	}
	defer orderClient.Close()

	orderWorker := deliveryService.NewOrderWorker(deliverySvc, orderClient.Gateway, clock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	releaseInterval := resolveReleaseInterval()
	worker := deliveryService.NewWorker(deliverySvc, releaseInterval)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Start(ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		orderWorker.Start(ctx)
	}()

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

	waitGracefulShutdown(cancel, srv, dbPool, serverErr, &wg)

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

func waitGracefulShutdown(
	cancel context.CancelFunc,
	srv *http.Server,
	dbPool *pgxpool.Pool,
	serverErr <-chan error,
	wg *sync.WaitGroup,
) {
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		log.Printf("Server error occurred: %v\n", err)
	case <-sigCtx.Done():
		log.Println("Shutdown initiated by signal")
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v\n", err)
	} else {
		log.Println("HTTP server stopped")
	}

	log.Println("Waiting for workers to stop...")
	workerDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workerDone)
	}()

	select {
	case <-workerDone:
		log.Println("Workers stopped")
	case <-time.After(5 * time.Second):
		log.Println("Workers shutdown timeout - proceeding anyway")
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

func resolveReleaseInterval() time.Duration {
	env := os.Getenv("RELEASE_INTERVAL_SECONDS")
	if env == "" {
		return 10 * time.Second
	}
	sec, err := strconv.Atoi(env)
	if err != nil || sec <= 0 {
		return 10 * time.Second
	}
	return time.Duration(sec) * time.Second
}

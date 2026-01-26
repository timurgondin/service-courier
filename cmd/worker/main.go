package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderChangedHandler "service-courier/internal/handler/queues/order/changed"
	"service-courier/internal/pkg/db"
	courierRepo "service-courier/internal/repository/courier"
	deliveryRepo "service-courier/internal/repository/delivery"
	deliveryService "service-courier/internal/service/delivery"
	orderChangedUC "service-courier/internal/service/order/changed"

	"github.com/IBM/sarama"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/joho/godotenv"
)

func main() {
	unused := "variable for test linter"

	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %s", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_ORDER_TOPIC")
	if topic == "" {
		topic = "order.status.changed"
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "my-group-id"
	}

	// Инициализируем клиента кафки
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	kafkaClient, err := sarama.NewConsumerGroup([]string{broker}, groupID, config)
	if err != nil {
		log.Printf("unable to create kafka consumer group: %v", err)
		cancel()
		return
	}
	defer func() {
		if err := kafkaClient.Close(); err != nil {
			log.Printf("kafka client close error: %v", err)
		}
	}()

	dbPool := db.MustInitDB()

	ctxGetter := trmpgx.DefaultCtxGetter

	courierRepository := courierRepo.NewCourierRepository(dbPool)

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

	// usecase
	orderChangedUsecase := orderChangedUC.NewUsecase(deliverySvc)

	// handler
	orderChangeHandler := orderChangedHandler.NewHandler(orderChangedUsecase)

	go func() {
		for {
			err := kafkaClient.Consume(ctx, []string{topic}, orderChangeHandler)
			if err != nil {
				log.Printf("consume error: %v", err)
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	log.Println("Assign worker started")

	waitGracefulShutdown(cancel)

	log.Println("Assign worker stopped gracefully.")

}

func waitGracefulShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var signal string
	select {
	case sig := <-sigChan:
		signal = sig.String()
	default:
	}

	cancel()

	log.Printf("Shutdown signal (%s)", signal)
}

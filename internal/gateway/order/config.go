package order

import (
	"os"
	"time"
)

type Config struct {
	Addr     string
	Timeout  time.Duration
	Lookback time.Duration
}

func LoadConfig() Config {
	addr := os.Getenv("ORDER_SERVICE_GRPC_ADDR")
	if addr == "" {
		addr = "service-order:50051"
	}

	return Config{
		Addr:     addr,
		Timeout:  3 * time.Second,
		Lookback: 5 * time.Second,
	}
}

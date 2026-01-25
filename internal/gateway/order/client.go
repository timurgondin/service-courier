package order

import (
	"fmt"

	pb "service-courier/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	Gateway *Gateway
}

func NewClient(cfg Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to gRPC server: %v", err)
	}

	return &Client{
		conn:    conn,
		Gateway: NewGateway(pb.NewOrdersServiceClient(conn)),
	}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

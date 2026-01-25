package order

import (
	"context"
	"fmt"
	"service-courier/internal/model/order"
	"service-courier/internal/metrics"
	"service-courier/internal/pkg/retry"
	pb "service-courier/internal/proto"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Gateway struct {
	client client
}

func NewGateway(c client) *Gateway {
	return &Gateway{
		client: c,
	}
}

func (g *Gateway) GetOrders(ctx context.Context, from time.Time) ([]order.Order, error) {
	pbReq := &pb.GetOrdersRequest{From: timestamppb.New(from)}
	exec := retry.NewRetryExecutor(retry.RetryConfig{
		MaxAttempts: 3,
		Strategy:    retry.NewExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0),
		ShouldRetry: isRetryable,
	})

	var resp *pb.GetOrdersResponse
	err := exec.ExecuteWithCallback(
		func() error {
			r, err := g.client.GetOrders(ctx, pbReq)
			if err != nil {
				return err
			}
			resp = r
			return nil
		},
		func(attempt int, err error, delay time.Duration) {
			metrics.GatewayRetriesTotal.Inc()
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get orders failed: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("get orders failed: response is nil")
	}
	orders := make([]order.Order, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		orders = append(orders, order.Order{
			ID:        o.GetId(),
			CreatedAt: o.GetCreatedAt().AsTime(),
		})
	}
	return orders, nil
}

func isRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

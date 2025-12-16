package order

import (
	"context"
	"fmt"
	"service-courier/internal/model/order"
	pb "service-courier/internal/proto"
	"time"

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
	resp, err := g.client.GetOrders(ctx, pbReq)
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

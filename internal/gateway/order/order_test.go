package order_test

import (
	"context"
	"testing"
	"time"

	order "service-courier/internal/gateway/order"
	"service-courier/internal/metrics"
	pb "service-courier/internal/proto"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type stubClient struct {
	responses []stubResponse
	calls     int
}

type stubResponse struct {
	resp *pb.GetOrdersResponse
	err  error
}

func (s *stubClient) GetOrders(ctx context.Context, in *pb.GetOrdersRequest, opts ...grpc.CallOption) (*pb.GetOrdersResponse, error) {
	if s.calls >= len(s.responses) {
		s.calls++
		return nil, status.Error(codes.Unavailable, "no response configured")
	}
	r := s.responses[s.calls]
	s.calls++
	return r.resp, r.err
}

func (s *stubClient) GetOrderById(ctx context.Context, in *pb.GetOrderByIdRequest, opts ...grpc.CallOption) (*pb.GetOrderByIdResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}

func TestGatewayGetOrders_RetryOnTemporary(t *testing.T) {
	now := time.Now()
	resp := &pb.GetOrdersResponse{
		Orders: []*pb.Order{{Id: "1", CreatedAt: timestamppb.New(now)}},
	}
	client := &stubClient{
		responses: []stubResponse{
			{resp: nil, err: status.Error(codes.Unavailable, "temporary")},
			{resp: resp, err: nil},
		},
	}

	gw := order.NewGateway(client)

	before := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	orders, err := gw.GetOrders(context.Background(), now)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if client.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", client.calls)
	}
	after := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	if after != before+1 {
		t.Fatalf("expected retries metric to increment by 1, before: %v, after: %v", before, after)
	}
}

func TestGatewayGetOrders_NoRetryOnPermanent(t *testing.T) {
	client := &stubClient{
		responses: []stubResponse{
			{resp: nil, err: status.Error(codes.InvalidArgument, "bad request")},
		},
	}

	gw := order.NewGateway(client)

	before := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	_, err := gw.GetOrders(context.Background(), time.Now())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if client.calls != 1 {
		t.Fatalf("expected 1 call, got %d", client.calls)
	}
	after := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	if after != before {
		t.Fatalf("expected retries metric to stay the same, before: %v, after: %v", before, after)
	}
}

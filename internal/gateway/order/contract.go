package order

import (
	"context"

	pb "service-courier/internal/proto"

	"google.golang.org/grpc"
)

type client interface {
	GetOrders(ctx context.Context, in *pb.GetOrdersRequest, opts ...grpc.CallOption) (*pb.GetOrdersResponse, error)
	GetOrderById(ctx context.Context, in *pb.GetOrderByIdRequest, opts ...grpc.CallOption) (*pb.GetOrderByIdResponse, error)
}

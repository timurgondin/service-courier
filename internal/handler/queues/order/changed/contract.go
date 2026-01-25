package changed

import (
	"context"
	"service-courier/internal/model/order"
)

type usecase interface {
	Process(ctx context.Context, order order.Order) error
}

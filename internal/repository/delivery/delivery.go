package delivery

import (
	"context"
	"errors"
	"fmt"
	"service-courier/internal/model/delivery"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool         *pgxpool.Pool
	getter       *trmpgx.CtxGetter
	queryBuilder squirrel.StatementBuilderType
}

func NewDeliveryRepository(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{
		pool:         pool,
		getter:       getter,
		queryBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *Repository) exec(ctx context.Context) trmpgx.Tr {
	return r.getter.DefaultTrOrDB(ctx, r.pool)
}

func (r *Repository) Create(ctx context.Context, deliveryData delivery.Delivery) error {
	query, args, err := r.queryBuilder.
		Insert("delivery").
		Columns("courier_id", "order_id", "assigned_at", "deadline").
		Values(deliveryData.CourierID, deliveryData.OrderID, deliveryData.AssignedAt, deliveryData.Deadline).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	err = r.exec(ctx).QueryRow(ctx, query, args...).Scan(&deliveryData.ID)

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*delivery.Delivery, error) {
	query, args, err := r.queryBuilder.
		Select("id", "courier_id", "order_id", "assigned_at", "deadline").
		From("delivery").
		Where(squirrel.Eq{"order_id": orderID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	var deliveryData delivery.Delivery
	err = r.exec(ctx).QueryRow(ctx, query, args...).Scan(
		&deliveryData.ID,
		&deliveryData.CourierID,
		&deliveryData.OrderID,
		&deliveryData.AssignedAt,
		&deliveryData.Deadline,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, delivery.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("query delivery: %w", err)
	}
	return &deliveryData, nil
}

func (r *Repository) DeleteByOrderID(ctx context.Context, orderID string) error {
	query, args, err := r.queryBuilder.
		Delete("delivery").
		Where(squirrel.Eq{"order_id": orderID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete: %w", err)
	}

	result, err := r.exec(ctx).Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete delivery: %w", err)
	}
	if result.RowsAffected() == 0 {
		return delivery.ErrDeliveryNotFound
	}
	return nil
}

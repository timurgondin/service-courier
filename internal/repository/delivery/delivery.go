package delivery

import (
	"context"
	"errors"
	"fmt"
	"service-courier/internal/model/delivery"
	"time"

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
		Columns("courier_id", "order_id", "status", "assigned_at", "deadline").
		Values(
			deliveryData.CourierID,
			deliveryData.OrderID,
			delivery.StatusActive,
			deliveryData.AssignedAt,
			deliveryData.Deadline,
		).
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
		Select("id", "courier_id", "order_id", "status", "assigned_at", "deadline").
		From("delivery").
		Where(squirrel.Eq{"order_id": orderID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var deliveryData delivery.Delivery
	err = r.exec(ctx).QueryRow(ctx, query, args...).Scan(
		&deliveryData.ID,
		&deliveryData.CourierID,
		&deliveryData.OrderID,
		&deliveryData.Status,
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
		return fmt.Errorf("build query: %w", err)
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

func (r *Repository) ListExpired(ctx context.Context, now time.Time) ([]delivery.Delivery, error) {
	query, args, err := r.queryBuilder.
		Select("id", "courier_id", "order_id", "status", "assigned_at", "deadline").
		From("delivery").
		Where(squirrel.And{
			squirrel.Lt{"deadline": now},
			squirrel.Eq{"status": delivery.StatusActive},
		}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.exec(ctx).Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	expired := make([]delivery.Delivery, 0)
	for rows.Next() {
		var deliveryData delivery.Delivery
		err := rows.Scan(
			&deliveryData.ID,
			&deliveryData.CourierID,
			&deliveryData.OrderID,
			&deliveryData.Status,
			&deliveryData.AssignedAt,
			&deliveryData.Deadline,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading data: %w", err)
		}
		expired = append(expired, deliveryData)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return expired, nil
}

func (r *Repository) UpdateStatusByIDs(ctx context.Context, ids []int64, status delivery.DeliveryStatus) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := r.queryBuilder.
		Update("delivery").
		Set("status", string(status)).
		Where(squirrel.Eq{"id": ids}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.exec(ctx).Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

package courier

import (
	"context"
	"errors"
	"fmt"
	"service-courier/internal/model/courier"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool         *pgxpool.Pool
	queryBuilder squirrel.StatementBuilderType
}

func NewCourierRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:         pool,
		queryBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*courier.Courier, error) {
	query := r.queryBuilder.
		Select("id", "name", "phone", "status", "transport_type", "created_at", "updated_at").
		From("couriers").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var courierData courier.Courier
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&courierData.ID,
		&courierData.Name,
		&courierData.Phone,
		&courierData.Status,
		&courierData.TransportType,
		&courierData.CreatedAt,
		&courierData.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, courier.ErrCourierNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &courierData, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]courier.Courier, error) {
	query, args, err := r.queryBuilder.
		Select("id", "name", "phone", "status", "transport_type", "created_at", "updated_at").
		From("couriers").
		OrderBy("id").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	couriers := make([]courier.Courier, 0)
	for rows.Next() {
		var courierData courier.Courier
		err := rows.Scan(
			&courierData.ID,
			&courierData.Name,
			&courierData.Phone,
			&courierData.Status,
			&courierData.TransportType,
			&courierData.CreatedAt,
			&courierData.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading data: %w", err)
		}
		couriers = append(couriers, courierData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return couriers, nil
}

func (r *Repository) Create(ctx context.Context, courierData courier.Courier) (id int64, err error) {
	query, args, err := r.queryBuilder.
		Insert("couriers").
		Columns("name", "phone", "status", "transport_type").
		Values(courierData.Name, courierData.Phone, courierData.Status, courierData.TransportType).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return id, fmt.Errorf("build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, query, args...).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, courier.ErrPhoneExists
		}
		return 0, fmt.Errorf("database error: %w", err)
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, courierData courier.Courier) error {
	updateBuilder := r.queryBuilder.
		Update("couriers").
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": courierData.ID})

	if courierData.Name != "" {
		updateBuilder = updateBuilder.Set("name", courierData.Name)
	}
	if courierData.Phone != "" {
		updateBuilder = updateBuilder.Set("phone", courierData.Phone)
	}
	if courierData.Status != "" {
		updateBuilder = updateBuilder.Set("status", courierData.Status)
	}
	if courierData.TransportType != "" {
		updateBuilder = updateBuilder.Set("transport_type", courierData.TransportType)
	}

	query, args, err := updateBuilder.ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	result, err := r.pool.Exec(ctx, query, args...)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return courier.ErrPhoneExists
		}
		return fmt.Errorf("database error: %w", err)
	}

	if result.RowsAffected() == 0 {
		return courier.ErrCourierNotFound
	}

	return nil
}

func (r *Repository) GetAvailableWithMinDeliveries(ctx context.Context) (*courier.Courier, error) {
	query, args, err := r.queryBuilder.
		Select(
			"c.id",
			"c.name",
			"c.phone",
			"c.status",
			"c.transport_type",
			"c.created_at",
			"c.updated_at",
		).
		From("couriers c").
		LeftJoin("delivery d ON d.courier_id = c.id AND d.status = 'completed'").
		Where(squirrel.Eq{"c.status": "available"}).
		GroupBy(
			"c.id",
			"c.name",
			"c.phone",
			"c.status",
			"c.transport_type",
			"c.created_at",
			"c.updated_at",
		).
		OrderBy("COUNT(d.id) ASC").
		Limit(1).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var courierData courier.Courier
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&courierData.ID,
		&courierData.Name,
		&courierData.Phone,
		&courierData.Status,
		&courierData.TransportType,
		&courierData.CreatedAt,
		&courierData.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, courier.ErrNoAvailableCouriers
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &courierData, nil
}

func (r *Repository) UpdateStatusBatch(ctx context.Context, ids []int64, status courier.CourierStatus) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := r.queryBuilder.
		Update("couriers").
		Set("status", status).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": ids}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update status batch: %w", err)
	}

	return nil
}

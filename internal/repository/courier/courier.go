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
		Select("id", "name", "phone", "status", "created_at", "updated_at").
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
		Select("id", "name", "phone", "status", "created_at", "updated_at").
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
		var courier courier.Courier
		err := rows.Scan(
			&courier.ID,
			&courier.Name,
			&courier.Phone,
			&courier.Status,
			&courier.CreatedAt,
			&courier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading data: %w", err)
		}
		couriers = append(couriers, courier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return couriers, nil
}

func (r *Repository) Create(ctx context.Context, courierData courier.Courier) (id int64, err error) {
	query, args, err := r.queryBuilder.
		Insert("couriers").
		Columns("name", "phone", "status").
		Values(courierData.Name, courierData.Phone, courierData.Status).
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

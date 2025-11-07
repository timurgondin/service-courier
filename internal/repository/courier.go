package repository

import (
	"context"
	"fmt"
	"service-courier/internal/model"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourierRepository struct {
	pool *pgxpool.Pool
}

func NewCourierRepository(pool *pgxpool.Pool) *CourierRepository {
	return &CourierRepository{pool: pool}
}

func (r *CourierRepository) GetByID(ctx context.Context, id int64) (*model.CourierDB, error) {
	query, args, err := sq.
		Select("id", "name", "phone", "status").
		From("couriers").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	var courier model.CourierDB
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&courier.ID,
		&courier.Name,
		&courier.Phone,
		&courier.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrCourierNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &courier, nil
}

func (r *CourierRepository) GetAll(ctx context.Context) ([]model.CourierDB, error) {
	query, args, err := sq.
		Select("id", "name", "phone", "status").
		From("couriers").
		OrderBy("id ASC").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var couriers []model.CourierDB
	for rows.Next() {
		var courier model.CourierDB
		err := rows.Scan(
			&courier.ID,
			&courier.Name,
			&courier.Phone,
			&courier.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading data: %w", err)
		}
		couriers = append(couriers, courier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if couriers == nil {
		couriers = []model.CourierDB{}
	}

	return couriers, nil
}

func (r *CourierRepository) Create(ctx context.Context, courier *model.CourierDB) (int64, error) {
	var id int64
	query, args, err := sq.
		Insert("couriers").Columns("name", "phone", "status").
		Values(courier.Name, courier.Phone, courier.Status).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, query, args...).Scan(&id)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return 0, model.ErrPhoneExists
		}
		return 0, fmt.Errorf("database error: %w", err)
	}
	return id, nil
}

func (r *CourierRepository) Update(ctx context.Context, courier *model.CourierUpdateDB) error {

	updateBuilder := sq.Update("couriers")

	if courier.Name != nil {
		updateBuilder = updateBuilder.Set("name", *courier.Name)
	}
	if courier.Phone != nil {
		updateBuilder = updateBuilder.Set("phone", *courier.Phone)
	}
	if courier.Status != nil {
		updateBuilder = updateBuilder.Set("status", *courier.Status)
	}

	query, args, err := updateBuilder.
		Set("updated_at", courier.UpdatedAt).
		Where((sq.Eq{"id": courier.ID})).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	cmdTag, err := r.pool.Exec(ctx, query, args...)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return model.ErrPhoneExists
		}
		return fmt.Errorf("database error: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return model.ErrCourierNotFound
	}

	return nil
}

package repository

import (
	"context"
	"time"

	"subscription/internal/subscription"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *subscription.Subscription) (uuid.UUID, error) {
	var id uuid.UUID
	const q = `
	INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`
	err := r.pool.QueryRow(ctx, q, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).Scan(&id)
	return id, err
}

func (r *SubscriptionRepository) Get(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	const q = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions WHERE id = $1;
	`
	row := r.pool.QueryRow(ctx, q, id)
	var out subscription.Subscription
	if err := row.Scan(&out.ID, &out.ServiceName, &out.Price, &out.UserID, &out.StartDate, &out.EndDate, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *subscription.Subscription) error {
	const q = `
		UPDATE subscriptions
		SET service_name=$2, price=$3, user_id=$4, start_date=$5, end_date=$6, updated_at=now()
		WHERE id=$1;
	`
	_, err := r.pool.Exec(ctx, q, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	return err
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id=$1;`, id)
	return err
}

type ListFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

func (r *SubscriptionRepository) List(ctx context.Context, f ListFilter) ([]subscription.Subscription, error) {
	const q = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
		  AND ($2::text IS NULL OR service_name ILIKE '%'||$2||'%')
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4;
	`
	rows, err := r.pool.Query(ctx, q, f.UserID, f.ServiceName, f.Limit, f.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []subscription.Subscription
	for rows.Next() {
		var s subscription.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type SummaryFilter struct {
	Start       time.Time
	End         time.Time
	UserID      *uuid.UUID
	ServiceName *string
}

func (r *SubscriptionRepository) Summary(ctx context.Context, f SummaryFilter) (int, error) {
	const q = `
WITH months AS (
  SELECT generate_series($1::date, $2::date, interval '1 month')::date AS m
)
SELECT COALESCE(SUM(s.price), 0)::int AS total
FROM months m
JOIN subscriptions s
  ON s.start_date <= m.m
 AND (s.end_date IS NULL OR s.end_date >= m.m)
WHERE ($3::uuid IS NULL OR s.user_id = $3)
  AND ($4::text IS NULL OR s.service_name ILIKE '%'||$4||'%');
`
	var total int
	err := r.pool.QueryRow(ctx, q, f.Start, f.End, f.UserID, f.ServiceName).Scan(&total)
	return total, err
}

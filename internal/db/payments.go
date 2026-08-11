package db

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xdars/grpc-tasks/internal/domain"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Add(ctx context.Context, p *domain.Payment) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO payments (
			id, 
			user_id, 
			type, 
			status, 
			amount, 
			currency
		) VALUES ($1, $2, $3, $4, $5, $6)`,

		p.ID, p.UserID, p.Type, p.Status, p.Amount, p.Currency)

	if err != nil {
		log.Printf("PaymentRepository.Add error: %v", err)
	}
	return err
}

func (r *PaymentRepository) Get(ctx context.Context, id string) (*domain.Payment, bool, error) {
	p := &domain.Payment{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, type, status, amount, currency FROM payments WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.UserID, &p.Type, &p.Status, &p.Amount, &p.Currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return p, true, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status=$1, updated_at=NOW() WHERE id=$2`,
		status, id,
	)
	return err
}

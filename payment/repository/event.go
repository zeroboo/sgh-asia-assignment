package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zeroboo.payment/model"
)

// EventRepo handles CRUD operations for the events table.
type EventRepo struct {
	db *sql.DB
}

// NewEventRepo creates a new EventRepo.
func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

// Create inserts a new event row.
func (r *EventRepo) Create(ctx context.Context, e *model.Event) error {
	return createEvent(ctx, r.db, e)
}

// CreateTx inserts a new event row within a database transaction.
func (r *EventRepo) CreateTx(ctx context.Context, dbTx *sql.Tx, e *model.Event) error {
	return createEvent(ctx, dbTx, e)
}

func createEvent(ctx context.Context, q interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, e *model.Event) error {
	query := `INSERT INTO events (id, transaction_id, type, payload, created_at)
		VALUES (?, ?, ?, ?, ?)`

	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}

	_, err := q.ExecContext(ctx, query,
		e.ID, e.TransactionID, e.Type, e.Payload, e.CreatedAt,
	)
	return err
}

// GetByID fetches a single event by its primary key.
func (r *EventRepo) GetByID(ctx context.Context, id string) (*model.Event, error) {
	query := `SELECT id, transaction_id, type, payload, created_at
		FROM events WHERE id = ?`

	var e model.Event
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.TransactionID, &e.Type, &e.Payload, &e.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get event %s: %w", id, err)
	}
	return &e, nil
}

// ListByTransactionID returns all events for a given transaction.
func (r *EventRepo) ListByTransactionID(ctx context.Context, transactionID string) ([]*model.Event, error) {
	query := `SELECT id, transaction_id, type, payload, created_at
		FROM events WHERE transaction_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("list events for transaction %s: %w", transactionID, err)
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(
			&e.ID, &e.TransactionID, &e.Type, &e.Payload, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event row: %w", err)
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}

// Delete removes an event by ID.
func (r *EventRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete event %s: %w", id, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("event %s not found", id)
	}
	return nil
}

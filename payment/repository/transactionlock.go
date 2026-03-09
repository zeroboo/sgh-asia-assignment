package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zeroboo.payment/model"
)

// TransactionLockRepo handles CRUD operations for the transaction_locks table.
type TransactionLockRepo struct {
	db *sql.DB
}

// NewTransactionLockRepo creates a new TransactionLockRepo.
func NewTransactionLockRepo(db *sql.DB) *TransactionLockRepo {
	return &TransactionLockRepo{db: db}
}

// Create inserts a new lock row (acquires the lock) with an expiration duration.
func (r *TransactionLockRepo) Create(ctx context.Context, transactionID string, expiration time.Duration) error {
	now := time.Now()
	expiresAt := now.Add(expiration)
	query := `INSERT INTO transaction_locks (transaction_id, created_at, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, transactionID, now, expiresAt)
	if err != nil {
		return fmt.Errorf("acquire lock for transaction %s: %w", transactionID, err)
	}
	return nil
}

// GetByID fetches a lock by transaction ID.
func (r *TransactionLockRepo) GetByID(ctx context.Context, transactionID string) (*model.TransactionLock, error) {
	query := `SELECT transaction_id, created_at, deleted_at
		FROM transaction_locks WHERE transaction_id = ?`

	var lock model.TransactionLock
	var deletedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, transactionID).Scan(
		&lock.ID, &lock.CreatedAt, &deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get lock for transaction %s: %w", transactionID, err)
	}
	if deletedAt.Valid {
		lock.DeletedAt = deletedAt.Time
	}
	return &lock, nil
}

// Delete soft-deletes a lock by setting deleted_at (releases the lock).
func (r *TransactionLockRepo) Delete(ctx context.Context, transactionID string) error {
	query := `UPDATE transaction_locks SET deleted_at = ? WHERE transaction_id = ? AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, time.Now(), transactionID)
	if err != nil {
		return fmt.Errorf("release lock for transaction %s: %w", transactionID, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("lock for transaction %s not found or already released", transactionID)
	}
	return nil
}

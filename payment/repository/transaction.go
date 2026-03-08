package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zeroboo.payment/model"
)

// TransactionRepo handles CRUD operations for the transactions table.
type TransactionRepo struct {
	db *sql.DB
}

// NewTransactionRepo creates a new TransactionRepo.
func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

// Create inserts a new transaction row.
func (r *TransactionRepo) Create(ctx context.Context, tx *model.Transaction) error {
	return createTransaction(ctx, r.db, tx)
}

// CreateTx inserts a new transaction row within a database transaction.
func (r *TransactionRepo) CreateTx(ctx context.Context, dbTx *sql.Tx, tx *model.Transaction) error {
	return createTransaction(ctx, dbTx, tx)
}

func createTransaction(ctx context.Context, q interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, tx *model.Transaction) error {
	query := `INSERT INTO transactions
		(transaction_id, user_id, amount, type, status, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = now
	}
	if tx.UpdatedAt.IsZero() {
		tx.UpdatedAt = now
	}

	_, err := q.ExecContext(ctx, query,
		tx.ID, tx.UserID, tx.Amount, tx.Type, tx.Status, tx.Description,
		tx.CreatedAt, tx.UpdatedAt,
	)
	return err
}

// GetByID fetches a single transaction by its primary key.
func (r *TransactionRepo) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	return getTransactionByID(ctx, r.db, id)
}

// GetByIDTx fetches a single transaction by its primary key within a database transaction.
func (r *TransactionRepo) GetByIDTx(ctx context.Context, dbTx *sql.Tx, id string) (*model.Transaction, error) {
	return getTransactionByID(ctx, dbTx, id)
}

func getTransactionByID(ctx context.Context, q interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}, id string) (*model.Transaction, error) {
	query := `SELECT transaction_id, user_id, amount, type, status, description, created_at, updated_at
		FROM transactions WHERE transaction_id = ?`

	var tx model.Transaction
	err := q.QueryRowContext(ctx, query, id).Scan(
		&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Status, &tx.Description,
		&tx.CreatedAt, &tx.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction %s: %w", id, err)
	}
	return &tx, nil
}

// UpdateStatus sets a new status for a transaction.
func (r *TransactionRepo) UpdateStatus(ctx context.Context, id string, status model.TransactionStatus) error {
	return updateTransactionStatus(ctx, r.db, id, status)
}

// UpdateStatusTx sets a new status for a transaction within a database transaction.
func (r *TransactionRepo) UpdateStatusTx(ctx context.Context, dbTx *sql.Tx, id string, status model.TransactionStatus) error {
	return updateTransactionStatus(ctx, dbTx, id, status)
}

func updateTransactionStatus(ctx context.Context, q interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, id string, status model.TransactionStatus) error {
	query := `UPDATE transactions SET status = ?, updated_at = ? WHERE transaction_id = ?`
	res, err := q.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update transaction status %s: %w", id, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("transaction %s not found", id)
	}
	return nil
}

// ListByUserID returns all transactions for a given user, ordered by created_at desc.
func (r *TransactionRepo) ListByUserID(ctx context.Context, userID string) ([]*model.Transaction, error) {
	query := `SELECT transaction_id, user_id, amount, type, status, description, created_at, updated_at
		FROM transactions WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list transactions for user %s: %w", userID, err)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var tx model.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Status, &tx.Description,
			&tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		transactions = append(transactions, &tx)
	}
	return transactions, rows.Err()
}

// Delete removes a transaction by ID.
func (r *TransactionRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM transactions WHERE transaction_id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete transaction %s: %w", id, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("transaction %s not found", id)
	}
	return nil
}

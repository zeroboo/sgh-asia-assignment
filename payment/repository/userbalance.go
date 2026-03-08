package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zeroboo.payment/model"
)

// UserBalanceRepo handles CRUD operations for the user_balances table.
type UserBalanceRepo struct {
	db *sql.DB
}

// NewUserBalanceRepo creates a new UserBalanceRepo.
func NewUserBalanceRepo(db *sql.DB) *UserBalanceRepo {
	return &UserBalanceRepo{db: db}
}

// GetByUserID fetches the balance for a given user.
func (r *UserBalanceRepo) GetByUserID(ctx context.Context, userID string) (*model.UserBalance, error) {
	return getUserBalance(ctx, r.db, userID)
}

// GetByUserIDTx fetches the balance for a given user within a database transaction.
func (r *UserBalanceRepo) GetByUserIDTx(ctx context.Context, dbTx *sql.Tx, userID string) (*model.UserBalance, error) {
	return getUserBalance(ctx, dbTx, userID)
}

func getUserBalance(ctx context.Context, q interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}, userID string) (*model.UserBalance, error) {
	query := `SELECT user_id, balance, updated_at FROM user_balances WHERE user_id = ?`

	var ub model.UserBalance
	err := q.QueryRowContext(ctx, query, userID).Scan(&ub.UserID, &ub.Balance, &ub.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get balance for user %s: %w", userID, err)
	}
	return &ub, nil
}

// Upsert inserts a new balance row or updates the existing one.
func (r *UserBalanceRepo) Upsert(ctx context.Context, ub *model.UserBalance) error {
	return upsertUserBalance(ctx, r.db, ub)
}

// UpsertTx inserts or updates within a database transaction.
func (r *UserBalanceRepo) UpsertTx(ctx context.Context, dbTx *sql.Tx, ub *model.UserBalance) error {
	return upsertUserBalance(ctx, dbTx, ub)
}

func upsertUserBalance(ctx context.Context, q interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, ub *model.UserBalance) error {
	query := `INSERT INTO user_balances (user_id, balance, updated_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE balance = VALUES(balance), updated_at = VALUES(updated_at)`

	if ub.UpdatedAt.IsZero() {
		ub.UpdatedAt = time.Now()
	}

	_, err := q.ExecContext(ctx, query, ub.UserID, ub.Balance, ub.UpdatedAt)
	return err
}

// UpdateBalance sets a new balance for a user.
func (r *UserBalanceRepo) UpdateBalance(ctx context.Context, userID string, newBalance int64) error {
	return updateUserBalance(ctx, r.db, userID, newBalance)
}

// UpdateBalanceTx sets a new balance within a database transaction.
func (r *UserBalanceRepo) UpdateBalanceTx(ctx context.Context, dbTx *sql.Tx, userID string, newBalance int64) error {
	return updateUserBalance(ctx, dbTx, userID, newBalance)
}

func updateUserBalance(ctx context.Context, q interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, userID string, newBalance int64) error {
	query := `UPDATE user_balances SET balance = ?, updated_at = ? WHERE user_id = ?`
	res, err := q.ExecContext(ctx, query, newBalance, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("update balance for user %s: %w", userID, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user balance %s not found", userID)
	}
	return nil
}

// Delete removes a user balance row.
func (r *UserBalanceRepo) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM user_balances WHERE user_id = ?`
	res, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete balance for user %s: %w", userID, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user balance %s not found", userID)
	}
	return nil
}

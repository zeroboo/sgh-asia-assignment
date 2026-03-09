package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zeroboo.payment/model"
	"zeroboo.payment/repository"
)

// PaymentService implements handler.IPaymentService using MySQL repositories.
type PaymentService struct {
	db                  *sql.DB
	transactionRepo     *repository.TransactionRepo
	transactionLockRepo *repository.TransactionLockRepo
	userBalanceRepo     *repository.UserBalanceRepo
	eventRepo           *repository.EventRepo
}

// NewPaymentService creates a PaymentService backed by the given *sql.DB.
func NewPaymentService(db *sql.DB) *PaymentService {
	return &PaymentService{
		db:                  db,
		transactionRepo:     repository.NewTransactionRepo(db),
		transactionLockRepo: repository.NewTransactionLockRepo(db),
		userBalanceRepo:     repository.NewUserBalanceRepo(db),
		eventRepo:           repository.NewEventRepo(db),
	}
}

// GetPayment retrieves a transaction by its ID. Returns (nil, nil) when not found.
func (s *PaymentService) GetPayment(transactionID string) (*model.Transaction, error) {
	ctx := context.Background()
	return s.transactionRepo.GetByID(ctx, transactionID)
}

// GetUserBalance retrieves the current wallet balance for a user.
func (s *PaymentService) GetUserBalance(userID string) (*model.UserBalance, error) {
	ctx := context.Background()
	return s.userBalanceRepo.GetByUserID(ctx, userID)
}

// LockPayment creates an advisory lock row to prevent duplicate processing.
func (s *PaymentService) LockPayment(transactionID string, expiration time.Duration) error {
	ctx := context.Background()
	return s.transactionLockRepo.Create(ctx, transactionID, expiration)
}

// UnlockPayment releases the advisory lock for a transaction.
func (s *PaymentService) UnlockPayment(transactionID string) error {
	ctx := context.Background()
	return s.transactionLockRepo.Delete(ctx, transactionID)
}

// CreatePayment creates a new debit transaction, updates the user balance,
func (s *PaymentService) CreatePayment(userID string, transactionID string, amount int64) error {
	ctx := context.Background()

	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback() // no-op after Commit

	// 1. Create the transaction record with status "created".
	tx := &model.Transaction{
		ID:             transactionID,
		UserID:         userID,
		Amount:         amount,
		Type:           model.TypeDebit,
		Status:         model.StatusCreated,
		IdempotencyKey: transactionID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.transactionRepo.CreateTx(ctx, dbTx, tx); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	// 2. Record the "payment.created" event.
	if err := s.eventRepo.CreateTx(ctx, dbTx, &model.Event{
		ID:            fmt.Sprintf("%s-%s", transactionID, string(model.EventPaymentCreated)),
		TransactionID: transactionID,
		Type:          model.EventPaymentCreated,
		Payload:       fmt.Sprintf(`{"userId":"%s","amount":%d}`, userID, amount),
	}); err != nil {
		return fmt.Errorf("record created event: %w", err)
	}

	// 3. Move to "processing".
	if err := s.transactionRepo.UpdateStatusTx(ctx, dbTx, transactionID, model.StatusProcessing); err != nil {
		return fmt.Errorf("set processing status: %w", err)
	}

	// 4. Get or initialise user balance – then debit.
	balance, err := s.userBalanceRepo.GetByUserIDTx(ctx, dbTx, userID)
	if err != nil {
		return fmt.Errorf("get user balance: %w", err)
	}

	if balance == nil {
		balance = &model.UserBalance{
			UserID:    userID,
			Balance:   0,
			UpdatedAt: time.Now(),
		}
		if err := s.userBalanceRepo.UpsertTx(ctx, dbTx, balance); err != nil {
			return fmt.Errorf("initialise user balance: %w", err)
		}
	}

	newBalance := balance.Balance - amount
	if newBalance < 0 {
		_ = s.transactionRepo.UpdateStatusTx(ctx, dbTx, transactionID, model.StatusFailed)
		_ = s.eventRepo.CreateTx(ctx, dbTx, &model.Event{
			ID:            fmt.Sprintf("%s-%s", transactionID, string(model.EventPaymentFailed)),
			TransactionID: transactionID,
			Type:          model.EventPaymentFailed,
			Payload:       `{"reason":"insufficient funds"}`,
		})
		// Commit the failed status so it is persisted.
		if commitErr := dbTx.Commit(); commitErr != nil {
			return fmt.Errorf("commit failed-status tx: %w", commitErr)
		}
		return fmt.Errorf("insufficient funds: balance %d, amount %d", balance.Balance, amount)
	}

	if err := s.userBalanceRepo.UpdateBalanceTx(ctx, dbTx, userID, newBalance); err != nil {
		return fmt.Errorf("update user balance: %w", err)
	}

	// 5. Mark complete.
	if err := s.transactionRepo.UpdateStatusTx(ctx, dbTx, transactionID, model.StatusCompleted); err != nil {
		return fmt.Errorf("set completed status: %w", err)
	}

	_ = s.eventRepo.CreateTx(ctx, dbTx, &model.Event{
		ID:            fmt.Sprintf("%s-%s", transactionID, string(model.EventPaymentCompleted)),
		TransactionID: transactionID,
		Type:          model.EventPaymentCompleted,
		Payload:       fmt.Sprintf(`{"userId":"%s","amount":%d,"newBalance":%d}`, userID, amount, newBalance),
	})

	// 6. Commit the entire database transaction.
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

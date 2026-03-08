package model

import "time"

// TransactionStatus represents the lifecycle state of a payment transaction.
type TransactionStatus string

const (
	StatusCreated    TransactionStatus = "created"
	StatusProcessing TransactionStatus = "processing"
	StatusCompleted  TransactionStatus = "completed"
	StatusFailed     TransactionStatus = "failed"
	StatusRefunded   TransactionStatus = "refunded"
)

// TransactionType debit (money out of wallet) from credit (money into wallet)
type TransactionType string

const (
	TypeDebit  TransactionType = "debit"  //decrease
	TypeCredit TransactionType = "credit" //increase
)

// Transaction represents a single paymen
type Transaction struct {
	ID             string            `json:"id" db:"transaction_id"`
	UserID         string            `json:"user_id" db:"user_id"`
	Amount         float64           `json:"amount" db:"amount"`
	Type           TransactionType   `json:"type" db:"type"`
	Status         TransactionStatus `json:"status" db:"status"`
	Description    string            `json:"description,omitempty" db:"description"`
	IdempotencyKey string            `json:"idempotency_key" db:"idempotency_key"` // same as TransactionID for this service
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
}

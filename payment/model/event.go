package model

import "time"

// EventType labels what happened.
type EventType string

const (
	EventPaymentCreated    EventType = "payment.created"
	EventPaymentProcessing EventType = "payment.processing"
	EventPaymentCompleted  EventType = "payment.completed"
	EventPaymentFailed     EventType = "payment.failed"
	EventPaymentDuplicate  EventType = "payment.duplicate"
)

// Event records an immutable fact that happened in the system
// Implement event sourcing
type Event struct {
	ID            string    `json:"id" db:"id"`
	TransactionID string    `json:"transaction_id" db:"transaction_id"`
	Type          EventType `json:"type" db:"type"`
	Payload       string    `json:"payload" db:"payload"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

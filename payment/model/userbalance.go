package model

import "time"

// UserBalance tracks a user's current wallet balance.
type UserBalance struct {
	UserID    string    `json:"user_id" db:"user_id"`
	Balance   int64     `json:"balance" db:"balance"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

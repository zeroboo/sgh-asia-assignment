package handler

import "fmt"

// PayRequest pay request content.
type PayRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	UserID        string `json:"user_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency,omitempty"`
	Description   string `json:"description,omitempty"`
}

// PayResponse response for payrequest
type PayResponse struct {
	BaseResponse
	TransactionID string `json:"transaction_id"`
	UserID        string `json:"user_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	NewBalance    int64  `json:"new_balance"`
	Message       string `json:"message,omitempty"`
	ProcessedAt   string `json:"processed_at"`
}

type BaseResponse struct {
	Status       int
	ErrorMessage string `json:"error_msg"`
	ErrorCode    string `json:"error"`
}

var ErrInvalidRequest error = fmt.Errorf("Invalid request")

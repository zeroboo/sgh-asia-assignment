package test

import (
	"fmt"
	"net/http"
	"testing"

	"zeroboo.payment/handler"
)

// ---------------------------------------------------------------------------
// Integration Tests
// ---------------------------------------------------------------------------

// go test -timeout 30s -run ^TestAPI_InvalidRequest_ResponseError$ zeroboo.payment/test -v
func TestAPI_InvalidRequest_ResponseError(t *testing.T) {
	cleanTables(t)

	tests := []struct {
		name           string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name:           "empty request",
			payload:        map[string]any{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing transaction_id",
			payload:        map[string]any{"user_id": "user-1", "amount": 100},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing user_id",
			payload:        map[string]any{"transaction_id": "tx-1", "amount": 100},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing amount",
			payload:        map[string]any{"transaction_id": "tx-1", "user_id": "user-1"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "zero amount",
			payload:        map[string]any{"transaction_id": "tx-1", "user_id": "user-1", "amount": 0},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := postPay(t, tc.payload)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
			}

			var body handler.BaseResponse
			decodeBody(t, resp, &body)
			if body.Status != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, body.Status)
			}
		})
	}
}

// go test -timeout 30s -run ^TestAPI_UserNotExists_ResponseError$ zeroboo.payment/test -v
func TestAPI_UserNotExists_ResponseError(t *testing.T) {
	cleanTables(t)

	resp := postPay(t, handler.PayRequest{
		TransactionID: "tx-no-user-1",
		UserID:        "ghost-user",
		Amount:        100,
	})

	var body handler.BaseResponse
	decodeBody(t, resp, &body)
	if body.Status != http.StatusBadRequest {
		t.Errorf("expected status %d for missing user, got %d", http.StatusBadRequest, body.Status)
	}
}

// go test -timeout 30s -run ^TestAPI_InsufficientBalance_ResponseError$ zeroboo.payment/test -v
func TestAPI_InsufficientBalance_ResponseError(t *testing.T) {
	cleanTables(t)
	seedUserBalance(t, "user-poor", 50)

	resp := postPay(t, handler.PayRequest{
		TransactionID: "tx-poor-1",
		UserID:        "user-poor",
		Amount:        200,
	})

	var body handler.BaseResponse
	decodeBody(t, resp, &body)
	if body.Status != http.StatusBadRequest {
		t.Errorf("expected status %d for insufficient balance, got %d", http.StatusBadRequest, body.Status)
	}
}

// go test -timeout 30s -run ^TestAPI_InsufficientBalance_ResponseError$ zeroboo.payment/test -v
func TestAPI_ValidRequest_ResponseSuccess(t *testing.T) {
	cleanTables(t)
	seedUserBalance(t, "user-rich", 1000)

	resp := postPay(t, handler.PayRequest{
		TransactionID: "tx-ok-1",
		UserID:        "user-rich",
		Amount:        300,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}

	var body handler.PayResponse
	decodeBody(t, resp, &body)
	if body.Status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, body.Status)
	}

	// Verify balance was deducted in the database.
	var newBalance int64
	err := suiteDB.QueryRow("SELECT balance FROM user_balances WHERE user_id = ?", "user-rich").Scan(&newBalance)
	if err != nil {
		t.Fatalf("query balance: %v", err)
	}
	if newBalance != 700 {
		t.Errorf("expected balance 700 after payment, got %d", newBalance)
	}

	// Verify transaction record exists with status completed.
	var status string
	err = suiteDB.QueryRow("SELECT status FROM transactions WHERE transaction_id = ?", "tx-ok-1").Scan(&status)
	if err != nil {
		t.Fatalf("query transaction: %v", err)
	}
	if status != "completed" {
		t.Errorf("expected transaction status 'completed', got '%s'", status)
	}
}

// go test -timeout 30s -run ^TestAPI_DuplicateTransaction_ResponseError$ zeroboo.payment/test -v
func TestAPI_DuplicateTransaction_ResponseError(t *testing.T) {
	cleanTables(t)
	seedUserBalance(t, "user-dup", 1000)

	// First payment should succeed.
	req := handler.PayRequest{
		TransactionID: "tx-dup-1",
		UserID:        "user-dup",
		Amount:        100,
	}
	resp1 := postPay(t, req)
	var body1 handler.PayResponse
	decodeBody(t, resp1, &body1)
	if body1.Status != http.StatusOK {
		t.Fatalf("first payment: expected status %d, got %d", http.StatusOK, body1.Status)
	}

	// Same transaction ID again → should return the existing transaction.
	resp2 := postPay(t, req)
	var body2 handler.PayResponse
	decodeBody(t, resp2, &body2)
	if body2.Status != http.StatusOK {
		t.Errorf("duplicate payment: expected status %d, got %d", http.StatusOK, body2.Status)
	}
	if body2.TransactionID != "tx-dup-1" {
		t.Errorf("expected transaction_id tx-dup-1, got %s", body2.TransactionID)
	}

	// Balance should only have been deducted once (1000 - 100 = 900).
	var balance int64
	err := suiteDB.QueryRow("SELECT balance FROM user_balances WHERE user_id = ?", "user-dup").Scan(&balance)
	if err != nil {
		t.Fatalf("query balance: %v", err)
	}
	if balance != 900 {
		t.Errorf("expected balance 900 (deducted once), got %d", balance)
	}
}

// go test -timeout 30s -run ^TestAPI_MultiplePayments_CorrectBalance$ zeroboo.payment/test -v
func TestAPI_MultiplePayments_CorrectBalance(t *testing.T) {
	cleanTables(t)
	seedUserBalance(t, "user-multi", 1000)

	// Make 3 successive payments.
	amounts := []int64{200, 300, 100}
	for i, amount := range amounts {
		txID := fmt.Sprintf("tx-multi-%d", i+1)
		resp := postPay(t, handler.PayRequest{
			TransactionID: txID,
			UserID:        "user-multi",
			Amount:        amount,
		})
		var body handler.PayResponse
		decodeBody(t, resp, &body)
		if body.Status != http.StatusOK {
			t.Fatalf("payment %d: expected status %d, got %d", i+1, http.StatusOK, body.Status)
		}
	}

	// Balance: 1000 - 200 - 300 - 100 = 400
	var balance int64
	err := suiteDB.QueryRow("SELECT balance FROM user_balances WHERE user_id = ?", "user-multi").Scan(&balance)
	if err != nil {
		t.Fatalf("query balance: %v", err)
	}
	if balance != 400 {
		t.Errorf("expected balance 400, got %d", balance)
	}
}

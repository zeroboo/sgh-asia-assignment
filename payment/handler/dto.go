package handler

// PayRequestDTO pay request content.
type PayRequestDTO struct {
	TransactionID string  `json:"transactionID" binding:"required"`
	UserID        string  `json:"userID" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	Currency      string  `json:"currency,omitempty"`
	Description   string  `json:"description,omitempty"`
}

// PayResponseDTO response for payrequest
type PayResponseDTO struct {
	Status        string  `json:"status"`
	TransactionID string  `json:"transactionID"`
	UserID        string  `json:"userID"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	NewBalance    float64 `json:"newBalance"`
	Message       string  `json:"message,omitempty"`
	ProcessedAt   string  `json:"processedAt"`
}

// ErrorDTO is a standard error response.
type ErrorDTO struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

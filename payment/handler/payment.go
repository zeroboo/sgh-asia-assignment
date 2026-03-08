package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"zeroboo.payment/model"
)

type ITransactionRepository interface {
	GetTransaction(transactionId string) (*model.Transaction, error)
}

type IEventLogRepository interface {
}

type PaymentHandler struct {
	transactionRepo ITransactionRepository
	eventLogRepo    IEventLogRepository
}

func (handler *PaymentHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/pay", handler.handlePay)
}

func IsTransactionProcessing(transactionId string) bool {
	return false
}

// Accquired lock on a transaction id.
// returns nil if locking success, any error if locking failed
func LockTransaction(transactionId string) error {
	return nil
}

// Release lock on transaction id
// returns nil if unlocking success, any error if fail
func UnlockTransaction(transactionId string) error {
	return nil
}

func (h *PaymentHandler) handlePay(c *gin.Context) {
	//Parse request
	var req PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorDTO{
			Error:   "malformed request body",
			Code:    "INVALID_JSON",
			Details: err.Error(),
		})
		return
	}

	//Is transation existed?
	transaction, errLoad := h.transactionRepo.GetTransaction(req.TransactionID)
	if errLoad != nil {
		//Response 500
	}
	if transaction != nil {
		//Response current transaction status
		//
	}

	// Create new transaction
	errLock := LockTransaction(req.TransactionID)
	if errLock != nil {
		//Response 500
	}

	// Create new transaction

	// Write transaction

	// Write event log

	errUnlock := UnlockTransaction(req.TransactionID)
	if errUnlock != nil {
		//Log
	}

	// Response
	resp := &PayResponse{
		Status: "success",
	}
	c.JSON(http.StatusCreated, resp)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"zeroboo.payment/model"
)

type IPaymentService interface {
	GetUserBalance(userId string) (*model.UserBalance, error)
	GetPayment(transactionId string) (*model.Transaction, error)
	LockPayment(transactionId string) error
	UnlockPayment(transactionId string) error
	CreatePayment(userId string, transactionId string, amount int64) error
}

type PaymentHandler struct {
	paymentService IPaymentService
}

func (handler *PaymentHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/pay", handler.handlePay)
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
	transaction, errLoad := h.paymentService.GetPayment(req.TransactionID)
	if errLoad != nil {
		//Response 500
	}

	if transaction != nil {
		//Response current transaction status
		//
	}

	// Create a lock on transaction
	errLock := h.paymentService.LockPayment(req.TransactionID)
	defer h.paymentService.UnlockPayment(req.TransactionID)
	if errLock != nil {
		//Response 500
	}

	// Validate user has enough money
	userBalance, errBalance := h.paymentService.GetUserBalance(req.UserID)
	if errBalance != nil {
		//Response 500
	}
	if userBalance.Balance < req.Amount {
		//Response 400
	}

	// Create new transaction
	errPay := h.paymentService.CreatePayment(req.UserID, req.TransactionID, req.Amount)
	if errPay != nil {
		//Unlock payment
		//Response 500
	}

	// Response
	resp := &PayResponse{
		Status: "success",
	}
	c.JSON(http.StatusCreated, resp)
}

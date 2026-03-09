package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"zeroboo.payment/model"
)

type IPaymentService interface {
	GetUserBalance(userId string) (*model.UserBalance, error)
	GetPayment(transactionId string) (*model.Transaction, error)
	LockPayment(transactionId string, expiration time.Duration) error
	UnlockPayment(transactionId string) error
	CreatePayment(userId string, transactionId string, amount int64) error
}

type PaymentHandler struct {
	paymentService IPaymentService
	logger         *slog.Logger
}

func NewPaymentHandler(paymentService IPaymentService, logger *slog.Logger) *PaymentHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger.With("component", "payment-handler"),
	}
}

func (handler *PaymentHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/pay", handler.handlePay)
}

// TransactionLockSeconds Expiration of lock on transaction id
const TransactionLockSeconds time.Duration = 300

func (h *PaymentHandler) handlePay(c *gin.Context) {

	//Parse request
	var req PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid pay request", "error", err)
		c.JSON(http.StatusOK, &BaseResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: ErrInvalidRequest.Error(),
		})
		return
	}

	log := h.logger.With("transactionID", req.TransactionID, "userID", req.UserID, "amount", req.Amount)
	log.Info("processing payment", "amount", req.Amount)

	//Is transation existed?
	transaction, errLoad := h.paymentService.GetPayment(req.TransactionID)
	if errLoad != nil {
		log.Error("failed to get payment", "error", errLoad)
		//Response 500
		c.JSON(http.StatusOK, &BaseResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: "",
		})
		return
	}

	if transaction != nil {
		log.Info("transaction already exists", "status", transaction.Status)
		//Response current transaction status
		c.JSON(http.StatusOK, &PayResponse{
			BaseResponse:  BaseResponse{Status: http.StatusOK},
			TransactionID: transaction.ID,
			UserID:        transaction.UserID,
			Amount:        int64(transaction.Amount),
		})
		return
	}

	// Create a lock on transaction
	errLock := h.paymentService.LockPayment(req.TransactionID, TransactionLockSeconds*time.Second)
	defer h.paymentService.UnlockPayment(req.TransactionID)
	if errLock != nil {
		log.Error("failed to lock payment", "error", errLock)
		//Response 500
		c.JSON(http.StatusOK, &BaseResponse{
			Status: http.StatusInternalServerError,
		})
		return
	}
	log.Debug("transaction lock acquired")

	// Validate user has enough money
	userBalance, errBalance := h.paymentService.GetUserBalance(req.UserID)
	if errBalance != nil {
		log.Error("failed to get user balance", "error", errBalance)
		//Response 500
		c.JSON(http.StatusOK, &BaseResponse{
			Status: http.StatusInternalServerError,
		})
		return
	}

	//User balance not found
	if userBalance == nil {
		log.Error("user balance not exists")
		c.JSON(http.StatusOK, &BaseResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: "user not exists",
		})
		return
	}

	if userBalance.Balance < req.Amount {
		log.Warn("insufficient balance", "balance", userBalance.Balance, "amount", req.Amount)
		//Response 400
		c.JSON(http.StatusOK, &BaseResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: "insufficient balance",
		})
		return
	}

	// Create new transaction
	errPay := h.paymentService.CreatePayment(req.UserID, req.TransactionID, req.Amount)
	if errPay != nil {
		log.Error("failed to create payment", "error", errPay)
		//Response 500
		c.JSON(http.StatusOK, &BaseResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: "payment failed",
		})
		return
	}

	//Get user balance
	log.Info("payment created successfully")

	// Response
	c.JSON(http.StatusOK, &PayResponse{
		BaseResponse: BaseResponse{Status: http.StatusOK},
	})
}

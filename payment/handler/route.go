package handler

import "github.com/gin-gonic/gin"

func (handler *PaymentHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/pay", handler.handlePay)
}

type PaymentHandler struct {
}

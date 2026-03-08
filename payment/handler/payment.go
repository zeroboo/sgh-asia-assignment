package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) handlePay(c *gin.Context) {
	//Parse request
	var dto PayRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, ErrorDTO{
			Error:   "malformed request body",
			Code:    "INVALID_JSON",
			Details: err.Error(),
		})
		return
	}

	// Check idempotency

	// Get status of transaction

	// Process transaction

	// Response
	resp := &PayResponseDTO{
		Status: "success",
	}
	c.JSON(http.StatusCreated, resp)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type PaymentHandler struct {
	service service.PaymentService
}

func NewPaymentHandler(service service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Create(c *gin.Context) {
	orderID, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)

	payment, err := h.service.Create(c.Request.Context(), orderID, req, actorID, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Created(c, "Payment created successfully", payment)
}

func (h *PaymentHandler) ListByOrder(c *gin.Context) {
	orderID, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	payments, err := h.service.ListByOrder(c.Request.Context(), orderID)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Payments retrieved successfully", payments)
}

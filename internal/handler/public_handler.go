package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type PublicHandler struct {
	service service.PublicService
}

func NewPublicHandler(service service.PublicService) *PublicHandler {
	return &PublicHandler{service: service}
}

func (h *PublicHandler) GetInvoiceByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "Invoice token is required")
		return
	}

	invoice, err := h.service.GetInvoiceByToken(c.Request.Context(), token)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Invoice retrieved successfully", invoice)
}

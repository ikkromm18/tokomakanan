package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type AuditHandler struct {
	service service.AuditService
}

func NewAuditHandler(service service.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 {
		limit = 20
	}

	entityType := c.Query("entity_type")
	action := c.Query("action")

	logs, meta, err := h.service.List(c.Request.Context(), page, limit, entityType, action)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Paginated(c, "Audit logs retrieved successfully", logs, meta)
}

package handler

import (
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
	page, limit := parsePaginationParams(c)
	entityType := c.Query("entity_type")
	action := c.Query("action")

	logs, meta, err := h.service.List(c.Request.Context(), page, limit, entityType, action)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Paginated(c, "Audit logs retrieved successfully", logs, meta)
}

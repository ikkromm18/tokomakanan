package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetSummary(c *gin.Context) {
	role, _ := c.Get("role")
	userRole, _ := role.(string)

	summary, err := h.service.GetSummary(c.Request.Context(), userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Dashboard summary retrieved successfully", summary)
}

func (h *DashboardHandler) GetChart(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	chart, err := h.service.GetChart(c.Request.Context(), startDate, endDate)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Sales chart retrieved successfully", chart)
}

func (h *DashboardHandler) GetPOReminders(c *gin.Context) {
	role, _ := c.Get("role")
	userRole, _ := role.(string)

	reminders, err := h.service.GetPOReminders(c.Request.Context(), userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "PO reminders retrieved successfully", reminders)
}

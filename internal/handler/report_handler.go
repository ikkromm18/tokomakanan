package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type ReportHandler struct {
	service service.ReportService
}

func NewReportHandler(service service.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) GetSalesReport(c *gin.Context) {
	var filter dto.ReportFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.ValidationError(c, err)
		return
	}

	report, err := h.service.GetSalesReport(c.Request.Context(), filter)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Sales report retrieved successfully", report)
}

func (h *ReportHandler) GetProfitReport(c *gin.Context) {
	var filter dto.ReportFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.ValidationError(c, err)
		return
	}

	report, err := h.service.GetProfitReport(c.Request.Context(), filter)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Profit report retrieved successfully", report)
}

func (h *ReportHandler) ExportCSV(c *gin.Context) {
	var filter dto.ReportFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.ValidationError(c, err)
		return
	}

	data, filename, err := h.service.ExportCSV(c.Request.Context(), filter)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

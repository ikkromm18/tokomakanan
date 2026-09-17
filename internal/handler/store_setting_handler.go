package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type StoreSettingHandler struct {
	service service.StoreSettingService
}

func NewStoreSettingHandler(service service.StoreSettingService) *StoreSettingHandler {
	return &StoreSettingHandler{service: service}
}

func (h *StoreSettingHandler) Get(c *gin.Context) {
	setting, err := h.service.Get(c.Request.Context())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Store settings retrieved successfully", setting)
}

func (h *StoreSettingHandler) Update(c *gin.Context) {
	var req dto.UpdateStoreSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)

	setting, err := h.service.Update(c.Request.Context(), actorID, req, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Store settings updated successfully", setting)
}

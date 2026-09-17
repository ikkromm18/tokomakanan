package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := h.service.Login(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", res)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userInfo, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User profile retrieved successfully", userInfo)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID, req, c.ClientIP()); err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Password changed successfully", nil)
}

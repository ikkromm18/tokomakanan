package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	role := c.Query("role")
	search := c.Query("search")

	users, meta, err := h.service.List(c.Request.Context(), page, limit, role, search)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Paginated(c, "Users retrieved successfully", users, meta)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)

	user, err := h.service.Create(c.Request.Context(), req, actorID, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Created(c, "User created successfully", user)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User retrieved successfully", user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)

	user, err := h.service.Update(c.Request.Context(), id, req, actorID, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully", user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	actorID, _ := getUserID(c)

	if err := h.service.Delete(c.Request.Context(), id, actorID, c.ClientIP()); err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.NoContent(c)
}

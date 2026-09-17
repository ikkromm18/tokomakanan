package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type CustomerHandler struct {
	service service.CustomerService
}

func NewCustomerHandler(service service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) List(c *gin.Context) {
	page, limit := parsePaginationParams(c)
	search := c.Query("search")

	customers, meta, err := h.service.List(c.Request.Context(), page, limit, search)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Paginated(c, "Customers retrieved successfully", customers, meta)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)

	customer, err := h.service.Create(c.Request.Context(), req, actorID, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Created(c, "Customer created successfully", customer)
}

func (h *CustomerHandler) GetByID(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	customer, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Customer retrieved successfully", customer)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)

	customer, err := h.service.Update(c.Request.Context(), id, req, actorID, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Customer updated successfully", customer)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type OrderHandler struct {
	service service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) List(c *gin.Context) {
	page, limit := parsePaginationParams(c)
	status := c.Query("status")
	orderType := c.Query("order_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	role, _ := c.Get("role")
	userRole, _ := role.(string)

	orders, meta, err := h.service.List(c.Request.Context(), page, limit, status, orderType, startDate, endDate, userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Paginated(c, "Orders retrieved successfully", orders, meta)
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID, _ := getUserID(c)
	role, _ := c.Get("role")
	userRole, _ := role.(string)

	order, err := h.service.Create(c.Request.Context(), userID, userRole, req, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Created(c, "Order created successfully", order)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	role, _ := c.Get("role")
	userRole, _ := role.(string)

	order, err := h.service.GetByID(c.Request.Context(), id, userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Order retrieved successfully", order)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)

	if err := h.service.UpdateStatus(c.Request.Context(), id, req.Status, actorID, c.ClientIP()); err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Order status updated successfully", nil)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) List(c *gin.Context) {
	page, limit := parsePaginationParams(c)
	category := c.Query("category")
	search := c.Query("search")

	var isActivePtr *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		active := activeStr == "true" || activeStr == "1"
		isActivePtr = &active
	}

	role, _ := c.Get("role")
	userRole, _ := role.(string)

	products, meta, err := h.service.List(c.Request.Context(), page, limit, category, isActivePtr, search, userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Paginated(c, "Products retrieved successfully", products, meta)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)
	role, _ := c.Get("role")
	userRole, _ := role.(string)

	product, err := h.service.Create(c.Request.Context(), req, actorID, userRole, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Created(c, "Product created successfully", product)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	role, _ := c.Get("role")
	userRole, _ := role.(string)

	product, err := h.service.GetByID(c.Request.Context(), id, userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Product retrieved successfully", product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)
	role, _ := c.Get("role")
	userRole, _ := role.(string)

	product, err := h.service.Update(c.Request.Context(), id, req, actorID, userRole, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Product updated successfully", product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	actorID, _ := getUserID(c)

	if err := h.service.Delete(c.Request.Context(), id, actorID, c.ClientIP()); err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Product deleted successfully", nil)
}

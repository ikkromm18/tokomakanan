package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
)

type PackageHandler struct {
	service service.PackageService
}

func NewPackageHandler(service service.PackageService) *PackageHandler {
	return &PackageHandler{service: service}
}

func (h *PackageHandler) List(c *gin.Context) {
	page, limit := parsePaginationParams(c)
	search := c.Query("search")

	var isActivePtr *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		active := activeStr == "true" || activeStr == "1"
		isActivePtr = &active
	}

	role, _ := c.Get("role")
	userRole, _ := role.(string)

	packages, meta, err := h.service.List(c.Request.Context(), page, limit, isActivePtr, search, userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Paginated(c, "Packages retrieved successfully", packages, meta)
}

func (h *PackageHandler) Create(c *gin.Context) {
	var req dto.CreatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)
	role, _ := c.Get("role")
	userRole, _ := role.(string)

	pkg, err := h.service.Create(c.Request.Context(), req, actorID, userRole, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Created(c, "Package created successfully", pkg)
}

func (h *PackageHandler) GetByID(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid package ID")
		return
	}

	role, _ := c.Get("role")
	userRole, _ := role.(string)

	pkg, err := h.service.GetByID(c.Request.Context(), id, userRole)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Package retrieved successfully", pkg)
}

func (h *PackageHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid package ID")
		return
	}

	var req dto.UpdatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	actorID, _ := getUserID(c)
	role, _ := c.Get("role")
	userRole, _ := role.(string)

	pkg, err := h.service.Update(c.Request.Context(), id, req, actorID, userRole, c.ClientIP())
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Package updated successfully", pkg)
}

func (h *PackageHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid package ID")
		return
	}

	actorID, _ := getUserID(c)

	if err := h.service.Delete(c.Request.Context(), id, actorID, c.ClientIP()); err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Package deleted successfully", nil)
}

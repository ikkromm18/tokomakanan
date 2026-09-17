package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
)

func getUserID(c *gin.Context) (uint64, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case uint64:
		return v, true
	case int:
		return uint64(v), true
	case int64:
		return uint64(v), true
	case float64:
		return uint64(v), true
	default:
		return 0, false
	}
}

func parseIDParam(c *gin.Context, paramName string) (uint64, error) {
	return strconv.ParseUint(c.Param(paramName), 10, 64)
}

func parsePaginationParams(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = pagination.DefaultPage
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		limit = pagination.DefaultLimit
	}
	if limit > pagination.MaxLimit {
		limit = pagination.MaxLimit
	}

	return page, limit
}

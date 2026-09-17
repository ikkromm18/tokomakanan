package pagination

import (
	"math"

	"github.com/ikkromm18/tokomakanan/internal/dto"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

func FormatPagination(page, limit int, totalRows int64) dto.PaginationMeta {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))
	if totalPages == 0 && totalRows > 0 {
		totalPages = 1
	}

	return dto.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}
}

func GetOffset(page, limit int) int {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return (page - 1) * limit
}

func GetLimit(limit int) int {
	if limit < 1 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

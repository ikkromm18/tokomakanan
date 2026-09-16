package pagination

import (
	"math"

	"github.com/ikkromm18/tokomakanan/internal/dto"
)

func FormatPagination(page, limit int, totalRows int64) dto.PaginationMeta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
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
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return (page - 1) * limit
}

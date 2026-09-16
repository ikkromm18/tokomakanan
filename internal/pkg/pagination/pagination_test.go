package pagination_test

import (
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"github.com/stretchr/testify/assert"
)

func TestFormatPagination(t *testing.T) {
	t.Run("normal pagination", func(t *testing.T) {
		meta := pagination.FormatPagination(1, 20, 45)
		assert.Equal(t, 1, meta.Page)
		assert.Equal(t, 20, meta.Limit)
		assert.Equal(t, int64(45), meta.TotalRows)
		assert.Equal(t, 3, meta.TotalPages)
	})

	t.Run("default page and limit when invalid", func(t *testing.T) {
		meta := pagination.FormatPagination(0, 0, 0)
		assert.Equal(t, 1, meta.Page)
		assert.Equal(t, 20, meta.Limit)
		assert.Equal(t, int64(0), meta.TotalRows)
		assert.Equal(t, 0, meta.TotalPages)
	})

	t.Run("max limit cap at 100", func(t *testing.T) {
		meta := pagination.FormatPagination(1, 150, 250)
		assert.Equal(t, 100, meta.Limit)
		assert.Equal(t, 3, meta.TotalPages)
	})

	t.Run("single page with few rows", func(t *testing.T) {
		meta := pagination.FormatPagination(1, 20, 5)
		assert.Equal(t, 1, meta.TotalPages)
	})
}

func TestGetOffset(t *testing.T) {
	t.Run("page 1 returns 0", func(t *testing.T) {
		offset := pagination.GetOffset(1, 20)
		assert.Equal(t, 0, offset)
	})

	t.Run("page 2 returns 20", func(t *testing.T) {
		offset := pagination.GetOffset(2, 20)
		assert.Equal(t, 20, offset)
	})

	t.Run("defaults when non-positive inputs", func(t *testing.T) {
		offset := pagination.GetOffset(-1, -5)
		assert.Equal(t, 0, offset)
	})
}

package invoice_test

import (
	"testing"
	"time"

	invoicepkg "github.com/ikkromm18/tokomakanan/internal/pkg/invoice"
	"github.com/stretchr/testify/assert"
)

func TestFormatInvoiceNumber(t *testing.T) {
	date := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	invNo := invoicepkg.FormatInvoiceNumber(date, 1)
	assert.Equal(t, "INV/20260917/00001", invNo)

	invNo42 := invoicepkg.FormatInvoiceNumber(date, 42)
	assert.Equal(t, "INV/20260917/00042", invNo42)

	invNo99999 := invoicepkg.FormatInvoiceNumber(date, 99999)
	assert.Equal(t, "INV/20260917/99999", invNo99999)
}

package invoice

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// FormatInvoiceNumber formats invoice string: INV/YYYYMMDD/XXXXX
func FormatInvoiceNumber(date time.Time, counter int64) string {
	return fmt.Sprintf("INV/%s/%05d", date.Format("20060102"), counter)
}

// GenerateInvoiceNumber retrieves the next sequential number for today within tx/db and formats it
func GenerateInvoiceNumber(ctx context.Context, db *gorm.DB, date time.Time) (string, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var count int64
	err := db.WithContext(ctx).
		Table("orders").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&count).Error
	if err != nil {
		return "", fmt.Errorf("failed to generate invoice counter: %w", err)
	}

	return FormatInvoiceNumber(date, count+1), nil
}

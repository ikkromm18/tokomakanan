package repository_test

import (
	"context"
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestStoreSettingDB(t *testing.T) *gorm.DB {
	tx := setupTestDB(t)
	err := tx.AutoMigrate(&model.StoreSetting{})
	require.NoError(t, err)

	// Clean table for this test transaction
	err = tx.Exec("DELETE FROM store_settings").Error
	require.NoError(t, err)

	return tx
}

func stringPtr(s string) *string {
	return &s
}

func TestStoreSettingRepository_Get_Empty(t *testing.T) {
	tx := setupTestStoreSettingDB(t)
	repo := repository.NewStoreSettingRepository(tx)

	setting, err := repo.Get(context.Background())
	require.NoError(t, err)
	assert.Nil(t, setting)
}

func TestStoreSettingRepository_Get_Found(t *testing.T) {
	tx := setupTestStoreSettingDB(t)
	repo := repository.NewStoreSettingRepository(tx)

	seed := &model.StoreSetting{
		Name:          "Bakery Lezat",
		Address:       stringPtr("Jl. Mawar No. 12"),
		Phone:         stringPtr("081234567890"),
		LogoURL:       stringPtr("https://example.com/logo.png"),
		ReceiptFooter: stringPtr("Terima kasih atas kunjungan Anda"),
	}
	err := tx.Create(seed).Error
	require.NoError(t, err)

	setting, err := repo.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, setting)
	assert.Equal(t, seed.ID, setting.ID)
	assert.Equal(t, "Bakery Lezat", setting.Name)
	assert.Equal(t, "Jl. Mawar No. 12", *setting.Address)
	assert.Equal(t, "081234567890", *setting.Phone)
	assert.Equal(t, "https://example.com/logo.png", *setting.LogoURL)
	assert.Equal(t, "Terima kasih atas kunjungan Anda", *setting.ReceiptFooter)
}

func TestStoreSettingRepository_Update_InsertWhenEmpty(t *testing.T) {
	tx := setupTestStoreSettingDB(t)
	repo := repository.NewStoreSettingRepository(tx)

	newSetting := &model.StoreSetting{
		Name:    "Toko Makanan",
		Address: stringPtr("Jl. Baru"),
	}

	err := repo.Update(context.Background(), newSetting)
	require.NoError(t, err)
	assert.NotZero(t, newSetting.ID)

	// Verify in DB
	found, err := repo.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Toko Makanan", found.Name)
	assert.Equal(t, "Jl. Baru", *found.Address)
}

func TestStoreSettingRepository_Update_UpdateExisting(t *testing.T) {
	tx := setupTestStoreSettingDB(t)
	repo := repository.NewStoreSettingRepository(tx)

	seed := &model.StoreSetting{
		Name: "Original Name",
	}
	err := tx.Create(seed).Error
	require.NoError(t, err)

	seed.Name = "Updated Name"
	seed.Phone = stringPtr("0899999999")
	err = repo.Update(context.Background(), seed)
	require.NoError(t, err)

	found, err := repo.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "0899999999", *found.Phone)
}

package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestUserDB(t *testing.T) *gorm.DB {
	tx := setupTestDB(t)
	err := tx.AutoMigrate(&model.User{})
	require.NoError(t, err)
	return tx
}

func TestUserRepository_FindByEmail_Found(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	user := &model.User{
		Name:         "John Doe",
		Email:        email,
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	found, err := repo.FindByEmail(context.Background(), email)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, email, found.Email)
	assert.Equal(t, "John Doe", found.Name)
	assert.Equal(t, model.RoleAdmin, found.Role)
	assert.True(t, found.IsActive)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	found, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_FindByEmail_SoftDeletedIgnored(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	email := fmt.Sprintf("deleted_%d@example.com", time.Now().UnixNano())
	user := &model.User{
		Name:         "Deleted User",
		Email:        email,
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	// Soft delete user
	err = tx.Delete(user).Error
	require.NoError(t, err)

	found, err := repo.FindByEmail(context.Background(), email)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_FindByID_Found(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	email := fmt.Sprintf("findbyid_%d@example.com", time.Now().UnixNano())
	user := &model.User{
		Name:         "Jane Doe",
		Email:        email,
		PasswordHash: "hashedpassword",
		Role:         model.RoleOwner,
		IsActive:     true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Jane Doe", found.Name)
	assert.Equal(t, model.RoleOwner, found.Role)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	found, err := repo.FindByID(context.Background(), 99999999)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_FindByID_SoftDeletedIgnored(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	email := fmt.Sprintf("deletedbyid_%d@example.com", time.Now().UnixNano())
	user := &model.User{
		Name:         "Deleted User",
		Email:        email,
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	// Soft delete user
	err = tx.Delete(user).Error
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_Update_Success(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	email := fmt.Sprintf("update_%d@example.com", time.Now().UnixNano())
	user := &model.User{
		Name:         "Original Name",
		Email:        email,
		PasswordHash: "oldhash",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	// Update password hash and name
	user.Name = "Updated Name"
	user.PasswordHash = "newhash"
	err = repo.Update(context.Background(), user)
	require.NoError(t, err)

	// Reload from DB and verify
	var reloaded model.User
	err = tx.First(&reloaded, user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", reloaded.Name)
	assert.Equal(t, "newhash", reloaded.PasswordHash)
}

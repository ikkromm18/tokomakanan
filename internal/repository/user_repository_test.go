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

func TestUserRepository_Create_Success(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	email := fmt.Sprintf("created_%d@example.com", time.Now().UnixNano())
	newUser := &model.User{
		Name:         "Alice Admin",
		Email:        email,
		PasswordHash: "hashedsecret",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}

	err := repo.Create(context.Background(), newUser)
	require.NoError(t, err)
	assert.NotZero(t, newUser.ID)

	var fetched model.User
	err = tx.First(&fetched, newUser.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Alice Admin", fetched.Name)
	assert.Equal(t, email, fetched.Email)
	assert.Equal(t, model.RoleAdmin, fetched.Role)
	assert.True(t, fetched.IsActive)
}

func TestUserRepository_FindAll_PaginationAndFilters(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	ts := time.Now().UnixNano()
	user1 := &model.User{Name: "FilterAlpha Owner", Email: fmt.Sprintf("alpha_%d@test.com", ts), PasswordHash: "h", Role: model.RoleOwner, IsActive: true}
	user2 := &model.User{Name: "FilterBeta Admin", Email: fmt.Sprintf("beta_%d@test.com", ts), PasswordHash: "h", Role: model.RoleAdmin, IsActive: true}
	user3 := &model.User{Name: "FilterGamma Owner", Email: fmt.Sprintf("gamma_%d@test.com", ts), PasswordHash: "h", Role: model.RoleOwner, IsActive: true}

	require.NoError(t, tx.Create(user1).Error)
	require.NoError(t, tx.Create(user2).Error)
	require.NoError(t, tx.Create(user3).Error)

	// Test 1: Filter by role = owner
	owners, total, err := repo.FindAll(context.Background(), 1, 10, model.RoleOwner, "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(2))
	for _, u := range owners {
		assert.Equal(t, model.RoleOwner, u.Role)
	}

	// Test 2: Filter by search = "FilterBeta"
	searched, totalSearch, err := repo.FindAll(context.Background(), 1, 10, "", "FilterBeta")
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalSearch)
	require.Len(t, searched, 1)
	assert.Equal(t, user2.ID, searched[0].ID)

	// Test 3: Pagination with limit = 1
	page1, totalPage, err := repo.FindAll(context.Background(), 1, 1, "", fmt.Sprintf("%d@test.com", ts))
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalPage)
	assert.Len(t, page1, 1)

	page2, _, err := repo.FindAll(context.Background(), 2, 1, "", fmt.Sprintf("%d@test.com", ts))
	require.NoError(t, err)
	assert.Len(t, page2, 1)
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestUserRepository_Delete_SoftDelete(t *testing.T) {
	tx := setupTestUserDB(t)
	repo := repository.NewUserRepository(tx)

	email := fmt.Sprintf("delete_target_%d@example.com", time.Now().UnixNano())
	user := &model.User{
		Name:         "Delete Target",
		Email:        email,
		PasswordHash: "hashed",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}
	require.NoError(t, tx.Create(user).Error)

	err := repo.Delete(context.Background(), user.ID)
	require.NoError(t, err)

	// Normal FindByID should return nil
	found, err := repo.FindByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Nil(t, found)

	// But record still exists with DeletedAt populated
	var unscopedUser model.User
	err = tx.Unscoped().First(&unscopedUser, user.ID).Error
	require.NoError(t, err)
	assert.True(t, unscopedUser.DeletedAt.Valid)
}


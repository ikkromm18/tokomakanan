package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/database"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	cfg := &config.Config{
		AppEnv:     "development",
		DBHost:     "127.0.0.1",
		DBPort:     "3306",
		DBUser:     "root",
		DBPassword: "ikrom1214",
		DBName:     "tokomakanan_test",
	}

	db, err := database.NewMySQLConnection(cfg)
	if err != nil {
		cfg.DBName = "tokomakanan"
		db, err = database.NewMySQLConnection(cfg)
		require.NoError(t, err)
	}

	err = db.AutoMigrate(&model.AuditLog{})
	require.NoError(t, err)

	tx := db.Begin()
	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}

func TestAuditRepository_Create_AllFields(t *testing.T) {
	tx := setupTestDB(t)
	repo := repository.NewAuditRepository(tx)

	userID := uint64(1)
	entityID := uint64(42)
	oldVal := `{"name":"Old Bread","price":10000}`
	newVal := `{"name":"New Bread","price":12000}`
	ip := "192.168.1.100"

	log := &model.AuditLog{
		UserID:     &userID,
		Action:     "UPDATE",
		EntityType: "product",
		EntityID:   &entityID,
		OldValue:   &oldVal,
		NewValue:   &newVal,
		IPAddress:  &ip,
		CreatedAt:  time.Now(),
	}

	err := repo.Create(context.Background(), log)
	require.NoError(t, err)
	assert.NotZero(t, log.ID)
}

func TestAuditRepository_Create_NullableFields(t *testing.T) {
	tx := setupTestDB(t)
	repo := repository.NewAuditRepository(tx)

	log := &model.AuditLog{
		UserID:     nil,
		Action:     "SYSTEM_CLEANUP",
		EntityType: "system",
		EntityID:   nil,
		OldValue:   nil,
		NewValue:   nil,
		IPAddress:  nil,
		CreatedAt:  time.Now(),
	}

	err := repo.Create(context.Background(), log)
	require.NoError(t, err)
	assert.NotZero(t, log.ID)
}

func TestAuditRepository_FindAll_PaginationAndFilters(t *testing.T) {
	tx := setupTestDB(t)
	repo := repository.NewAuditRepository(tx)

	actionPrefix := "TEST_" + time.Now().Format("150405.000000")
	action1 := actionPrefix + "_CREATE"
	action2 := actionPrefix + "_UPDATE"

	userID := uint64(1)
	e1 := uint64(101)
	e2 := uint64(102)
	e3 := uint64(103)
	val := `{"test":true}`

	logsToCreate := []*model.AuditLog{
		{UserID: &userID, Action: action1, EntityType: "product", EntityID: &e1, NewValue: &val, CreatedAt: time.Now().Add(-3 * time.Minute)},
		{UserID: &userID, Action: action2, EntityType: "product", EntityID: &e2, NewValue: &val, CreatedAt: time.Now().Add(-2 * time.Minute)},
		{UserID: &userID, Action: action1, EntityType: "order", EntityID: &e3, NewValue: &val, CreatedAt: time.Now().Add(-1 * time.Minute)},
	}

	for _, l := range logsToCreate {
		err := repo.Create(context.Background(), l)
		require.NoError(t, err)
	}

	// 1. Filter by entityType "product"
	logs, total, err := repo.FindAll(context.Background(), 1, 10, "product", "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(2))
	for _, l := range logs {
		assert.Equal(t, "product", l.EntityType)
	}

	// 2. Filter by action
	logs, total, err = repo.FindAll(context.Background(), 1, 10, "", action1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	for _, l := range logs {
		assert.Equal(t, action1, l.Action)
	}

	// 3. Filter by both entityType and action
	logs, total, err = repo.FindAll(context.Background(), 1, 10, "order", action1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "order", logs[0].EntityType)
	assert.Equal(t, action1, logs[0].Action)

	// 4. Pagination (limit 1)
	logs, total, err = repo.FindAll(context.Background(), 1, 1, "", action1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, logs, 1)

	// Page 2
	logsPage2, total2, err := repo.FindAll(context.Background(), 2, 1, "", action1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total2)
	assert.Len(t, logsPage2, 1)
	assert.NotEqual(t, logs[0].ID, logsPage2[0].ID)

	// Check ordering (newest first)
	assert.True(t, logs[0].CreatedAt.After(logsPage2[0].CreatedAt) || logs[0].CreatedAt.Equal(logsPage2[0].CreatedAt))

	// 5. Default boundary fallback: page < 1 and limit < 1
	logsDefault, totalDefault, err := repo.FindAll(context.Background(), 0, 0, "", action1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalDefault)
	assert.Len(t, logsDefault, 2)

	// 6. Max limit clamp: limit > 100
	logsMax, totalMax, err := repo.FindAll(context.Background(), 1, 150, "", action1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalMax)
	assert.Len(t, logsMax, 2)
}

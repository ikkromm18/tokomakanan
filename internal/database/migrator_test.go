package database_test

import (
	"path/filepath"
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRunMigrations_And_Rollback(t *testing.T) {
	cfg := &config.Config{
		AppEnv:     "development",
		DBHost:     "127.0.0.1",
		DBPort:     "3306",
		DBUser:     "root",
		DBPassword: "ikrom1214",
		DBName:     "tokomakanan",
	}

	gormDB, err := database.NewMySQLConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, gormDB)

	sqlDB, err := gormDB.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)

	// Path to migrations relative to this test file
	migrationsDir, err := filepath.Abs("../../migrations")
	require.NoError(t, err)

	// First ensure clean state by rolling back if already migrated
	_ = database.RollbackMigrations(sqlDB, migrationsDir)

	// Run up migrations
	err = database.RunMigrations(sqlDB, migrationsDir)
	require.NoError(t, err)

	// Running up migrations again should be idempotent (ErrNoChange handled)
	err = database.RunMigrations(sqlDB, migrationsDir)
	assert.NoError(t, err)

	// Verify all 10 domain tables exist
	expectedTables := []string{
		"users",
		"store_settings",
		"customers",
		"products",
		"product_packages",
		"package_items",
		"orders",
		"order_items",
		"order_payments",
		"audit_logs",
	}

	for _, tbl := range expectedTables {
		var tableName string
		query := "SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_name = ?;"
		err := sqlDB.QueryRow(query, cfg.DBName, tbl).Scan(&tableName)
		assert.NoError(t, err, "table %s should exist in database", tbl)
		assert.Equal(t, tbl, tableName)
	}

	// Verify superadmin seed
	var email, role, passwordHash string
	var isActive int
	query := "SELECT email, role, password_hash, is_active FROM users WHERE email = ?;"
	err = sqlDB.QueryRow(query, "superadmin@tokomakanan.com").Scan(&email, &role, &passwordHash, &isActive)
	require.NoError(t, err)
	assert.Equal(t, "superadmin@tokomakanan.com", email)
	assert.Equal(t, "superadmin", role)
	assert.Equal(t, 1, isActive)

	// Verify password hash matches SuperAdmin123!
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("SuperAdmin123!"))
	assert.NoError(t, err, "password hash must match SuperAdmin123!")

	// Test rollback (down migrations)
	err = database.RollbackMigrations(sqlDB, migrationsDir)
	require.NoError(t, err)

	// Verify domain tables are dropped after rollback
	for _, tbl := range expectedTables {
		var tableName string
		query := "SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_name = ?;"
		err := sqlDB.QueryRow(query, cfg.DBName, tbl).Scan(&tableName)
		assert.Error(t, err, "table %s should not exist after rollback", tbl)
	}

	// Re-run up migrations to leave the database in migrated state
	err = database.RunMigrations(sqlDB, migrationsDir)
	require.NoError(t, err)

	// Verify superadmin exists again
	err = sqlDB.QueryRow(query, "superadmin@tokomakanan.com").Scan(&email, &role, &passwordHash, &isActive)
	assert.NoError(t, err)
}

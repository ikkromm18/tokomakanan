package database_test

import (
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMySQLConnection_Success(t *testing.T) {
	cfg := &config.Config{
		AppEnv:     "development",
		DBHost:     "127.0.0.1",
		DBPort:     "3306",
		DBUser:     "root",
		DBPassword: "ikrom1214",
		DBName:     "tokomakanan",
	}

	db, err := database.NewMySQLConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)

	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestNewMySQLConnection_ProductionMode(t *testing.T) {
	cfg := &config.Config{
		AppEnv:     "production",
		DBHost:     "127.0.0.1",
		DBPort:     "3306",
		DBUser:     "root",
		DBPassword: "ikrom1214",
		DBName:     "tokomakanan",
	}

	db, err := database.NewMySQLConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
}

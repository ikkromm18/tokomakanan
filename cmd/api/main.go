package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/database"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
)

func main() {
	migrateFlag := flag.Bool("migrate", false, "Run database schema migrations")
	rollbackFlag := flag.Bool("rollback", false, "Rollback database schema migrations")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	gormDB, err := database.NewMySQLConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to retrieve sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if *migrateFlag {
		log.Println("Running database migrations...")
		if err := database.RunMigrations(sqlDB, "migrations"); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Database migrations applied successfully.")
		return
	}

	if *rollbackFlag {
		log.Println("Rolling back database migrations...")
		if err := database.RollbackMigrations(sqlDB, "migrations"); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Database migrations rolled back successfully.")
		return
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Service is healthy", gin.H{
			"app": cfg.AppName,
			"env": cfg.AppEnv,
		})
	})

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

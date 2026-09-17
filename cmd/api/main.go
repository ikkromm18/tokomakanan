package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/database"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/ikkromm18/tokomakanan/internal/router"
	"github.com/ikkromm18/tokomakanan/internal/service"
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

	// 1. Initialize Repositories
	userRepo := repository.NewUserRepository(gormDB)
	storeRepo := repository.NewStoreSettingRepository(gormDB)
	auditRepo := repository.NewAuditRepository(gormDB)
	productRepo := repository.NewProductRepository(gormDB)
	packageRepo := repository.NewPackageRepository(gormDB)
	customerRepo := repository.NewCustomerRepository(gormDB)
	orderRepo := repository.NewOrderRepository(gormDB)
	paymentRepo := repository.NewPaymentRepository(gormDB)
	dashboardRepo := repository.NewDashboardRepository(gormDB)
	reportRepo := repository.NewReportRepository(gormDB)

	// 2. Initialize Services
	auditService := service.NewAuditService(auditRepo)
	authService := service.NewAuthService(userRepo, auditService, cfg)
	userService := service.NewUserService(userRepo, auditService, cfg)
	storeService := service.NewStoreSettingService(storeRepo, auditService)
	productService := service.NewProductService(productRepo, auditService)
	packageService := service.NewPackageService(packageRepo, productRepo, auditService)
	customerService := service.NewCustomerService(customerRepo, auditService)
	orderService := service.NewOrderService(orderRepo, productRepo, packageRepo, auditService)
	paymentService := service.NewPaymentService(paymentRepo, orderRepo, auditService)
	publicService := service.NewPublicService(orderRepo, storeRepo)
	dashboardService := service.NewDashboardService(dashboardRepo)
	reportService := service.NewReportService(reportRepo)

	// 3. Initialize Handlers
	handlers := &router.Handlers{
		Auth:         handler.NewAuthHandler(authService),
		User:         handler.NewUserHandler(userService),
		StoreSetting: handler.NewStoreSettingHandler(storeService),
		Audit:        handler.NewAuditHandler(auditService),
		Product:      handler.NewProductHandler(productService),
		Package:      handler.NewPackageHandler(packageService),
		Customer:     handler.NewCustomerHandler(customerService),
		Order:        handler.NewOrderHandler(orderService),
		Payment:      handler.NewPaymentHandler(paymentService),
		Public:       handler.NewPublicHandler(publicService),
		Dashboard:    handler.NewDashboardHandler(dashboardService),
		Report:       handler.NewReportHandler(reportService),
	}

	// 4. Setup Router
	r := router.SetupRouter(cfg, handlers)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/middleware"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	StoreSetting *handler.StoreSettingHandler
	Audit        *handler.AuditHandler
	Product      *handler.ProductHandler
	Package      *handler.PackageHandler
	Customer     *handler.CustomerHandler
	Order        *handler.OrderHandler
	Payment      *handler.PaymentHandler
	Public       *handler.PublicHandler
	Dashboard    *handler.DashboardHandler
	Report       *handler.ReportHandler
}

func SetupRouter(cfg *config.Config, h *Handlers) *gin.Engine {
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global Middlewares
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.RateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Service is healthy", gin.H{
			"app": cfg.AppName,
			"env": cfg.AppEnv,
		})
	})

	api := r.Group("/api/v1")

	// Public Routes
	if h.Public != nil {
		public := api.Group("/public")
		public.GET("/invoice/:token", h.Public.GetInvoiceByToken)
	}

	if h.Auth != nil {
		auth := api.Group("/auth")
		auth.POST("/login", h.Auth.Login)
	}

	// Protected Routes
	protected := api.Group("")
	protected.Use(middleware.Auth(cfg.JWTSecret))

	// Auth (Protected)
	if h.Auth != nil {
		protected.GET("/auth/me", h.Auth.Me)
		protected.PUT("/auth/change-password", h.Auth.ChangePassword)
	}

	// Users (Superadmin only)
	if h.User != nil {
		users := protected.Group("/users")
		users.Use(middleware.RequireRoles(model.RoleSuperadmin))
		users.GET("", h.User.List)
		users.POST("", h.User.Create)
		users.GET("/:id", h.User.GetByID)
		users.PUT("/:id", h.User.Update)
		users.DELETE("/:id", h.User.Delete)
	}

	// Store Settings
	if h.StoreSetting != nil {
		settings := protected.Group("/store-settings")
		settings.GET("", h.StoreSetting.Get)
		settings.PUT("", middleware.RequireRoles(model.RoleSuperadmin, model.RoleOwner), h.StoreSetting.Update)
	}

	// Audit Logs (Superadmin only)
	if h.Audit != nil {
		audits := protected.Group("/audit-logs")
		audits.Use(middleware.RequireRoles(model.RoleSuperadmin))
		audits.GET("", h.Audit.List)
	}

	// Products
	if h.Product != nil {
		products := protected.Group("/products")
		products.GET("", h.Product.List)
		products.GET("/:id", h.Product.GetByID)

		management := products.Group("")
		management.Use(middleware.RequireRoles(model.RoleSuperadmin, model.RoleOwner))
		management.POST("", h.Product.Create)
		management.PUT("/:id", h.Product.Update)
		management.DELETE("/:id", h.Product.Delete)
	}

	// Packages
	if h.Package != nil {
		packages := protected.Group("/packages")
		packages.GET("", h.Package.List)
		packages.GET("/:id", h.Package.GetByID)

		management := packages.Group("")
		management.Use(middleware.RequireRoles(model.RoleSuperadmin, model.RoleOwner))
		management.POST("", h.Package.Create)
		management.PUT("/:id", h.Package.Update)
		management.DELETE("/:id", h.Package.Delete)
	}

	// Customers
	if h.Customer != nil {
		customers := protected.Group("/customers")
		customers.GET("", h.Customer.List)
		customers.POST("", h.Customer.Create)
		customers.GET("/:id", h.Customer.GetByID)
		customers.PUT("/:id", h.Customer.Update)
	}

	// Orders & Payments
	if h.Order != nil {
		orders := protected.Group("/orders")
		orders.GET("", h.Order.List)
		orders.POST("", h.Order.Create)
		orders.GET("/:id", h.Order.GetByID)
		orders.PATCH("/:id/status", h.Order.UpdateStatus)

		if h.Payment != nil {
			orders.POST("/:id/payments", h.Payment.Create)
			orders.GET("/:id/payments", h.Payment.ListByOrder)
		}
	}

	// Dashboard
	if h.Dashboard != nil {
		dash := protected.Group("/dashboard")
		dash.GET("/summary", h.Dashboard.GetSummary)
		dash.GET("/chart", h.Dashboard.GetChart)
		dash.GET("/po-reminders", h.Dashboard.GetPOReminders)
	}

	// Reports (Superadmin & Owner only)
	if h.Report != nil {
		reports := protected.Group("/reports")
		reports.Use(middleware.RequireRoles(model.RoleSuperadmin, model.RoleOwner))
		reports.GET("/sales", h.Report.GetSalesReport)
		reports.GET("/profit", h.Report.GetProfitReport)
		reports.GET("/export", h.Report.ExportCSV)
	}

	return r
}

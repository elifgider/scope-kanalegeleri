package handler

import (
	"kanalegeleri/go-app/internal/config"
	"kanalegeleri/go-app/internal/service"

	"github.com/gin-gonic/gin"
)

// Load creates and configures the Gin engine with all application routes.
func Load(
	cfg config.Settings,
	catalog *service.CatalogService,
	orders *service.OrderService,
	auth *service.AuthService,
	uploads *service.UploadService,
) (*gin.Engine, error) {
	if cfg.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Static dosyalar
	r.Static("/static", cfg.Paths.StaticDir)
	r.Static("/uploads", cfg.Paths.UploadsDir)

	h, err := NewHTTPHandler(cfg, catalog, orders, auth, uploads)
	if err != nil {
		return nil, err
	}

	// Public
	r.GET("/", h.handleHomePage)
	r.GET("/admin/login", h.handleAdminLogin)
	r.POST("/admin/login", h.handleAdminLogin)

	// Admin (auth middleware ile korunuyor)
	admin := r.Group("/admin", h.requireAuth)
	{
		admin.GET("", h.handleAdminPage)
		admin.POST("/products", h.handleCreateProduct)
		admin.POST("/products/update", h.handleUpdateProduct)
		admin.POST("/products/delete", h.handleDeleteProduct)
		admin.POST("/uploads", h.handleUploadImage)
		admin.POST("/logout", h.handleLogout)
	}

	// API
	api := r.Group("/api")
	{
		api.GET("/products", h.handleProductsAPI)
		api.POST("/orders", h.handleOrdersAPI)
	}

	return r, nil
}

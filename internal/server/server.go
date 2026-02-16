package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joefazee/learning-go-shop/internal/config"
	"github.com/joefazee/learning-go-shop/internal/services"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Server struct {
	config *config.Config
	db     *gorm.DB
	logger *zerolog.Logger

	// Injected services
	authService    *services.AuthService
	productService *services.ProductService
	userService    *services.UserService
	uploadService  *services.UploadService
}

func New(
	cfg *config.Config,
	db *gorm.DB,
	logger *zerolog.Logger,
	authService *services.AuthService,
	productService *services.ProductService,
	userService *services.UserService,
	uploadService *services.UploadService,
) *Server {
	return &Server{
		config:         cfg,
		db:             db,
		logger:         logger,
		authService:    authService,
		productService: productService,
		userService:    userService,
		uploadService:  uploadService,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.New()

	// Marker log
	if s.logger != nil {
		s.logger.Info().Msg("routes loaded: /api/v1/* enabled")
	}

	// Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())

	// Static files for uploaded images
	router.Static("/uploads", "./uploads")

	// Health
	router.GET("/health", s.healthCheck)

	// Debug: list all routes
	router.GET("/__routes", func(c *gin.Context) {
		routes := router.Routes()
		list := make([]map[string]string, 0, len(routes))
		for _, r := range routes {
			list = append(list, map[string]string{
				"method": r.Method,
				"path":   r.Path,
			})
		}
		c.JSON(http.StatusOK, list)
	})

	// API v1
	api := router.Group("/api/v1")
	{
		// ===== AUTH (PUBLIC) =====
		auth := api.Group("/auth")
		{
			auth.POST("/register", s.register)
			auth.POST("/login", s.login)
			auth.POST("/refresh", s.refreshToken)
			auth.POST("/logout", s.logout)
		}

		// ===== PROTECTED =====
		protected := api.Group("/")
		protected.Use(s.authMiddleware())
		{
			// ---- USERS ----
			users := protected.Group("/users")
			{
				users.GET("/profile", s.getProfile)
				users.PUT("/profile", s.updateProfile)
			}

			// ---- CATEGORIES (ADMIN ONLY WRITE) ----
			categories := protected.Group("/categories")
			{
				categories.POST("", s.adminMiddleware(), s.createCategory)
				categories.PUT("/:id", s.adminMiddleware(), s.updateCategory)
				categories.DELETE("/:id", s.adminMiddleware(), s.deleteCategory)
			}

			// ---- PRODUCTS (ADMIN ONLY WRITE) ----
			products := protected.Group("/products")
			{
				products.POST("", s.adminMiddleware(), s.createProduct)
				products.PUT("/:id", s.adminMiddleware(), s.updateProduct)
				products.DELETE("/:id", s.adminMiddleware(), s.deleteProduct)

				// Upload product image
				products.POST("/:id/images", s.adminMiddleware(), s.uploadProductImage)
			}
		}

		// ===== PUBLIC READ =====
		api.GET("/categories", s.getCategories)
		api.GET("/products", s.getProducts)
		api.GET("/products/:id", s.getProduct)
	}

	// Custom 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "route not found",
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		})
	})

	return router
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

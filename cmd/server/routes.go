package main

import (
	"github.com/gin-gonic/gin"

	"github.com/ronaldocristover/app-monitoring/internal/config"
	"github.com/ronaldocristover/app-monitoring/internal/middleware"
)

func setupRoutes(router *gin.Engine, h *handlers, cfg *config.Config) {
	// Middleware
	router.Use(middleware.Recovery(nil))
	router.Use(middleware.Logger(nil))
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           int(cfg.CORS.MaxAge.Seconds()),
	}))
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit())

	// Public routes
	registerHealthRoutes(router, h)
	registerAuthRoutes(router, h)

	// Protected routes (JWT auth)
	protected := router.Group("")
	protected.Use(middleware.Auth(cfg.JWT.Secret))
	{
		protected.GET("/api/v1/auth/me", h.Auth.Me)
		registerDashboardRoutes(protected, h)
		registerUserRoutes(protected, h)
		registerAppRoutes(protected, h)
		registerEnvironmentRoutes(protected, h)
		registerServerRoutes(protected, h)
		registerServiceRoutes(protected, h)
	}
}

func registerHealthRoutes(router *gin.Engine, h *handlers) {
	router.GET("/health", h.Health.HealthCheck)
	router.GET("/health/status", h.Health.HealthStatusCheck)
	router.GET("/health/detailed", h.Health.HealthDetailedCheck)
	router.GET("/health/live", h.Health.HealthLiveCheck)
	router.GET("/health/ready", h.Health.HealthReadyCheck)
}

func registerAuthRoutes(router *gin.Engine, h *handlers) {
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.RefreshToken)
	}
}

func registerDashboardRoutes(protected *gin.RouterGroup, h *handlers) {
	protected.GET("/api/v1/dashboard", h.Dashboard.GetStats)
}

func registerUserRoutes(protected *gin.RouterGroup, h *handlers) {
	users := protected.Group("/api/v1/users")
	{
		users.GET("", h.User.List)
		users.GET("/:id", h.User.Get)
		users.PUT("/:id", h.User.Update)
		users.DELETE("/:id", h.User.Delete)
	}
}

func registerAppRoutes(protected *gin.RouterGroup, h *handlers) {
	apps := protected.Group("/api/v1/apps")
	{
		apps.POST("", h.App.Create)
		apps.POST("/full", h.App.CreateFull)
		apps.GET("", h.App.List)
		apps.GET("/:id", h.App.Get)
		apps.GET("/:id/detail", h.App.GetDetail)
		apps.PUT("/:id", h.App.Update)
		apps.PUT("/:id/full", h.App.UpdateFull)
		apps.DELETE("/:id", h.App.Delete)
	}
}

func registerEnvironmentRoutes(protected *gin.RouterGroup, h *handlers) {
	envs := protected.Group("/api/v1/environments")
	{
		envs.POST("", h.Environment.Create)
		envs.GET("", h.Environment.List)
		envs.GET("/:id", h.Environment.Get)
		envs.PUT("/:id", h.Environment.Update)
		envs.DELETE("/:id", h.Environment.Delete)
	}
}

func registerServerRoutes(protected *gin.RouterGroup, h *handlers) {
	servers := protected.Group("/api/v1/servers")
	{
		servers.POST("", h.Server.Create)
		servers.GET("", h.Server.List)
		servers.GET("/:id", h.Server.Get)
		servers.PUT("/:id", h.Server.Update)
		servers.DELETE("/:id", h.Server.Delete)
	}
}

func registerServiceRoutes(protected *gin.RouterGroup, h *handlers) {
	services := protected.Group("/api/v1/services")
	{
		services.POST("", h.Service.Create)
		services.GET("", h.Service.List)
		services.GET("/:id", h.Service.Get)
		services.PUT("/:id", h.Service.Update)
		services.DELETE("/:id", h.Service.Delete)
		services.POST("/:id/ping", h.Service.ManualPing)

		// Monitoring config
		services.GET("/:id/monitoring", h.MonitoringConfig.Get)
		services.PUT("/:id/monitoring", h.MonitoringConfig.Upsert)

		// Monitoring logs
		services.GET("/:id/logs", h.MonitoringLog.ListByService)

		// Deployments
		services.POST("/:id/deployments", h.Deployment.Create)
		services.GET("/:id/deployments", h.Deployment.ListByService)
	}

	// Standalone deployment routes
	deployments := protected.Group("/api/v1/deployments")
	{
		deployments.GET("/:id", h.Deployment.Get)
		deployments.PUT("/:id", h.Deployment.Update)
		deployments.DELETE("/:id", h.Deployment.Delete)
	}

	// Standalone backup routes
	backups := protected.Group("/api/v1/backups")
	{
		backups.GET("/:id", h.Backup.Get)
		backups.PUT("/:id", h.Backup.Update)
		backups.DELETE("/:id", h.Backup.Delete)
	}

	// Backup routes under services
	services.POST("/:id/backups", h.Backup.Create)
	services.GET("/:id/backups", h.Backup.ListByService)
}

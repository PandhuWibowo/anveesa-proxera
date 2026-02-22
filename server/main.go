package main

import (
	"log"
	"net/http"

	"github.com/anveesa/proxera/config"
	"github.com/anveesa/proxera/crypto"
	"github.com/anveesa/proxera/database"
	"github.com/anveesa/proxera/handlers"
	"github.com/anveesa/proxera/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.Load()

	// Initialize encryption
	crypto.SetKey(config.C.EncryptionKey)

	// Initialize database
	if err := database.Init(config.C.DatabasePath); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Set Gin mode
	if config.C.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(middleware.CORS(config.C.AllowOrigins))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "1.0.0"})
	})

	// WebSocket endpoint
	r.GET("/ws", handlers.HandleWS)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Servers
		servers := v1.Group("/servers")
		{
			servers.GET("", handlers.ListServers)
			servers.POST("", handlers.CreateServer)
			servers.GET("/:id", handlers.GetServer)
			servers.PUT("/:id", handlers.UpdateServer)
			servers.PATCH("/:id", handlers.PatchServer)
			servers.DELETE("/:id", handlers.DeleteServer)
			servers.GET("/:id/health", handlers.ServerHealth)
			servers.GET("/:id/metrics", handlers.ServerMetrics)
			servers.GET("/:id/config", handlers.GetServerConfig)
			servers.PUT("/:id/config", handlers.PutServerConfig)
			servers.POST("/:id/reload", handlers.ReloadServer)
			servers.GET("/:id/logs", handlers.StreamServerLogs)
		}

		// Routes
		routes := v1.Group("/routes")
		{
			routes.GET("", handlers.ListRoutes)
			routes.POST("", handlers.CreateRoute)
			routes.GET("/:id", handlers.GetRoute)
			routes.PUT("/:id", handlers.UpdateRoute)
			routes.PATCH("/:id", handlers.PatchRoute)
			routes.DELETE("/:id", handlers.DeleteRoute)
			routes.POST("/:id/toggle", handlers.ToggleRoute)
		}

		// Alerts
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", handlers.ListAlerts)
			alerts.POST("", handlers.CreateAlert)
			alerts.GET("/:id", handlers.GetAlert)
			alerts.PATCH("/:id", handlers.UpdateAlert)
			alerts.DELETE("/:id", handlers.DeleteAlert)
			alerts.POST("/:id/acknowledge", handlers.AcknowledgeAlert)
			alerts.POST("/:id/resolve", handlers.ResolveAlert)
			alerts.POST("/bulk/acknowledge", handlers.BulkAcknowledgeAlerts)
		}

		// Dashboard
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/stats", handlers.DashboardStats)
			dashboard.GET("/traffic", handlers.DashboardTraffic)
		}
	}

	addr := ":" + config.C.Port
	log.Printf("Proxera backend listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

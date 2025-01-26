package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lukelittle/claroz/claroz-backend/docs" // Import generated Swagger docs
	"github.com/lukelittle/claroz/claroz-backend/internal/api/handlers"
	"github.com/lukelittle/claroz/claroz-backend/internal/api/routes"
	"github.com/lukelittle/claroz/claroz-backend/internal/config"
	"github.com/lukelittle/claroz/claroz-backend/internal/utils"
)

// @title           Claroz API
// @version         1.0
// @description     A RESTful API for the Claroz social platform
// @termsOfService  https://claroz.com/terms

// @contact.name   API Support
// @contact.url    https://claroz.com/support
// @contact.email  lucas.little@claroz.com

// @license.name  Proprietary
// @license.url   https://claroz.com/license

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.NewConfig()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize database connection
	db, err := utils.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Setup routes
	routes.SetupRoutes(router, db)

	// Setup Swagger documentation
	router.GET("/swagger.json", handlers.ServeSwaggerJSON)
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", handlers.ServeSwaggerUI)

	// Redirect root to Swagger UI
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Start server
	serverAddr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s", serverAddr)
	log.Printf("Swagger documentation available at http://localhost%s/swagger", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

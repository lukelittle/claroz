package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lukelittle/claroz/claroz-backend/internal/api/handlers"
	"github.com/lukelittle/claroz/claroz-backend/internal/api/middleware"
	"github.com/lukelittle/claroz/claroz-backend/internal/config"
	"github.com/lukelittle/claroz/claroz-backend/internal/federation"
	"github.com/lukelittle/claroz/claroz-backend/internal/repository"
	"github.com/lukelittle/claroz/claroz-backend/internal/utils"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)

	// Initialize storage
	cfg := config.NewConfig()
	storage, err := utils.NewFileStorage(&cfg.Storage)
	if err != nil {
		panic(err)
	}

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo)
	authHandler := handlers.NewAuthHandler(userRepo)
	postHandler := handlers.NewPostHandler(postRepo, storage)
	atpClient, err := federation.NewATProtoClient(cfg.Federation.PDSHost)
	if err != nil {
		panic(err)
	}
	federationHandler := handlers.NewFederationHandler(userRepo, atpClient)

	// Serve static files for uploads
	router.Static("/uploads", cfg.Storage.LocalPath)

	// API routes group
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Federation routes
		federation := api.Group("/federation")
		{
			federation.GET("/resolve/:handle", federationHandler.ResolveRemoteProfile)
			federation.POST("/sync/:did", federationHandler.SyncRemoteProfile)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetCurrentUser)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
				users.GET("/:id/posts", postHandler.GetUserPosts)
			}

			// Post routes
			posts := protected.Group("/posts")
			{
				posts.POST("", postHandler.CreatePost)
				posts.GET("", postHandler.GetPosts)
				posts.GET("/:id", postHandler.GetPost)
				posts.DELETE("/:id", postHandler.DeletePost)
				posts.POST("/:id/comments", postHandler.AddComment)
				posts.POST("/:id/like", postHandler.LikePost)
				posts.DELETE("/:id/like", postHandler.UnlikePost)
			}

			// Follow routes
			users.POST("/:id/follow", postHandler.FollowUser)
			users.DELETE("/:id/follow", postHandler.UnfollowUser)
		}
	}
}

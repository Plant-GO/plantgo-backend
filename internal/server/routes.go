package server

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	_ "plantgo-backend/cmd/api/docs"
	"plantgo-backend/internal/database"
	"plantgo-backend/internal/modules/auth"
	"plantgo-backend/internal/modules/level"
	"plantgo-backend/internal/modules/level/infrastructure"
	"plantgo-backend/internal/modules/notification"
	notificationinfra "plantgo-backend/internal/modules/notification/infrastructure"
	"plantgo-backend/internal/modules/plant"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

	// Initialize services and handlers
	authService := auth.NewAuthService(database.NewGormDB())
	
	// Initialize repositories
	plantRepository := infrastructure.NewPlantRepository(database.NewGormDB())
	notificationRepository := notificationinfra.NewNotificationRepository(database.NewGormDB())
	
	// Initialize Firebase service
	firebaseService, err := notification.NewFirebaseService(notificationRepository)
	if err != nil {
		log.Printf("Failed to initialize Firebase service: %v", err)
		// Continue without Firebase service - notifications will still work but without push notifications
		firebaseService = nil
	}
	
	// Initialize services
	notificationService := notification.NewNotificationService(notificationRepository, firebaseService)
	scanService := plant.NewScanService(notificationService)
	
	// Initialize handlers
	plantHandler := level.NewPlantHandler(plantRepository, notificationService)
	notificationHandler := notification.NewNotificationHandler(notificationService)

	// API v1 routes
	api := r.Group("/api/v1")
	
	// Auth routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/guest", authService.GuestLoginHandler)
		authGroup.POST("/google", authService.GoogleLoginHandler)
		authGroup.POST("/google/callback", authService.GoogleCallbackHandler)
		authGroup.POST("/register", authService.RegisterHandler)
		authGroup.POST("/login", authService.LoginHandler)
		authGroup.GET("/profile", authService.GetProfileHandler)
	}

	// Plant/Level routes
	levelGroup := api.Group("/levels")
	{
		levelGroup.GET("/", plantHandler.GetAllLevels)
		levelGroup.GET("/:id", plantHandler.GetLevel)
		levelGroup.GET("/number/:number", plantHandler.GetLevelByNumber)
		levelGroup.POST("/complete", plantHandler.CompleteLevel)
		levelGroup.POST("/complete-by-number", plantHandler.CompleteLevelByNumber)
		levelGroup.GET("/user/:userId/progress", plantHandler.GetUserProgress)
		levelGroup.GET("/user/:userId/completed", plantHandler.GetCompletedLevels)
		levelGroup.GET("/user/:userId/reward", plantHandler.GetUserReward)
		levelGroup.GET("/details/:id", plantHandler.GetLevelDetails)
		levelGroup.GET("/game-data", plantHandler.GetGameData)
		levelGroup.POST("/", plantHandler.CreateLevel)
		levelGroup.PUT("/:id", plantHandler.UpdateLevel)
		levelGroup.DELETE("/:id", plantHandler.DeleteLevel)
	}

	// Plant scanning routes
	plantGroup := api.Group("/plants")
	{
		plantGroup.POST("/scan", scanService.ScanImageHandler)
	}

	// Notification routes
	notificationGroup := api.Group("/notifications")
	{
		notificationGroup.GET("/:userId", notificationHandler.GetUserNotifications)
		notificationGroup.GET("/:userId/unread", notificationHandler.GetUnreadNotifications)
		notificationGroup.GET("/:userId/unread/count", notificationHandler.GetUnreadCount)
		notificationGroup.PUT("/:id/read", notificationHandler.MarkAsRead)
		notificationGroup.PUT("/:userId/read-all", notificationHandler.MarkAllAsRead)
		notificationGroup.DELETE("/:id", notificationHandler.DeleteNotification)
		notificationGroup.POST("/fcm-token", notificationHandler.UpdateFCMToken)
		notificationGroup.GET("/:userId/preferences", notificationHandler.GetUserPreferences)
		notificationGroup.PUT("/:userId/preferences", notificationHandler.UpdateUserPreferences)
	}

	// Protected routes
	authorized := r.Group("/")
	// authorized.Use(AuthMiddleware()) // Uncomment when you have auth middleware
	{
		// User profile
		authorized.GET("/profile", authService.GetProfileHandler)

		// Game routes (user-facing)
		gameGroup := authorized.Group("/game")
		{
			gameGroup.GET("/data/:userId", plantHandler.GetGameData)
			gameGroup.GET("/level/:userId/:number", plantHandler.GetLevelDetails)
			gameGroup.GET("/progress/:userId", plantHandler.GetUserProgress)
			gameGroup.GET("/completed/:userId", plantHandler.GetCompletedLevels)
			gameGroup.GET("/rewards/:userId", plantHandler.GetUserReward)
			gameGroup.POST("/complete", plantHandler.CompleteLevel)
			gameGroup.POST("/complete-by-number", plantHandler.CompleteLevelByNumber)
		}

		// Level routes (general access)
		levelGroup := authorized.Group("/levels")
		{
			levelGroup.GET("/", plantHandler.GetAllLevels)
			levelGroup.GET("/:id", plantHandler.GetLevel)
			levelGroup.GET("/number/:number", plantHandler.GetLevelByNumber)
		}

		// Admin routes (level management)
		adminGroup := authorized.Group("/admin")
		// adminGroup.Use(AdminMiddleware()) // Add admin middleware when available
		{
			adminGroup.POST("/levels", plantHandler.CreateLevel)
			adminGroup.PUT("/levels/:id", plantHandler.UpdateLevel)
			adminGroup.DELETE("/levels/:id", plantHandler.DeleteLevel)
		}

		// Notification routes
		notificationGroup := authorized.Group("/notifications")
		{
			notificationGroup.GET("/:userId", notificationHandler.GetUserNotifications)
			notificationGroup.GET("/:userId/unread", notificationHandler.GetUnreadNotifications)
			notificationGroup.GET("/:userId/count", notificationHandler.GetUnreadCount)
			notificationGroup.PUT("/:id/read", notificationHandler.MarkAsRead)
			notificationGroup.PUT("/:userId/read-all", notificationHandler.MarkAllAsRead)
			notificationGroup.DELETE("/:id", notificationHandler.DeleteNotification)
			notificationGroup.POST("/fcm-token", notificationHandler.UpdateFCMToken)
			notificationGroup.GET("/:userId/preferences", notificationHandler.GetUserPreferences)
			notificationGroup.PUT("/:userId/preferences", notificationHandler.UpdateUserPreferences)
		}
	}

	// Health check for the plant service
	r.GET("/plant/health", plantHandler.HealthCheck)

	return r
}

// HelloWorldHandler godoc
// @Summary      Hello World
// @Description  Basic test route
// @Tags         Utility
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       / [get]
func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"
	c.JSON(http.StatusOK, resp)
}

// healthHandler godoc
// @Summary      Health Check
// @Description  Returns database and service health
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /health [get]
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
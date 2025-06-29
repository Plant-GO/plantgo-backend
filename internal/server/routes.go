package server

import (
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
	scanService := plant.NewScanService()
	
	// Initialize PlantHandler with repository
	plantRepository := infrastructure.NewPlantRepository(database.NewGormDB())
	plantHandler := level.NewPlantHandler(plantRepository)

	// Auth routes
	r.GET("/auth/google/login", authService.GoogleLoginHandler)
	r.GET("/auth/google/callback", authService.GoogleCallbackHandler)
	r.POST("/auth/guest/login", authService.GuestLoginHandler)
	r.POST("/auth/register", authService.RegisterHandler)
	r.POST("/auth/login", authService.LoginHandler)

	// Scan routes
	r.POST("/scan/image", scanService.ScanImageHandler)
	r.GET("/scan/video", scanService.ScanVideoHandler)

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
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

	authService := auth.NewAuthService(database.NewGormDB())

	r.GET("/auth/google/login", authService.GoogleLoginHandler)
	r.GET("/auth/google/callback", authService.GoogleCallbackHandler)
	r.POST("/auth/guest/login", authService.GuestLoginHandler)
	r.POST("/auth/register", authService.RegisterHandler)
	r.POST("/auth/login", authService.LoginHandler)

	authorized := r.Group("/")
	// authorized.Use(AuthMiddleware())
	{
		authorized.GET("/profile", authService.GetProfileHandler)
	}

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

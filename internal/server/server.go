package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"plantgo-backend/internal/database"
	_ "plantgo-backend/cmd/api/docs"
)

type Server struct {
	port int

	db database.Service
}

func NewServer() *Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	return &Server{
		port: port,
		db:   database.New(),
	}
}

func (s *Server) HttpServer() *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

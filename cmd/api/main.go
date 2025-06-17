// @title           PlantGo API
// @version         1.0
// @description     Backend API for the PlantGo system
// @termsOfService  http://wtfisterms.com/terms/

// @contact.name   Sushil Awasthi
// @contact.email  sushilawasthi2999@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"plantgo-backend/internal/server"
	_ "plantgo-backend/cmd/api/docs" 
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true
}

func main() {
	srv := server.NewServer()         
	apiServer := srv.HttpServer()     

	done := make(chan bool, 1)
	go gracefulShutdown(apiServer, done)

	err := apiServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	log.Println("Graceful shutdown complete.")
}


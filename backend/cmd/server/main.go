package main

import (
	"fmt"
	"log"
	"os"

	"suppliment/backend/internal/config"
	"suppliment/backend/internal/database"
	"suppliment/backend/internal/handlers"
	"suppliment/backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	if err := os.MkdirAll("./logs", 0o755); err != nil {
		log.Fatalf("failed to create logs dir: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	r.Use(middleware.RateLimiter(cfg))

	handlers.RegisterRoutes(r, db, cfg)

	log.Printf("Suppliment API server running on port %s", cfg.ServerPort)
	if err := r.Run(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

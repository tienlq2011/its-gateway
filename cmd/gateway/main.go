package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"its-gateway/internal/config"
	"its-gateway/internal/handlers"
	"its-gateway/internal/mq"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize RabbitMQ publisher
	rabbit, err := mq.NewPublisher(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer rabbit.Close()

	// Initialize handler with config
	h := &handlers.ITSHandler{
		Publisher: rabbit,
		Config:    cfg,
	}

	// Setup Gin router and endpoints
	r := gin.Default()
	r.POST("/", h.HandleITSEvent)
	r.POST("/notification/its-event", h.HandleITSEvent)
	r.POST("/internal/gate/:lane/open", h.OpenBarrier)

	// Start server
	log.Printf("Starting ITS Gateway on port %s...", cfg.Server.Port)
	if err := r.Run("0.0.0.0:" + cfg.Server.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

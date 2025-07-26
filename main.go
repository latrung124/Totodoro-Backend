/*
File: main.go
Author: trung.la
Date: 07/25/2025
Description: Main entry point for the BackEnd Monolith.
*/

package main

import (
	"log"

	"github.com/latrung124/Totodoro-Backend/internal/config"
	"github.com/latrung124/Totodoro-Backend/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.Load()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get configuration: %v", err)
	}

	// Initialize database connections
	connections, err := database.NewConnections(
		cfg.UserDBURL,
		cfg.PomodoroDBURL,
		cfg.StatisticDBURL,
		cfg.NotificationDBURL,
		cfg.TaskDBURL,
	)

	if err != nil {
		log.Fatalf("Failed to initialize database connections: %v", err)
	}
	defer connections.Close()

	port := cfg.Port
	router := gin.Default()

	router.Run(":" + port)
}

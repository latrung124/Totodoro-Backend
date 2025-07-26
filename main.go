/*
File: main.go
Author: trung.la
Date: 07/25/2025
Description: Main entry point for the BackEnd Monolith.
*/

package main

import (
	"github.com/latrung124/Totodoro-Backend/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()
	cfg, err := config.GetConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	port := cfg.Port
	router := gin.Default()

	router.Run(":" + port)
}

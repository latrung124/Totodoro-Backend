/*
File: config.go
Author: trung.la
Date: 07/26/2025
Description: Configuration management for the BackEnd Monolith.
*/

package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	UserDBURL         string
	PomodoroDBURL     string
	StatisticsDBURL   string
	NotificationDBURL string
	TaskDBURL         string
}

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	}
}

func GetConfig() (*Config, error) {
	return &Config{
		Port:              os.Getenv("PORT"),
		UserDBURL:         os.Getenv("USER_DB_URL"),
		PomodoroDBURL:     os.Getenv("POMODORO_DB_URL"),
		StatisticsDBURL:   os.Getenv("STATISTICS_DB_URL"),
		NotificationDBURL: os.Getenv("NOTIFICATION_DB_URL"),
		TaskDBURL:         os.Getenv("TASK_DB_URL"),
	}, nil
}

/*
File: testconfig.go
Author: trung.la
Date: 07/26/2025
Description: Test Configuration management for the BackEnd Monolith.
*/

package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func LoadTestConfig(envPath string) {
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	}
}

func GetTestConfig() (*Config, error) {
	return &Config{
		Port:              os.Getenv("TEST_PORT"),
		UserDBURL:         os.Getenv("TEST_USER_DB_URL"),
		PomodoroDBURL:     os.Getenv("TEST_POMODORO_DB_URL"),
		StatisticDBURL:    os.Getenv("TEST_STATISTIC_DB_URL"),
		NotificationDBURL: os.Getenv("TEST_NOTIFICATION_DB_URL"),
		TaskDBURL:         os.Getenv("TEST_TASK_DB_URL"),
	}, nil
}

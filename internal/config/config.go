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
	Port              string // Port for gateway
	Host              string
	UserDBURL         string
	PomodoroDBURL     string
	StatisticDBURL    string
	NotificationDBURL string
	TaskDBURL         string
	UserPort          string
	PomodoroPort      string
	StatisticPort     string
	NotificationPort  string
	TaskPort          string
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
		Host:              os.Getenv("HOST"),
		UserDBURL:         os.Getenv("USER_DB_URL"),
		PomodoroDBURL:     os.Getenv("POMODORO_DB_URL"),
		StatisticDBURL:    os.Getenv("STATISTIC_DB_URL"),
		NotificationDBURL: os.Getenv("NOTIFICATION_DB_URL"),
		TaskDBURL:         os.Getenv("TASK_DB_URL"),
		UserPort:          os.Getenv("USER_PORT"),
		PomodoroPort:      os.Getenv("POMODORO_PORT"),
		StatisticPort:     os.Getenv("STATISTIC_PORT"),
		NotificationPort:  os.Getenv("NOTIFICATION_PORT"),
		TaskPort:          os.Getenv("TASK_PORT"),
	}, nil
}

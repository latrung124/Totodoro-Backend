/*
File: main.go
Author: trung.la
Date: 07/25/2025
Description: Main entry point for the BackEnd Monolith.
*/

package main

import (
	"log"
	"net"

	"github.com/latrung124/Totodoro-Backend/internal/config"
	"github.com/latrung124/Totodoro-Backend/internal/database"
	"github.com/latrung124/Totodoro-Backend/internal/notification"
	"github.com/latrung124/Totodoro-Backend/internal/pomodoro"
	notificationpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/notification_service"
	pomodoropb "github.com/latrung124/Totodoro-Backend/internal/proto_package/pomodoro_service"
	statisticpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/statistic_service"
	taskmanagementpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/task_management_service"
	userpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
	"github.com/latrung124/Totodoro-Backend/internal/statistic"
	"github.com/latrung124/Totodoro-Backend/internal/task_management"
	"github.com/latrung124/Totodoro-Backend/internal/user"
	"google.golang.org/grpc"
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

	listen, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Port, err)
	}

	grpcServer := grpc.NewServer()
	userService := user.NewService(connections)
	userpb.RegisterUserServiceServer(grpcServer, userService)

	log.Printf("Starting gRPC server on %s", cfg.Port)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}

	pomodoroService := pomodoro.NewService(connections)
	pomodoropb.RegisterPomodoroServiceServer(grpcServer, pomodoroService)

	log.Printf("Pomodoro service registered successfully")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve Pomodoro service: %v", err)
	}

	statisticService := statistic.NewService(connections)
	statisticpb.RegisterStatisticServiceServer(grpcServer, statisticService)
	log.Printf("Statistic service registered successfully")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve Statistic service: %v", err)
	}

	taskmanagerService := task_management.NewService(connections)
	taskmanagementpb.RegisterTaskManagementServiceServer(grpcServer, taskmanagerService)
	log.Printf("Task Management service registered successfully")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve Task Management service: %v", err)
	}

	notificationService := notification.NewService(connections)
	notificationpb.RegisterNotificationServiceServer(grpcServer, notificationService)
	log.Printf("Notification service registered successfully")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve Notification service: %v", err)
	}
}

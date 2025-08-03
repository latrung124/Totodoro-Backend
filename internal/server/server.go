/*
File: internal/server/server.go
Author: trung.la
Date: 08/03/2025
Description: Server initialization and gRPC service registration.
*/

package server

import (
	"context"
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

type Server struct {
	grpcServer  *grpc.Server
	connections *database.Connections
}

func NewServer() *Server {
	s := &Server{
		grpcServer: grpc.NewServer(),
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	config.Load()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get configuration: %v", err)
		return err
	}

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

	s.connections = connections

	// Register services
	userService := user.NewService(connections)
	userpb.RegisterUserServiceServer(s.grpcServer, userService)
	log.Printf("User service registered successfully")

	pomodoroService := pomodoro.NewService(connections)
	pomodoropb.RegisterPomodoroServiceServer(s.grpcServer, pomodoroService)
	log.Printf("Pomodoro service registered successfully")

	statisticService := statistic.NewService(connections)
	statisticpb.RegisterStatisticServiceServer(s.grpcServer, statisticService)
	log.Printf("Statistic service registered successfully")

	taskmanagerService := task_management.NewService(connections)
	taskmanagementpb.RegisterTaskManagementServiceServer(s.grpcServer, taskmanagerService)
	log.Printf("Task Management service registered successfully")

	notificationService := notification.NewService(connections)
	notificationpb.RegisterNotificationServiceServer(s.grpcServer, notificationService)
	log.Printf("Notification service registered successfully")

	listenAddr := ":" + cfg.Port
	listen, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Port, err)
		return err
	}

	log.Printf("Starting gRPC server on %s", cfg.Port)
	if err := s.grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
		return err
	}

	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		log.Println("gRPC server stopped gracefully")
	}

	if s.connections != nil {
		s.connections.Close()
		log.Println("Database connections closed successfully")
	}
}

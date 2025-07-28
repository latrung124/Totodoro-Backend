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
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
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
	pb.RegisterUserServiceServer(grpcServer, userService)

	log.Printf("Starting gRPC server on %s", cfg.Port)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

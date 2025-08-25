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
	// one grpc.Server and listener per service
	servers     map[string]*grpc.Server
	listeners   map[string]net.Listener
	connections *database.Connections
}

func NewServer() *Server {
	return &Server{
		servers:   make(map[string]*grpc.Server),
		listeners: make(map[string]net.Listener),
	}
}

func (s *Server) Start(ctx context.Context) error {
	config.Load()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Printf("Failed to get configuration: %v", err)
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
		log.Printf("Failed to initialize database connections: %v", err)
		return err
	}
	s.connections = connections

	// Construct service implementations once (they can share DB connections)
	userService := user.NewService(connections)
	pomodoroService := pomodoro.NewService(connections)
	statisticService := statistic.NewService(connections)
	taskmanagerService := task_management.NewService(connections)
	notificationService := notification.NewService(connections)

	// Build listen addresses with host + port
	userAddr := net.JoinHostPort(cfg.Host, cfg.UserPort)
	pomodoroAddr := net.JoinHostPort(cfg.Host, cfg.PomodoroPort)
	statisticAddr := net.JoinHostPort(cfg.Host, cfg.StatisticPort)
	taskAddr := net.JoinHostPort(cfg.Host, cfg.TaskPort)
	notificationAddr := net.JoinHostPort(cfg.Host, cfg.NotificationPort)

	// Start each service on its own port (host-aware)
	if err := s.startService("user", userAddr, func(gs *grpc.Server) {
		userpb.RegisterUserServiceServer(gs, userService)
	}); err != nil {
		return err
	}

	if err := s.startService("pomodoro", pomodoroAddr, func(gs *grpc.Server) {
		pomodoropb.RegisterPomodoroServiceServer(gs, pomodoroService)
	}); err != nil {
		return err
	}

	if err := s.startService("statistic", statisticAddr, func(gs *grpc.Server) {
		statisticpb.RegisterStatisticServiceServer(gs, statisticService)
	}); err != nil {
		return err
	}

	if err := s.startService("task_management", taskAddr, func(gs *grpc.Server) {
		taskmanagementpb.RegisterTaskManagementServiceServer(gs, taskmanagerService)
	}); err != nil {
		return err
	}

	if err := s.startService("notification", notificationAddr, func(gs *grpc.Server) {
		notificationpb.RegisterNotificationServiceServer(gs, notificationService)
	}); err != nil {
		return err
	}

	// All services started asynchronously; return to caller.
	log.Printf("All gRPC services started: user:%s pomodoro:%s statistic:%s task:%s notification:%s",
		userAddr, pomodoroAddr, statisticAddr, taskAddr, notificationAddr)

	return nil
}

func (s *Server) startService(name, addr string, register func(*grpc.Server)) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("Failed to listen %s service on %s: %v", name, addr, err)
		return err
	}
	gs := grpc.NewServer()
	register(gs)

	s.servers[name] = gs
	s.listeners[name] = lis

	go func(n, a string, srv *grpc.Server, l net.Listener) {
		log.Printf("Starting %s gRPC service on %s", n, a)
		if err := srv.Serve(l); err != nil {
			// Serve returns non-nil on Stop/GracefulStop as well; just log it.
			log.Printf("%s gRPC service exited: %v", n, err)
		}
	}(name, addr, gs, lis)

	return nil
}

func (s *Server) Stop() {
	// Stop gRPC servers gracefully
	for name, srv := range s.servers {
		if srv != nil {
			log.Printf("Stopping %s gRPC service", name)
			srv.GracefulStop()
		}
	}
	// Close listeners
	for name, lis := range s.listeners {
		if lis != nil {
			_ = lis.Close()
			log.Printf("Closed listener for %s service", name)
		}
	}
	// Close shared DB connections
	if s.connections != nil {
		s.connections.Close()
		log.Println("Database connections closed successfully")
	}
}

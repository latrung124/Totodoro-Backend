/*
File: internal/api_gateway/api_gateway.go
Author: trung.la
Date: 08/24/2025
Description: API Gateway initialization and gRPC service registration.
*/

package api_gateway

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/api_gateway/handler"
	pomodoropb "github.com/latrung124/Totodoro-Backend/internal/proto_package/pomodoro_service"
	taskmanagementpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/task_management_service"
	userpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	// HTTP
	Mux    *http.ServeMux
	Server *http.Server

	// gRPC clients
	UserConn   *grpc.ClientConn
	UserClient userpb.UserServiceClient

	TaskManagementConn   *grpc.ClientConn
	TaskManagementClient taskmanagementpb.TaskManagementServiceClient

	PomodoroConn   *grpc.ClientConn
	PomodoroClient pomodoropb.PomodoroServiceClient
}

type Options struct {
	// HTTP listen address, e.g. ":8080"
	HTTPAddr string
	// gRPC addresses
	UserServiceAddr           string // e.g. "localhost:50051"
	TaskManagementServiceAddr string
	PomodoroServiceAddr       string
	// OIDC Client ID for authentication
	OIDCClientID string
}

func New(ctx context.Context, opt Options) (*Gateway, error) {
	// Create gRPC client connection (non-blocking; recommended)
	log.Printf("[ApiGateway] Connecting to UserService at %s", opt.UserServiceAddr)
	userConn, err := grpc.NewClient(
		opt.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		log.Printf("[ApiGateway] Failed to connect to UserService: %v", err)
		return nil, err
	}

	taskmanagementConn, err := grpc.NewClient(
		opt.TaskManagementServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		log.Printf("[ApiGateway] Failed to connect to TaskManagementService: %v", err)
		return nil, err
	}

	pomodoroConn, err := grpc.NewClient(
		opt.PomodoroServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		log.Printf("[ApiGateway] Failed to connect to PomodoroService: %v", err)
		return nil, err
	}

	gw := &Gateway{
		Mux:                  http.NewServeMux(),
		UserConn:             userConn,
		UserClient:           userpb.NewUserServiceClient(userConn),
		TaskManagementConn:   taskmanagementConn,
		TaskManagementClient: taskmanagementpb.NewTaskManagementServiceClient(taskmanagementConn),
		PomodoroConn:         pomodoroConn,
		PomodoroClient:       pomodoropb.NewPomodoroServiceClient(pomodoroConn),
	}

	// Register HTTP handlers
	uh := handler.NewUserHandler(gw.UserClient)
	handler.RegisterUserRoutes(gw.Mux, uh)

	tmh := handler.NewTaskManagementHandler(gw.TaskManagementClient)
	handler.RegisterTaskManagementRoutes(gw.Mux, tmh)

	ph := handler.NewPomodoroHandler(gw.PomodoroClient)
	handler.RegisterPomodoroRoutes(gw.Mux, ph)

	gw.Server = &http.Server{
		Addr:         opt.HTTPAddr,
		Handler:      gw.Mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("[ApiGateway] Listening on %s", opt.HTTPAddr)

	return gw, nil
}

func (g *Gateway) ListenAndServe() error {
	log.Printf("[ApiGateway] starting http server on %s", g.Server.Addr)
	return g.Server.ListenAndServe()
}

func (g *Gateway) Shutdown(ctx context.Context) error {
	log.Printf("[ApiGateway] shutting down http server")
	if g.Server != nil {
		_ = g.Server.Shutdown(ctx)
	}
	if g.UserConn != nil {
		log.Printf("[ApiGateway] closing user gRPC connection")
		_ = g.UserConn.Close()
	}
	return nil
}

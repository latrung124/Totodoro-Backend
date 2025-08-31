/*
File: main.go
Author: trung.la
Date: 07/25/2025
Description: Main entry point for the BackEnd Monolith.
*/

package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/api_gateway"
	"github.com/latrung124/Totodoro-Backend/internal/config"
	"github.com/latrung124/Totodoro-Backend/internal/config/google_config"
	"github.com/latrung124/Totodoro-Backend/internal/server"
)

func main() {
	// Root context cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load config once here
	config.Load()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	// Validate required ports
	if cfg.Port == "" {
		log.Fatal("HTTP gateway port (cfg.Port) is empty")
	}
	if cfg.UserPort == "" {
		log.Fatal("User gRPC port (cfg.UserPort) is empty")
	}

	httpAddr := net.JoinHostPort(cfg.Host, cfg.Port)
	userGRPCAddr := net.JoinHostPort(cfg.Host, cfg.UserPort)
	taskmanagementGRPCAddr := net.JoinHostPort(cfg.Host, cfg.TaskPort)
	pomodoroGRPCAddr := net.JoinHostPort(cfg.Host, cfg.PomodoroPort)

	// Start gRPC server(s)
	srv := server.NewServer()
	go func() {
		if err := srv.Start(ctx); err != nil {
			log.Printf("gRPC server exited: %v", err)
		}
	}()

	// Optionally wait for the User gRPC service to be ready before starting the gateway
	if err := waitForPort(ctx, userGRPCAddr, 10*time.Second); err != nil {
		log.Printf("warning: user gRPC not ready after timeout (%v); starting gateway anyway", err)
	}

	// Set Google Client Secret key for OIDC
	googleClientSecret, err := google_config.GetGoogleClientSecret()
	if err != nil {
		log.Fatalf("failed to load google client secret: %v", err)
	}

	// Start API Gateway (REST -> gRPC)
	gw, err := api_gateway.New(ctx, api_gateway.Options{
		HTTPAddr:                  httpAddr,
		UserServiceAddr:           userGRPCAddr,
		OIDCClientID:              googleClientSecret.ClientID,
		TaskManagementServiceAddr: taskmanagementGRPCAddr,
		PomodoroServiceAddr:       pomodoroGRPCAddr,
	})
	if err != nil {
		log.Fatalf("failed to init API gateway: %v", err)
	}
	go func() {
		if err := gw.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("api gateway exited: %v", err)
		}
	}()

	// Wait for signal then shutdown gracefully
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = gw.Shutdown(shutdownCtx)
	srv.Stop()
	log.Println("Shutdown complete")
}

// waitForPort attempts to connect to addr until it succeeds or the timeout/context cancels.
func waitForPort(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		d := time.Until(deadline)
		if d <= 0 {
			return context.DeadlineExceeded
		}
		dialer := net.Dialer{Timeout: 500 * time.Millisecond}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

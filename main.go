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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/api_gateway"
	"github.com/latrung124/Totodoro-Backend/internal/config"
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

	// Start gRPC server
	srv := server.NewServer()
	go func() {
		if err := srv.Start(ctx); err != nil {
			log.Printf("gRPC server exited: %v", err)
		}
	}()

	// Start API Gateway (REST -> gRPC)
	gw, err := api_gateway.New(ctx, api_gateway.Options{
		HTTPAddr:        ":8081", // adjust if you expose this in config
		UserServiceAddr: "localhost:" + cfg.Port,
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

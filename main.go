/*
File: main.go
Author: trung.la
Date: 07/25/2025
Description: Main entry point for the BackEnd Monolith.
*/

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/latrung124/Totodoro-Backend/internal/server"
)

func main() {
	srv := server.NewServer()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	srv.Stop()
	log.Println("Server shutting down gracefully")
}

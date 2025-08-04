/*
File: service_test.go
Author: trung.la
Date: 08-04-2025
Description: Test cases for user service functions.
*/

package user

import (
	"context"
	"testing"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
)

func TestCreateUser(t *testing.T) {
	// Mock database connections
	db := &database.Connections{
		UserDB: nil, // Replace with a mock or in-memory database
	}

	service := NewService(db)

	// Create a test request
	req := &pb.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
	}

	// Call the CreateUser method
	resp, err := service.CreateUser(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Validate the response
	if resp.User.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.User.Email)
	}
	if resp.User.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, resp.User.Username)
	}
	if resp.User.CreatedAt.AsTime().After(time.Now()) {
		t.Errorf("Invalid CreatedAt timestamp")
	}
}

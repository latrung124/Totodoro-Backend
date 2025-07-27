/*
File: service.go
Author: trung.la
Date: 07/27/2025
Description: This file contains the service layer for user management.
*/

package user

import (
	"context"
	"log"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service represents the user service implementation.
type Service struct {
	pb.UnimplementedUserServiceServer
	db *database.Connections
}

// NewService creates a new user service instance.
func NewService(db *database.Connections) *Service {
	return &Service{db: db}
}

// GetUser retrieves a user by user_id.
func (s *Service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Simulate database query (replace with actual SQL)
	var user pb.User
	err := s.db.UserDB.QueryRowContext(ctx, "SELECT user_id, email, username, created_at, updated_at FROM users WHERE user_id = $1", req.UserId).Scan(&user.UserId, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetUserResponse{User: &user}, nil
}

// CreateUser creates a new user.
func (s *Service) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.Email == "" || req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "email and username are required")
	}

	// Generate a unique user_id (e.g., UUID would be better in production)
	userId := "user_" + time.Now().Format("20060102150405")

	// Set current timestamp
	now := time.Now()
	newUser := &pb.User{
		UserId:    userId,
		Email:     req.Email,
		Username:  req.Username,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	// Simulate database insert (replace with actual SQL)
	_, err := s.db.UserDB.ExecContext(ctx, "INSERT INTO users (user_id, email, username, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		newUser.UserId, newUser.Email, newUser.Username, newUser.CreatedAt, newUser.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.CreateUserResponse{User: newUser}, nil
}

// UpdateUser updates an existing user's profile.
func (s *Service) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Fetch existing user
	var existingUser pb.User
	err := s.db.UserDB.QueryRowContext(ctx, "SELECT user_id, email, username, created_at, updated_at FROM users WHERE user_id = $1", req.UserId).Scan(&existingUser.UserId, &existingUser.Email, &existingUser.Username, &existingUser.CreatedAt, &existingUser.UpdatedAt)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Update fields if provided
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	existingUser.UpdatedAt = time.Now()

	// Simulate database update (replace with actual SQL)
	_, err = s.db.UserDB.ExecContext(ctx, "UPDATE users SET email = $1, username = $2, updated_at = $3 WHERE user_id = $4",
		existingUser.Email, existingUser.Username, existingUser.UpdatedAt, existingUser.UserId)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &pb.UpdateUserResponse{User: &existingUser}, nil
}

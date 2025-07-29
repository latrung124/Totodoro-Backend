/*
File: service.go
Author: trung.la
Date: 07/29/2025
Package: github.com/latrung124/Totodoro-Backend/internal/task_management
Description: This file contains the service layer for task management.
*/

package task_management

import (
	"context"
	"log"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/task_management_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedTaskManagementServiceServer
	db *database.Connections
}

func NewService(db *database.Connections) *Service {
	return &Service{db: db}
}

// CreateTask creates a new task for a user.
func (s *Service) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.GroupId == "" {
		return nil, status.Error(codes.InvalidArgument, "group_id is required")
	}

	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Simulate database insert (replace with actual SQL)
	taskId := "task_" + time.Now().Format("20060102150405")
	now := time.Now()
	newTask := &pb.Task{
		TaskId:      taskId,
		UserId:      req.UserId,
		GroupId:     req.GroupId,
		Description: req.Description,
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	_, err := s.db.TaskDB.ExecContext(ctx, "INSERT INTO tasks (task_id, user_id, group_id, title, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)", newTask.TaskId, newTask.UserId, newTask.GroupId, req.Title, req.Description, now, now)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	return &pb.CreateTaskResponse{Task: newTask}, nil
}

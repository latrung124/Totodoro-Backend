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

func (s *Service) GetTasks(ctx context.Context, req *pb.GetTasksRequest) (*pb.GetTasksResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	rows, err := s.db.TaskDB.QueryContext(ctx, "SELECT task_id, group_id, title, description, created_at, updated_at FROM tasks WHERE user_id = $1", req.UserId)
	if err != nil {
		log.Printf("Error fetching tasks: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch tasks")
	}
	defer rows.Close()

	var tasks []*pb.Task
	for rows.Next() {
		var task pb.Task
		task.UserId = req.UserId
		if err := rows.Scan(&task.TaskId, &task.GroupId, &task.Title, &task.Description, &task.CreatedAt, &task.UpdatedAt); err != nil {
			log.Printf("Error scanning task: %v", err)
			return nil, status.Error(codes.Internal, "failed to scan task")
		}
		tasks = append(tasks, &task)
	}

	return &pb.GetTasksResponse{Tasks: tasks}, nil
}

func (s *Service) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	// update the task in the database
	_, err := s.db.TaskDB.ExecContext(ctx, "UPDATE tasks SET title = $1, description = $2, deadline = $3, isCompleted = $4 WHERE task_id = $5", req.Title, req.Description, req.Deadline, req.IsCompleted, req.TaskId)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		return nil, status.Error(codes.Internal, "failed to update task")
	}

	// Assuming the task is updated successfully, we can return the updated task
	// Query the updated task from the database
	var updatedTask pb.Task
	err = s.db.TaskDB.QueryRowContext(ctx, "SELECT task_id, user_id, group_id, title, description, created_at, updated_at FROM tasks WHERE task_id = $1", req.TaskId).Scan(
		&updatedTask.TaskId,
		&updatedTask.UserId,
		&updatedTask.GroupId,
		&updatedTask.Title,
		&updatedTask.Description,
		&updatedTask.CreatedAt,
		&updatedTask.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error fetching updated task: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch updated task")
	}
	if updatedTask.TaskId == "" {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	return &pb.UpdateTaskResponse{Task: &updatedTask}, nil
}

func (s *Service) CreateTaskGroup(ctx context.Context, req *pb.CreateTaskGroupRequest) (*pb.CreateTaskGroupResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	groupId := "group_" + time.Now().Format("20060102150405")
	now := time.Now()
	newGroup := &pb.TaskGroup{
		GroupId:     groupId,
		UserId:      req.UserId,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	_, err := s.db.TaskDB.ExecContext(ctx, "INSERT INTO task_groups (group_id, user_id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)", newGroup.GroupId, newGroup.UserId, newGroup.Name, newGroup.Description, now, now)
	if err != nil {
		log.Printf("Error creating task group: %v", err)
		return nil, status.Error(codes.Internal, "failed to create task group")
	}

	return &pb.CreateTaskGroupResponse{Group: newGroup}, nil
}

func (s *Service) GetTaskGroups(ctx context.Context, req *pb.GetTaskGroupsRequest) (*pb.GetTaskGroupsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	rows, err := s.db.TaskDB.QueryContext(ctx, "SELECT group_id, name, description, created_at, updated_at FROM task_groups WHERE user_id = $1", req.UserId)
	if err != nil {
		log.Printf("Error fetching task groups: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch task groups")
	}
	defer rows.Close()

	var groups []*pb.TaskGroup
	for rows.Next() {
		var group pb.TaskGroup
		group.UserId = req.UserId
		if err := rows.Scan(&group.GroupId, &group.Name, &group.Description, &group.CreatedAt, &group.UpdatedAt); err != nil {
			log.Printf("Error scanning task group: %v", err)
			return nil, status.Error(codes.Internal, "failed to scan task group")
		}
		groups = append(groups, &group)
	}

	return &pb.GetTaskGroupsResponse{Groups: groups}, nil
}

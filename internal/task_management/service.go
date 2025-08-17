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
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
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

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	if req.TotalPomodoros <= 0 {
		return nil, status.Error(codes.InvalidArgument, "total_pomodoros must be greater than 0")
	}

	taskId := uuid.NewString()
	now := time.Now()

	var deadlineVal any
	if req.Deadline != nil {
		deadlineVal = req.Deadline.AsTime()
	} else {
		deadlineVal = nil // Use NULL for deadline if not provided
	}

	newTask := &pb.Task{
		TaskId:             taskId,
		UserId:             req.UserId,
		GroupId:            req.GroupId,
		Name:               req.Name,
		Description:        req.Description,
		CompletedPomodoros: 0,
		TotalPomodoros:     req.TotalPomodoros,
		Progress:           0,
		Priority:           req.Priority,
		Status:             req.Status,
		Deadline:           req.Deadline,
		CreatedAt:          timestamppb.New(now),
		UpdatedAt:          timestamppb.New(now),
	}

	_, err := s.db.TaskDB.ExecContext(
		ctx,
		`INSERT INTO tasks (
            task_id, user_id, group_id, name, description,
            priority, status, total_pomodoros, completed_pomodoros, progress,
            deadline, created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		newTask.TaskId,
		newTask.UserId,
		newTask.GroupId,
		newTask.Name,
		newTask.Description,
		newTask.Priority, // enum stored as int
		newTask.Status,   // enum stored as int
		newTask.TotalPomodoros,
		newTask.CompletedPomodoros,
		newTask.Progress,
		deadlineVal, // nil/NULL or time.Time
		now,
		now,
	)

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

	rows, err := s.db.TaskDB.QueryContext(ctx, `
        SELECT
            task_id, group_id, name, description,
            priority, status, total_pomodoros, completed_pomodoros, progress,
            deadline, created_at, updated_at
        FROM tasks
        WHERE user_id = $1
    `, req.UserId)
	if err != nil {
		log.Printf("Error fetching tasks: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch tasks")
	}
	defer rows.Close()

	var tasks []*pb.Task
	for rows.Next() {
		var (
			task                 pb.Task
			priorityInt          int32
			statusInt            int32
			totalPomodoros       int32
			completedPomodoros   int32
			progress             int32
			deadlineNT           sql.NullTime
			createdAt, updatedAt time.Time
		)

		if err := rows.Scan(
			&task.TaskId,
			&task.GroupId,
			&task.Name,
			&task.Description,
			&priorityInt,
			&statusInt,
			&totalPomodoros,
			&completedPomodoros,
			&progress,
			&deadlineNT,
			&createdAt,
			&updatedAt,
		); err != nil {
			log.Printf("Error scanning task: %v", err)
			return nil, status.Error(codes.Internal, "failed to scan task")
		}

		task.UserId = req.UserId
		task.Priority = pb.TaskPriority(priorityInt)
		task.Status = pb.TaskStatus(statusInt)
		task.TotalPomodoros = totalPomodoros
		task.CompletedPomodoros = completedPomodoros
		task.Progress = progress
		if deadlineNT.Valid {
			task.Deadline = timestamppb.New(deadlineNT.Time)
		} else {
			task.Deadline = nil
		}
		task.CreatedAt = timestamppb.New(createdAt)
		task.UpdatedAt = timestamppb.New(updatedAt)

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch tasks")
	}

	return &pb.GetTasksResponse{Tasks: tasks}, nil
}

func (s *Service) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	// update the task in the database
	_, err := s.db.TaskDB.ExecContext(ctx, "UPDATE tasks SET title = $1, description = $2, deadline = $3 WHERE task_id = $4", req.Name, req.Description, req.Deadline, req.TaskId)
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
		&updatedTask.Name,
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

	now := time.Now()
	newGroup := &pb.TaskGroup{
		GroupId:     uuid.NewString(),
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
		var (
			group     pb.TaskGroup
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&group.GroupId, &group.Name, &group.Description, &createdAt, &updatedAt); err != nil {
			log.Printf("Error scanning task group: %v", err)
			return nil, status.Error(codes.Internal, "failed to scan task group")
		}
		group.CreatedAt = timestamppb.New(createdAt)
		group.UpdatedAt = timestamppb.New(updatedAt)
		groups = append(groups, &group)
	}

	return &pb.GetTaskGroupsResponse{Groups: groups}, nil
}

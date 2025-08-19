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

func toDBTaskGroupPriority(p pb.TaskGroupPriority) string {
	switch p {
	case pb.TaskGroupPriority_TASK_GROUP_PRIORITY_LOW:
		return "low"
	case pb.TaskGroupPriority_TASK_GROUP_PRIORITY_MEDIUM:
		return "medium"
	case pb.TaskGroupPriority_TASK_GROUP_PRIORITY_HIGH:
		return "high"
	default:
		return "medium"
	}
}

func toDBTaskGroupStatus(s pb.TaskGroupStatus) string {
	switch s {
	case pb.TaskGroupStatus_TASK_GROUP_STATUS_IDLE:
		return "idle"
	case pb.TaskGroupStatus_TASK_GROUP_STATUS_COMPLETED:
		return "completed"
	case pb.TaskGroupStatus_TASK_GROUP_STATUS_PENDING:
		return "pending"
	case pb.TaskGroupStatus_TASK_GROUP_STATUS_IN_PROGRESS:
		return "in progress"
	default:
		return "idle"
	}
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

	now := time.Now()

	// Handle optional deadline
	var deadlineVal any
	if req.Deadline != nil {
		deadlineVal = req.Deadline.AsTime()
	} else {
		deadlineVal = nil // store NULL when not provided
	}

	// Update task using the correct columns
	res, err := s.db.TaskDB.ExecContext(ctx, `
        UPDATE tasks SET
            name = $1,
            description = $2,
            priority = $3,
            status = $4,
            total_pomodoros = $5,
            completed_pomodoros = $6,
            progress = $7,
            deadline = $8,
            updated_at = $9
        WHERE task_id = $10
    `,
		req.Name,
		req.Description,
		int32(req.Priority),
		int32(req.Status),
		req.TotalPomodoros,
		req.CompletedPomodoros,
		req.Progress,
		deadlineVal,
		now,
		req.TaskId,
	)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		return nil, status.Error(codes.Internal, "failed to update task")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	// Fetch and return the updated task (convert DB types to protobuf types)
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
	err = s.db.TaskDB.QueryRowContext(ctx, `
        SELECT
            task_id, user_id, group_id, name, description,
            priority, status, total_pomodoros, completed_pomodoros, progress,
            deadline, created_at, updated_at
        FROM tasks
        WHERE task_id = $1
    `, req.TaskId).Scan(
		&task.TaskId,
		&task.UserId,
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
	)
	if err != nil {
		log.Printf("Error fetching updated task: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch updated task")
	}

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

	return &pb.UpdateTaskResponse{Task: &task}, nil
}

func (s *Service) CreateTaskGroup(ctx context.Context, req *pb.CreateTaskGroupRequest) (*pb.CreateTaskGroupResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	now := time.Now()
	groupID := uuid.NewString()

	var deadlineVal any
	if req.Deadline != nil {
		deadlineVal = req.Deadline.AsTime()
	} else {
		deadlineVal = nil
	}

	priorityLabel := toDBTaskGroupPriority(req.Priority)
	statusLabel := toDBTaskGroupStatus(req.Status)

	newGroup := &pb.TaskGroup{
		GroupId:        groupID,
		UserId:         req.UserId,
		Icon:           req.Icon,
		Name:           req.Name,
		Description:    req.Description,
		Deadline:       req.Deadline,
		Priority:       req.Priority,
		Status:         req.Status,
		CompletedTasks: 0,
		TotalTasks:     0,
		CreatedAt:      timestamppb.New(now),
		UpdatedAt:      timestamppb.New(now),
	}

	_, err := s.db.TaskDB.ExecContext(
		ctx,
		`INSERT INTO task_groups (
            group_id, user_id, icon, name, description, deadline,
            priority, status, completed_tasks, total_tasks,
            created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		newGroup.GroupId,
		newGroup.UserId,
		newGroup.Icon,
		newGroup.Name,
		newGroup.Description,
		deadlineVal,
		priorityLabel,
		statusLabel,
		newGroup.CompletedTasks,
		newGroup.TotalTasks,
		now,
		now,
	)

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

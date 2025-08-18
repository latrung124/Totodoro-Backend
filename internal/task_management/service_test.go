/*
File: internal/task_management/service_test.go
Author: trung.la
Date: 08-13-2025
Description: Test cases for task management service functions.
*/

package task_management

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/latrung124/Totodoro-Backend/internal/config"
	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/task_management_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupTestDB() (*database.Connections, error) {
	envPath := "../../.env"
	config.LoadTestConfig(envPath)
	testCfg, err := config.GetTestConfig()
	if err != nil {
		return nil, err
	}

	connections, err := database.NewConnections(
		testCfg.UserDBURL,
		testCfg.PomodoroDBURL,
		testCfg.StatisticDBURL,
		testCfg.NotificationDBURL,
		testCfg.TaskDBURL,
	)
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func seedTaskGroup(t *testing.T, db *sql.DB, groupId string, userId string, name string, description string) {
	_, err := db.Exec(
		`INSERT INTO task_groups (group_id, user_id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		groupId, userId, name, description, time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to seed task group: %v", err)
	}
}

func RemoveTaskGroup(connections *database.Connections, groupId string) {
	// Remove test rows from the task_groups table
	_, err := connections.TaskDB.Exec("DELETE FROM task_groups WHERE group_id = $1", groupId)
	if err != nil {
		log.Printf("Failed to clean up test task group: %v", err)
	} else {
		log.Println("Test task group cleaned up successfully")
	}
}

func TestCreateTaskGroup(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	userId := uuid.NewString()
	name := "Test Task Group"
	description := "This is a test task group"

	service := NewService(connections)

	req := &pb.CreateTaskGroupRequest{
		UserId:      userId,
		Name:        name,
		Description: description,
	}

	resp, err := service.CreateTaskGroup(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTaskGroup failed: %v", err)
	}

	if resp.Group.Name != name {
		t.Errorf("Expected name %s, got ID %s and name %s",
			name, resp.Group.GroupId, resp.Group.Name)
	}

	RemoveTaskGroup(connections, resp.Group.GroupId)
}

func TestGetTaskGroups(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	userId := uuid.NewString()
	groupId := uuid.NewString()
	name := "Test Task Group"
	description := "This is a test task group"

	seedTaskGroup(t, connections.TaskDB, groupId, userId, name, description)
	defer RemoveTaskGroup(connections, groupId)

	service := NewService(connections)

	req := &pb.GetTaskGroupsRequest{
		UserId: userId,
	}

	resp, err := service.GetTaskGroups(context.Background(), req)
	if err != nil {
		t.Fatalf("GetTaskGroups failed: %v", err)
	}

	if len(resp.Groups) == 0 {
		t.Fatal("Expected at least one task group")
	}

	if resp.Groups[0].GroupId != groupId || resp.Groups[0].Name != name {
		t.Errorf("Expected group ID %s and name %s, got ID %s and name %s",
			groupId, name, resp.Groups[0].GroupId, resp.Groups[0].Name)
	}
}

func TestCreateTaskGroup_InvalidUserId(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	req := &pb.CreateTaskGroupRequest{
		UserId:      "",
		Name:        "Invalid User ID Group",
		Description: "This group has an invalid user ID",
	}

	_, err = service.CreateTaskGroup(context.Background(), req)
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Expected InvalidArgument error, got %v", err)
	}
}

func TestGetTaskGroups_InvalidUserId(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	req := &pb.GetTaskGroupsRequest{
		UserId: "",
	}

	_, err = service.GetTaskGroups(context.Background(), req)
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Expected InvalidArgument error, got %v", err)
	}
}

func seedTask(
	t *testing.T,
	db *sql.DB,
	taskId, userId, groupId, name, description string,
	priority pb.TaskPriority,
	status pb.TaskStatus,
	totalPomodoros, completedPomodoros, progress int32,
	deadline *time.Time,
) {
	t.Helper()

	now := time.Now()

	var deadlineVal any
	if deadline != nil {
		deadlineVal = *deadline
	} else {
		deadlineVal = nil // INSERT NULL when no deadline provided
	}

	_, err := db.Exec(
		`INSERT INTO tasks (
            task_id, user_id, group_id, name, description,
            priority, status, total_pomodoros, completed_pomodoros, progress,
            deadline, created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		taskId,
		userId,
		groupId,
		name,
		description,
		int32(priority),
		int32(status),
		totalPomodoros,
		completedPomodoros,
		progress,
		deadlineVal,
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
}

func TestCreateTask(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	userId := uuid.NewString()
	groupId := uuid.NewString()
	taskId := uuid.NewString()
	name := "Test Task"
	description := "This is a test task"
	priority := pb.TaskPriority_TASK_PRIORITY_MEDIUM
	status := pb.TaskStatus_TASK_STATUS_IDLE
	totalPomodoros := int32(2)
	deadline := time.Now().Add(24 * time.Hour)

	seedTaskGroup(t, connections.TaskDB, groupId, userId, "Test Group", "This is a test group")
	defer RemoveTaskGroup(connections, groupId)

	req := &pb.CreateTaskRequest{
		UserId:         userId,
		GroupId:        groupId,
		Name:           name,
		Description:    description,
		Priority:       priority,
		Status:         status,
		TotalPomodoros: totalPomodoros,
		Deadline:       timestamppb.New(deadline),
	}

	resp, err := service.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if resp.Task.TaskId != taskId || resp.Task.Name != name {
		t.Errorf("Expected task ID %s and name %s, got ID %s and name %s",
			taskId, name, resp.Task.TaskId, resp.Task.Name)
	}
}

func TestGetTasks(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	userId := uuid.NewString()
	groupId := uuid.NewString()
	taskId := uuid.NewString()
	name := "Test Task"
	description := "This is a test task"
	priority := pb.TaskPriority_TASK_PRIORITY_MEDIUM
	status := pb.TaskStatus_TASK_STATUS_IDLE
	totalPomodoros := int32(2)
	deadline := time.Now().Add(24 * time.Hour)

	seedTaskGroup(t, connections.TaskDB, groupId, userId, "Test Group", "This is a test group")
	defer RemoveTaskGroup(connections, groupId)

	seedTask(t, connections.TaskDB, taskId, userId, groupId, name, description,
		priority, status, totalPomodoros, 0, 0, &deadline)

	req := &pb.GetTasksRequest{
		UserId:  userId,
		GroupId: groupId,
	}

	resp, err := service.GetTasks(context.Background(), req)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(resp.Tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	if resp.Tasks[0].TaskId != taskId || resp.Tasks[0].Name != name {
		t.Errorf("Expected task ID %s and name %s, got ID %s and name %s",
			taskId, name, resp.Tasks[0].TaskId, resp.Tasks[0].Name)
	}
}

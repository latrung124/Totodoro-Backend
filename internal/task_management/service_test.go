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
	"github.com/latrung124/Totodoro-Backend/internal/helper"
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

func timesClose(a, b time.Time, tol time.Duration) bool {
	d := a.Sub(b)
	if d < 0 {
		d = -d
	}
	return d <= tol
}

func seedTaskGroup(t *testing.T, db *sql.DB, groupId string, userId string, name string, description string) {
	t.Helper()

	now := time.Now()
	icon := ""
	var deadline any = nil
	priority := "medium"
	status := "idle"
	completedTasks := int32(0)
	totalTasks := int32(0)

	_, err := db.Exec(
		`INSERT INTO task_groups (
            group_id, user_id, icon, name, description, deadline,
            priority, status, completed_tasks, total_tasks,
            created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		groupId, userId, icon, name, description, deadline,
		priority, status, completedTasks, totalTasks,
		now, now,
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

	service := NewService(connections)

	userId := uuid.NewString()
	name := "Test Task Group"
	icon := "This is a test icon"
	description := "This is a test task group"
	deadline := time.Now().Add(24 * time.Hour)
	priority := pb.TaskGroupPriority_TASK_GROUP_PRIORITY_MEDIUM
	status := pb.TaskGroupStatus_TASK_GROUP_STATUS_UNSPECIFIED

	req := &pb.CreateTaskGroupRequest{
		UserId:      userId,
		Name:        name,
		Icon:        icon,
		Description: description,
		Deadline:    timestamppb.New(deadline),
		Priority:    priority,
		Status:      status,
		// Icon, Deadline, Priority, Status left as zero-values
	}

	resp, err := service.CreateTaskGroup(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTaskGroup failed: %v", err)
	}
	defer RemoveTaskGroup(connections, resp.Group.GroupId)

	if resp.Group == nil {
		t.Fatal("Response group is nil")
	}

	if resp.Group.UserId != userId {
		t.Errorf("Expected user_id %s, got %s", userId, resp.Group.UserId)
	}

	if resp.Group.Name != name {
		t.Errorf("Expected name %q, got %q", name, resp.Group.Name)
	}

	if resp.Group.Icon != icon {
		t.Errorf("Expected icon %q, got %q", icon, resp.Group.Icon)
	}

	if resp.Group.Deadline == nil || !resp.Group.Deadline.AsTime().Equal(deadline) {
		t.Errorf("Expected deadline %v, got %v", deadline, resp.Group.Deadline.AsTime())
	}

	if resp.Group.Priority != priority {
		t.Errorf("Expected priority MEDIUM, got %v", resp.Group.Priority)
	}

	if resp.Group.Description != description {
		t.Errorf("Expected description %q, got %q", description, resp.Group.Description)
	}

	if resp.Group.Status != pb.TaskGroupStatus_TASK_GROUP_STATUS_UNSPECIFIED {
		t.Errorf("Expected default status UNSPECIFIED, got %v", resp.Group.Status)
	}

	if resp.Group.CompletedTasks != 0 || resp.Group.TotalTasks != 0 {
		t.Errorf("Expected completed_tasks=0 and total_tasks=0, got %d and %d",
			resp.Group.CompletedTasks, resp.Group.TotalTasks)
	}

	if resp.Group.CreatedAt.AsTime().After(time.Now().Add(1*time.Second)) ||
		resp.Group.UpdatedAt.AsTime().After(time.Now().Add(1*time.Second)) {
		t.Errorf("Invalid timestamps in response")
	}

	var (
		gotName, gotIcon, gotDescription, gotPriority, gotStatus string
		gotCompleted, gotTotal                                   int32
		deadlineNT                                               sql.NullTime
	)

	err = connections.TaskDB.QueryRow(`
        SELECT name, icon, description, priority, status, completed_tasks, total_tasks, deadline
        FROM task_groups
        WHERE group_id = $1`, resp.Group.GroupId).Scan(
		&gotName, &gotIcon, &gotDescription, &gotPriority, &gotStatus, &gotCompleted, &gotTotal, &deadlineNT,
	)

	if err != nil {
		t.Fatalf("Failed to fetch created task group: %v", err)
	}

	if gotName != name || gotDescription != description || gotIcon != icon {
		t.Errorf("Persisted name/description/icon mismatch: expected %q/%q/%q, got %q/%q/%q",
			name, description, icon, gotName, gotDescription, gotIcon)
	}

	priorityLabel := helper.TaskGroupPriorityDbEnumToString(priority)
	if gotPriority != string(priorityLabel) {
		t.Errorf("Persisted priority expected %s, got %s", string(priorityLabel), gotPriority)
	}

	statusLabel := helper.TaskGroupStatusDbEnumToString(status)
	if gotStatus != string(statusLabel) {
		t.Errorf("Persisted status expected %s, got %s", string(statusLabel), gotStatus)
	}

	if gotCompleted != 0 || gotTotal != 0 {
		t.Errorf("Persisted completed/total expected 0/0, got %d/%d", gotCompleted, gotTotal)
	}

	if resp.Group.Deadline == nil || !timesClose(resp.Group.Deadline.AsTime().UTC(), deadline.UTC(), time.Second) {
		t.Errorf("Expected deadline ~%v, got %v", deadline, resp.Group.Deadline.AsTime())
	}
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
		string(helper.TaskPriorityDbEnumToString(priority)),
		string(helper.TaskStatusDbEnumToString(status)),
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

func RemoveTask(connections *database.Connections, taskId string) {
	// Remove test rows from the tasks table
	_, err := connections.TaskDB.Exec("DELETE FROM tasks WHERE task_id = $1", taskId)
	if err != nil {
		log.Printf("Failed to clean up test task: %v", err)
	} else {
		log.Println("Test task cleaned up successfully")
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

	if resp.Task.Name != name {
		t.Errorf("Expected name %s, got name %s",
			name, resp.Task.Name)
	}

	if resp.Task.Description != description {
		t.Errorf("Expected description %s, got description %s", resp.Task.Description, description)
	}

	if resp.Task.Priority != priority {
		t.Errorf("Expected priority %v, got %v", priority, resp.Task.Priority)
	}

	if resp.Task.Status != status {
		t.Errorf("Expected status %v, got %v", status, resp.Task.Status)
	}

	if resp.Task.TotalPomodoros != totalPomodoros {
		t.Errorf("Expected total pomodoros %d, got %d", totalPomodoros, resp.Task.TotalPomodoros)
	}

	if resp.Task.Deadline == nil || !timesClose(resp.Task.Deadline.AsTime().UTC(), deadline.UTC(), time.Second) {
		t.Errorf("Expected deadline ~%v, got %v", deadline, resp.Task.Deadline.AsTime())
	}

	// Clean up the task after the test
	RemoveTask(connections, resp.Task.TaskId)
	if err != nil {
		t.Fatalf("Failed to clean up test task: %v", err)
	} else {
		log.Println("Test task cleaned up successfully")
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
	defer RemoveTask(connections, taskId)

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

func TestGetTasks_InvalidUserId(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	req := &pb.GetTasksRequest{
		UserId:  "",
		GroupId: uuid.NewString(),
	}

	_, err = service.GetTasks(context.Background(), req)
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Expected InvalidArgument error, got %v", err)
	}
}

func TestUpdateTask(t *testing.T) {
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
	totalPomodoros := int32(3)
	completedPomodoros := int32(1)
	progress := int32(60)
	deadline := time.Now().Add(24 * time.Hour)

	seedTaskGroup(t, connections.TaskDB, groupId, userId, "Test Group", "This is a test group")
	defer RemoveTaskGroup(connections, groupId)

	seedTask(t, connections.TaskDB, taskId, userId, groupId, name, description,
		priority, status, totalPomodoros, 0, 0, &deadline)
	defer RemoveTask(connections, taskId)

	newName := "Updated Task Name"
	newDescription := "Updated task description"
	newPriority := pb.TaskPriority_TASK_PRIORITY_HIGH
	newStatus := pb.TaskStatus_TASK_STATUS_IN_PROGRESS

	req := &pb.UpdateTaskRequest{
		TaskId:             taskId,
		Name:               newName,
		Description:        newDescription,
		Priority:           newPriority,
		Status:             newStatus,
		TotalPomodoros:     totalPomodoros,
		CompletedPomodoros: completedPomodoros,
		Progress:           progress,
		Deadline:           timestamppb.New(deadline),
	}

	resp, err := service.UpdateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	if resp.Task.TaskId != taskId {
		t.Errorf("Expected task ID %s, got %s", taskId, resp.Task.TaskId)
	}

	if resp.Task.Name != newName {
		t.Errorf("Expected task name %s, got %s", newName, resp.Task.Name)
	}

	if resp.Task.Description != newDescription {
		t.Errorf("Expected task description %s, got %s", newDescription, resp.Task.Description)
	}

	if resp.Task.Priority != newPriority {
		t.Errorf("Expected task priority %v, got %v", newPriority, resp.Task.Priority)
	}

	if resp.Task.Status != newStatus {
		t.Errorf("Expected task status %v, got %v", newStatus, resp.Task.Status)
	}

	if resp.Task.TotalPomodoros != totalPomodoros {
		t.Errorf("Expected total pomodoros %d, got %d", totalPomodoros, resp.Task.TotalPomodoros)
	}

	if resp.Task.CompletedPomodoros != completedPomodoros {
		t.Errorf("Expected completed pomodoros %d, got %d", completedPomodoros, resp.Task.CompletedPomodoros)
	}

	if resp.Task.Progress != progress {
		t.Errorf("Expected progress %d, got %d", progress, resp.Task.Progress)
	}

	if resp.Task.Deadline == nil || !timesClose(resp.Task.Deadline.AsTime().UTC(), deadline.UTC(), time.Second) {
		t.Errorf("Expected deadline ~%v, got %v", deadline, resp.Task.Deadline.AsTime())
	}
}

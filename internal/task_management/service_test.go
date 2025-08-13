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

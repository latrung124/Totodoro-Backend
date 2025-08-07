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

	"github.com/latrung124/Totodoro-Backend/internal/config"
	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
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

func CleanupTestDB(connections *database.Connections) {
	// Close all database connections
	if connections.UserDB != nil {
		connections.UserDB.Exec("DROP TABLE IF EXISTS user_db")
		connections.UserDB.Close()
	}
	if connections.PomodoroDB != nil {
		connections.PomodoroDB.Exec("DROP TABLE IF EXISTS pomodoro_db")
		connections.PomodoroDB.Close()
	}
	if connections.StatisticDB != nil {
		connections.StatisticDB.Exec("DROP TABLE IF EXISTS statistic_db")
		connections.StatisticDB.Close()
	}
	if connections.NotificationDB != nil {
		connections.NotificationDB.Exec("DROP TABLE IF EXISTS notification_db")
		connections.NotificationDB.Close()
	}
	if connections.TaskDB != nil {
		connections.TaskDB.Exec("DROP TABLE IF EXISTS task_db")
		connections.TaskDB.Close()
	}
}

func TestCreateUser(t *testing.T) {
	// Test configuration loading
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}

	service := NewService(connections)

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

	var count int
	err = connections.UserDB.QueryRow("SELECT COUNT(*) FROM users WHERE user_id = $1", resp.User.UserId).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query user count: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user in database, found %d", count)
	}
}

func TestGetUser(t *testing.T) {
	//Test configuration loading
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}

	service := NewService(connections)

	req := &pb.GetUserRequest{
		UserId: "50f58fc8-c980-4ba6-9fcc-1e6f69367f94id-12345",
	}

	resp, err := service.GetUser(context.Background(), req)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	// Compare the response with expected values
	if resp.User.UserId != req.UserId {
		t.Errorf("Expected UserId %s, got %s", req.UserId, resp.User.UserId)
	}

	if resp.User.Email != "test@gmail.com" {
		t.Errorf("Expected Email test@gmail.com")
	}

	if resp.User.Username != "testuser" {
		t.Errorf("Expected Username testuser, got %s", resp.User.Username)
	}

	// Check timestamps
	if resp.User.CreatedAt.AsTime().After(time.Now()) {
		t.Errorf("Invalid CreatedAt timestamp")
	}

	if resp.User.UpdatedAt.AsTime().After(time.Now()) {
		t.Errorf("Invalid UpdatedAt timestamp")
	}
}

func TestGetUserNotFound(t *testing.T) {
	// Test configuration loading
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}

	service := NewService(connections)

	req := &pb.GetUserRequest{
		UserId: "50f58fc8-c980-4ba6-9fcc-1e6f69367f4",
	}

	_, err = service.GetUser(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

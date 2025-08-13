/*
File: internal/user/service_test.go
Author: trung.la
Date: 08-04-2025
Description: Test cases for user service functions.
*/

package user

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/latrung124/Totodoro-Backend/internal/config"
	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func RemoveUserId(connections *database.Connections, userId string) {
	// Remove test rows from the users table
	_, err := connections.UserDB.Exec("DELETE FROM users WHERE user_id = $1",
		&userId)
	if err != nil {
		log.Printf("Failed to clean up test user: %v", err)
	} else {
		log.Println("Test user cleaned up successfully")
	}
}

func seedTestUser(t *testing.T, db *sql.DB, userId string, email string, username string) {
	_, err := db.Exec(
		`INSERT INTO users (user_id, email, username, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW()) 
		 ON CONFLICT (user_id) DO NOTHING`,
		userId, email, username,
	)
	if err != nil {
		t.Fatalf("Failed to seed test user: %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}

	service := NewService(connections)

	req := &pb.CreateUserRequest{
		Email:    "CreateUserRequest@example.com",
		Username: "CreateUserRequest",
	}

	resp, err := service.CreateUser(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if resp.User.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.User.Email)
	}
	if resp.User.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, resp.User.Username)
	}

	// Check CreatedAt is not in the future
	if resp.User.CreatedAt.AsTime().After(time.Now().Add(1 * time.Second)) {
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

	//TODO: Check createdAt and updatedAt timestamps

	// Clean up the test user after the test
	RemoveUserId(connections, resp.User.UserId)
	t.Logf("Test UserId: %s", resp.User.UserId)
}

func TestGetUser(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}

	service := NewService(connections)
	testUserID := uuid.NewString()
	seedTestUser(t, connections.UserDB, testUserID, "TestGetUser@gmail.com", "TestGetUser")

	req := &pb.GetUserRequest{
		UserId: testUserID,
	}

	resp, err := service.GetUser(context.Background(), req)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if resp.User.UserId != req.UserId {
		t.Errorf("Expected UserId %s, got %s", req.UserId, resp.User.UserId)
	}
	if resp.User.Email != "TestGetUser@gmail.com" {
		t.Errorf("Expected Email TestGetUser@gmail.com, got %s", resp.User.Email)
	}
	if resp.User.Username != "TestGetUser" {
		t.Errorf("Expected Username TestGetUser, got %s", resp.User.Username)
	}

	// TODO: Check CreatedAt and UpdatedAt timestamps

	// Clean up the test user after the test
	RemoveUserId(connections, testUserID)
}

func TestGetUserNotFound(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}

	service := NewService(connections)

	req := &pb.GetUserRequest{
		UserId: "12345678-1234-1234-1234-123456789012", // Non-existent user ID
	}

	_, err = service.GetUser(context.Background(), req)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("Expected NotFound error, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}

	service := NewService(connections)
	testUserID := uuid.NewString()
	seedTestUser(t, connections.UserDB, testUserID, "TestUpdateUser@gmail.com", "TestUpdateUser")

	req := &pb.UpdateUserRequest{
		UserId:   testUserID,
		Email:    "TestUpdateUser_After@gmail.com",
		Username: "TestUpdateUser_After",
	}

	resp, err := service.UpdateUser(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if resp.User.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.User.Email)
	}
	if resp.User.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, resp.User.Username)
	}

	// TODO: Check CreatedAt and UpdatedAt timestamps

	var count int
	err = connections.UserDB.QueryRow("SELECT COUNT(*) FROM users WHERE user_id = $1", req.UserId).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query user count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user in database, found %d", count)
	}
	t.Logf("Test UserId: %s", req.UserId)

	// Clean up the test user after the test
	RemoveUserId(connections, req.UserId)
}

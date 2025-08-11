/*
File: internal/pomodoro/service_test.go
Author: trung.la
Date: 08-11-2025
Description: Unit tests for the pomodoro service functions.
*/

package pomodoro

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/config"
	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/pomodoro_service"
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

func SeedPomodoroSession(t *testing.T, db *sql.DB, sessionId string, userId string, taskId string, startTime time.Time, endTime time.Time) {
	_, err := db.Exec(
		`INSERT INTO pomodoro_sessions (session_id, user_id, task_id, start_time, end_time, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		sessionId,
		userId,
		taskId,
		startTime,
		endTime,
		pb.SessionStatus_SESSION_STATUS_UNSPECIFIED,
	)
	if err != nil {
		t.Fatalf("Failed to seed test pomodoro session: %v", err)
	}
}

func RemovePomodoroSession(connections *database.Connections, sessionId string) {
	// Remove test rows from the pomodoro_sessions table
	_, err := connections.PomodoroDB.Exec("DELETE FROM pomodoro_sessions WHERE session_id = $1", sessionId)
	if err != nil {
		log.Printf("Failed to clean up test pomodoro session: %v", err)
	} else {
		log.Println("Test pomodoro session cleaned up successfully")
	}
}

func TestCreateSession(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	now := time.Now()

	req := &pb.CreateSessionRequest{
		UserId:    "test_user",
		TaskId:    "test_task",
		StartTime: timestamppb.New(now),
		EndTime:   timestamppb.New(now.Add(25 * time.Minute)),
		Status:    pb.SessionStatus_SESSION_STATUS_UNSPECIFIED,
	}

	resp, err := service.CreateSession(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if resp.Session.UserId != req.UserId || resp.Session.TaskId != req.TaskId {
		t.Errorf("Expected UserId %s and TaskId %s, got UserId %s and TaskId %s", req.UserId, req.TaskId, resp.Session.UserId, resp.Session.TaskId)
	}

	// Check StartTime and EndTime
	if resp.Session.StartTime.AsTime().After(resp.Session.EndTime.AsTime()) {
		t.Errorf("StartTime must be before EndTime")
	}

	// Check if session was inserted into the database
	var count int
	err = connections.PomodoroDB.QueryRow("SELECT COUNT(*) FROM pomodoro_sessions WHERE session_id = $1", resp.Session.SessionId).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query session count: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 session in the database, found %d", count)
	}

	// Clean up test session
	RemovePomodoroSession(connections, resp.Session.SessionId)
}

func TestGetSessions(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	userId := "test_user"
	taskId := "test_task"
	sessionId := "session_test_123"
	startTime := time.Now()
	endTime := startTime.Add(25 * time.Minute)

	// Seed a test session
	SeedPomodoroSession(t, connections.PomodoroDB, sessionId, userId, taskId, startTime, endTime)

	req := &pb.GetSessionsRequest{
		UserId: userId,
	}

	resp, err := service.GetSessions(context.Background(), req)
	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}

	if len(resp.Sessions) == 0 {
		t.Error("Expected at least one session, got none")
	}

	for _, session := range resp.Sessions {
		if session.UserId != userId {
			t.Errorf("Expected UserId %s, got %s", userId, session.UserId)
		}
		if session.TaskId != taskId {
			t.Errorf("Expected TaskId %s, got %s", taskId, session.TaskId)
		}
	}

	// Clean up test session
	RemovePomodoroSession(connections, sessionId)
}

func TestUpdateSessionResponse(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	userId := "test_user"
	taskId := "test_task"
	sessionId := "session_test_update_123"
	startTime := time.Now()
	endTime := startTime.Add(25 * time.Minute)

	// Seed a test session
	SeedPomodoroSession(t, connections.PomodoroDB, sessionId, userId, taskId, startTime, endTime)

	req := &pb.UpdateSessionRequest{
		SessionId: sessionId,
		Status:    pb.SessionStatus_COMPLETED,
	}

	resp, err := service.UpdateSessionResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateSessionResponse failed: %v", err)
	}

	if resp.Session.SessionId != sessionId {
		t.Errorf("Expected SessionId %s, got %s", sessionId, resp.Session.SessionId)
	}
	if resp.Session.Status != pb.SessionStatus_COMPLETED {
		t.Errorf("Expected Status %s, got %s", pb.SessionStatus_COMPLETED, resp.Session.Status)
	}

	// Check if the session was updated in the database
	var status pb.SessionStatus
	err = connections.PomodoroDB.QueryRow("SELECT status FROM pomodoro_sessions WHERE session_id = $1", sessionId).Scan(&status)
	if err != nil {
		t.Fatalf("Failed to query updated session status: %v", err)
	}

	if status != pb.SessionStatus_COMPLETED {
		t.Errorf("Expected updated status %s, got %s", pb.SessionStatus_COMPLETED, status)
	}

	// Clean up test session
	RemovePomodoroSession(connections, sessionId)
}

func TestDeleteSession(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer connections.Close()

	service := NewService(connections)

	userId := "test_user"
	taskId := "test_task"
	sessionId := "session_test_delete_123"
	startTime := time.Now()
	endTime := startTime.Add(25 * time.Minute)

	// Seed a test session
	SeedPomodoroSession(t, connections.PomodoroDB, sessionId, userId, taskId, startTime, endTime)

	req := &pb.DeleteSessionRequest{
		SessionId: sessionId,
	}

	resp, err := service.DeleteSession(context.Background(), req)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	if resp.Success != true {
		t.Error("Expected success response, got false")
	}

	// Check if the session was deleted from the database
	var count int
	err = connections.PomodoroDB.QueryRow("SELECT COUNT(*) FROM pomodoro_sessions WHERE session_id = $1", sessionId).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query deleted session count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 sessions in the database after deletion, found %d", count)
	}
}

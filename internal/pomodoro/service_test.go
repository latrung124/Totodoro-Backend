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

	"github.com/google/uuid"
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
	progress := 0
	statusStr := pb.SessionStatus_name[int32(pb.SessionStatus_SESSION_STATUS_IDLE)]
	sessionTypeStr := pb.SessionType_name[int32(pb.SessionType_SESSION_TYPE_SHORT_BREAK)]
	numberInCycle := 0
	lastUpdate := time.Now()

	_, err := db.Exec(
		`INSERT INTO sessions (
            session_id, user_id, task_id, start_time, progress, end_time,
            status, session_type, number_in_cycle, last_update
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		sessionId,
		userId,
		taskId,
		startTime,
		progress,
		endTime,
		statusStr,
		sessionTypeStr,
		numberInCycle,
		lastUpdate,
	)
	if err != nil {
		t.Fatalf("Failed to seed test pomodoro session: %v", err)
	}
}

func RemovePomodoroSession(connections *database.Connections, sessionId string) {
	// Remove test rows from the pomodoro_sessions table
	_, err := connections.PomodoroDB.Exec("DELETE FROM sessions WHERE session_id = $1", sessionId)
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
		UserId:        uuid.NewString(),
		TaskId:        uuid.NewString(),
		StartTime:     timestamppb.New(now),
		SessionType:   pb.SessionType_SESSION_TYPE_SHORT_BREAK, // per new proto
		NumberInCycle: 0,
	}

	resp, err := service.CreateSession(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if resp.Session.UserId != req.UserId || resp.Session.TaskId != req.TaskId {
		t.Errorf("Expected UserId %s and TaskId %s, got UserId %s and TaskId %s",
			req.UserId, req.TaskId, resp.Session.UserId, resp.Session.TaskId)
	}

	// StartTime must be set
	if resp.Session.StartTime == nil {
		t.Fatalf("StartTime is nil")
	}

	// EndTime is not part of CreateSessionRequest anymore; service may set it or leave nil.
	// If present, it must be equal/after StartTime.
	if resp.Session.EndTime != nil && resp.Session.StartTime.AsTime().After(resp.Session.EndTime.AsTime()) {
		t.Errorf("StartTime must be before or equal to EndTime when EndTime is set")
	}

	// Progress should start at 0
	if resp.Session.Progress != 0 {
		t.Errorf("Expected initial progress 0, got %d", resp.Session.Progress)
	}

	// Check if session was inserted into the database
	var count int
	err = connections.PomodoroDB.QueryRow(
		"SELECT COUNT(*) FROM pomodoro_sessions WHERE session_id = $1",
		resp.Session.SessionId,
	).Scan(&count)
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

	userId := uuid.NewString()
	taskId := uuid.NewString()
	sessionId := uuid.NewString()
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

	userId := uuid.NewString()
	taskId := uuid.NewString()
	sessionId := uuid.NewString()
	startTime := time.Now().Add(-25 * time.Minute)
	endTime := time.Now()

	// Seed a test session
	SeedPomodoroSession(t, connections.PomodoroDB, sessionId, userId, taskId, startTime, endTime)

	// Prepare update values
	newProgress := int32(120)
	newStatus := pb.SessionStatus_SESSION_STATUS_COMPLETED
	newType := pb.SessionType_SESSION_TYPE_LONG_BREAK
	newNumberInCycle := int32(2)
	newLastUpdate := time.Now()

	req := &pb.UpdateSessionRequest{
		SessionId:     sessionId,
		Progress:      newProgress,
		EndTime:       timestamppb.New(endTime),
		Status:        newStatus,
		SessionType:   newType,
		NumberInCycle: newNumberInCycle,
		LastUpdate:    timestamppb.New(newLastUpdate),
	}

	resp, err := service.UpdateSessionResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateSessionResponse failed: %v", err)
	}

	if resp.Session.SessionId != sessionId {
		t.Errorf("Expected SessionId %s, got %s", sessionId, resp.Session.SessionId)
	}
	if resp.Session.Progress != newProgress {
		t.Errorf("Expected Progress %d, got %d", newProgress, resp.Session.Progress)
	}
	if resp.Session.Status != newStatus {
		t.Errorf("Expected Status %v, got %v", newStatus, resp.Session.Status)
	}
	if resp.Session.SessionType != newType {
		t.Errorf("Expected SessionType %v, got %v", newType, resp.Session.SessionType)
	}
	if resp.Session.NumberInCycle != newNumberInCycle {
		t.Errorf("Expected NumberInCycle %d, got %d", newNumberInCycle, resp.Session.NumberInCycle)
	}
	if resp.Session.EndTime == nil {
		t.Fatalf("EndTime is nil in response")
	}
	et := resp.Session.EndTime.AsTime().UTC()
	if et.Before(endTime.UTC().Add(-time.Second)) || et.After(endTime.UTC().Add(time.Second)) {
		t.Errorf("Expected EndTime ~%v, got %v", endTime, resp.Session.EndTime.AsTime())
	}
	if resp.Session.LastUpdate == nil {
		t.Fatalf("LastUpdate is nil in response")
	}

	// Check if the session was updated in the database
	var (
		gotProgress      int32
		gotEndTime       time.Time
		gotStatusStr     string
		gotTypeStr       string
		gotNumberInCycle int32
		gotLastUpdate    time.Time
	)
	err = connections.PomodoroDB.QueryRow(`
        SELECT progress, end_time, status, session_type, number_in_cycle, last_update
        FROM pomodoro_sessions
        WHERE session_id = $1`, sessionId).Scan(
		&gotProgress, &gotEndTime, &gotStatusStr, &gotTypeStr, &gotNumberInCycle, &gotLastUpdate,
	)
	if err != nil {
		t.Fatalf("Failed to query updated session: %v", err)
	}

	// DB stores status/session_type as strings using proto names (e.g., "SESSION_STATUS_COMPLETED")
	expStatusStr := pb.SessionStatus_name[int32(newStatus)]
	expTypeStr := pb.SessionType_name[int32(newType)]

	if gotProgress != newProgress {
		t.Errorf("DB progress mismatch: expected %d, got %d", newProgress, gotProgress)
	}
	if gotStatusStr != expStatusStr {
		t.Errorf("DB status mismatch: expected %s, got %s", expStatusStr, gotStatusStr)
	}
	if gotTypeStr != expTypeStr {
		t.Errorf("DB session_type mismatch: expected %s, got %s", expTypeStr, gotTypeStr)
	}
	if gotNumberInCycle != newNumberInCycle {
		t.Errorf("DB number_in_cycle mismatch: expected %d, got %d", newNumberInCycle, gotNumberInCycle)
	}
	// Compare times in UTC with small tolerance
	if gotEndTime.UTC().Before(endTime.UTC().Add(-time.Second)) || gotEndTime.UTC().After(endTime.UTC().Add(time.Second)) {
		t.Errorf("DB end_time mismatch: expected ~%v, got %v", endTime, gotEndTime)
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

	userId := uuid.NewString()
	taskId := uuid.NewString()
	sessionId := uuid.NewString()
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

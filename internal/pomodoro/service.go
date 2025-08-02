/*
File: internal/pomodoro/service.go
Author: trung.la
Date: 07/30/2025
Package: github.com/latrung124/Totodoro-Backend/internal/pomodoro
Description: This file contains the service layer for pomodoro management.
*/

package pomodoro

import (
	"context"
	"log"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/pomodoro_service"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedPomodoroServiceServer
	db *database.Connections
}

func NewService(db *database.Connections) *Service {
	return &Service{db: db}
}

// CreatePomodoro creates a new pomodoro session for a user.
func (s *Service) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	if req.StartTime.AsTime().After(req.EndTime.AsTime()) {
		return nil, status.Error(codes.InvalidArgument, "StartTime must be before EndTime")
	}

	//Start session

	sessionId := "session_" + time.Now().Format("20060102150405")
	newSession := &pb.PomodoroSession{
		SessionId: sessionId,
		UserId:    req.UserId,
		TaskId:    req.TaskId,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    pb.SessionStatus_SESSION_STATUS_UNSPECIFIED,
	}

	_, err := s.db.PomodoroDB.ExecContext(ctx, "INSERT INTO pomodoro_sessions (session_id, user_id, task_id, start_time, end_time, status) VALUES ($1, $2, $3, $4, $5, $6)", newSession.SessionId, newSession.UserId, newSession.TaskId, newSession.StartTime.AsTime(), newSession.EndTime.AsTime(), newSession.Status)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // Unique violation
			return nil, status.Error(codes.AlreadyExists, "session already exists")
		}
		log.Printf("Failed to insert session: %v", err)
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	return &pb.CreateSessionResponse{Session: newSession}, nil
}

func (s *Service) GetSessions(ctx context.Context, req *pb.GetSessionsRequest) (*pb.GetSessionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	rows, err := s.db.PomodoroDB.QueryContext(ctx, "SELECT session_id, user_id, task_id, start_time, end_time, status FROM pomodoro_sessions WHERE user_id = $1", req.UserId)
	if err != nil {
		log.Printf("Failed to query sessions: %v", err)
		return nil, status.Error(codes.Internal, "failed to retrieve sessions")
	}
	defer rows.Close()

	var sessions []*pb.PomodoroSession
	for rows.Next() {
		var session pb.PomodoroSession
		if err := rows.Scan(&session.SessionId, &session.UserId, &session.TaskId, &session.StartTime, &session.EndTime, &session.Status); err != nil {
			log.Printf("Failed to scan session: %v", err)
			return nil, status.Error(codes.Internal, "failed to retrieve sessions")
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over sessions: %v", err)
		return nil, status.Error(codes.Internal, "failed to retrieve sessions")
	}

	return &pb.GetSessionsResponse{Sessions: sessions}, nil
}

func (s *Service) UpdateSessionResponse(ctx context.Context, req *pb.UpdateSessionRequest) (*pb.UpdateSessionResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	if req.Status == pb.SessionStatus_SESSION_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	_, err := s.db.PomodoroDB.ExecContext(ctx, "UPDATE pomodoro_sessions SET status = $1 WHERE session_id = $2", req.Status, req.SessionId)
	if err != nil {
		log.Printf("Failed to update session: %v", err)
		return nil, status.Error(codes.Internal, "failed to update session")
	}

	// Read the updated session
	session := &pb.PomodoroSession{
		SessionId: req.SessionId,
		Status:    req.Status,
	}

	return &pb.UpdateSessionResponse{Session: session}, nil
}

func (s *Service) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	_, err := s.db.PomodoroDB.ExecContext(ctx, "DELETE FROM pomodoro_sessions WHERE session_id = $1", req.SessionId)
	if err != nil {
		log.Printf("Failed to delete session: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete session")
	}

	return &pb.DeleteSessionResponse{Success: true}, nil
}

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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedPomodoroServiceServer
	db *database.Connections
}

func NewService(db *database.Connections) *Service {
	return &Service{db: db}
}

// CreatePomodoro creates a new pomodoro session for a user.
func (s *Service) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (pb *pb.CreateSessionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	if req.StartTime > req.EndTime {
		return nil, status.Error(codes.InvalidArgument, "StartTime must be before EndTime")
	}

	//Start session

	sessionId := "session_" + time.Now().Format("20060102150405")
	now := time.Now()
	newSession := &pb.PomodoroSession{
		SessionId: sessionId,
		UserId:    req.UserId,
		TaskId:    req.TaskId,
		StartTime: timestamppb.New(time.Unix(req.StartTime, 0)),
		EndTime:   timestamppb.New(time.Unix(req.EndTime, 0)),
		Status:   pb.SessionStatus_SESSION_STATUS_UNSPECIFIED,
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

func (s *Service) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	var session pb.PomodoroSession
	err := s.db.PomodoroDB.QueryRowContext(ctx, "SELECT session_id, user_id, task_id, start_time, end_time, status FROM pomodoro_sessions WHERE session_id = $1", req.SessionId).Scan(
		&session.SessionId,
		&session.UserId,
		&session.TaskId,
		&session.StartTime,
		&session.EndTime,
		&session.Status,
	)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return nil, status.Error(codes.NotFound, "session not found")
	}

	return &pb.GetSessionResponse{Session: &session}, nil
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

	return &pb.UpdateSessionResponse{Success: true}, nil
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
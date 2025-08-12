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

	"github.com/google/uuid"
	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/pomodoro_service"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedPomodoroServiceServer
	db *database.Connections
}

var sessionStatusEnumToStringMap = map[int32]string{
	0: "SESSION_STATUS_UNSPECIFIED",
	1: "COMPLETED",
	2: "PAUSED",
	3: "RUNNING",
	4: "ERROR",
}

var sessionStatusStringToEnumMap = map[string]pb.SessionStatus{
	"SESSION_STATUS_UNSPECIFIED": pb.SessionStatus_SESSION_STATUS_UNSPECIFIED,
	"COMPLETED":                  pb.SessionStatus_COMPLETED,
	"PAUSED":                     pb.SessionStatus_PAUSED,
	"RUNNING":                    pb.SessionStatus_RUNNING,
	"ERROR":                      pb.SessionStatus_ERROR,
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
	sessionId := uuid.NewString()

	newSession := &pb.PomodoroSession{
		SessionId: sessionId,
		UserId:    req.UserId,
		TaskId:    req.TaskId,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    pb.SessionStatus_SESSION_STATUS_UNSPECIFIED,
	}

	statusStr := sessionStatusEnumToStringMap[int32(newSession.Status)]

	_, err := s.db.PomodoroDB.ExecContext(ctx, "INSERT INTO pomodoro_sessions (session_id, user_id, task_id, start_time, end_time, status) VALUES ($1, $2, $3, $4, $5, $6)", newSession.SessionId, newSession.UserId, newSession.TaskId, newSession.StartTime.AsTime(), newSession.EndTime.AsTime(), statusStr)
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
		return nil, status.Error(codes.InvalidArgument, "user_id is required") // fixed message
	}

	rows, err := s.db.PomodoroDB.QueryContext(ctx, `
        SELECT session_id, user_id, task_id, start_time, end_time, status
        FROM pomodoro_sessions
        WHERE user_id = $1`, req.UserId)
	if err != nil {
		log.Printf("Failed to query sessions: %v", err)
		return nil, status.Error(codes.Internal, "failed to retrieve sessions")
	}
	defer rows.Close()

	var sessions []*pb.PomodoroSession
	for rows.Next() {
		var (
			session   pb.PomodoroSession
			start     time.Time
			end       time.Time
			statusStr string
		)
		if err := rows.Scan(&session.SessionId, &session.UserId, &session.TaskId, &start, &end, &statusStr); err != nil {
			log.Printf("Failed to scan session: %v", err)
			return nil, status.Error(codes.Internal, "failed to retrieve sessions")
		}
		session.StartTime = timestamppb.New(start)
		session.EndTime = timestamppb.New(end)
		if v, ok := sessionStatusStringToEnumMap[statusStr]; ok {
			session.Status = v
		} else {
			session.Status = pb.SessionStatus_SESSION_STATUS_UNSPECIFIED
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

	statusStr := sessionStatusEnumToStringMap[int32(req.Status)]
	_, err := s.db.PomodoroDB.ExecContext(ctx, "UPDATE pomodoro_sessions SET status = $1 WHERE session_id = $2", statusStr, req.SessionId)
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

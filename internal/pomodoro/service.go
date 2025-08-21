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

	"database/sql"

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

	//Start session
	sessionId := uuid.NewString()

	newSession := &pb.PomodoroSession{
		SessionId:     sessionId,
		UserId:        req.UserId,
		TaskId:        req.TaskId,
		StartTime:     req.StartTime,
		Progress:      0,
		EndTime:       timestamppb.New(req.StartTime.AsTime().Add(25 * time.Minute)),
		Status:        pb.SessionStatus_SESSION_STATUS_IDLE,
		SessionType:   pb.SessionType_SESSION_TYPE_SHORT_BREAK,
		NumberInCycle: 0,
		LastUpdate:    timestamppb.Now(),
	}

	statusStr := pb.SessionStatus_name[int32(newSession.Status)]
	sessionTypeStr := pb.SessionType_name[int32(newSession.SessionType)]

	_, err := s.db.PomodoroDB.ExecContext(ctx, "INSERT INTO sessions (session_id, user_id, task_id, start_time, progress, end_time, status, session_type, number_in_cycle, last_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", newSession.SessionId, newSession.UserId, newSession.TaskId, newSession.StartTime.AsTime(), newSession.Progress, newSession.EndTime.AsTime(), statusStr, sessionTypeStr, newSession.NumberInCycle, newSession.LastUpdate.AsTime())
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
        SELECT session_id, task_id, start_time, progress, end_time, status,
		session_type, number_in_cycle, last_update
        FROM sessions
        WHERE user_id = $1`, req.UserId)
	if err != nil {
		log.Printf("Failed to query sessions: %v", err)
		return nil, status.Error(codes.Internal, "failed to retrieve sessions")
	}
	defer rows.Close()

	var sessions []*pb.PomodoroSession
	for rows.Next() {
		var (
			session        pb.PomodoroSession
			start          time.Time
			end            time.Time
			statusStr      string
			sessionTypeStr string
			lastUpdate     time.Time
		)
		if err := rows.Scan(&session.SessionId, &session.UserId, &session.TaskId, &start, &session.Progress, &end, &statusStr, &sessionTypeStr, &lastUpdate); err != nil {
			log.Printf("Failed to scan session: %v", err)
			return nil, status.Error(codes.Internal, "failed to retrieve sessions")
		}
		session.StartTime = timestamppb.New(start)
		session.EndTime = timestamppb.New(end)
		if v, ok := pb.SessionStatus_value[statusStr]; ok {
			session.Status = pb.SessionStatus(v)
		} else {
			session.Status = pb.SessionStatus_SESSION_STATUS_UNSPECIFIED
		}

		if v, ok := pb.SessionType_value[sessionTypeStr]; ok {
			session.SessionType = pb.SessionType(v)
		} else {
			session.SessionType = pb.SessionType_SESSION_TYPE_UNSPECIFIED
		}
		session.LastUpdate = timestamppb.New(lastUpdate)
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

	statusStr, ok := pb.SessionStatus_name[int32(req.Status)]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	sessionTypeStr, ok := pb.SessionType_name[int32(req.SessionType)]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid session type")
	}

	_, err := s.db.PomodoroDB.ExecContext(ctx, `UPDATE sessions SET
		progress = $1,
		end_time = $2,
		status = $3,
		session_type = $4,
		number_in_cycle = $5,
		last_update = $6
		WHERE session_id = $7`, req.Progress, req.EndTime, statusStr, sessionTypeStr, req.NumberInCycle, req.LastUpdate, req.SessionId)

	if err != nil {
		log.Printf("Failed to update session: %v", err)
		return nil, status.Error(codes.Internal, "failed to update session")
	}

	// Retrieve the updated session
	var session pb.PomodoroSession
	row := s.db.PomodoroDB.QueryRowContext(ctx, `SELECT session_id, user_id, task_id, start_time, progress, end_time, status,
		session_type, number_in_cycle, last_update
		FROM sessions WHERE session_id = $1`, req.SessionId)

	var start, end, lastUpdate time.Time
	if err := row.Scan(&session.SessionId, &session.UserId, &session.TaskId, &start, &session.Progress, &end, &statusStr, &sessionTypeStr, &session.NumberInCycle, &lastUpdate); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		log.Printf("Failed to retrieve updated session: %v", err)
		return nil, status.Error(codes.Internal, "failed to retrieve updated session")
	}

	session.StartTime = timestamppb.New(start)
	session.EndTime = timestamppb.New(end)
	if v, ok := pb.SessionStatus_value[statusStr]; ok {
		session.Status = pb.SessionStatus(v)
	} else {
		session.Status = pb.SessionStatus_SESSION_STATUS_IDLE
	}
	if v, ok := pb.SessionType_value[sessionTypeStr]; ok {
		session.SessionType = pb.SessionType(v)
	} else {
		session.SessionType = pb.SessionType_SESSION_TYPE_SHORT_BREAK
	}
	session.LastUpdate = timestamppb.New(lastUpdate)

	// Return the updated session
	log.Printf("Session updated successfully: %s", session.SessionId)

	return &pb.UpdateSessionResponse{Session: &session}, nil
}

func (s *Service) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	_, err := s.db.PomodoroDB.ExecContext(ctx, "DELETE FROM sessions WHERE session_id = $1", req.SessionId)
	if err != nil {
		log.Printf("Failed to delete session: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete session")
	}

	return &pb.DeleteSessionResponse{Success: true}, nil
}

/*
File: service.go
Author: trung.la
Date: 07/29/2025
Package: github.com/latrung124/Totodoro-Backend/internal/statistic
Description: This file contains the service layer for statistic management.
*/

package statistic

import (
	"context"
	"log"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/statistic_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedStatisticServiceServer
	db *database.Connections
}

func NewService(db *database.Connections) *Service {
	return &Service{db: db}
}

func (s *Service) GetStatistic(ctx context.Context, req *pb.GetStatisticRequest) (*pb.GetStatisticResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var stats pb.Statistic
	err := s.db.StatisticDB.QueryRowContext(ctx, "SELECT stats_id, user_id, total_sessions, total_time, tasks_completed, last_active created_at FROM statistics WHERE user_id = $1", req.UserId).Scan(&stats.StatsId, &stats.UserId, &stats.TotalSessions, &stats.TotalTime, &stats.TasksCompleted, &stats.LastActive)
	if err != nil {
		return nil, status.Error(codes.NotFound, "statistics not found")
	}

	return &pb.GetStatisticResponse{Statistic: &stats}, nil
}

func (s *Service) CreateStatistic(ctx context.Context, req *pb.CreateStatisticRequest) (*pb.CreateStatisticResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	statsId := "stats_" + time.Now().Format("20060102150405")

	now := time.Now()
	newStats := &pb.Statistic{
		StatsId:        statsId,
		UserId:         req.UserId,
		TotalSessions:  0,
		TotalTime:      0,
		TasksCompleted: 0,
		LastActive:     timestamppb.New(now),
	}

	_, err := s.db.StatisticDB.ExecContext(ctx, "INSERT INTO statistics (stats_id, user_id, total_sessions, total_time, tasks_completed, last_active) VALUES ($1, $2, $3, $4, $5, $6)", newStats.StatsId, newStats.UserId, newStats.TotalSessions, newStats.TotalTime, newStats.TasksCompleted, now)
	if err != nil {
		log.Printf("Error creating statistic: %v", err)
		return nil, status.Error(codes.Internal, "failed to create statistic")
	}

	return &pb.CreateStatisticResponse{Statistic: newStats}, nil
}

func (s *Service) UpdateStatistic(ctx context.Context, req *pb.UpdateStatisticRequest) (*pb.UpdateStatisticResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "stats_id is required")
	}

	_, err := s.db.StatisticDB.ExecContext(ctx, "UPDATE statistics SET total_sessions = $1, total_time = $2, tasks_completed = $3, last_active = $4 WHERE user_id = $5",
		req.TotalSessions, req.TotalTime, req.TasksCompleted, time.Now(), req.UserId)
	if err != nil {
		log.Printf("Error updating statistic: %v", err)
		return nil, status.Error(codes.Internal, "failed to update statistic")
	}

	resultStatistic := &pb.Statistic{
		StatsId:        "",
		UserId:         req.UserId,
		TotalSessions:  req.TotalSessions,
		TotalTime:      req.TotalTime,
		TasksCompleted: req.TasksCompleted,
		LastActive:     req.LastActive,
	}

	errEntry := s.db.StatisticDB.QueryRowContext(ctx,
		"SELECT stats_id, user_id, total_sessions, total_time, tasks_completed, last_active FROM statistics WHERE user_id = $1",
		req.UserId).Scan(&resultStatistic.StatsId, &resultStatistic.UserId, &resultStatistic.TotalSessions, &resultStatistic.TotalTime, &resultStatistic.TasksCompleted, &resultStatistic.LastActive)
	if errEntry != nil {
		return nil, status.Error(codes.NotFound, "statistics not found")
	}
	return &pb.UpdateStatisticResponse{Statistic: resultStatistic}, nil
}

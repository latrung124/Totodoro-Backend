/*
* File: internal/notification/service.go
* Author: trung.la
* Date: 01/08/2025
* Package: github.com/latrung124/Totodoro-Backend/internal/notification
* Description: This file contains the service layer for notification management.
 */

package notification

import (
	"context"
	"log"
	"time"

	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/notification_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedNotificationServiceServer
	db *database.Connections
}

func NewService(db *database.Connections) *Service {
	return &Service{db: db}
}

func (s *Service) CreateNotification(ctx context.Context, req *pb.CreateNotificationRequest) (*pb.CreateNotificationResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.Message == "" {
		return nil, status.Error(codes.InvalidArgument, "message is required")
	}

	notificationId := "notification_" + time.Now().Format("20060102150405")
	newNotification := &pb.Notification{
		NotificationId: notificationId,
		UserId:         req.UserId,
		Message:        req.Message,
		Type:           req.Type,
		ScheduledTime:  req.ScheduledTime,
		Status:         pb.NotificationStatus_SENT,
	}

	_, err := s.db.NotificationDB.ExecContext(ctx, "INSERT INTO notifications (notification_id, user_id, message, type, scheduled_time, status) VALUES ($1, $2, $3, $4, $5)", newNotification.NotificationId, newNotification.UserId, newNotification.Message, newNotification.ScheduledTime, newNotification.Status)
	if err != nil {
		log.Printf("Error inserting notification into database: %v", err)
		return nil, status.Error(codes.Internal, "failed to create notification")
	}

	return &pb.CreateNotificationResponse{Notification: newNotification}, nil
}

func (s *Service) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	rows, err := s.db.NotificationDB.QueryContext(ctx, "SELECT notification_id, user_id, message, type, scheduled_time, status FROM notifications WHERE user_id = $1", req.UserId)
	if err != nil {
		log.Printf("Error querying notifications: %v", err)
		return nil, status.Error(codes.Internal, "failed to retrieve notifications")
	}
	defer rows.Close()

	var notifications []*pb.Notification
	for rows.Next() {
		var n pb.Notification
		if err := rows.Scan(&n.NotificationId, &n.UserId, &n.Message, &n.Type, &n.ScheduledTime, &n.Status); err != nil {
			log.Printf("Error scanning notification: %v", err)
			return nil, status.Error(codes.Internal, "failed to retrieve notifications")
		}
		notifications = append(notifications, &n)
	}

	return &pb.GetNotificationsResponse{Notifications: notifications}, nil
}

func (s *Service) UpdateNotificationStatus(ctx context.Context, req *pb.UpdateNotificationStatusRequest) (*pb.UpdateNotificationStatusResponse, error) {
	if req.NotificationId == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}

	if req.Status == pb.NotificationStatus_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	_, err := s.db.NotificationDB.ExecContext(ctx, "UPDATE notifications SET status = $1 WHERE notification_id = $2", req.Status, req.NotificationId)
	if err != nil {
		log.Printf("Error updating notification status: %v", err)
		return nil, status.Error(codes.Internal, "failed to update notification status")
	}

	// Get the updated notification
	var updatedNotification pb.Notification
	err = s.db.NotificationDB.QueryRowContext(ctx, "SELECT notification_id, user_id, message, type, scheduled_time, status FROM notifications WHERE notification_id = $1", req.NotificationId).Scan(&updatedNotification.NotificationId, &updatedNotification.UserId, &updatedNotification.Message, &updatedNotification.Type, &updatedNotification.ScheduledTime, &updatedNotification.Status)
	if err != nil {
		log.Printf("Error retrieving updated notification: %v", err)
		return nil, status.Error(codes.Internal, "failed to retrieve updated notification")
	}

	return &pb.UpdateNotificationStatusResponse{Notification: &updatedNotification}, nil
}

/*
File: service.go
Author: trung.la
Date: 07/27/2025
Description: This file contains the service layer for user management.
*/

package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/latrung124/Totodoro-Backend/internal/database"
	pb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service represents the user service implementation.
type Service struct {
	pb.UnimplementedUserServiceServer
	db *database.Connections
}

// NewService creates a new user service instance.
func NewService(db *database.Connections) *Service {
	return &Service{db: db}
}

// GetUser retrieves a user by user_id.
func (s *Service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	var (
		userID    string
		email     string
		username  string
		createdAt time.Time
		updatedAt time.Time
	)

	err := s.db.UserDB.QueryRowContext(
		ctx,
		"SELECT user_id, email, username, created_at, updated_at FROM users WHERE user_id = $1",
		req.UserId,
	).Scan(&userID, &email, &username, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to retrieve user: %v", err))
	}

	user := &pb.User{
		UserId:    userID,
		Email:     email,
		Username:  username,
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &pb.GetUserResponse{User: user}, nil
}

// CreateUser creates a new user with the provided details.
func (s *Service) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	email := strings.TrimSpace(req.GetEmail())
	username := strings.TrimSpace(req.GetUsername())

	if email == "" || username == "" {
		return nil, status.Error(codes.InvalidArgument, "email and username are required")
	}

	var (
		id                   uuid.UUID
		createdAt, updatedAt sql.NullTime
	)
	id = uuid.New()
	createdAt = sql.NullTime{Time: time.Now(), Valid: true}
	updatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err := s.db.UserDB.QueryRowContext(ctx, `
        INSERT INTO users (user_id, email, username)
        VALUES ($1, $2, $3)
        RETURNING user_id, created_at, updated_at
    `, id, email, username).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code == "23505":
			// unique_violation
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		default:
			log.Printf("CreateUser insert failed: %v", err)
			return nil, status.Error(codes.Internal, "failed to create user")
		}
	}

	u := &pb.User{
		UserId:   id.String(),
		Email:    email,
		Username: username,
	}
	if createdAt.Valid {
		u.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		u.UpdatedAt = timestamppb.New(updatedAt.Time)
	}

	// Create default settings for the new user
	if err := s.CreateDefaultSettings(ctx, u.UserId); err != nil {
		log.Printf("Failed to create default settings for user %s: %v", u.UserId, err)
	} else {
		log.Printf("Default settings created for user %s", u.UserId)
	}

	return &pb.CreateUserResponse{User: u}, nil
}

// UpdateUser updates an existing user's profile.
func (s *Service) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Temporary variables for DB scan
	var createdAt, updatedAt time.Time
	var existingUser pb.User

	// Fetch existing user
	err := s.db.UserDB.QueryRowContext(
		ctx,
		"SELECT user_id, email, username, created_at, updated_at FROM users WHERE user_id = $1",
		req.UserId,
	).Scan(&existingUser.UserId, &existingUser.Email, &existingUser.Username, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
	}

	// Convert to protobuf timestamp
	existingUser.CreatedAt = timestamppb.New(createdAt)
	existingUser.UpdatedAt = timestamppb.New(updatedAt)

	// Update fields if provided
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	newUpdatedAt := time.Now()
	existingUser.UpdatedAt = timestamppb.New(newUpdatedAt)

	// Update DB
	_, err = s.db.UserDB.ExecContext(
		ctx,
		"UPDATE users SET email = $1, username = $2, updated_at = $3 WHERE user_id = $4",
		existingUser.Email, existingUser.Username, newUpdatedAt, existingUser.UserId,
	)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &pb.UpdateUserResponse{User: &existingUser}, nil
}

func (s *Service) GetSettings(ctx context.Context, req *pb.GetSettingsRequest) (*pb.GetSettingsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var (
		userID                 string
		pomodoroDuration       int32
		shortBreakDuration     int32
		longBreakDuration      int32
		autoStartShortBreak    bool
		autoStartLongBreak     bool
		autoStartPomodoro      bool
		pomodoroInterval       int32
		theme                  string
		shortBreakNotification bool
		longBreakNotification  bool
		pomodoroNotification   bool
		autoStartMusic         bool
		language               string
	)

	err := s.db.UserDB.QueryRowContext(ctx, `
        SELECT user_id,
               pomodoro_duration, short_break_duration, long_break_duration,
               auto_start_short_break, auto_start_long_break, auto_start_pomodoro,
               pomodoro_interval, theme,
               short_break_notification, long_break_notification, pomodoro_notification,
               auto_start_music, language
        FROM settings
        WHERE user_id = $1
    `, req.UserId).Scan(
		&userID,
		&pomodoroDuration, &shortBreakDuration, &longBreakDuration,
		&autoStartShortBreak, &autoStartLongBreak, &autoStartPomodoro,
		&pomodoroInterval, &theme,
		&shortBreakNotification, &longBreakNotification, &pomodoroNotification,
		&autoStartMusic, &language,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "settings not found")
		}
		log.Printf("Failed to get settings: %v", err)
		return nil, status.Error(codes.Internal, "failed to get settings")
	}

	settings := &pb.Settings{
		UserId:                 userID,
		PomodoroDuration:       pomodoroDuration,
		ShortBreakDuration:     shortBreakDuration,
		LongBreakDuration:      longBreakDuration,
		AutoStartShortBreak:    autoStartShortBreak,
		AutoStartLongBreak:     autoStartLongBreak,
		AutoStartPomodoro:      autoStartPomodoro,
		PomodoroInterval:       pomodoroInterval,
		Theme:                  theme,
		ShortBreakNotification: shortBreakNotification,
		LongBreakNotification:  longBreakNotification,
		PomodoroNotification:   pomodoroNotification,
		AutoStartMusic:         autoStartMusic,
		Language:               language,
	}

	return &pb.GetSettingsResponse{Settings: settings}, nil
}

func (s *Service) CreateDefaultSettings(ctx context.Context, userId string) error {
	var (
		pomodoroDuration       int32  = 25
		shortBreakDuration     int32  = 5
		longBreakDuration      int32  = 15
		autoStartShortBreak           = false
		autoStartLongBreak            = false
		autoStartPomodoro             = false
		pomodoroInterval       int32  = 4
		theme                  string = "#222222"
		shortBreakNotification        = true
		longBreakNotification         = true
		pomodoroNotification          = true
		autoStartMusic                = false
		language               string = "en"
	)
	_, err := s.db.UserDB.ExecContext(ctx, `
		INSERT INTO settings (
			user_id,
			pomodoro_duration, short_break_duration, long_break_duration,
			auto_start_short_break, auto_start_long_break, auto_start_pomodoro,
			pomodoro_interval, theme,
			short_break_notification, long_break_notification, pomodoro_notification,
			auto_start_music, language
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14
		)
	`, userId,
		pomodoroDuration, shortBreakDuration, longBreakDuration,
		autoStartShortBreak, autoStartLongBreak, autoStartPomodoro,
		pomodoroInterval, theme,
		shortBreakNotification, longBreakNotification, pomodoroNotification,
		autoStartMusic, language,
	)

	if err != nil {
		log.Printf("Failed to create default settings: %v", err)
		return status.Error(codes.Internal, "failed to create default settings")
	}

	log.Printf("Default settings created for user_id: %s", userId)
	return nil
}

func (s *Service) UpdateSettings(ctx context.Context, req *pb.UpdateSettingsRequest) (*pb.UpdateSettingsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	selectErr := s.db.UserDB.QueryRowContext(ctx, "SELECT 1 FROM settings WHERE user_id = $1 LIMIT 1", req.UserId).Err()
	if selectErr != nil {
		if errors.Is(selectErr, sql.ErrNoRows) {
			log.Printf("No existing settings for user_id %s, creating default settings", req.UserId)
			if err := s.CreateDefaultSettings(ctx, req.UserId); err != nil {
				return nil, err
			}
		}
		log.Printf("Failed to check existing settings: %v", selectErr)
	} else {
		log.Printf("Existing settings found for user_id %s, proceeding to update", req.UserId)
	}

	_, err := s.db.UserDB.ExecContext(ctx, `
        INSERT INTO settings (
            user_id,
            pomodoro_duration, short_break_duration, long_break_duration,
            auto_start_short_break, auto_start_long_break, auto_start_pomodoro,
            pomodoro_interval, theme,
            short_break_notification, long_break_notification, pomodoro_notification,
            auto_start_music, language
        ) VALUES (
            $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14
        )
        ON CONFLICT (user_id) DO UPDATE SET
            pomodoro_duration       = EXCLUDED.pomodoro_duration,
            short_break_duration    = EXCLUDED.short_break_duration,
            long_break_duration     = EXCLUDED.long_break_duration,
            auto_start_short_break  = EXCLUDED.auto_start_short_break,
            auto_start_long_break   = EXCLUDED.auto_start_long_break,
            auto_start_pomodoro     = EXCLUDED.auto_start_pomodoro,
            pomodoro_interval       = EXCLUDED.pomodoro_interval,
            theme                   = EXCLUDED.theme,
            short_break_notification= EXCLUDED.short_break_notification,
            long_break_notification = EXCLUDED.long_break_notification,
            pomodoro_notification   = EXCLUDED.pomodoro_notification,
            auto_start_music        = EXCLUDED.auto_start_music,
            language                = EXCLUDED.language
    `,
		req.UserId,
		req.PomodoroDuration, req.ShortBreakDuration, req.LongBreakDuration,
		req.AutoStartShortBreak, req.AutoStartLongBreak, req.AutoStartPomodoro,
		req.PomodoroInterval, req.Theme,
		req.ShortBreakNotification, req.LongBreakNotification, req.PomodoroNotification,
		req.AutoStartMusic, req.Language,
	)
	if err != nil {
		log.Printf("Failed to upsert settings: %v", err)
		return nil, status.Error(codes.Internal, "failed to update settings")
	}

	var (
		userID                 string
		pomodoroDuration       int32
		shortBreakDuration     int32
		longBreakDuration      int32
		autoStartShortBreak    bool
		autoStartLongBreak     bool
		autoStartPomodoro      bool
		pomodoroInterval       int32
		theme                  string
		shortBreakNotification bool
		longBreakNotification  bool
		pomodoroNotification   bool
		autoStartMusic         bool
		language               string
	)
	err = s.db.UserDB.QueryRowContext(ctx, `
        SELECT user_id,
               pomodoro_duration, short_break_duration, long_break_duration,
               auto_start_short_break, auto_start_long_break, auto_start_pomodoro,
               pomodoro_interval, theme,
               short_break_notification, long_break_notification, pomodoro_notification,
               auto_start_music, language
        FROM settings
        WHERE user_id = $1
    `, req.UserId).Scan(
		&userID,
		&pomodoroDuration, &shortBreakDuration, &longBreakDuration,
		&autoStartShortBreak, &autoStartLongBreak, &autoStartPomodoro,
		&pomodoroInterval, &theme,
		&shortBreakNotification, &longBreakNotification, &pomodoroNotification,
		&autoStartMusic, &language,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "settings not found")
		}
		log.Printf("Failed to fetch updated settings: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch updated settings")
	}

	settings := &pb.Settings{
		UserId:                 userID,
		PomodoroDuration:       pomodoroDuration,
		ShortBreakDuration:     shortBreakDuration,
		LongBreakDuration:      longBreakDuration,
		AutoStartShortBreak:    autoStartShortBreak,
		AutoStartLongBreak:     autoStartLongBreak,
		AutoStartPomodoro:      autoStartPomodoro,
		PomodoroInterval:       pomodoroInterval,
		Theme:                  theme,
		ShortBreakNotification: shortBreakNotification,
		LongBreakNotification:  longBreakNotification,
		PomodoroNotification:   pomodoroNotification,
		AutoStartMusic:         autoStartMusic,
		Language:               language,
	}

	return &pb.UpdateSettingsResponse{Settings: settings}, nil
}

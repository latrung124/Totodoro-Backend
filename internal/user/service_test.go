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

// ...existing code...

func SeedSettings(t *testing.T, db *sql.DB, s *pb.Settings) {
	t.Helper()
	if s == nil || s.UserId == "" {
		t.Fatal("SeedSettings requires non-nil settings with UserId")
	}

	_, err := db.Exec(
		`INSERT INTO settings (
            user_id,
            pomodoro_duration, short_break_duration, long_break_duration,
            auto_start_short_break, auto_start_long_break, auto_start_pomodoro,
            pomodoro_interval, theme,
            short_break_notification, long_break_notification, pomodoro_notification,
            auto_start_music, language
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
        ON CONFLICT (user_id) DO UPDATE SET
            pomodoro_duration = EXCLUDED.pomodoro_duration,
            short_break_duration = EXCLUDED.short_break_duration,
            long_break_duration = EXCLUDED.long_break_duration,
            auto_start_short_break = EXCLUDED.auto_start_short_break,
            auto_start_long_break = EXCLUDED.auto_start_long_break,
            auto_start_pomodoro = EXCLUDED.auto_start_pomodoro,
            pomodoro_interval = EXCLUDED.pomodoro_interval,
            theme = EXCLUDED.theme,
            short_break_notification = EXCLUDED.short_break_notification,
            long_break_notification = EXCLUDED.long_break_notification,
            pomodoro_notification = EXCLUDED.pomodoro_notification,
            auto_start_music = EXCLUDED.auto_start_music,
            language = EXCLUDED.language`,
		s.UserId,
		s.PomodoroDuration, s.ShortBreakDuration, s.LongBreakDuration,
		s.AutoStartShortBreak, s.AutoStartLongBreak, s.AutoStartPomodoro,
		s.PomodoroInterval, s.Theme,
		s.ShortBreakNotification, s.LongBreakNotification, s.PomodoroNotification,
		s.AutoStartMusic, s.Language,
	)
	if err != nil {
		t.Fatalf("Failed to seed settings: %v", err)
	}
}

func RemoveSettings(connections *database.Connections, userID string) {
	_, _ = connections.UserDB.Exec("DELETE FROM settings WHERE user_id = $1", userID)
}

// ...existing code...

func TestGetSettings(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}
	defer connections.Close()

	service := NewService(connections)

	userID := uuid.NewString()
	seedTestUser(t, connections.UserDB, userID, "getsettings@example.com", "GetSettingsUser")

	expPomodoroDuration := int32(25)
	expShortBreakDuration := int32(5)
	expLongBreakDuration := int32(15)
	expAutoStartShortBreak := true
	expAutoStartLongBreak := true
	expAutoStartPomodoro := false
	expPomodoroInterval := int32(4)
	expTheme := "#222222"
	expShortBreakNotification := true
	expLongBreakNotification := true
	expPomodoroNotification := true
	expAutoStartMusic := false
	expLanguage := "en"

	SeedSettings(t, connections.UserDB, &pb.Settings{
		UserId:                 userID,
		PomodoroDuration:       expPomodoroDuration,
		ShortBreakDuration:     expShortBreakDuration,
		LongBreakDuration:      expLongBreakDuration,
		AutoStartShortBreak:    expAutoStartShortBreak,
		AutoStartLongBreak:     expAutoStartLongBreak,
		AutoStartPomodoro:      expAutoStartPomodoro,
		PomodoroInterval:       expPomodoroInterval,
		Theme:                  expTheme,
		ShortBreakNotification: expShortBreakNotification,
		LongBreakNotification:  expLongBreakNotification,
		PomodoroNotification:   expPomodoroNotification,
		AutoStartMusic:         expAutoStartMusic,
		Language:               expLanguage,
	})

	resp, err := service.GetSettings(context.Background(), &pb.GetSettingsRequest{UserId: userID})
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if resp.Settings == nil {
		t.Fatal("Settings is nil")
	}
	got := resp.Settings

	if got.UserId != userID {
		t.Errorf("user_id: want %s, got %s", userID, got.UserId)
	}
	if got.PomodoroDuration != expPomodoroDuration {
		t.Errorf("pomodoro_duration: want %d, got %d", expPomodoroDuration, got.PomodoroDuration)
	}
	if got.ShortBreakDuration != expShortBreakDuration {
		t.Errorf("short_break_duration: want %d, got %d", expShortBreakDuration, got.ShortBreakDuration)
	}
	if got.LongBreakDuration != expLongBreakDuration {
		t.Errorf("long_break_duration: want %d, got %d", expLongBreakDuration, got.LongBreakDuration)
	}
	if got.AutoStartShortBreak != expAutoStartShortBreak {
		t.Errorf("auto_start_short_break: want %v, got %v", expAutoStartShortBreak, got.AutoStartShortBreak)
	}
	if got.AutoStartLongBreak != expAutoStartLongBreak {
		t.Errorf("auto_start_long_break: want %v, got %v", expAutoStartLongBreak, got.AutoStartLongBreak)
	}
	if got.AutoStartPomodoro != expAutoStartPomodoro {
		t.Errorf("auto_start_pomodoro: want %v, got %v", expAutoStartPomodoro, got.AutoStartPomodoro)
	}
	if got.PomodoroInterval != expPomodoroInterval {
		t.Errorf("pomodoro_interval: want %d, got %d", expPomodoroInterval, got.PomodoroInterval)
	}
	if got.Theme != expTheme {
		t.Errorf("theme: want %q, got %q", expTheme, got.Theme)
	}
	if got.ShortBreakNotification != expShortBreakNotification {
		t.Errorf("short_break_notification: want %v, got %v", expShortBreakNotification, got.ShortBreakNotification)
	}
	if got.LongBreakNotification != expLongBreakNotification {
		t.Errorf("long_break_notification: want %v, got %v", expLongBreakNotification, got.LongBreakNotification)
	}
	if got.PomodoroNotification != expPomodoroNotification {
		t.Errorf("pomodoro_notification: want %v, got %v", expPomodoroNotification, got.PomodoroNotification)
	}
	if got.AutoStartMusic != expAutoStartMusic {
		t.Errorf("auto_start_music: want %v, got %v", expAutoStartMusic, got.AutoStartMusic)
	}
	if got.Language != expLanguage {
		t.Errorf("language: want %q, got %q", expLanguage, got.Language)
	}

	_, _ = connections.UserDB.Exec("DELETE FROM settings WHERE user_id = $1", userID)
	RemoveUserId(connections, userID)
	RemoveSettings(connections, userID)
}

func TestUpdateSettings(t *testing.T) {
	connections, err := setupTestDB()
	if err != nil {
		t.Fatal("Failed to set up test database connections")
	}
	defer connections.Close()

	service := NewService(connections)

	userID := uuid.NewString()
	seedTestUser(t, connections.UserDB, userID, "updateSettingsUser@gmail.com", "UpdateSettingsUser")
	SeedSettings(t, connections.UserDB, &pb.Settings{
		UserId:                 userID,
		PomodoroDuration:       25,
		ShortBreakDuration:     5,
		LongBreakDuration:      15,
		AutoStartShortBreak:    true,
		AutoStartLongBreak:     true,
		AutoStartPomodoro:      false,
		PomodoroInterval:       4,
		Theme:                  "#222222",
		ShortBreakNotification: true,
		LongBreakNotification:  true,
		PomodoroNotification:   true,
		AutoStartMusic:         false,
		Language:               "en",
	})

	req := &pb.UpdateSettingsRequest{
		UserId:                 userID,
		PomodoroDuration:       30,
		ShortBreakDuration:     10,
		LongBreakDuration:      20,
		AutoStartShortBreak:    false,
		AutoStartLongBreak:     false,
		AutoStartPomodoro:      true,
		PomodoroInterval:       3,
		Theme:                  "#ffffff",
		ShortBreakNotification: false,
		LongBreakNotification:  false,
		PomodoroNotification:   false,
		AutoStartMusic:         true,
		Language:               "es",
	}

	resp, err := service.UpdateSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	if resp.Settings == nil {
		t.Fatal("Settings is nil")
	}

	got := resp.Settings
	if got.UserId != req.UserId {
		t.Errorf("user_id: want %s, got %s", req.UserId, got.UserId)
	}

	if got.PomodoroDuration != req.PomodoroDuration {
		t.Errorf("pomodoro_duration: want %d, got %d", req.PomodoroDuration, got.PomodoroDuration)
	}

	if got.ShortBreakDuration != req.ShortBreakDuration {
		t.Errorf("short_break_duration: want %d, got %d", req.ShortBreakDuration, got.ShortBreakDuration)
	}

	if got.LongBreakDuration != req.LongBreakDuration {
		t.Errorf("long_break_duration: want %d, got %d", req.LongBreakDuration, got.LongBreakDuration)
	}

	if got.AutoStartShortBreak != req.AutoStartShortBreak {
		t.Errorf("auto_start_short_break: want %v, got %v", req.AutoStartShortBreak, got.AutoStartShortBreak)
	}

	if got.AutoStartLongBreak != req.AutoStartLongBreak {
		t.Errorf("auto_start_long_break: want %v, got %v", req.AutoStartLongBreak, got.AutoStartLongBreak)
	}

	if got.AutoStartPomodoro != req.AutoStartPomodoro {
		t.Errorf("auto_start_pomodoro: want %v, got %v", req.AutoStartPomodoro, got.AutoStartPomodoro)
	}

	if got.PomodoroInterval != req.PomodoroInterval {
		t.Errorf("pomodoro_interval: want %d, got %d", req.PomodoroInterval, got.PomodoroInterval)
	}

	if got.Theme != req.Theme {
		t.Errorf("theme: want %q, got %q", req.Theme, got.Theme)
	}

	if got.ShortBreakNotification != req.ShortBreakNotification {
		t.Errorf("short_break_notification: want %v, got %v", req.ShortBreakNotification, got.ShortBreakNotification)
	}

	if got.LongBreakNotification != req.LongBreakNotification {
		t.Errorf("long_break_notification: want %v, got %v", req.LongBreakNotification, got.LongBreakNotification)
	}

	if got.PomodoroNotification != req.PomodoroNotification {
		t.Errorf("pomodoro_notification: want %v, got %v", req.PomodoroNotification, got.PomodoroNotification)
	}

	if got.AutoStartMusic != req.AutoStartMusic {
		t.Errorf("auto_start_music: want %v, got %v", req.AutoStartMusic, got.AutoStartMusic)
	}

	if got.Language != req.Language {
		t.Errorf("language: want %q, got %q", req.Language, got.Language)
	}

	_, _ = connections.UserDB.Exec("DELETE FROM settings WHERE user_id = $1", userID)
	RemoveUserId(connections, userID)
	RemoveSettings(connections, userID)
}

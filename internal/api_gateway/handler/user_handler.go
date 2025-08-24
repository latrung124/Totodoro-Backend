/*
File: internal/api_gateway/handler/user_handler.go
Author: trung.la
Date: 08/24/2025
Description: User service gRPC handler implementation.
*/

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	userpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type UserHandler struct {
	client userpb.UserServiceClient
	jsonMO protojson.MarshalOptions
}

func NewUserHandler(client userpb.UserServiceClient) *UserHandler {
	return &UserHandler{
		client: client,
		jsonMO: protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseEnumNumbers:  false,
			UseProtoNames:   false,
			Indent:          "",
		},
	}
}

// Routes:
//
//	GET    /api/users/{userId}
//	POST   /api/users
//	PUT    /api/users/{userId}
//	GET    /api/users/{userId}/settings
//	PUT    /api/users/{userId}/settings
func RegisterUserRoutes(mux *http.ServeMux, h *UserHandler) {
	mux.HandleFunc("/api/users/", h.usersWithID)     // GET/PUT /api/users/{id}
	mux.HandleFunc("/api/users", h.usersRoot)        // POST /api/users
	mux.HandleFunc("/api/users/", h.userSettingsMux) // GET/PUT /api/users/{id}/settings (handled in usersWithID for simplicity)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
}

func (handler *UserHandler) usersRoot(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		handler.createUser(writer, request)
	default:
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) usersWithID(w http.ResponseWriter, r *http.Request) {
	// Expect /api/users/{id} or /api/users/{id}/settings
	trim := strings.TrimPrefix(r.URL.Path, "/api/users/")
	parts := strings.Split(trim, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	userID := parts[0]

	// Settings sub-route
	if len(parts) >= 2 && parts[1] == "settings" {
		switch r.Method {
		case http.MethodGet:
			h.getSettings(w, r.WithContext(context.WithValue(r.Context(), ctxUserIDKey{}, userID)))
			return
		case http.MethodPut:
			h.updateSettings(w, r.WithContext(context.WithValue(r.Context(), ctxUserIDKey{}, userID)))
			return
		default:
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.getUser(w, r.WithContext(context.WithValue(r.Context(), ctxUserIDKey{}, userID)))
	case http.MethodPut:
		h.updateUser(w, r.WithContext(context.WithValue(r.Context(), ctxUserIDKey{}, userID)))
	default:
		http.NotFound(w, r)
	}
}

func (h *UserHandler) userSettingsMux(w http.ResponseWriter, r *http.Request) {
	// Kept for compatibility; logic handled in usersWithID
	http.NotFound(w, r)
}

type ctxUserIDKey struct{}

// DTOs for incoming JSON bodies
type createUserBody struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}
type updateUserBody struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}
type updateSettingsBody struct {
	PomodoroDuration       int32  `json:"pomodoroDuration"`
	ShortBreakDuration     int32  `json:"shortBreakDuration"`
	LongBreakDuration      int32  `json:"longBreakDuration"`
	AutoStartShortBreak    bool   `json:"autoStartShortBreak"`
	AutoStartLongBreak     bool   `json:"autoStartLongBreak"`
	AutoStartPomodoro      bool   `json:"autoStartPomodoro"`
	PomodoroInterval       int32  `json:"pomodoroInterval"`
	Theme                  string `json:"theme"`
	ShortBreakNotification bool   `json:"shortBreakNotification"`
	LongBreakNotification  bool   `json:"longBreakNotification"`
	PomodoroNotification   bool   `json:"pomodoroNotification"`
	AutoStartMusic         bool   `json:"autoStartMusic"`
	Language               string `json:"language"`
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var body createUserBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req := &userpb.CreateUserRequest{
		Email:    body.Email,
		Username: body.Username,
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.CreateUser(ctx, req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeProto(w, http.StatusCreated, h.jsonMO, resp.User)
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(ctxUserIDKey{})
	userID, _ := val.(string)
	if userID == "" {
		writeJSONError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeProto(w, http.StatusOK, h.jsonMO, resp.User)
}

func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(ctxUserIDKey{})
	userID, _ := val.(string)
	if userID == "" {
		writeJSONError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	var body updateUserBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req := &userpb.UpdateUserRequest{
		UserId:   userID,
		Email:    body.Email,
		Username: body.Username,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	resp, err := h.client.UpdateUser(ctx, req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeProto(w, http.StatusOK, h.jsonMO, resp.User)
}

func (h *UserHandler) getSettings(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(ctxUserIDKey{})
	userID, _ := val.(string)
	if userID == "" {
		writeJSONError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetSettings(ctx, &userpb.GetSettingsRequest{UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeProto(w, http.StatusOK, h.jsonMO, resp.Settings)
}

func (h *UserHandler) updateSettings(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(ctxUserIDKey{})
	userID, _ := val.(string)
	if userID == "" {
		writeJSONError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	var body updateSettingsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req := &userpb.UpdateSettingsRequest{
		UserId:                 userID,
		PomodoroDuration:       body.PomodoroDuration,
		ShortBreakDuration:     body.ShortBreakDuration,
		LongBreakDuration:      body.LongBreakDuration,
		AutoStartShortBreak:    body.AutoStartShortBreak,
		AutoStartLongBreak:     body.AutoStartLongBreak,
		AutoStartPomodoro:      body.AutoStartPomodoro,
		PomodoroInterval:       body.PomodoroInterval,
		Theme:                  body.Theme,
		ShortBreakNotification: body.ShortBreakNotification,
		LongBreakNotification:  body.LongBreakNotification,
		PomodoroNotification:   body.PomodoroNotification,
		AutoStartMusic:         body.AutoStartMusic,
		Language:               body.Language,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	resp, err := h.client.UpdateSettings(ctx, req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeProto(w, http.StatusOK, h.jsonMO, resp.Settings)
}

func writeProto(w http.ResponseWriter, status int, mo protojson.MarshalOptions, msg proto.Message) {
	b, err := mo.Marshal(msg)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to marshal response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	code := http.StatusInternalServerError
	switch st.Code() {
	case 3: // InvalidArgument
		code = http.StatusBadRequest
	case 5: // NotFound
		code = http.StatusNotFound
	case 7: // PermissionDenied
		code = http.StatusForbidden
	case 16: // Unauthenticated
		code = http.StatusUnauthorized
	case 9, 10, 11, 14: // FailedPrecondition, Aborted, OutOfRange, Unavailable
		code = http.StatusServiceUnavailable
	}
	writeJSONError(w, code, st.Message())
}

// Defensive helper to ensure body is present
func requireBody(r *http.Request) error {
	if r.Body == nil {
		return errors.New("request body required")
	}
	return nil
}

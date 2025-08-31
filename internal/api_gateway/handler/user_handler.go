/*
File: internal/api_gateway/handler/user_handler.go
Author: trung.la
Date: 08/24/2025
Description: User service gRPC handler implementation.
*/

package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	userpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/user_service"
	"google.golang.org/protobuf/encoding/protojson"
)

type UserHandler struct {
	client userpb.UserServiceClient
}

func NewUserHandler(client userpb.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

// RegisterUserRoutes mounts the grpc-gateway runtime mux that is generated in user_service.pb.gw.go.
// It registers versioned routes (/api/v1/users, etc.) onto the given ServeMux.
// Health endpoint remains a simple net/http handler.
func RegisterUserRoutes(mux *http.ServeMux, h *UserHandler) {
	// Configure JSON marshaling behavior
	jsonpb := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseEnumNumbers:  false,
			UseProtoNames:   false,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	// Create grpc-gateway mux and register generated routes using the existing UserServiceClient
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
	)

	if err := userpb.RegisterUserServiceHandlerClient(context.Background(), gwmux, h.client); err != nil {
		log.Printf("[gateway][user] failed to register grpc-gateway handlers: %v", err)
	}

	mux.Handle("/v1/users", gwmux)

	// Health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
}

/*
File: internal/api_gateway/handler/task_management_handler.go
Author: trung.la
Date: 08/26/2025
Description: This file contains the handler functions for task management operations in the API gateway.
*/

package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	taskpb "github.com/latrung124/Totodoro-Backend/internal/proto_package/task_management_service"
	"google.golang.org/protobuf/encoding/protojson"
)

type TaskManagementHandler struct {
	client taskpb.TaskManagementServiceClient
}

func NewTaskManagementHandler(client taskpb.TaskManagementServiceClient) *TaskManagementHandler {
	return &TaskManagementHandler{client: client}
}

// RegisterTaskManagementRoutes mounts the generated grpc-gateway mux for TaskManagementService.
func RegisterTaskManagementRoutes(mux *http.ServeMux, h *TaskManagementHandler) {
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

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
	)

	if err := taskpb.RegisterTaskManagementServiceHandlerClient(context.Background(), gwmux, h.client); err != nil {
		log.Printf("[gateway][task] failed to register grpc-gateway handlers: %v", err)
	}

	// Support both /v1/... (native) and /api/v1/... (prefixed) routes.
	mux.Handle("/v1/", gwmux)
	mux.Handle("/api/", http.StripPrefix("/api", gwmux))
}

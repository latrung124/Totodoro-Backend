/*
File: internal/api_gateway/handler/pomodoro_handler.go
Author: trung.la
Date: 08/27/2025
Description: This file contains the handler functions for pomodoro operations in the API gateway.
*/

package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pomodoropb "github.com/latrung124/Totodoro-Backend/internal/proto_package/pomodoro_service"
	"google.golang.org/protobuf/encoding/protojson"
)

type PomodoroHandler struct {
	client pomodoropb.PomodoroServiceClient
}

func NewPomodoroHandler(client pomodoropb.PomodoroServiceClient) *PomodoroHandler {
	return &PomodoroHandler{client: client}
}

// RegisterPomodoroRoutes mounts the generated grpc-gateway mux for PomodoroService.
func RegisterPomodoroRoutes(mux *http.ServeMux, h *PomodoroHandler) {
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

	if err := pomodoropb.RegisterPomodoroServiceHandlerClient(context.Background(), gwmux, h.client); err != nil {
		log.Printf("[gateway][pomodoro] failed to register grpc-gateway handlers: %v", err)
	}

	mux.Handle("/v1/sessions/", gwmux)
}

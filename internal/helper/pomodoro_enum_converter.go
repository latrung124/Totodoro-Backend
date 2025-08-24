/*
File: internal/helper/pomodoro_enum_converter.go
Author: trung.la
Date: 08/24/2025
Package: github.com/latrung124/Totodoro-Backend/internal/helper
Description: This file contains helper functions to convert between protobuf enums and database enums for pomodoro sessions
*/

package helper

import (
	pomodoroPb "github.com/latrung124/Totodoro-Backend/internal/proto_package/pomodoro_service"
)

func SessionStatusDbEnumToString(status pomodoroPb.SessionStatus) string {
	switch status {
	case pomodoroPb.SessionStatus_SESSION_STATUS_IDLE:
		return "idle"
	case pomodoroPb.SessionStatus_SESSION_STATUS_IN_PROGRESS:
		return "in progress"
	case pomodoroPb.SessionStatus_SESSION_STATUS_PENDING:
		return "pending"
	case pomodoroPb.SessionStatus_SESSION_STATUS_COMPLETED:
		return "completed"
	default:
		return "idle"
	}
}

func SessionStatusDbStringToEnum(status string) pomodoroPb.SessionStatus {
	switch status {
	case "idle":
		return pomodoroPb.SessionStatus_SESSION_STATUS_IDLE
	case "in progress":
		return pomodoroPb.SessionStatus_SESSION_STATUS_IN_PROGRESS
	case "pending":
		return pomodoroPb.SessionStatus_SESSION_STATUS_PENDING
	case "completed":
		return pomodoroPb.SessionStatus_SESSION_STATUS_COMPLETED
	default:
		return pomodoroPb.SessionStatus_SESSION_STATUS_IDLE
	}
}

func SessionTypeDbEnumToString(sessionType pomodoroPb.SessionType) string {
	switch sessionType {
	case pomodoroPb.SessionType_SESSION_TYPE_SHORT_BREAK:
		return "short break"
	case pomodoroPb.SessionType_SESSION_TYPE_LONG_BREAK:
		return "long break"
	default:
		return "short break"
	}
}

func SessionTypeDbStringToEnum(sessionType string) pomodoroPb.SessionType {
	switch sessionType {
	case "short break":
		return pomodoroPb.SessionType_SESSION_TYPE_SHORT_BREAK
	case "long break":
		return pomodoroPb.SessionType_SESSION_TYPE_LONG_BREAK
	default:
		return pomodoroPb.SessionType_SESSION_TYPE_SHORT_BREAK
	}
}

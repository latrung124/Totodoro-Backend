/*
* File: internal/helper/enum_converter.go
* Author: trung.la
* Date: 23/08/2025
* Package: github.com/latrung124/Totodoro-Backend/internal/helper
* Description: This file contains helper functions for converting between different enum types.
 */

package helper

import (
	taskManagementPb "github.com/latrung124/Totodoro-Backend/internal/proto_package/task_management_service"
)

func TaskGroupPriorityDbEnumToString(p taskManagementPb.TaskGroupPriority) string {
	switch p {
	case taskManagementPb.TaskGroupPriority_TASK_GROUP_PRIORITY_LOW:
		return "low"
	case taskManagementPb.TaskGroupPriority_TASK_GROUP_PRIORITY_MEDIUM:
		return "medium"
	case taskManagementPb.TaskGroupPriority_TASK_GROUP_PRIORITY_HIGH:
		return "high"
	default:
		return "medium"
	}
}

func TaskGroupPriorityDbStringToEnum(p string) taskManagementPb.TaskGroupPriority {
	switch p {
	case "low":
		return taskManagementPb.TaskGroupPriority_TASK_GROUP_PRIORITY_LOW
	case "medium":
		return taskManagementPb.TaskGroupPriority_TASK_GROUP_PRIORITY_MEDIUM
	case "high":
		return taskManagementPb.TaskGroupPriority_TASK_GROUP_PRIORITY_HIGH
	default:
		return taskManagementPb.TaskGroupPriority_TASK_GROUP_PRIORITY_MEDIUM
	}
}

func TaskGroupStatusDbEnumToString(status taskManagementPb.TaskGroupStatus) string {
	switch status {
	case taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_IDLE:
		return "idle"
	case taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_COMPLETED:
		return "completed"
	case taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_PENDING:
		return "pending"
	case taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_IN_PROGRESS:
		return "in progress"
	default:
		return "idle"
	}
}

func TaskGroupStatusDbStringToEnum(status string) taskManagementPb.TaskGroupStatus {
	switch status {
	case "idle":
		return taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_IDLE
	case "completed":
		return taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_COMPLETED
	case "pending":
		return taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_PENDING
	case "in progress":
		return taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_IN_PROGRESS
	default:
		return taskManagementPb.TaskGroupStatus_TASK_GROUP_STATUS_IDLE
	}
}

func TaskPriorityDbEnumToString(priority taskManagementPb.TaskPriority) string {
	switch priority {
	case taskManagementPb.TaskPriority_TASK_PRIORITY_LOW:
		return "low"
	case taskManagementPb.TaskPriority_TASK_PRIORITY_MEDIUM:
		return "medium"
	case taskManagementPb.TaskPriority_TASK_PRIORITY_HIGH:
		return "high"
	default:
		return "medium"
	}
}

func TaskPriorityDbStringToEnum(priority string) taskManagementPb.TaskPriority {
	switch priority {
	case "low":
		return taskManagementPb.TaskPriority_TASK_PRIORITY_LOW
	case "medium":
		return taskManagementPb.TaskPriority_TASK_PRIORITY_MEDIUM
	case "high":
		return taskManagementPb.TaskPriority_TASK_PRIORITY_HIGH
	default:
		return taskManagementPb.TaskPriority_TASK_PRIORITY_MEDIUM
	}
}

func TaskStatusDbEnumToString(status taskManagementPb.TaskStatus) string {
	switch status {
	case taskManagementPb.TaskStatus_TASK_STATUS_IDLE:
		return "idle"
	case taskManagementPb.TaskStatus_TASK_STATUS_COMPLETED:
		return "completed"
	case taskManagementPb.TaskStatus_TASK_STATUS_PENDING:
		return "pending"
	case taskManagementPb.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return "in progress"
	default:
		return "idle"
	}
}

func TaskStatusDbStringToEnum(status string) taskManagementPb.TaskStatus {
	switch status {
	case "idle":
		return taskManagementPb.TaskStatus_TASK_STATUS_IDLE
	case "completed":
		return taskManagementPb.TaskStatus_TASK_STATUS_COMPLETED
	case "pending":
		return taskManagementPb.TaskStatus_TASK_STATUS_PENDING
	case "in progress":
		return taskManagementPb.TaskStatus_TASK_STATUS_IN_PROGRESS
	default:
		return taskManagementPb.TaskStatus_TASK_STATUS_IDLE // Default to idle if unknown
	}
}

package database

import (
	"go_services/models"
	"time"
)

// CreateTask creates a new task in the database
func CreateTask(taskID, containerID string, args models.JSONMap, isRetry bool, parentTaskID string) error {
	task := models.Task{
		TaskID:         taskID,
		ContainerID:    containerID,
		DateTime:       time.Now().Format(time.RFC3339),
		Args:           args,
		Status:         models.Started,
		SuccessJobIDs:  []string{},
		FailedJobIDs:   []string{},
		CompletionRate: 0,
		IsRetryTask:    isRetry,
		ParentTaskID:   parentTaskID,
	}

	if err := DB.Create(&task).Error; err != nil {
		return err
	}
	return nil
}

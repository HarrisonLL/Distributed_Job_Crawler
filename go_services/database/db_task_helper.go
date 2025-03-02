package database

import (
	"go_services/models"
	"time"
)

// CreateTask creates a new task in the database
func CreateTask(taskID, RunID string, company string, jobType string, location string) error {
	task := models.Task{
		TaskID:         taskID,
		RunID:          RunID,
		DateTime:       time.Now().Format(time.RFC3339),
		Company:        company,
		JobType:        jobType,
		Location:       location,
		Status:         models.Started,
		NumbersOfJobs:  0,
		SuccessJobIDs:  []string{},
		CompletionRate: 0,
	}

	if err := DB.Create(&task).Error; err != nil {
		return err
	}
	return nil
}

// Update task status either by taskID or containerID
func UpdateTaskStatus(taskID string, containerID string, status models.TaskStatus) error {
	if taskID != "" {
		var task models.Task
		if err := DB.First(&task, "task_id = ?", taskID).Error; err != nil {
			return err
		}
		task.Status = status
		if err := DB.Save(&task).Error; err != nil {
			return err
		}
		return nil
	} else {
		var task models.Task
		if err := DB.First(&task, "run_id = ?", containerID).Error; err != nil {
			return err
		}
		task.Status = status
		if err := DB.Save(&task).Error; err != nil {
			return err
		}
		return nil
	}
}

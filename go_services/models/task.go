package models

import (
	"github.com/lib/pq"
)

type TaskStatus int

const (
	Started    TaskStatus = 1
	InProgress TaskStatus = 2
	Error      TaskStatus = 3
	Completed  TaskStatus = 4
)

func (ts TaskStatus) String() string {
	return [...]string{"", "started", "in_progress", "error", "completed"}[ts]
}

type Task struct {
	TaskID         string         `gorm:"primaryKey" json:"task_id"`          // Task identifier
	ContainerID    string         `json:"container_id"`                       // Docker Container ID
	DateTime       string         `json:"date_time"`                          // Task creation time
	Args           JSONMap        `gorm:"type:json" json:"args"`              // Task arguments as JSON
	Status         TaskStatus     `json:"status"`                             // Task status
	SuccessJobIDs  pq.StringArray `gorm:"type:text[]" json:"success_job_ids"` // IDs of successful jobs
	FailedJobIDs   pq.StringArray `gorm:"type:text[]" json:"failed_job_ids"`  // IDs of failed jobs
	CompletionRate float64        `json:"completion_rate"`                    // Completion rate
	IsRetryTask    bool           `json:"is_retry"`                           // Task is retry
	ParentTaskID   string         `json:"parent_task_id"`                     // Only if task is retried
}

// JSONMap is a custom type to handle JSON encoding/decoding
type JSONMap map[string]interface{}

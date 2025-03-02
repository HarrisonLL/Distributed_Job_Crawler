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
	HostType       string         `json:"host_type"`                          // Host type: "local" or "docker" or "k8s"
	RunID          string         `json:"run_id"`                             // Process ID, or Docker Container ID, or K8S Pod ID
	DateTime       string         `json:"date_time"`                          // Task creation time
	Company        string         `json: company`                             // Flattened task arguments company field
	JobType        string         `json: job_type`                            // Flattened task arguments job type field
	Location       string         `json: location`                            // Flattened task arguments location field
	Status         TaskStatus     `json:"status"`                             // Task status
	SuccessJobIDs  pq.StringArray `gorm:"type:text[]" json:"success_job_ids"` // IDs of successful jobs
	NumbersOfJobs  int            `json:"numbers_of_jobs"`                    // Number of jobs
	CompletionRate float64        `json:"completion_rate"`                    // Completion rate
}

// JSONMap is a custom type to handle JSON encoding/decoding
type JSONMap map[string]interface{}
